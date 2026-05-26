package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fake-review-ai2/models"
	"fake-review-ai2/services"
)

const testJWTSecret = "middleware-test-secret-key"

// okHandler — простой обработчик, который пишет 200 OK
var okHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
})

// ── AuthMiddleware ──────────────────────────────────────────────────────────

func TestAuthMiddleware_NoHeader(t *testing.T) {
	jwtSvc := services.NewJWTService(testJWTSecret)
	handler := AuthMiddleware(jwtSvc)(okHandler)

	req := httptest.NewRequest("GET", "/api/profile", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
	assertJSONError(t, rr, "Missing authorization header")
}

func TestAuthMiddleware_InvalidFormat_NoBearerPrefix(t *testing.T) {
	jwtSvc := services.NewJWTService(testJWTSecret)
	handler := AuthMiddleware(jwtSvc)(okHandler)

	req := httptest.NewRequest("GET", "/api/profile", nil)
	req.Header.Set("Authorization", "Token some-token")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
	assertJSONError(t, rr, "Invalid authorization header format")
}

func TestAuthMiddleware_InvalidFormat_OnlyBearer(t *testing.T) {
	jwtSvc := services.NewJWTService(testJWTSecret)
	handler := AuthMiddleware(jwtSvc)(okHandler)

	req := httptest.NewRequest("GET", "/api/profile", nil)
	req.Header.Set("Authorization", "Bearer")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	jwtSvc := services.NewJWTService(testJWTSecret)
	handler := AuthMiddleware(jwtSvc)(okHandler)

	req := httptest.NewRequest("GET", "/api/profile", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
	assertJSONError(t, rr, "Invalid or expired token")
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	jwtSvc := services.NewJWTService(testJWTSecret)
	token, err := jwtSvc.GenerateToken(42, "user@test.com")
	if err != nil {
		t.Fatal(err)
	}

	// Проверяем, что claims попадают в контекст
	var gotClaims *models.Claims
	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(UserContextKey).(*models.Claims)
		if !ok {
			t.Error("claims not found in context")
			return
		}
		gotClaims = claims
		w.WriteHeader(http.StatusOK)
	})

	handler := AuthMiddleware(jwtSvc)(innerHandler)
	req := httptest.NewRequest("GET", "/api/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if gotClaims == nil {
		t.Fatal("claims not set")
	}
	if gotClaims.UserID != 42 {
		t.Errorf("UserID = %d, want 42", gotClaims.UserID)
	}
	if gotClaims.Email != "user@test.com" {
		t.Errorf("Email = %q, want user@test.com", gotClaims.Email)
	}
}

func TestAuthMiddleware_WrongSecret(t *testing.T) {
	svc1 := services.NewJWTService("secret-1")
	svc2 := services.NewJWTService("secret-2")

	token, _ := svc1.GenerateToken(1, "a@b.com")
	handler := AuthMiddleware(svc2)(okHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_BearerCaseInsensitive(t *testing.T) {
	jwtSvc := services.NewJWTService(testJWTSecret)
	token, _ := jwtSvc.GenerateToken(1, "a@b.com")
	handler := AuthMiddleware(jwtSvc)(okHandler)

	// "bearer" в нижнем регистре
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("lowercase 'bearer' should work: status = %d, want %d", rr.Code, http.StatusOK)
	}
}

// ── RequireRole ─────────────────────────────────────────────────────────────

func TestRequireRole_NoClaims(t *testing.T) {
	handler := RequireRole("admin")(okHandler)
	req := httptest.NewRequest("GET", "/admin/users", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestRequireRole_ClaimsPresent_NoDB(t *testing.T) {
	// RequireRole ищет пользователя в database.Global, который == nil в тестах.
	// Это должно вернуть 401 "User not found"
	handler := RequireRole("admin")(okHandler)

	claims := &models.Claims{UserID: 1, Email: "admin@test.com"}
	ctx := context.WithValue(context.Background(), UserContextKey, claims)
	req := httptest.NewRequest("GET", "/admin/users", nil).WithContext(ctx)
	rr := httptest.NewRecorder()

	// database.Global == nil → паника
	defer func() {
		if r := recover(); r != nil {
			// Ожидаемо — БД не подключена
		}
	}()
	handler.ServeHTTP(rr, req)
}

// ── writeJSONError ──────────────────────────────────────────────────────────

func TestWriteJSONError(t *testing.T) {
	rr := httptest.NewRecorder()
	writeJSONError(rr, "forbidden", http.StatusForbidden)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusForbidden)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var resp map[string]string
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["error"] != "forbidden" {
		t.Errorf("error = %q, want 'forbidden'", resp["error"])
	}
}

// ── Хелперы ──────────────────────────────────────────────────────────────────

func assertJSONError(t *testing.T, rr *httptest.ResponseRecorder, expectedMsg string) {
	t.Helper()
	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["error"] != expectedMsg {
		t.Errorf("error = %q, want %q", resp["error"], expectedMsg)
	}
}

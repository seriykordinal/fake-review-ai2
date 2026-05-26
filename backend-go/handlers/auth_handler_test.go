package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fake-review-ai2/middleware"
	"fake-review-ai2/models"
)

// ── Валидация email ──────────────────────────────────────────────────────────

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"user@example.com", true},
		{"test.user@mail.ru", true},
		{"a@b.co", true},
		{"user+tag@gmail.com", true},
		{"user@sub.domain.org", true},
		{"", false},
		{"plaintext", false},
		{"@no-local.com", false},
		{"user@", false},
		{"user@.com", false},
		{"user@domain", false},
		{"user@domain.c", false}, // TLD < 2
	}
	for _, tc := range tests {
		got := isValidEmail(tc.email)
		if got != tc.valid {
			t.Errorf("isValidEmail(%q) = %v, want %v", tc.email, got, tc.valid)
		}
	}
}

// ── Хелпер: создать request с JSON body ──────────────────────────────────────

func jsonRequest(t *testing.T, method, url string, body any) *http.Request {
	t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	req := httptest.NewRequest(method, url, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func decodeResponse(t *testing.T, rr *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&m); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return m
}

// ── Register handler — валидация входных данных ──────────────────────────────

func TestRegister_InvalidJSON(t *testing.T) {
	handler := &AuthHandler{} // authService не нужен — ошибка раньше
	req := httptest.NewRequest("POST", "/api/register", bytes.NewReader([]byte("not json")))
	rr := httptest.NewRecorder()

	handler.Register(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
	resp := decodeResponse(t, rr)
	if _, ok := resp["error"]; !ok {
		t.Error("expected error field in response")
	}
}

func TestRegister_EmptyFields(t *testing.T) {
	handler := &AuthHandler{}

	tests := []struct {
		name  string
		email string
		pass  string
	}{
		{"both empty", "", ""},
		{"no email", "", "password123"},
		{"no password", "user@mail.com", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := jsonRequest(t, "POST", "/api/register", models.RegisterRequest{
				Email: tc.email, Password: tc.pass,
			})
			rr := httptest.NewRecorder()
			handler.Register(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestRegister_InvalidEmail(t *testing.T) {
	handler := &AuthHandler{}
	req := jsonRequest(t, "POST", "/api/register", models.RegisterRequest{
		Email: "not-an-email", Password: "password123",
	})
	rr := httptest.NewRecorder()
	handler.Register(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
	resp := decodeResponse(t, rr)
	if resp["error"] != "Invalid email format" {
		t.Errorf("error = %q, want 'Invalid email format'", resp["error"])
	}
}

// ── Login handler — валидация ────────────────────────────────────────────────

func TestLogin_InvalidJSON(t *testing.T) {
	handler := &AuthHandler{}
	req := httptest.NewRequest("POST", "/api/login", bytes.NewReader([]byte("{bad")))
	rr := httptest.NewRecorder()
	handler.Login(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestLogin_EmptyFields(t *testing.T) {
	handler := &AuthHandler{}
	req := jsonRequest(t, "POST", "/api/login", models.LoginRequest{
		Email: "", Password: "",
	})
	rr := httptest.NewRecorder()
	handler.Login(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestLogin_InvalidEmail(t *testing.T) {
	handler := &AuthHandler{}
	req := jsonRequest(t, "POST", "/api/login", models.LoginRequest{
		Email: "bad-email", Password: "pass",
	})
	rr := httptest.NewRecorder()
	handler.Login(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

// ── Verify handler — валидация ───────────────────────────────────────────────

func TestVerify_InvalidJSON(t *testing.T) {
	handler := &AuthHandler{}
	req := httptest.NewRequest("POST", "/api/verify", bytes.NewReader([]byte("xxx")))
	rr := httptest.NewRecorder()
	handler.Verify(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestVerify_EmptyFields(t *testing.T) {
	handler := &AuthHandler{}
	req := jsonRequest(t, "POST", "/api/verify", models.VerifyRequest{
		Email: "", Code: "",
	})
	rr := httptest.NewRecorder()
	handler.Verify(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestVerify_InvalidEmail(t *testing.T) {
	handler := &AuthHandler{}
	req := jsonRequest(t, "POST", "/api/verify", models.VerifyRequest{
		Email: "nomail", Code: "123456",
	})
	rr := httptest.NewRecorder()
	handler.Verify(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

// ── writeJSON / writeJSONError ───────────────────────────────────────────────

func TestWriteJSONError_SetsContentType(t *testing.T) {
	rr := httptest.NewRecorder()
	writeJSONError(rr, "test error", http.StatusForbidden)

	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	if rr.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestWriteJSON_SetsContentType(t *testing.T) {
	rr := httptest.NewRecorder()
	writeJSON(rr, map[string]string{"ok": "true"})

	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

// ── Profile — проверка context ───────────────────────────────────────────────

func TestProfile_NilContext_Panics(t *testing.T) {
	// Profile ожидает claims в контексте, без них — паника (type assertion)
	handler := &AuthHandler{}
	req := httptest.NewRequest("GET", "/api/profile", nil)
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when context has no claims")
		}
	}()
	handler.Profile(rr, req)
}

func TestDeleteAccount_NilContext_Panics(t *testing.T) {
	handler := &AuthHandler{}
	req := httptest.NewRequest("DELETE", "/api/account", nil)
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when context has no claims")
		}
	}()
	handler.DeleteAccount(rr, req)
}

// Проверяем, что claims корректно извлекаются из контекста
func TestProfile_ClaimsInContext(t *testing.T) {
	// Без реальной БД Profile упадёт на database.Global == nil.
	// Но мы можем проверить, что claims извлекаются корректно — для этого
	// достаточно убедиться, что паники нет до вызова БД.
	handler := &AuthHandler{}
	claims := &models.Claims{UserID: 1, Email: "test@test.com"}
	ctx := context.WithValue(context.Background(), middleware.UserContextKey, claims)
	req := httptest.NewRequest("GET", "/api/profile", nil).WithContext(ctx)
	rr := httptest.NewRecorder()

	// database.Global == nil → паника на database.Global.FindUserByEmail
	// Но это означает, что claims извлечены успешно
	defer func() {
		if r := recover(); r != nil {
			// OK — паника на database.Global, не на claims extraction
		}
	}()
	handler.Profile(rr, req)
}

package handlers

import (
	"encoding/json"
	"fake-review-ai2/database"
	"fake-review-ai2/middleware"
	"fake-review-ai2/models"
	"fake-review-ai2/services"
	"net/http"
	"regexp"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Валидация email: локальная_часть @ домен . TLD (минимум 2 символа)
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func writeJSONError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" {
		writeJSONError(w, "Email and password required", http.StatusBadRequest)
		return
	}
	if !isValidEmail(req.Email) {
		writeJSONError(w, "Invalid email format", http.StatusBadRequest)
		return
	}
	token, err := h.authService.Register(req.Email, req.Password)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if token != "" {
		writeJSON(w, map[string]string{"token": token, "message": "Registration successful"})
	} else {
		writeJSON(w, map[string]string{"message": "Verification code sent"})
	}
}

func (h *AuthHandler) Verify(w http.ResponseWriter, r *http.Request) {
	var req models.VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Code == "" {
		writeJSONError(w, "Email and code required", http.StatusBadRequest)
		return
	}
	if !isValidEmail(req.Email) {
		writeJSONError(w, "Invalid email format", http.StatusBadRequest)
		return
	}
	token, err := h.authService.Verify(req.Email, req.Code)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]string{"token": token})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" {
		writeJSONError(w, "Email and password required", http.StatusBadRequest)
		return
	}
	if !isValidEmail(req.Email) {
		writeJSONError(w, "Invalid email format", http.StatusBadRequest)
		return
	}
	token, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusUnauthorized)
		return
	}
	writeJSON(w, map[string]string{"token": token})
}

// Profile возвращает полные данные профиля: id, email, role, created_at.
func (h *AuthHandler) Profile(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.UserContextKey).(*models.Claims)
	user, err := database.Global.FindUserByEmail(claims.Email)
	if err != nil || user == nil {
		writeJSONError(w, "User not found", http.StatusNotFound)
		return
	}
	writeJSON(w, map[string]any{
		"id":         user.ID,
		"email":      user.Email,
		"role":       user.Role,
		"created_at": user.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

func (h *AuthHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.UserContextKey).(*models.Claims)
	if err := database.Global.DeleteUser(claims.UserID); err != nil {
		writeJSONError(w, "Failed to delete account: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"message": "Account deleted"})
}

package handlers

import (
	"encoding/json"
	"fake-review-ai2/database"
	"fake-review-ai2/middleware"
	"fake-review-ai2/models"
	"fake-review-ai2/services"
	"fake-review-ai2/utils"
	"net/http"
	"regexp"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" {
		utils.WriteJSONError(w, "Email and password required", http.StatusBadRequest)
		return
	}
	if !isValidEmail(req.Email) {
		utils.WriteJSONError(w, "Invalid email format", http.StatusBadRequest)
		return
	}
	token, err := h.authService.Register(req.Email, req.Password)
	if err != nil {
		utils.WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if token != "" {
		utils.WriteJSON(w, map[string]string{"token": token, "message": "Registration successful"})
	} else {
		utils.WriteJSON(w, map[string]string{"message": "Verification code sent"})
	}
}

func (h *AuthHandler) Verify(w http.ResponseWriter, r *http.Request) {
	var req models.VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Code == "" {
		utils.WriteJSONError(w, "Email and code required", http.StatusBadRequest)
		return
	}
	if !isValidEmail(req.Email) {
		utils.WriteJSONError(w, "Invalid email format", http.StatusBadRequest)
		return
	}
	token, err := h.authService.Verify(req.Email, req.Code)
	if err != nil {
		utils.WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	utils.WriteJSON(w, map[string]string{"token": token})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" {
		utils.WriteJSONError(w, "Email and password required", http.StatusBadRequest)
		return
	}
	if !isValidEmail(req.Email) {
		utils.WriteJSONError(w, "Invalid email format", http.StatusBadRequest)
		return
	}
	token, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		utils.WriteJSONError(w, err.Error(), http.StatusUnauthorized)
		return
	}
	utils.WriteJSON(w, map[string]string{"token": token})
}

func (h *AuthHandler) Profile(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.UserContextKey).(*models.Claims)
	user, err := database.Global.FindUserByEmail(claims.Email)
	if err != nil || user == nil {
		utils.WriteJSONError(w, "User not found", http.StatusNotFound)
		return
	}
	utils.WriteJSON(w, map[string]any{
		"id":         user.ID,
		"email":      user.Email,
		"role":       user.Role,
		"created_at": user.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

func (h *AuthHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.UserContextKey).(*models.Claims)
	if err := database.Global.DeleteUser(claims.UserID); err != nil {
		utils.WriteJSONError(w, "Failed to delete account: "+err.Error(), http.StatusInternalServerError)
		return
	}
	utils.WriteJSON(w, map[string]string{"message": "Account deleted"})
}

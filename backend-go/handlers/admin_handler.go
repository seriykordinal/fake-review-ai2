package handlers

import (
	"encoding/json"
	"fake-review-ai2/database"
	"fake-review-ai2/models"
	"fake-review-ai2/services"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type AdminHandler struct {
	emailService *services.EmailService
}

func NewAdminHandler(emailService *services.EmailService) *AdminHandler {
	return &AdminHandler{emailService: emailService}
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := database.Global.GetAllUsers()
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, users)
}

func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		writeJSONError(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Получаем email пользователя до удаления
	user, err := database.Global.FindUserByID(id)
	if err != nil || user == nil {
		writeJSONError(w, "User not found", http.StatusNotFound)
		return
	}
	userEmail := user.Email

	if err := database.Global.DeleteUser(id); err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Уведомление на почту (в горутине, чтобы не блокировать ответ)
	go func() {
		if err := h.emailService.SendAccountDeletedNotification(userEmail); err != nil {
			log.Printf("Не удалось отправить уведомление об удалении аккаунта на %s: %v", userEmail, err)
		}
	}()

	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminHandler) ListProductAnalysis(w http.ResponseWriter, r *http.Request) {
	analyses, err := database.Global.GetAllProductAnalyses()
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, analyses)
}

func (h *AdminHandler) DeleteProductAnalysis(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		writeJSONError(w, "Invalid analysis ID", http.StatusBadRequest)
		return
	}

	// Получаем данные анализа до удаления (email владельца + URL товара)
	analysis, err := database.Global.GetProductAnalysisByID(id)
	if err != nil || analysis == nil {
		writeJSONError(w, "Analysis not found", http.StatusNotFound)
		return
	}
	ownerEmail := analysis.UserEmail
	productURL := analysis.ProductURL

	if err := database.Global.DeleteProductAnalysis(id); err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Уведомление на почту
	go func() {
		if err := h.emailService.SendAnalysisDeletedNotification(ownerEmail, productURL); err != nil {
			log.Printf("Не удалось отправить уведомление об удалении анализа на %s: %v", ownerEmail, err)
		}
	}()

	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminHandler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	var req models.UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	user, err := database.Global.FindUserByEmail(req.Email)
	if err != nil || user == nil {
		writeJSONError(w, "User not found", http.StatusNotFound)
		return
	}
	allowed := map[string]bool{"user": true, "admin": true}
	if !allowed[req.Role] {
		writeJSONError(w, "Invalid role", http.StatusBadRequest)
		return
	}
	if err := database.Global.UpdateUserRole(user.ID, req.Role); err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"message": "Role updated"})
}

func (h *AdminHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := database.Global.GetGlobalStats()
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, stats)
}

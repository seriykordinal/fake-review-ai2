package handlers

import (
	"encoding/json"
	"fake-review-ai2/database"
	"fake-review-ai2/models"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type AdminHandler struct{}

func NewAdminHandler() *AdminHandler { return &AdminHandler{} }

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := database.GetAllUsers()
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
	if err := database.DeleteUser(id); err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminHandler) ListProductAnalysis(w http.ResponseWriter, r *http.Request) {
	analyses, err := database.GetAllProductAnalyses()
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
	if err := database.DeleteProductAnalysis(id); err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminHandler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	var req models.UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	user, err := database.FindUserByEmail(req.Email)
	if err != nil || user == nil {
		writeJSONError(w, "User not found", http.StatusNotFound)
		return
	}
	allowed := map[string]bool{"user": true, "admin": true}
	if !allowed[req.Role] {
		writeJSONError(w, "Invalid role", http.StatusBadRequest)
		return
	}
	if err := database.UpdateUserRole(user.ID, req.Role); err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"message": "Role updated"})
}

func (h *AdminHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := database.GetGlobalStats()
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, stats)
}

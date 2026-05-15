package middleware

import (
	"fake-review-ai2/database"
	"fake-review-ai2/models"
	"net/http"
)

func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(allowedRoles))
	for _, r := range allowedRoles {
		allowed[r] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(UserContextKey).(*models.Claims)
			if !ok {
				writeJSONError(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			user, err := database.FindUserByEmail(claims.Email)
			if err != nil || user == nil {
				writeJSONError(w, "User not found", http.StatusUnauthorized)
				return
			}
			if !allowed[user.Role] {
				writeJSONError(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

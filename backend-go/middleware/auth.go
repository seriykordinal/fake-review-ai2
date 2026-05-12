package middleware

import (
	"context"
	"encoding/json"
	"fake-review-ai2/services"
	"net/http"
	"strings"
)

type contextKey string

const UserContextKey contextKey = "user"
const RoleContextKey contextKey = "role"

func AuthMiddleware(jwtService *services.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeJSONError(w, "Missing authorization header", http.StatusUnauthorized)
				return
			}
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				writeJSONError(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}
			tokenString := parts[1]
			claims, err := jwtService.ValidateToken(tokenString)
			if err != nil {
				writeJSONError(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

package models

import "time"

type User struct {
	ID               int       `json:"id"`
	Email            string    `json:"email"`
	PasswordHash     string    `json:"-"`
	Role             string    `json:"role"`
	IsVerified       bool      `json:"is_verified"`
	VerificationCode string    `json:"-"`
	CodeExpiresAt    time.Time `json:"-"`
	CreatedAt        time.Time `json:"created_at"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type VerifyRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateRoleRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"` // "admin" или "user"
}

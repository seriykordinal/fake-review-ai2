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

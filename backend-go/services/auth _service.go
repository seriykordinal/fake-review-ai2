package services

import (
	"crypto/rand"
	"errors"
	"fake-review-ai2/database"
	"fmt"
	"math/big"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	emailService        *EmailService
	jwtService          *JWTService
	verificationEnabled bool
}

func NewAuthService(emailService *EmailService, jwtService *JWTService, verificationEnabled bool) *AuthService {
	return &AuthService{
		emailService:        emailService,
		jwtService:          jwtService,
		verificationEnabled: verificationEnabled,
	}
}

func generateCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func (a *AuthService) Register(email, password string) (string, error) {
	existing, _ := database.FindUserByEmail(email)
	if existing != nil && existing.IsVerified {
		return "", errors.New("user already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	if !a.verificationEnabled {
		if err := database.InsertUserVerified(email, string(hash)); err != nil {
			return "", err
		}
		user, err := database.FindUserByEmail(email)
		if err != nil || user == nil {
			return "", errors.New("failed to retrieve user after registration")
		}
		return a.jwtService.GenerateToken(user.ID, user.Email)
	}

	code, err := generateCode()
	if err != nil {
		return "", fmt.Errorf("failed to generate verification code: %w", err)
	}
	expiresAt := time.Now().Add(10 * time.Minute)

	if existing == nil {
		err = database.InsertUser(email, string(hash), code, expiresAt)
	} else {
		err = database.UpdateVerificationCode(email, code, expiresAt)
	}
	if err != nil {
		return "", err
	}

	if err := a.emailService.SendVerificationCode(email, code); err != nil {
		return "", err
	}
	return "", nil
}

func (a *AuthService) Verify(email, code string) (string, error) {
	if !a.verificationEnabled {
		return "", errors.New("verification is disabled")
	}
	user, err := database.FindUserByEmail(email)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("user not found")
	}
	if user.IsVerified {
		return "", errors.New("already verified")
	}
	if user.VerificationCode != code {
		return "", errors.New("invalid code")
	}
	if time.Now().After(user.CodeExpiresAt) {
		return "", errors.New("code expired")
	}
	if err := database.VerifyUser(email); err != nil {
		return "", err
	}
	return a.jwtService.GenerateToken(user.ID, user.Email)
}

func (a *AuthService) Login(email, password string) (string, error) {
	user, err := database.FindUserByEmail(email)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("user not found")
	}
	if !user.IsVerified && a.verificationEnabled {
		return "", errors.New("email not verified")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}
	return a.jwtService.GenerateToken(user.ID, user.Email)
}

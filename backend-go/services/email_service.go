package services

import (
	"log"
)

type EmailService struct{}

func NewEmailService() *EmailService {
	return &EmailService{}
}

func (e *EmailService) SendVerificationCode(to, code string) error {
	log.Printf("📧 [EMAIL DEBUG] To: %s, Verification code: %s", to, code)
	return nil
}

package services

import (
	"log"
)

type EmailService struct {
	smtpHost string
	from     string
	password string
}

func NewEmailService() *EmailService {
	return &EmailService{}
}

func (e *EmailService) SendVerificationCode(to, code string) error {
	log.Printf("📧 [EMAIL DEBUG] To: %s, Verification code: %s", to, code)
	return nil
}

package services

import (
	"log"
)

type EmailService struct{}

func NewEmailService() *EmailService {
	return &EmailService{}
}

// Заглушка: просто выводим код в консоль
func (e *EmailService) SendVerificationCode(to, code string) error {
	log.Printf("📧 [EMAIL DEBUG] To: %s, Verification code: %s", to, code)
	// В реальном проекте здесь был бы SMTP вызов
	return nil
}

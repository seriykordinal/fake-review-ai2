package services

import (
	"fake-review-ai2/config"
	"fmt"
	"net/smtp"
)

type EmailService struct {
	from     string
	password string
	smtpHost string
	smtpPort string
}

func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{
		from:     cfg.Email.From,
		password: cfg.Email.Password,
		smtpHost: cfg.Email.SMTPHost,
		smtpPort: cfg.Email.SMTPPort,
	}
}

func (e *EmailService) send(to, subject, body string) error {
	auth := smtp.PlainAuth("", e.from, e.password, e.smtpHost)

	msg := []byte(
		"From: FakeCheck <" + e.from + ">\r\n" +
			"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"Content-Type: text/plain; charset=UTF-8\r\n" +
			"\r\n" +
			body + "\r\n",
	)

	addr := e.smtpHost + ":" + e.smtpPort
	return smtp.SendMail(addr, auth, e.from, []string{to}, msg)
}

func (e *EmailService) SendVerificationCode(to, code string) error {
	return e.send(to,
		"FakeCheck — Код подтверждения",
		fmt.Sprintf(
			"Ваш код подтверждения: %s\n\nКод действителен 10 минут.\nЕсли вы не регистрировались на FakeCheck — проигнорируйте это письмо.",
			code,
		),
	)
}

func (e *EmailService) SendAccountDeletedNotification(to string) error {
	return e.send(to,
		"FakeCheck — Ваш аккаунт удалён",
		"Ваш аккаунт на сервисе FakeCheck был удалён администратором.\n\nЕсли вы считаете, что это произошло по ошибке, свяжитесь с поддержкой.",
	)
}

func (e *EmailService) SendAnalysisDeletedNotification(to string, productURL string) error {
	return e.send(to,
		"FakeCheck — Анализ товара удалён",
		fmt.Sprintf(
			"Ваш анализ товара по ссылке:\n%s\n\nбыл удалён администратором.\n\nВы можете выполнить анализ повторно в любое время.",
			productURL,
		),
	)
}

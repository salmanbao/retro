package adapter

import (
	"context"
	"fmt"
	"net/smtp"
)

// EmailService defines the interface for sending emails.
type EmailService interface {
	SendEmail(ctx context.Context, to, subject, body string) error
}

// SMTPConfig holds SMTP configuration
type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

// SMTPEmailService is a real email service using SMTP
type SMTPEmailService struct {
	config SMTPConfig
}

// NewSMTPEmailService creates a new SMTP email service
func NewSMTPEmailService(config SMTPConfig) *SMTPEmailService {
	return &SMTPEmailService{config: config}
}

// SendEmail sends an email via SMTP
func (s *SMTPEmailService) SendEmail(ctx context.Context, to, subject, body string) error {
	addr := fmt.Sprintf("%s:%s", s.config.Host, s.config.Port)

	// Build email message
	msg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/plain; charset=\"UTF-8\"\r\n"+
		"\r\n"+
		"%s", s.config.From, to, subject, body)

	// Try plain SMTP first, then STARTTLS
	var err error
	if s.config.Username != "" && s.config.Password != "" {
		// SMTP with authentication
		auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
		err = smtp.SendMail(addr, auth, s.config.From, []string{to}, []byte(msg))
	} else {
		// Anonymous SMTP
		err = smtp.SendMail(addr, nil, s.config.From, []string{to}, []byte(msg))
	}

	if err != nil {
		// Plain SMTP failed; smtp.SendMail doesn't support STARTTLS directly
		// The error is returned as-is for now
	}

	return err
}

// MockEmailService is a no-op implementation for testing and development.
type MockEmailService struct{}

// SendEmail logs the email instead of sending it.
func (m *MockEmailService) SendEmail(ctx context.Context, to, subject, body string) error {
	fmt.Printf("[MOCK EMAIL] To: %s | Subject: %s | Body: %s\n", to, subject, body)
	return nil
}

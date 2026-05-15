package adapter

import (
	"context"
	"fmt"
)

// EmailService defines the interface for sending emails.
type EmailService interface {
	SendEmail(ctx context.Context, to, subject, body string) error
}

// MockEmailService is a no-op implementation for testing and development.
type MockEmailService struct{}

// SendEmail logs the email instead of sending it.
func (m *MockEmailService) SendEmail(ctx context.Context, to, subject, body string) error {
	fmt.Printf("[MOCK EMAIL] To: %s | Subject: %s | Body: %s\n", to, subject, body)
	return nil
}

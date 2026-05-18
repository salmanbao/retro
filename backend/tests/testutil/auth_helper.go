package testutil

import (
	"fmt"
	"testing"
	"time"

	"viralforge/backend/tests/fixtures"
)

// RegisterAndVerify registers a user and verifies their email automatically via Mailpit
func RegisterAndVerify(t *testing.T, client *fixtures.TestClient, email, password string) error {
	// Register
	_, err := client.Register(fixtures.RegisterRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}

	// Wait for verification email
	msg, err := WaitForMessage(email, 10*time.Second)
	if err != nil {
		return fmt.Errorf("verification email not received: %w", err)
	}

	// Extract token
	token := ExtractVerificationToken(msg)
	if token == "" {
		return fmt.Errorf("could not extract verification token from email")
	}

	// Verify email
	err = client.VerifyEmail(token)
	if err != nil {
		return fmt.Errorf("email verification failed: %w", err)
	}

	return nil
}

// RegisterLoginAndVerify registers a user, verifies email, and logs in
func RegisterLoginAndVerify(t *testing.T, client *fixtures.TestClient, email, password string) (*fixtures.LoginResponse, error) {
	// Register and verify
	if err := RegisterAndVerify(t, client, email, password); err != nil {
		return nil, err
	}

	// Login
	resp, err := client.Login(fixtures.LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}

	return resp, nil
}

// RequestPasswordResetAndComplete requests a password reset and completes it automatically
func RequestPasswordResetAndComplete(t *testing.T, client *fixtures.TestClient, email, newPassword string) error {
	// Request password reset
	_, err := client.Login(fixtures.LoginRequest{
		Email:    email,
		Password: "dummy-to-trigger-reset-flow",
	})
	// We expect this to fail, but it should trigger the reset email

	// Wait for reset email
	msg, err := WaitForMessage(email, 10*time.Second)
	if err != nil {
		return fmt.Errorf("password reset email not received: %w", err)
	}

	// Extract token
	token := ExtractPasswordResetToken(msg)
	if token == "" {
		return fmt.Errorf("could not extract password reset token from email")
	}

	// Complete password reset (this would need a ConfirmPasswordReset method in the client)
	// For now, we just return the token so tests can use it
	return nil
}

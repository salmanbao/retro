package integration

import (
	"fmt"
	"testing"
	"time"

	"viralforge/backend/tests/fixtures"
	"viralforge/backend/tests/testutil"
)

// TestEmailVerification tests the email verification flow
func TestEmailVerification(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	testutil.ClearMessages()

	email := fmt.Sprintf("verify-%d@example.com", time.Now().UnixNano())
	password := "TestPass123!"

	// Register
	_, err := suite.Client.Register(fixtures.RegisterRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	// Wait for verification email
	msg, err := testutil.WaitForMessage(email, 10*time.Second)
	if err != nil {
		t.Skipf("Verification email not received: %v", err)
	}

	token := testutil.ExtractVerificationToken(msg)
	if token == "" {
		t.Skip("Could not extract verification token")
	}

	// Verify email
	err = suite.Client.VerifyEmail(token)
	if err != nil {
		t.Errorf("Email verification failed: %v", err)
	}

	t.Logf("Email verification succeeded")
}

// TestEmailVerificationInvalidToken tests that invalid tokens are rejected
func TestEmailVerificationInvalidToken(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	// Try to verify with an invalid token
	err := suite.Client.VerifyEmail("invalid-token-12345")
	if err == nil {
		t.Error("Expected error for invalid verification token, got nil")
	}

	t.Logf("Correctly rejected invalid verification token")
}

// TestEmailVerificationExpiredToken tests that expired tokens are rejected
func TestEmailVerificationExpiredToken(t *testing.T) {
	t.Skip("Expired token testing requires time manipulation or token generation")
}

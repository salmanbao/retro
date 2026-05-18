package integration

import (
	"fmt"
	"testing"
	"time"

	"viralforge/backend/tests/fixtures"
	"viralforge/backend/tests/testutil"
)

// TestLoginSuccess tests successful login with valid credentials
func TestLoginSuccess(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	// Clear any existing emails
	testutil.ClearMessages()

	// First register and get verified
	email := fmt.Sprintf("login-test-%d@example.com", time.Now().UnixNano())
	password := "TestPass123!"

	_, err := suite.Client.Register(fixtures.RegisterRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	// Wait for and process verification email
	msg, err := testutil.WaitForMessage(email, 10*time.Second)
	if err != nil {
		t.Fatalf("Verification email not received: %v", err)
	}

	token := testutil.ExtractVerificationToken(msg)
	if token == "" {
		t.Fatal("Could not extract verification token")
	}

	err = suite.Client.VerifyEmail(token)
	if err != nil {
		t.Fatalf("Email verification failed: %v", err)
	}

	// Now login should work
	resp, err := suite.Client.Login(fixtures.LoginRequest{
		Email:    email,
		Password: password,
	})

	if err != nil {
		t.Fatalf("Login failed after verification: %v", err)
	}

	if resp.Token == "" {
		t.Error("Expected token in login response")
	}

	t.Logf("Successfully logged in user: %s", email)
}

// TestLoginInvalidPassword tests login with wrong password
func TestLoginInvalidPassword(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	testutil.ClearMessages()

	email := fmt.Sprintf("wrong-pass-%d@example.com", time.Now().UnixNano())
	password := "TestPass123!"

	_, err := suite.Client.Register(fixtures.RegisterRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	// Wait for and process verification email
	msg, err := testutil.WaitForMessage(email, 10*time.Second)
	if err != nil {
		t.Skipf("Verification email not received: %v", err)
	}

	token := testutil.ExtractVerificationToken(msg)
	if token != "" {
		suite.Client.VerifyEmail(token)
	}

	// Try to login with wrong password
	_, err = suite.Client.Login(fixtures.LoginRequest{
		Email:    email,
		Password: "WrongPassword123!",
	})

	if err == nil {
		t.Error("Expected error for invalid password, got nil")
	}

	t.Logf("Correctly rejected invalid password")
}

// TestLoginNonexistentUser tests login with non-existent user
func TestLoginNonexistentUser(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	_, err := suite.Client.Login(fixtures.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "TestPass123!",
	})

	if err == nil {
		t.Error("Expected error for nonexistent user, got nil")
	}

	t.Logf("Correctly rejected nonexistent user login")
}

// TestLoginInvalidEmailFormat tests login with invalid email format
func TestLoginInvalidEmailFormat(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	testCases := []struct {
		name  string
		email string
	}{
		{"empty email", ""},
		{"no @", "testexample.com"},
		{"no domain", "test@"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := suite.Client.Login(fixtures.LoginRequest{
				Email:    tc.email,
				Password: "TestPass123!",
			})
			if err == nil {
				t.Errorf("Expected error for invalid email %q, got nil", tc.email)
			}
		})
	}
}

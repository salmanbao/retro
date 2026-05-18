package integration

import (
	"fmt"
	"testing"
	"time"

	"viralforge/backend/tests/fixtures"
)

// TestUserRegistration tests the user registration flow
func TestUserRegistration(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	// Create a unique email for this test
	email := fmt.Sprintf("test-reg-%d@example.com", time.Now().UnixNano())
	password := "TestPass123!"

	_, err := suite.Client.Register(fixtures.RegisterRequest{
		Email:    email,
		Password: password,
	})

	// Should succeed
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	t.Logf("Successfully registered user: %s", email)
}

// TestUserRegistrationDuplicate tests that duplicate registration is rejected
func TestUserRegistrationDuplicate(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	email := fmt.Sprintf("test-dup-%d@example.com", time.Now().UnixNano())
	password := "TestPass123!"

	// First registration should succeed
	_, err := suite.Client.Register(fixtures.RegisterRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		t.Fatalf("First registration failed: %v", err)
	}

	// Second registration with same email should fail
	_, err = suite.Client.Register(fixtures.RegisterRequest{
		Email:    email,
		Password: password,
	})
	if err == nil {
		t.Error("Expected error for duplicate registration, got nil")
	}

	t.Logf("Correctly rejected duplicate registration for: %s", email)
}

// TestUserRegistrationInvalidEmail tests that invalid emails are rejected
func TestUserRegistrationInvalidEmail(t *testing.T) {
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
		{"no local part", "@example.com"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := suite.Client.Register(fixtures.RegisterRequest{
				Email:    tc.email,
				Password: "TestPass123!",
			})
			if err == nil {
				t.Errorf("Expected error for invalid email %q, got nil", tc.email)
			}
		})
	}
}

// TestUserRegistrationWeakPassword tests that weak passwords are rejected
func TestUserRegistrationWeakPassword(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	testCases := []struct {
		name     string
		password string
	}{
		{"too short", "short"},
		{"no numbers", "NoNumbers!"},
		{"no uppercase", "nouppercase123!"},
		{"empty", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := suite.Client.Register(fixtures.RegisterRequest{
				Email:    "test-" + tc.password + "@example.com",
				Password: tc.password,
			})
			if err == nil {
				t.Errorf("Expected error for weak password, got nil")
			}
		})
	}
}

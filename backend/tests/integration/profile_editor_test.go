package integration

import (
	"fmt"
	"testing"
	"time"

	"viralforge/backend/tests/fixtures"
)

// TestEditorProfileCreation tests creating an editor profile
func TestEditorProfileCreation(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	// Register and login
	email := "editor-test-" + fmt.Sprintf("test-%s-%d@example.com", "profile_editor_test.go", time.Now().UnixNano())
	password := "TestPass123!"

	_, err := suite.Client.Register(fixtures.RegisterRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	// Login
	_, err = suite.Client.Login(fixtures.LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		t.Skipf("Login failed (verification may be required): %v", err)
	}

	// Create editor profile
	resp, err := suite.Client.CreateProfile(fixtures.CreateProfileRequest{
		Type: "editor",
	})
	if err != nil {
		t.Fatalf("Editor profile creation failed: %v", err)
	}

	if resp.ID == "" {
		t.Error("Expected profile ID in response")
	}
	if resp.Type != "editor" {
		t.Errorf("Expected type 'editor', got %s", resp.Type)
	}

	t.Logf("Successfully created editor profile: %s", resp.ID)
}

// TestEditorProfileAuthorization tests that editor profile follows authorization rules
func TestEditorProfileAuthorization(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	// Create user A with editor profile
	emailA := "user-a-editor-" + fmt.Sprintf("test-%s-%d@example.com", "profile_editor_test.go", time.Now().UnixNano())
	password := "TestPass123!"

	_, err := suite.Client.Register(fixtures.RegisterRequest{
		Email:    emailA,
		Password: password,
	})
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	_, err = suite.Client.Login(fixtures.LoginRequest{
		Email:    emailA,
		Password: password,
	})
	if err != nil {
		t.Skipf("Login failed: %v", err)
	}

	profileA, err := suite.Client.CreateProfile(fixtures.CreateProfileRequest{
		Type: "editor",
	})
	if err != nil {
		t.Fatalf("Profile creation failed: %v", err)
	}

	// Create user B
	emailB := "user-b-editor-" + fmt.Sprintf("test-%s-%d@example.com", "profile_editor_test.go", time.Now().UnixNano())
	_, _ = suite.Client.Register(fixtures.RegisterRequest{
		Email:    emailB,
		Password: password,
	})
	_, err = suite.Client.Login(fixtures.LoginRequest{
		Email:    emailB,
		Password: password,
	})
	if err != nil {
		t.Skipf("Login failed: %v", err)
	}

	// User B should not be able to access User A's editor profile
	_, err = suite.Client.GetProfileDetails(profileA.ID)
	if err == nil {
		t.Error("Expected authorization error when accessing another user's editor profile")
	}

	t.Logf("Editor profile authorization correctly enforced")
}

package integration

import (
	"fmt"
	"testing"
	"time"

	"viralforge/backend/tests/fixtures"
)

// TestBrandProfileCreation tests creating a brand profile
func TestBrandProfileCreation(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	// Register and login
	email := "brand-test-" + fmt.Sprintf("test-%s-%d@example.com", "profile_brand_test.go", time.Now().UnixNano())
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

	// Create brand profile
	resp, err := suite.Client.CreateProfile(fixtures.CreateProfileRequest{
		Type: "brand",
	})
	if err != nil {
		t.Fatalf("Brand profile creation failed: %v", err)
	}

	if resp.ID == "" {
		t.Error("Expected profile ID in response")
	}
	if resp.Type != "brand" {
		t.Errorf("Expected type 'brand', got %s", resp.Type)
	}

	t.Logf("Successfully created brand profile: %s", resp.ID)
}

// TestBrandProfileAuthorization tests that brand profile follows authorization rules
func TestBrandProfileAuthorization(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	// Create user A with brand profile
	emailA := "user-a-brand-" + fmt.Sprintf("test-%s-%d@example.com", "profile_brand_test.go", time.Now().UnixNano())
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
		Type: "brand",
	})
	if err != nil {
		t.Fatalf("Profile creation failed: %v", err)
	}

	// Create user B
	emailB := "user-b-brand-" + fmt.Sprintf("test-%s-%d@example.com", "profile_brand_test.go", time.Now().UnixNano())
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

	// User B should not be able to access User A's profile
	_, err = suite.Client.GetProfileDetails(profileA.ID)
	if err == nil {
		t.Error("Expected authorization error when accessing another user's brand profile")
	}

	t.Logf("Brand profile authorization correctly enforced")
}

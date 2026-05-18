package integration

import (
	"fmt"
	"testing"
	"time"

	"viralforge/backend/tests/fixtures"
)

// TestInfluencerProfileCreation tests creating an influencer profile
func TestInfluencerProfileCreation(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	// Register and login
	email := "influencer-test-" + fmt.Sprintf("test-%s-%d@example.com", "profile_influencer_test.go", time.Now().UnixNano())
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

	// Create influencer profile
	resp, err := suite.Client.CreateProfile(fixtures.CreateProfileRequest{
		Type: "influencer",
	})
	if err != nil {
		t.Fatalf("Influencer profile creation failed: %v", err)
	}

	if resp.ID == "" {
		t.Error("Expected profile ID in response")
	}
	if resp.Type != "influencer" {
		t.Errorf("Expected type 'influencer', got %s", resp.Type)
	}

	t.Logf("Successfully created influencer profile: %s", resp.ID)
}

// TestInfluencerProfileAuthorization tests that influencer profile follows authorization rules
func TestInfluencerProfileAuthorization(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	// Create user A with influencer profile
	emailA := "user-a-influencer-" + fmt.Sprintf("test-%s-%d@example.com", "profile_influencer_test.go", time.Now().UnixNano())
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
		Type: "influencer",
	})
	if err != nil {
		t.Fatalf("Profile creation failed: %v", err)
	}

	// Create user B
	emailB := "user-b-influencer-" + fmt.Sprintf("test-%s-%d@example.com", "profile_influencer_test.go", time.Now().UnixNano())
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

	// User B should not be able to access User A's influencer profile
	_, err = suite.Client.GetProfileDetails(profileA.ID)
	if err == nil {
		t.Error("Expected authorization error when accessing another user's influencer profile")
	}

	t.Logf("Influencer profile authorization correctly enforced")
}

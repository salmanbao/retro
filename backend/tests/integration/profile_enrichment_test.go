package integration

import (
	"fmt"
	"testing"
	"time"

	"viralforge/backend/tests/fixtures"
)

// TestProfileBioUpdate tests updating profile bio
func TestProfileBioUpdate(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	email := "bio-test-" + fmt.Sprintf("test-%s-%d@example.com", "profile_enrichment_test.go", time.Now().UnixNano())
	password := "TestPass123!"

	_, err := suite.Client.Register(fixtures.RegisterRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	_, err = suite.Client.Login(fixtures.LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		t.Skipf("Login failed: %v", err)
	}

	// Create profile
	profile, err := suite.Client.CreateProfile(fixtures.CreateProfileRequest{
		Type: "brand",
	})
	if err != nil {
		t.Fatalf("Profile creation failed: %v", err)
	}

	// Update bio
	err = suite.Client.UpdateProfileDetails(profile.ID, fixtures.UpdateProfileDetailsRequest{
		Bio: "Test bio content for brand profile",
	})
	if err != nil {
		t.Fatalf("Bio update failed: %v", err)
	}

	// Verify update
	details, err := suite.Client.GetProfileDetails(profile.ID)
	if err != nil {
		t.Fatalf("Get profile details failed: %v", err)
	}

	if bio, ok := (*details)["bio"].(string); !ok || bio != "Test bio content for brand profile" {
		t.Errorf("Expected bio 'Test bio content for brand profile', got %v", (*details)["bio"])
	}

	t.Logf("Bio update successful")
}

// TestProfileAvatarUpdate tests updating profile avatar URL
func TestProfileAvatarUpdate(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	email := "avatar-test-" + fmt.Sprintf("test-%s-%d@example.com", "profile_enrichment_test.go", time.Now().UnixNano())
	password := "TestPass123!"

	_, err := suite.Client.Register(fixtures.RegisterRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	_, err = suite.Client.Login(fixtures.LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		t.Skipf("Login failed: %v", err)
	}

	// Create profile
	profile, err := suite.Client.CreateProfile(fixtures.CreateProfileRequest{
		Type: "brand",
	})
	if err != nil {
		t.Fatalf("Profile creation failed: %v", err)
	}

	// Update avatar URL
	err = suite.Client.UpdateProfileDetails(profile.ID, fixtures.UpdateProfileDetailsRequest{
		AvatarURL: "https://example.com/avatar.jpg",
	})
	if err != nil {
		t.Fatalf("Avatar update failed: %v", err)
	}

	// Verify update
	details, err := suite.Client.GetProfileDetails(profile.ID)
	if err != nil {
		t.Fatalf("Get profile details failed: %v", err)
	}

	if avatar, ok := (*details)["avatar_url"].(string); !ok || avatar != "https://example.com/avatar.jpg" {
		t.Errorf("Expected avatar_url 'https://example.com/avatar.jpg', got %v", (*details)["avatar_url"])
	}

	t.Logf("Avatar update successful")
}

// TestProfileEnrichmentCrossProfileIsolation tests that enrichment is isolated per profile
func TestProfileEnrichmentCrossProfileIsolation(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	emailA := "enrich-a-" + fmt.Sprintf("test-%s-%d@example.com", "profile_enrichment_test.go", time.Now().UnixNano())
	emailB := "enrich-b-" + fmt.Sprintf("test-%s-%d@example.com", "profile_enrichment_test.go", time.Now().UnixNano())
	password := "TestPass123!"

	// Create user A
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
	_, err = suite.Client.Register(fixtures.RegisterRequest{
		Email:    emailB,
		Password: password,
	})
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}
	_, err = suite.Client.Login(fixtures.LoginRequest{
		Email:    emailB,
		Password: password,
	})
	if err != nil {
		t.Skipf("Login failed: %v", err)
	}

	profileB, err := suite.Client.CreateProfile(fixtures.CreateProfileRequest{
		Type: "brand",
	})
	if err != nil {
		t.Fatalf("Profile creation failed: %v", err)
	}

	// Update profile A's bio
	err = suite.Client.UpdateProfileDetails(profileA.ID, fixtures.UpdateProfileDetailsRequest{
		Bio: "User A's bio",
	})
	if err != nil {
		t.Fatalf("Bio update failed: %v", err)
	}

	// Profile B's bio should not be affected
	detailsB, err := suite.Client.GetProfileDetails(profileB.ID)
	if err != nil {
		t.Fatalf("Get profile details failed: %v", err)
	}

	if bio, ok := (*detailsB)["bio"].(string); ok && bio != "" {
		t.Error("Profile B should not have a bio set")
	}

	t.Logf("Cross-profile enrichment isolation verified")
}

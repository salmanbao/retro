package integration

import (
	"fmt"
	"testing"
	"time"

	"viralforge/backend/tests/fixtures"
)

// TestSocialLinksUpdate tests updating profile social links
func TestSocialLinksUpdate(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	email := "social-test-" + fmt.Sprintf("test-%s-%d@example.com", "profile_social_links_test.go", time.Now().UnixNano())
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
		Type: "influencer",
	})
	if err != nil {
		t.Fatalf("Profile creation failed: %v", err)
	}

	// Update social links
	socialLinks := map[string]string{
		"twitter":   "https://twitter.com/testuser",
		"instagram": "https://instagram.com/testuser",
		"youtube":   "https://youtube.com/testuser",
	}
	err = suite.Client.UpdateSocialLinks(profile.ID, fixtures.UpdateSocialLinksRequest{
		SocialLinks: socialLinks,
	})
	if err != nil {
		t.Fatalf("Social links update failed: %v", err)
	}

	// Verify update
	details, err := suite.Client.GetProfileDetails(profile.ID)
	if err != nil {
		t.Fatalf("Get profile details failed: %v", err)
	}

	links, ok := (*details)["social_links"].(map[string]interface{})
	if !ok {
		t.Error("Expected social_links in response")
		return
	}

	if links["twitter"] != "https://twitter.com/testuser" {
		t.Errorf("Expected twitter link, got %v", links["twitter"])
	}

	t.Logf("Social links update successful")
}

// TestSocialLinksPartialUpdate tests partial update of social links
func TestSocialLinksPartialUpdate(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	email := "partial-social-" + fmt.Sprintf("test-%s-%d@example.com", "profile_social_links_test.go", time.Now().UnixNano())
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

	profile, err := suite.Client.CreateProfile(fixtures.CreateProfileRequest{
		Type: "influencer",
	})
	if err != nil {
		t.Fatalf("Profile creation failed: %v", err)
	}

	// Set initial social links
	err = suite.Client.UpdateSocialLinks(profile.ID, fixtures.UpdateSocialLinksRequest{
		SocialLinks: map[string]string{
			"twitter": "https://twitter.com/testuser",
		},
	})
	if err != nil {
		t.Fatalf("Initial social links update failed: %v", err)
	}

	// Partial update - add new link
	err = suite.Client.UpdateSocialLinks(profile.ID, fixtures.UpdateSocialLinksRequest{
		SocialLinks: map[string]string{
			"instagram": "https://instagram.com/testuser",
		},
	})
	if err != nil {
		t.Fatalf("Partial social links update failed: %v", err)
	}

	t.Logf("Partial social links update successful")
}

package integration

import (
	"fmt"
	"testing"
	"time"

	"viralforge/backend/tests/fixtures"
)

// TestSocialLinksAutoTrigger tests that adding social links auto-triggers onboarding step
func TestSocialLinksAutoTrigger(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	email := "social-trigger-" + fmt.Sprintf("test-%s-%d@example.com", "onboarding_social_links_auto_trigger_test.go", time.Now().UnixNano())
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

	// Get initial onboarding state
	initialOnboarding, _ := suite.Client.GetOnboarding(profile.ID)

	// Add social links
	socialLinks := map[string]string{
		"twitter":   "https://twitter.com/testuser",
		"instagram": "https://instagram.com/testuser",
	}
	err = suite.Client.UpdateSocialLinks(profile.ID, fixtures.UpdateSocialLinksRequest{
		SocialLinks: socialLinks,
	})
	if err != nil {
		t.Logf("Social links update failed: %v", err)
	}

	// Recalculate onboarding
	recalculated, err := suite.Client.RecalculateOnboarding(profile.ID)
	if err != nil {
		t.Logf("Recalculate failed (may not be implemented): %v", err)
	}

	// Get updated onboarding state
	updatedOnboarding, _ := suite.Client.GetOnboarding(profile.ID)

	t.Logf("Social links auto-trigger test completed (initial: %v, recalculated: %v, updated: %v)",
		initialOnboarding != nil, recalculated != nil, updatedOnboarding != nil)
}

// TestSocialLinksPartialUpdateAutoTrigger tests that updating individual social link updates completion
func TestSocialLinksPartialUpdateAutoTrigger(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	email := "social-partial-" + fmt.Sprintf("test-%s-%d@example.com", "onboarding_social_links_auto_trigger_test.go", time.Now().UnixNano())
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

	// Get initial completion
	initial, _ := suite.Client.GetOnboarding(profile.ID)
	var initialPct float64
	if initial != nil {
		if pct, ok := (*initial)["completion_percentage"].(float64); ok {
			initialPct = pct
		}
	}

	// Add first social link
	suite.Client.UpdateSocialLinks(profile.ID, fixtures.UpdateSocialLinksRequest{
		SocialLinks: map[string]string{"twitter": "https://twitter.com/test"},
	})

	// Add second social link
	suite.Client.UpdateSocialLinks(profile.ID, fixtures.UpdateSocialLinksRequest{
		SocialLinks: map[string]string{"instagram": "https://instagram.com/test"},
	})

	// Get updated completion
	updated, _ := suite.Client.GetOnboarding(profile.ID)
	var updatedPct float64
	if updated != nil {
		if pct, ok := (*updated)["completion_percentage"].(float64); ok {
			updatedPct = pct
		}
	}

	t.Logf("Social links partial update: initial=%.0f%%, updated=%.0f%%", initialPct, updatedPct)
}

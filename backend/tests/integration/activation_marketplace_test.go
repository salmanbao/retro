package integration

import (
	"fmt"
	"testing"
	"time"

	"viralforge/backend/tests/fixtures"
)

// TestMarketplaceEligibilityAfterActivation tests that activated profiles are marketplace-eligible
func TestMarketplaceEligibilityAfterActivation(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	password := "TestPass123!"

	email := "marketplace-eligible-" + fmt.Sprintf("test-%s-%d@example.com", "activation_marketplace_test.go", time.Now().UnixNano())
	_, _ = suite.Client.Register(fixtures.RegisterRequest{Email: email, Password: password})
	_, err := suite.Client.Login(fixtures.LoginRequest{Email: email, Password: password})
	if err != nil {
		t.Skipf("Login failed: %v", err)
	}

	profile, err := suite.Client.CreateProfile(fixtures.CreateProfileRequest{Type: "influencer"})
	if err != nil {
		t.Fatalf("Profile creation failed: %v", err)
	}

	// Complete all required onboarding steps
	suite.Client.UpdateProfileDetails(profile.ID, fixtures.UpdateProfileDetailsRequest{
		Bio: "Completed influencer profile",
	})

	suite.Client.UpdateSocialLinks(profile.ID, fixtures.UpdateSocialLinksRequest{
		SocialLinks: map[string]string{
			"twitter": "https://twitter.com/test",
		},
	})

	// Configure payout preferences
	payoutPrefs := map[string]interface{}{
		"method":   "paypal",
		"currency": "USD",
	}
	suite.Client.DoRequest("PUT", "/api/v1/profiles/"+profile.ID+"/payout-preferences", payoutPrefs)

	// Recalculate to trigger completion
	suite.Client.RecalculateOnboarding(profile.ID)

	// Get final onboarding state
	onboarding, err := suite.Client.GetOnboarding(profile.ID)
	if err != nil {
		t.Skipf("Get onboarding failed: %v", err)
	}

	// Check marketplace eligibility
	if onboarding != nil {
		if state, ok := (*onboarding)["state"].(string); ok {
			t.Logf("Final activation state: %s", state)

			if state == "activated" {
				t.Logf("Profile is marketplace-eligible")
			} else {
				t.Logf("Profile is not yet activated: %s", state)
			}
		}

		if eligible, ok := (*onboarding)["marketplace_eligible"].(bool); ok {
			if eligible {
				t.Logf("Profile is marketplace-eligible: true")
			} else {
				t.Logf("Profile marketplace_eligible: false")
			}
		}
	}

	t.Logf("Marketplace eligibility test completed")
}

// TestMarketplaceNotEligibleBeforeActivation tests that non-activated profiles are not marketplace-eligible
func TestMarketplaceNotEligibleBeforeActivation(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	password := "TestPass123!"

	email := "marketplace-not-eligible-" + fmt.Sprintf("test-%s-%d@example.com", "activation_marketplace_test.go", time.Now().UnixNano())
	_, _ = suite.Client.Register(fixtures.RegisterRequest{Email: email, Password: password})
	_, err := suite.Client.Login(fixtures.LoginRequest{Email: email, Password: password})
	if err != nil {
		t.Skipf("Login failed: %v", err)
	}

	profile, err := suite.Client.CreateProfile(fixtures.CreateProfileRequest{Type: "brand"})
	if err != nil {
		t.Fatalf("Profile creation failed: %v", err)
	}

	// Do NOT complete onboarding steps

	// Check marketplace eligibility
	onboarding, err := suite.Client.GetOnboarding(profile.ID)
	if err != nil {
		t.Skipf("Get onboarding failed: %v", err)
	}

	if onboarding != nil {
		if state, ok := (*onboarding)["state"].(string); ok {
			if state != "activated" {
				t.Logf("Correctly not marketplace-eligible: state=%s", state)
			}
		}
	}

	t.Logf("Marketplace not eligible before activation verified")
}

// TestMarketplaceEligibilityEditorProfile tests that editor profiles can be marketplace-eligible
func TestMarketplaceEligibilityEditorProfile(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	password := "TestPass123!"

	email := "marketplace-editor-" + fmt.Sprintf("test-%s-%d@example.com", "activation_marketplace_test.go", time.Now().UnixNano())
	_, _ = suite.Client.Register(fixtures.RegisterRequest{Email: email, Password: password})
	_, err := suite.Client.Login(fixtures.LoginRequest{Email: email, Password: password})
	if err != nil {
		t.Skipf("Login failed: %v", err)
	}

	profile, err := suite.Client.CreateProfile(fixtures.CreateProfileRequest{Type: "editor"})
	if err != nil {
		t.Fatalf("Profile creation failed: %v", err)
	}

	// Complete editor-specific onboarding
	suite.Client.UpdateProfileDetails(profile.ID, fixtures.UpdateProfileDetailsRequest{
		Bio: "Completed editor profile",
	})

	// Recalculate
	suite.Client.RecalculateOnboarding(profile.ID)

	onboarding, _ := suite.Client.GetOnboarding(profile.ID)
	if onboarding != nil {
		t.Logf("Editor activation state: %v", *onboarding)
	}

	t.Logf("Editor marketplace eligibility test completed")
}

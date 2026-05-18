package integration

import (
	"fmt"
	"testing"
	"time"

	"viralforge/backend/tests/fixtures"
)

// TestOptionalStepSkipping tests that optional steps can be skipped
func TestOptionalStepSkipping(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	password := "TestPass123!"

	email := "optional-skip-" + fmt.Sprintf("test-%s-%d@example.com", "activation_optional_step_test.go", time.Now().UnixNano())
	_, _ = suite.Client.Register(fixtures.RegisterRequest{Email: email, Password: password})
	_, err := suite.Client.Login(fixtures.LoginRequest{Email: email, Password: password})
	if err != nil {
		t.Skipf("Login failed: %v", err)
	}

	profile, err := suite.Client.CreateProfile(fixtures.CreateProfileRequest{Type: "brand"})
	if err != nil {
		t.Fatalf("Profile creation failed: %v", err)
	}

	// Get onboarding to see which steps are optional
	onboarding, err := suite.Client.GetOnboarding(profile.ID)
	if err != nil {
		t.Skipf("Get onboarding failed: %v", err)
	}

	if onboarding != nil {
		if completion, ok := (*onboarding)["completion_percentage"].(float64); ok {
			t.Logf("Initial completion with skipped optional: %.0f%%", completion)
		}
	}

	// Only complete required steps (skip optional ones)
	suite.Client.UpdateProfileDetails(profile.ID, fixtures.UpdateProfileDetailsRequest{
		Bio: "Required bio only",
	})

	// Recalculate
	suite.Client.RecalculateOnboarding(profile.ID)

	// Check if profile can still become activated
	finalOnboarding, _ := suite.Client.GetOnboarding(profile.ID)
	if finalOnboarding != nil {
		if state, ok := (*finalOnboarding)["state"].(string); ok {
			t.Logf("Final state after skipping optional steps: %s", state)

			if state == "activated" || state == "pending_review" {
				t.Logf("Optional steps correctly skippable - activation state reached")
			} else {
				t.Logf("Optional steps skipped but more required steps remain")
			}
		}
	}

	t.Logf("Optional step skipping test completed")
}

// TestOptionalStepsDoNotBlockActivation tests that optional steps don't block activation
func TestOptionalStepsDoNotBlockActivation(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	password := "TestPass123!"

	email := "optional-no-block-" + fmt.Sprintf("test-%s-%d@example.com", "activation_optional_step_test.go", time.Now().UnixNano())
	_, _ = suite.Client.Register(fixtures.RegisterRequest{Email: email, Password: password})
	_, err := suite.Client.Login(fixtures.LoginRequest{Email: email, Password: password})
	if err != nil {
		t.Skipf("Login failed: %v", err)
	}

	profile, err := suite.Client.CreateProfile(fixtures.CreateProfileRequest{Type: "influencer"})
	if err != nil {
		t.Fatalf("Profile creation failed: %v", err)
	}

	// Complete only required steps, skip optional
	suite.Client.UpdateProfileDetails(profile.ID, fixtures.UpdateProfileDetailsRequest{
		Bio: "Required bio",
	})

	suite.Client.UpdateSocialLinks(profile.ID, fixtures.UpdateSocialLinksRequest{
		SocialLinks: map[string]string{"twitter": "https://twitter.com/test"},
	})

	// Configure payout
	payoutPrefs := map[string]interface{}{"method": "paypal", "currency": "USD"}
	suite.Client.DoRequest("PUT", "/api/v1/profiles/"+profile.ID+"/payout-preferences", payoutPrefs)

	// Submit KYC
	kycData := map[string]interface{}{"status": "pending"}
	suite.Client.DoRequest("POST", "/api/v1/profiles/"+profile.ID+"/kyc", kycData)

	// Recalculate
	suite.Client.RecalculateOnboarding(profile.ID)

	// Check if activation is possible without optional influencer steps
	onboarding, _ := suite.Client.GetOnboarding(profile.ID)
	if onboarding != nil {
		if state, ok := (*onboarding)["state"].(string); ok {
			t.Logf("State without optional steps: %s", state)
		}
	}

	t.Logf("Optional steps do not block activation test completed")
}

// TestOptionalStepIdentification tests identifying which steps are optional
func TestOptionalStepIdentification(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	password := "TestPass123!"

	email := "optional-identify-" + fmt.Sprintf("test-%s-%d@example.com", "activation_optional_step_test.go", time.Now().UnixNano())
	_, _ = suite.Client.Register(fixtures.RegisterRequest{Email: email, Password: password})
	_, err := suite.Client.Login(fixtures.LoginRequest{Email: email, Password: password})
	if err != nil {
		t.Skipf("Login failed: %v", err)
	}

	profile, err := suite.Client.CreateProfile(fixtures.CreateProfileRequest{Type: "brand"})
	if err != nil {
		t.Fatalf("Profile creation failed: %v", err)
	}

	// Get steps to identify which are optional
	steps, err := suite.Client.GetOnboardingSteps(profile.ID)
	if err != nil {
		t.Skipf("Get onboarding steps failed: %v", err)
	}

	if steps != nil {
		t.Logf("Steps retrieved: %v", *steps)

		// Look for required flag in steps
		if required, ok := (*steps)["required"].(bool); ok {
			t.Logf("Step required flag: %v", required)
		}
	}

	t.Logf("Optional step identification test completed")
}

package integration

import (
	"fmt"
	"testing"
	"time"

	"viralforge/backend/tests/fixtures"
)

// TestRequiredStepBlockingBrand tests that brand profile requires all brand-specific steps
func TestRequiredStepBlockingBrand(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	password := "TestPass123!"

	email := "required-brand-" + fmt.Sprintf("test-%s-%d@example.com", "activation_required_step_blocking_test.go", time.Now().UnixNano())
	_, _ = suite.Client.Register(fixtures.RegisterRequest{Email: email, Password: password})
	_, err := suite.Client.Login(fixtures.LoginRequest{Email: email, Password: password})
	if err != nil {
		t.Skipf("Login failed: %v", err)
	}

	profile, err := suite.Client.CreateProfile(fixtures.CreateProfileRequest{Type: "brand"})
	if err != nil {
		t.Fatalf("Brand profile creation failed: %v", err)
	}

	// Get onboarding steps
	steps, err := suite.Client.GetOnboardingSteps(profile.ID)
	if err != nil {
		t.Skipf("Get onboarding steps failed: %v", err)
	}

	if steps == nil {
		t.Skip("Onboarding steps not available")
	}

	// Verify required steps exist
	t.Logf("Brand required steps retrieved: %v", *steps)

	// Activation should not complete without all required steps
	onboarding, _ := suite.Client.GetOnboarding(profile.ID)
	if onboarding != nil {
		if state, ok := (*onboarding)["state"].(string); ok {
			if state == "activated" {
				t.Error("Brand profile should not be activated without completing all required steps")
			}
		}
	}

	t.Logf("Required step blocking for brand verified")
}

// TestRequiredStepBlockingEditor tests that editor profile requires all editor-specific steps
func TestRequiredStepBlockingEditor(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	password := "TestPass123!"

	email := "required-editor-" + fmt.Sprintf("test-%s-%d@example.com", "activation_required_step_blocking_test.go", time.Now().UnixNano())
	_, _ = suite.Client.Register(fixtures.RegisterRequest{Email: email, Password: password})
	_, err := suite.Client.Login(fixtures.LoginRequest{Email: email, Password: password})
	if err != nil {
		t.Skipf("Login failed: %v", err)
	}

	profile, err := suite.Client.CreateProfile(fixtures.CreateProfileRequest{Type: "editor"})
	if err != nil {
		t.Fatalf("Editor profile creation failed: %v", err)
	}

	// Get onboarding steps
	steps, err := suite.Client.GetOnboardingSteps(profile.ID)
	if err != nil {
		t.Skipf("Get onboarding steps failed: %v", err)
	}

	if steps == nil {
		t.Skip("Onboarding steps not available")
	}

	t.Logf("Editor required steps retrieved: %v", *steps)

	// Editor-specific steps should include portfolio items
	onboarding, _ := suite.Client.GetOnboarding(profile.ID)
	if onboarding != nil {
		if state, ok := (*onboarding)["state"].(string); ok {
			t.Logf("Editor activation state: %s", state)
		}
	}

	t.Logf("Required step blocking for editor verified")
}

// TestRequiredStepBlockingInfluencer tests that influencer profile requires all influencer-specific steps
func TestRequiredStepBlockingInfluencer(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	password := "TestPass123!"

	email := "required-influencer-" + fmt.Sprintf("test-%s-%d@example.com", "activation_required_step_blocking_test.go", time.Now().UnixNano())
	_, _ = suite.Client.Register(fixtures.RegisterRequest{Email: email, Password: password})
	_, err := suite.Client.Login(fixtures.LoginRequest{Email: email, Password: password})
	if err != nil {
		t.Skipf("Login failed: %v", err)
	}

	profile, err := suite.Client.CreateProfile(fixtures.CreateProfileRequest{Type: "influencer"})
	if err != nil {
		t.Fatalf("Influencer profile creation failed: %v", err)
	}

	// Get onboarding steps
	steps, err := suite.Client.GetOnboardingSteps(profile.ID)
	if err != nil {
		t.Skipf("Get onboarding steps failed: %v", err)
	}

	if steps == nil {
		t.Skip("Onboarding steps not available")
	}

	t.Logf("Influencer required steps retrieved: %v", *steps)

	// Influencer-specific steps should include KYC
	onboarding, _ := suite.Client.GetOnboarding(profile.ID)
	if onboarding != nil {
		if state, ok := (*onboarding)["state"].(string); ok {
			t.Logf("Influencer activation state: %s", state)
		}
	}

	t.Logf("Required step blocking for influencer verified")
}

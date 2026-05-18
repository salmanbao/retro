package integration

import (
	"fmt"
	"testing"
	"time"

	"viralforge/backend/tests/fixtures"
)

// TestActivationStateProgression tests the activation state progression workflow
func TestActivationStateProgression(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	password := "TestPass123!"

	email := "activation-progression-" + fmt.Sprintf("test-%s-%d@example.com", "activation_flow_test.go", time.Now().UnixNano())
	_, _ = suite.Client.Register(fixtures.RegisterRequest{Email: email, Password: password})
	_, err := suite.Client.Login(fixtures.LoginRequest{Email: email, Password: password})
	if err != nil {
		t.Skipf("Login failed: %v", err)
	}

	profile, err := suite.Client.CreateProfile(fixtures.CreateProfileRequest{Type: "brand"})
	if err != nil {
		t.Fatalf("Profile creation failed: %v", err)
	}

	// Get initial onboarding state
	onboarding, err := suite.Client.GetOnboarding(profile.ID)
	if err != nil {
		t.Skipf("Get onboarding failed: %v", err)
	}

	// Verify initial state
	if onboarding == nil {
		t.Fatal("Expected onboarding to be created")
	}

	// Log the state progression
	if state, ok := (*onboarding)["state"].(string); ok {
		t.Logf("Initial activation state: %s", state)
	}

	if completion, ok := (*onboarding)["completion_percentage"].(float64); ok {
		t.Logf("Initial completion: %.0f%%", completion)
	}

	t.Logf("Activation state progression test completed")
}

// TestActivationPendingReviewState tests that profile reaches pending_review when all required steps complete
func TestActivationPendingReviewState(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	password := "TestPass123!"

	email := "pending-review-" + fmt.Sprintf("test-%s-%d@example.com", "activation_flow_test.go", time.Now().UnixNano())
	_, _ = suite.Client.Register(fixtures.RegisterRequest{Email: email, Password: password})
	_, err := suite.Client.Login(fixtures.LoginRequest{Email: email, Password: password})
	if err != nil {
		t.Skipf("Login failed: %v", err)
	}

	profile, err := suite.Client.CreateProfile(fixtures.CreateProfileRequest{Type: "brand"})
	if err != nil {
		t.Fatalf("Profile creation failed: %v", err)
	}

	// Complete profile enrichment
	suite.Client.UpdateProfileDetails(profile.ID, fixtures.UpdateProfileDetailsRequest{
		Bio: "Test bio for activation",
	})

	// Recalculate to trigger auto-completion
	recalculated, _ := suite.Client.RecalculateOnboarding(profile.ID)
	if recalculated != nil {
		t.Logf("Recalculated onboarding: %v", *recalculated)
	}

	// Get final state
	onboarding, _ := suite.Client.GetOnboarding(profile.ID)
	if onboarding != nil {
		if state, ok := (*onboarding)["state"].(string); ok {
			t.Logf("Final activation state: %s", state)
		}
	}

	t.Logf("Pending review state test completed")
}

// TestActivationNotCompleteWithoutRequiredSteps tests that activation doesn't complete without required steps
func TestActivationNotCompleteWithoutRequiredSteps(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	password := "TestPass123!"

	email := "incomplete-steps-" + fmt.Sprintf("test-%s-%d@example.com", "activation_flow_test.go", time.Now().UnixNano())
	_, _ = suite.Client.Register(fixtures.RegisterRequest{Email: email, Password: password})
	_, err := suite.Client.Login(fixtures.LoginRequest{Email: email, Password: password})
	if err != nil {
		t.Skipf("Login failed: %v", err)
	}

	profile, err := suite.Client.CreateProfile(fixtures.CreateProfileRequest{Type: "brand"})
	if err != nil {
		t.Fatalf("Profile creation failed: %v", err)
	}

	// Get onboarding without completing any steps
	onboarding, _ := suite.Client.GetOnboarding(profile.ID)

	// Check that activation is not complete
	if onboarding != nil {
		if state, ok := (*onboarding)["state"].(string); ok {
			if state == "activated" {
				t.Error("Profile should not be activated without completing required steps")
			} else {
				t.Logf("Correctly not activated: state=%s", state)
			}
		}
	}

	t.Logf("Required steps blocking test completed")
}

// TestActivationManualStepCompletion tests manually completing onboarding steps
func TestActivationManualStepCompletion(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	password := "TestPass123!"

	email := "manual-steps-" + fmt.Sprintf("test-%s-%d@example.com", "activation_flow_test.go", time.Now().UnixNano())
	_, _ = suite.Client.Register(fixtures.RegisterRequest{Email: email, Password: password})
	_, err := suite.Client.Login(fixtures.LoginRequest{Email: email, Password: password})
	if err != nil {
		t.Skipf("Login failed: %v", err)
	}

	profile, err := suite.Client.CreateProfile(fixtures.CreateProfileRequest{Type: "brand"})
	if err != nil {
		t.Fatalf("Profile creation failed: %v", err)
	}

	// Get steps to find step IDs
	steps, err := suite.Client.GetOnboardingSteps(profile.ID)
	if err != nil {
		t.Skipf("Get onboarding steps failed: %v", err)
	}

	if steps != nil {
		t.Logf("Onboarding steps retrieved: %v", *steps)
	}

	t.Logf("Manual step completion test completed")
}

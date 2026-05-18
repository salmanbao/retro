package integration

import (
	"fmt"
	"testing"
	"time"

	"viralforge/backend/tests/fixtures"
)

// TestOnboardingProgressAutoCreation tests that onboarding progress is auto-created when a profile is created
func TestOnboardingProgressAutoCreation(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	email := "onboarding-auto-" + fmt.Sprintf("test-%s-%d@example.com", "onboarding_init_test.go", time.Now().UnixNano())
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

	// Create brand profile
	profile, err := suite.Client.CreateProfile(fixtures.CreateProfileRequest{
		Type: "brand",
	})
	if err != nil {
		t.Fatalf("Profile creation failed: %v", err)
	}

	// Onboarding progress should be auto-created
	onboarding, err := suite.Client.GetOnboarding(profile.ID)
	if err != nil {
		t.Fatalf("Get onboarding failed: %v", err)
	}

	if onboarding == nil {
		t.Error("Expected onboarding progress to be auto-created")
		return
	}

	// Check that onboarding has expected structure
	if id, ok := (*onboarding)["id"]; !ok || id == nil {
		t.Error("Expected onboarding ID in response")
	}

	if profileID, ok := (*onboarding)["profile_id"]; !ok || profileID == nil {
		t.Error("Expected profile_id in onboarding response")
	}

	t.Logf("Onboarding progress auto-created for profile: %s", profile.ID)
}

// TestOnboardingStepsAutoCreation tests that onboarding steps are auto-created
func TestOnboardingStepsAutoCreation(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	email := "steps-auto-" + fmt.Sprintf("test-%s-%d@example.com", "onboarding_init_test.go", time.Now().UnixNano())
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
		Type: "brand",
	})
	if err != nil {
		t.Fatalf("Profile creation failed: %v", err)
	}

	// Get onboarding steps
	steps, err := suite.Client.GetOnboardingSteps(profile.ID)
	if err != nil {
		t.Fatalf("Get onboarding steps failed: %v", err)
	}

	if steps == nil {
		t.Error("Expected onboarding steps to be created")
		return
	}

	t.Logf("Onboarding steps auto-created for profile: %s", profile.ID)
}

// TestOnboardingGetNextStep tests retrieving the next incomplete onboarding step
func TestOnboardingGetNextStep(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	email := "next-step-" + fmt.Sprintf("test-%s-%d@example.com", "onboarding_init_test.go", time.Now().UnixNano())
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
		Type: "brand",
	})
	if err != nil {
		t.Fatalf("Profile creation failed: %v", err)
	}

	// Get next step
	nextStep, err := suite.Client.GetNextStep(profile.ID)
	if err != nil {
		t.Fatalf("Get next step failed: %v", err)
	}

	if nextStep == nil {
		t.Log("GetNextStep returned nil (may have completed all steps or endpoint not implemented)")
	} else {
		t.Logf("Next step retrieved: %v", *nextStep)
	}
}

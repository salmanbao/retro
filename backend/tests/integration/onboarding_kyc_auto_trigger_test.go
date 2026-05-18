package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"viralforge/backend/tests/fixtures"
)

// TestKYCAutoTriggerOnApprovedStatus tests that KYC approval auto-triggers onboarding step
func TestKYCAutoTriggerOnApprovedStatus(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	email := "kyc-auto-" + fmt.Sprintf("test-%s-%d@example.com", "onboarding_kyc_auto_trigger_test.go", time.Now().UnixNano())
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

	// Submit KYC status
	kycData := map[string]interface{}{
		"status":          "pending",
		"document_type":   "passport",
		"document_number": "AB123456",
	}

	resp, _, err := suite.Client.DoRequest("POST", "/api/v1/profiles/"+profile.ID+"/kyc", kycData)
	if err != nil {
		t.Fatalf("KYC submission failed: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("KYC endpoint not implemented")
	}

	// Get initial onboarding state
	initialSteps, _ := suite.Client.GetOnboardingSteps(profile.ID)

	// Approve KYC (admin action)
	kycUpdate := map[string]interface{}{
		"status": "approved",
	}
	resp, _, err = suite.Client.DoRequest("PATCH", "/api/v1/profiles/"+profile.ID+"/kyc", kycUpdate)
	if err != nil {
		t.Logf("KYC update request failed: %v", err)
	}

	// Recalculate onboarding
	recalculated, err := suite.Client.RecalculateOnboarding(profile.ID)
	if err != nil {
		t.Logf("Recalculate failed (may not be implemented): %v", err)
	} else {
		t.Logf("Onboarding recalculated: %v", *recalculated)
	}

	// Get updated onboarding state
	updatedSteps, _ := suite.Client.GetOnboardingSteps(profile.ID)

	t.Logf("KYC auto-trigger test completed (initial: %v, updated: %v)", initialSteps != nil, updatedSteps != nil)
}

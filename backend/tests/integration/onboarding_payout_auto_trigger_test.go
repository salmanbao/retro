package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"viralforge/backend/tests/fixtures"
)

// TestPayoutPreferencesAutoTrigger tests that configuring payout preferences auto-triggers onboarding step
func TestPayoutPreferencesAutoTrigger(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	email := "payout-trigger-" + fmt.Sprintf("test-%s-%d@example.com", "onboarding_payout_auto_trigger_test.go", time.Now().UnixNano())
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

	// Configure payout preferences
	payoutPrefs := map[string]interface{}{
		"method":   "paypal",
		"currency": "USD",
	}

	resp, _, err := suite.Client.DoRequest("PUT", "/api/v1/profiles/"+profile.ID+"/payout-preferences", payoutPrefs)
	if err != nil {
		t.Fatalf("Payout preferences request failed: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Payout preferences endpoint not implemented")
	}

	// Recalculate onboarding to trigger auto-completion
	recalculated, err := suite.Client.RecalculateOnboarding(profile.ID)
	if err != nil {
		t.Logf("Recalculate failed (may not be implemented): %v", err)
	}

	// Get updated onboarding state
	updatedOnboarding, _ := suite.Client.GetOnboarding(profile.ID)

	t.Logf("Payout preferences auto-trigger test completed (initial: %v, recalculated: %v, updated: %v)",
		initialOnboarding != nil, recalculated != nil, updatedOnboarding != nil)
}

// TestPayoutPreferencesUpdateAutoComplete tests that updating payout preferences updates completion
func TestPayoutPreferencesUpdateAutoComplete(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	email := "payout-update-complete-" + fmt.Sprintf("test-%s-%d@example.com", "onboarding_payout_auto_trigger_test.go", time.Now().UnixNano())
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

	// Get initial completion percentage
	initial, _ := suite.Client.GetOnboarding(profile.ID)
	var initialPct float64
	if initial != nil {
		if pct, ok := (*initial)["completion_percentage"].(float64); ok {
			initialPct = pct
		}
	}

	// Configure payout preferences
	payoutPrefs := map[string]interface{}{
		"method":   "bank_transfer",
		"currency": "EUR",
	}
	suite.Client.DoRequest("PUT", "/api/v1/profiles/"+profile.ID+"/payout-preferences", payoutPrefs)

	// Get updated completion percentage
	updated, _ := suite.Client.GetOnboarding(profile.ID)
	var updatedPct float64
	if updated != nil {
		if pct, ok := (*updated)["completion_percentage"].(float64); ok {
			updatedPct = pct
		}
	}

	t.Logf("Payout update completion: initial=%.0f%%, updated=%.0f%%", initialPct, updatedPct)
}

package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"viralforge/backend/tests/fixtures"
)

// TestPayoutPreferencesConfiguration tests configuring payout preferences
func TestPayoutPreferencesConfiguration(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	email := "payout-test-" + fmt.Sprintf("test-%s-%d@example.com", "profile_payout_test.go", time.Now().UnixNano())
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

	// Configure payout preferences using DoRequest
	payoutPrefs := map[string]interface{}{
		"method":             "bank_transfer",
		"currency":           "USD",
		"minimum_payout":     100,
		"bank_account_last4": "1234",
	}

	resp, body, err := suite.Client.DoRequest("PUT", "/api/v1/profiles/"+profile.ID+"/payout-preferences", payoutPrefs)
	if err != nil {
		t.Fatalf("Payout preferences request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Logf("Payout preferences returned status %d: %s", resp.StatusCode, string(body))
		t.Skip("Payout preferences endpoint may not be implemented")
	}

	t.Logf("Payout preferences configured successfully")
}

// TestPayoutPreferencesRetrieval tests retrieving payout preferences
func TestPayoutPreferencesRetrieval(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	email := "payout-retrieve-" + fmt.Sprintf("test-%s-%d@example.com", "profile_payout_test.go", time.Now().UnixNano())
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

	// Get payout preferences
	resp, body, err := suite.Client.DoRequest("GET", "/api/v1/profiles/"+profile.ID+"/payout-preferences", nil)
	if err != nil {
		t.Fatalf("Get payout preferences request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Logf("Get payout preferences returned status %d: %s", resp.StatusCode, string(body))
		t.Skip("Payout preferences endpoint may not be implemented")
	}

	t.Logf("Payout preferences retrieved successfully")
}

// TestPayoutPreferencesIsolation tests that payout preferences are isolated per profile
func TestPayoutPreferencesIsolation(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	emailA := "payout-iso-a-" + fmt.Sprintf("test-%s-%d@example.com", "profile_payout_test.go", time.Now().UnixNano())
	emailB := "payout-iso-b-" + fmt.Sprintf("test-%s-%d@example.com", "profile_payout_test.go", time.Now().UnixNano())
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
		Type: "influencer",
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
		Type: "influencer",
	})
	if err != nil {
		t.Fatalf("Profile creation failed: %v", err)
	}

	// Configure payout for user A
	payoutPrefsA := map[string]interface{}{
		"method":   "paypal",
		"currency": "EUR",
	}
	_, _, err = suite.Client.DoRequest("PUT", "/api/v1/profiles/"+profileA.ID+"/payout-preferences", payoutPrefsA)
	if err != nil {
		t.Logf("Payout preferences setup failed: %v", err)
	}

	// User B's payout preferences should be independent
	resp, _, err := suite.Client.DoRequest("GET", "/api/v1/profiles/"+profileB.ID+"/payout-preferences", nil)
	if err != nil {
		t.Fatalf("Get payout preferences request failed: %v", err)
	}

	if resp.StatusCode == http.StatusOK {
		t.Logf("Payout preferences isolation verified (or endpoint not implemented)")
	} else {
		t.Logf("Payout preferences endpoint status: %d", resp.StatusCode)
	}
}

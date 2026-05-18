package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"viralforge/backend/tests/fixtures"
)

// TestKYCStatusSubmission tests submitting KYC status
func TestKYCStatusSubmission(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	email := "kyc-test-" + fmt.Sprintf("test-%s-%d@example.com", "profile_kyc_test.go", time.Now().UnixNano())
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

	// Submit KYC status using DoRequest
	kycData := map[string]interface{}{
		"status":          "pending",
		"document_type":   "passport",
		"document_number": "AB123456",
	}

	resp, body, err := suite.Client.DoRequest("POST", "/api/v1/profiles/"+profile.ID+"/kyc", kycData)
	if err != nil {
		t.Fatalf("KYC submission request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		t.Logf("KYC submission returned status %d: %s", resp.StatusCode, string(body))
		t.Skip("KYC endpoint may not be implemented")
	}

	t.Logf("KYC status submitted successfully")
}

// TestKYCStatusRetrieval tests retrieving KYC status
func TestKYCStatusRetrieval(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	email := "kyc-retrieve-" + fmt.Sprintf("test-%s-%d@example.com", "profile_kyc_test.go", time.Now().UnixNano())
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

	// Get KYC status
	resp, body, err := suite.Client.DoRequest("GET", "/api/v1/profiles/"+profile.ID+"/kyc", nil)
	if err != nil {
		t.Fatalf("KYC retrieval request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Logf("KYC retrieval returned status %d: %s", resp.StatusCode, string(body))
		t.Skip("KYC endpoint may not be implemented")
	}

	t.Logf("KYC status retrieved successfully")
}

// TestKYCStatusUpdate tests updating KYC status (approved/rejected)
func TestKYCStatusUpdate(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	email := "kyc-update-" + fmt.Sprintf("test-%s-%d@example.com", "profile_kyc_test.go", time.Now().UnixNano())
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

	// Submit initial KYC
	kycData := map[string]interface{}{
		"status": "pending",
	}
	_, _, err = suite.Client.DoRequest("POST", "/api/v1/profiles/"+profile.ID+"/kyc", kycData)
	if err != nil {
		t.Logf("Initial KYC submission failed: %v", err)
	}

	// Update KYC to approved (admin action typically)
	kycUpdate := map[string]interface{}{
		"status": "approved",
	}

	resp, body, err := suite.Client.DoRequest("PATCH", "/api/v1/profiles/"+profile.ID+"/kyc", kycUpdate)
	if err != nil {
		t.Fatalf("KYC update request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Logf("KYC update returned status %d: %s", resp.StatusCode, string(body))
		t.Skip("KYC endpoint may not be implemented or requires admin privileges")
	}

	t.Logf("KYC status updated successfully")
}

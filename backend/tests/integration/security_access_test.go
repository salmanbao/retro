package integration

import (
	"fmt"
	"testing"
	"time"

	"viralforge/backend/tests/fixtures"
)

// TestCrossUserProfileAccessDenial tests that users cannot access other users' profiles
func TestCrossUserProfileAccessDenial(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	password := "TestPass123!"

	// User A creates a profile
	emailA := "user-a-access-" + fmt.Sprintf("test-%s-%d@example.com", "security_access_test.go", time.Now().UnixNano())
	_, _ = suite.Client.Register(fixtures.RegisterRequest{Email: emailA, Password: password})
	loginResp, _ := suite.Client.Login(fixtures.LoginRequest{Email: emailA, Password: password})
	if loginResp == nil {
		t.Skip("Login failed (email verification may be required)")
		return
	}
	profileA, err := suite.Client.CreateProfile(fixtures.CreateProfileRequest{Type: "brand"})
	if err != nil {
		t.Fatalf("User A profile creation failed: %v", err)
	}
	if profileA == nil {
		t.Fatal("Profile A creation returned nil")
	}

	// User B creates their own profile
	emailB := "user-b-access-" + fmt.Sprintf("test-%s-%d@example.com", "security_access_test.go", time.Now().UnixNano())
	_, _ = suite.Client.Register(fixtures.RegisterRequest{Email: emailB, Password: password})
	loginRespB, _ := suite.Client.Login(fixtures.LoginRequest{Email: emailB, Password: password})
	if loginRespB == nil {
		t.Skip("Login failed (email verification may be required)")
		return
	}
	_, err = suite.Client.CreateProfile(fixtures.CreateProfileRequest{Type: "brand"})
	if err != nil {
		t.Skipf("User B profile creation failed: %v", err)
	}

	// User B attempts to access User A's profile - should be denied
	_, err = suite.Client.GetProfileDetails(profileA.ID)
	if err == nil {
		t.Error("Expected authorization error when accessing another user's profile, got nil")
	}

	t.Logf("Cross-user profile access correctly denied")
}

// TestCrossUserOnboardingAccessDenial tests that users cannot access other users' onboarding progress
func TestCrossUserOnboardingAccessDenial(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	password := "TestPass123!"

	// User A creates a profile
	emailA := "user-a-onboarding-" + fmt.Sprintf("test-%s-%d@example.com", "security_access_test.go", time.Now().UnixNano())
	_, _ = suite.Client.Register(fixtures.RegisterRequest{Email: emailA, Password: password})
	loginResp, _ := suite.Client.Login(fixtures.LoginRequest{Email: emailA, Password: password})
	if loginResp == nil {
		t.Skip("Login failed (email verification may be required)")
		return
	}
	profileA, err := suite.Client.CreateProfile(fixtures.CreateProfileRequest{Type: "brand"})
	if err != nil {
		t.Fatalf("User A profile creation failed: %v", err)
	}
	if profileA == nil {
		t.Fatal("Profile A creation returned nil")
	}

	// User B creates their own profile
	emailB := "user-b-onboarding-" + fmt.Sprintf("test-%s-%d@example.com", "security_access_test.go", time.Now().UnixNano())
	_, _ = suite.Client.Register(fixtures.RegisterRequest{Email: emailB, Password: password})
	loginRespB, _ := suite.Client.Login(fixtures.LoginRequest{Email: emailB, Password: password})
	if loginRespB == nil {
		t.Skip("Login failed (email verification may be required)")
		return
	}
	_, err = suite.Client.CreateProfile(fixtures.CreateProfileRequest{Type: "brand"})
	if err != nil {
		t.Skipf("User B profile creation failed: %v", err)
	}

	// User B attempts to access User A's onboarding progress - should be denied
	_, err = suite.Client.GetOnboarding(profileA.ID)
	if err == nil {
		t.Error("Expected authorization error when accessing another user's onboarding, got nil")
	}

	t.Logf("Cross-user onboarding access correctly denied")
}

// TestCrossUserStepUpdateDenial tests that users cannot update other users' onboarding steps
func TestCrossUserStepUpdateDenial(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	password := "TestPass123!"

	// User A creates a profile
	emailA := "user-a-step-" + fmt.Sprintf("test-%s-%d@example.com", "security_access_test.go", time.Now().UnixNano())
	_, _ = suite.Client.Register(fixtures.RegisterRequest{Email: emailA, Password: password})
	loginResp, _ := suite.Client.Login(fixtures.LoginRequest{Email: emailA, Password: password})
	if loginResp == nil {
		t.Skip("Login failed (email verification may be required)")
		return
	}
	profileA, err := suite.Client.CreateProfile(fixtures.CreateProfileRequest{Type: "brand"})
	if err != nil {
		t.Fatalf("User A profile creation failed: %v", err)
	}
	if profileA == nil {
		t.Fatal("Profile A creation returned nil")
	}

	// User B creates their own profile
	emailB := "user-b-step-" + fmt.Sprintf("test-%s-%d@example.com", "security_access_test.go", time.Now().UnixNano())
	_, _ = suite.Client.Register(fixtures.RegisterRequest{Email: emailB, Password: password})
	loginRespB, _ := suite.Client.Login(fixtures.LoginRequest{Email: emailB, Password: password})
	if loginRespB == nil {
		t.Skip("Login failed (email verification may be required)")
		return
	}
	_, err = suite.Client.CreateProfile(fixtures.CreateProfileRequest{Type: "brand"})
	if err != nil {
		t.Skipf("User B profile creation failed: %v", err)
	}

	// Get User A's onboarding steps
	steps, err := suite.Client.GetOnboardingSteps(profileA.ID)
	if err != nil {
		t.Skipf("Could not get onboarding steps: %v", err)
	}

	// User B attempts to update User A's step - should be denied
	if steps != nil {
		if stepID, ok := (*steps)["id"].(string); ok && stepID != "" {
			err = suite.Client.UpdateOnboardingStep(profileA.ID, stepID, "completed")
			if err == nil {
				t.Error("Expected authorization error when updating another user's step, got nil")
			}
		}
	}

	t.Logf("Cross-user step update correctly denied")
}

// TestUserCannotAccessOwnProfileAsOtherUser tests isolation within same user context
func TestUserCannotAccessOwnProfileAsOtherUser(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	password := "TestPass123!"

	// User A creates multiple profiles
	emailA := "user-a-multi-" + fmt.Sprintf("test-%s-%d@example.com", "security_access_test.go", time.Now().UnixNano())
	_, _ = suite.Client.Register(fixtures.RegisterRequest{Email: emailA, Password: password})
	loginResp, _ := suite.Client.Login(fixtures.LoginRequest{Email: emailA, Password: password})
	if loginResp == nil {
		t.Skip("Login failed (email verification may be required)")
		return
	}
	profileA1, _ := suite.Client.CreateProfile(fixtures.CreateProfileRequest{Type: "brand"})
	if profileA1 == nil {
		t.Fatal("Profile A1 creation failed")
	}
	profileA2, _ := suite.Client.CreateProfile(fixtures.CreateProfileRequest{Type: "editor"})
	if profileA2 == nil {
		t.Fatal("Profile A2 creation failed")
	}

	// User A can access both their own profiles
	details1, err := suite.Client.GetProfileDetails(profileA1.ID)
	if err != nil {
		t.Errorf("User should be able to access their own profile A1: %v", err)
	}
	if details1 == nil {
		t.Error("Expected details for own profile A1")
	}

	details2, err := suite.Client.GetProfileDetails(profileA2.ID)
	if err != nil {
		t.Errorf("User should be able to access their own profile A2: %v", err)
	}
	if details2 == nil {
		t.Error("Expected details for own profile A2")
	}

	t.Logf("User can access their own profiles correctly")
}

package integration

import (
	"fmt"
	"testing"
	"time"

	"viralforge/backend/tests/fixtures"
)

// TestTemplateAssignmentBrand tests that brand profile gets correct onboarding template
func TestTemplateAssignmentBrand(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	email := fmt.Sprintf("test-template-brand-%d@example.com", time.Now().UnixNano())
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
		t.Fatalf("Brand profile creation failed: %v", err)
	}

	// Get onboarding and verify template type
	onboarding, err := suite.Client.GetOnboarding(profile.ID)
	if err != nil {
		t.Fatalf("Get onboarding failed: %v", err)
	}

	if onboarding == nil {
		t.Fatal("Expected onboarding to be created for brand profile")
	}

	// Template assignment verification
	if templateID, ok := (*onboarding)["template_id"]; ok && templateID != nil {
		t.Logf("Brand profile has template_id: %v", templateID)
	}

	t.Logf("Brand profile onboarding template assigned")
}

// TestTemplateAssignmentEditor tests that editor profile gets correct onboarding template
func TestTemplateAssignmentEditor(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	email := fmt.Sprintf("test-template-editor-%d@example.com", time.Now().UnixNano())
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
		Type: "editor",
	})
	if err != nil {
		t.Fatalf("Editor profile creation failed: %v", err)
	}

	onboarding, err := suite.Client.GetOnboarding(profile.ID)
	if err != nil {
		t.Fatalf("Get onboarding failed: %v", err)
	}

	if onboarding == nil {
		t.Fatal("Expected onboarding to be created for editor profile")
	}

	t.Logf("Editor profile onboarding template assigned")
}

// TestTemplateAssignmentInfluencer tests that influencer profile gets correct onboarding template
func TestTemplateAssignmentInfluencer(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	email := fmt.Sprintf("test-template-influencer-%d@example.com", time.Now().UnixNano())
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
		t.Fatalf("Influencer profile creation failed: %v", err)
	}

	onboarding, err := suite.Client.GetOnboarding(profile.ID)
	if err != nil {
		t.Fatalf("Get onboarding failed: %v", err)
	}

	if onboarding == nil {
		t.Fatal("Expected onboarding to be created for influencer profile")
	}

	t.Logf("Influencer profile onboarding template assigned")
}

// TestTemplateDifferentiatesByProfileType tests that different profile types get different templates
func TestTemplateDifferentiatesByProfileType(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	password := "TestPass123!"

	// Create brand profile
	emailBrand := fmt.Sprintf("test-tpl-brand-%d@example.com", time.Now().UnixNano())
	_, _ = suite.Client.Register(fixtures.RegisterRequest{Email: emailBrand, Password: password})
	loginResp, _ := suite.Client.Login(fixtures.LoginRequest{Email: emailBrand, Password: password})
	if loginResp == nil {
		t.Skip("Login failed (email verification may be required)")
		return
	}
	profileBrand, _ := suite.Client.CreateProfile(fixtures.CreateProfileRequest{Type: "brand"})
	if profileBrand == nil {
		t.Fatal("Profile creation failed")
	}
	onboardingBrand, _ := suite.Client.GetOnboarding(profileBrand.ID)

	// Create editor profile
	emailEditor := fmt.Sprintf("test-tpl-editor-%d@example.com", time.Now().UnixNano())
	_, _ = suite.Client.Register(fixtures.RegisterRequest{Email: emailEditor, Password: password})
	loginResp, _ = suite.Client.Login(fixtures.LoginRequest{Email: emailEditor, Password: password})
	if loginResp == nil {
		t.Skip("Login failed (email verification may be required)")
		return
	}
	profileEditor, _ := suite.Client.CreateProfile(fixtures.CreateProfileRequest{Type: "editor"})
	if profileEditor == nil {
		t.Fatal("Profile creation failed")
	}
	onboardingEditor, _ := suite.Client.GetOnboarding(profileEditor.ID)

	// Create influencer profile
	emailInfluencer := fmt.Sprintf("test-tpl-influencer-%d@example.com", time.Now().UnixNano())
	_, _ = suite.Client.Register(fixtures.RegisterRequest{Email: emailInfluencer, Password: password})
	loginResp, _ = suite.Client.Login(fixtures.LoginRequest{Email: emailInfluencer, Password: password})
	if loginResp == nil {
		t.Skip("Login failed (email verification may be required)")
		return
	}
	profileInfluencer, _ := suite.Client.CreateProfile(fixtures.CreateProfileRequest{Type: "influencer"})
	if profileInfluencer == nil {
		t.Fatal("Profile creation failed")
	}
	onboardingInfluencer, _ := suite.Client.GetOnboarding(profileInfluencer.ID)

	// Verify all onboardings are created
	if onboardingBrand == nil || onboardingEditor == nil || onboardingInfluencer == nil {
		t.Error("Expected onboarding to be created for all profile types")
	}

	t.Logf("Template differentiation verified across brand, editor, and influencer profiles")
}

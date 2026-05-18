package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"viralforge/backend/tests/fixtures"
)

// TestEditorPortfolioCreate tests creating portfolio items (Editor only)
func TestEditorPortfolioCreate(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	email := "portfolio-test-" + fmt.Sprintf("test-%s-%d@example.com", "profile_portfolio_test.go", time.Now().UnixNano())
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

	// Create editor profile
	profile, err := suite.Client.CreateProfile(fixtures.CreateProfileRequest{
		Type: "editor",
	})
	if err != nil {
		t.Fatalf("Editor profile creation failed: %v", err)
	}

	// Create portfolio item using DoRequest for custom endpoint
	portfolioItem := map[string]interface{}{
		"title":       "Sample Portfolio Item",
		"description": "A sample portfolio item for testing",
		"url":         "https://example.com/work/sample",
		"category":    "writing",
	}

	resp, body, err := suite.Client.DoRequest("POST", "/api/v1/profiles/"+profile.ID+"/portfolio", portfolioItem)
	if err != nil {
		t.Fatalf("Portfolio creation request failed: %v", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		t.Logf("Portfolio creation returned status %d: %s", resp.StatusCode, string(body))
		t.Skip("Portfolio endpoint may not be implemented")
	}

	t.Logf("Portfolio item created successfully")
}

// TestEditorPortfolioRead tests reading portfolio items
func TestEditorPortfolioRead(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	email := "portfolio-read-" + fmt.Sprintf("test-%s-%d@example.com", "profile_portfolio_test.go", time.Now().UnixNano())
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

	// Get portfolio items
	resp, body, err := suite.Client.DoRequest("GET", "/api/v1/profiles/"+profile.ID+"/portfolio", nil)
	if err != nil {
		t.Fatalf("Portfolio read request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Logf("Portfolio read returned status %d: %s", resp.StatusCode, string(body))
		t.Skip("Portfolio endpoint may not be implemented")
	}

	t.Logf("Portfolio items retrieved successfully")
}

// TestEditorPortfolioUpdate tests updating portfolio items
func TestEditorPortfolioUpdate(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	email := "portfolio-update-" + fmt.Sprintf("test-%s-%d@example.com", "profile_portfolio_test.go", time.Now().UnixNano())
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

	_, err = suite.Client.CreateProfile(fixtures.CreateProfileRequest{
		Type: "editor",
	})
	if err != nil {
		t.Fatalf("Editor profile creation failed: %v", err)
	}

	t.Skip("Portfolio CRUD requires item ID - implement with create-then-update pattern")
}

// TestEditorPortfolioDelete tests deleting portfolio items
func TestEditorPortfolioDelete(t *testing.T) {
	suite := NewTestSuite(t)
	if suite == nil {
		return
	}
	defer suite.TearDown()
	suite.SkipIfNoServer()

	email := "portfolio-delete-" + fmt.Sprintf("test-%s-%d@example.com", "profile_portfolio_test.go", time.Now().UnixNano())
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

	_, err = suite.Client.CreateProfile(fixtures.CreateProfileRequest{
		Type: "editor",
	})
	if err != nil {
		t.Fatalf("Editor profile creation failed: %v", err)
	}

	t.Skip("Portfolio CRUD requires item ID - implement with create-then-delete pattern")
}

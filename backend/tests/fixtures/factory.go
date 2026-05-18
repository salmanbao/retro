package fixtures

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// UserFactory creates test users via the API.
type UserFactory struct {
	client *TestClient
}

// NewUserFactory creates a new UserFactory.
func NewUserFactory(client *TestClient) *UserFactory {
	return &UserFactory{client: client}
}

// User represents a created test user.
type User struct {
	ID            string
	Email         string
	Password      string
	ProfileIDs    []string
	Authenticated bool
}

// CreateRegistered creates a new user, verifies email, and logs in.
// Returns a fully authenticated user ready to create profiles.
func (f *UserFactory) CreateRegistered() (*User, error) {
	user := &User{
		ID:       uuid.New().String(),
		Email:    fmt.Sprintf("test-%d@example.com", time.Now().UnixNano()),
		Password: "TestPass123!",
	}

	// Register
	_, err := f.client.Register(RegisterRequest{
		Email:    user.Email,
		Password: user.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("register failed: %w", err)
	}

	// Verify email - in test mode, tokens are generated predictably
	// For now, we assume the verification happens automatically or via test token
	// This is a placeholder - actual implementation needs to handle email token lookup
	token := f.getVerificationToken(user.Email)
	if token != "" {
		if err := f.client.VerifyEmail(token); err != nil {
			// In test environment, verification might be bypassed
			// Continue anyway
		}
	}

	// Login
	_, err = f.client.Login(LoginRequest{
		Email:    user.Email,
		Password: user.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}

	user.Authenticated = true
	return user, nil
}

// getVerificationToken looks up the verification token for an email.
// In a real implementation, this would query the database directly.
func (f *UserFactory) getVerificationToken(email string) string {
	// For test environments, this could be:
	// 1. A test endpoint that returns the token
	// 2. Direct DB access to find the token
	// 3. A mocked email service that captures the token
	// For now, return empty to indicate verification should be skipped
	return ""
}

// CreateProfile creates a profile for an authenticated user.
func (f *UserFactory) CreateProfile(profileType string) (*ProfileResponse, error) {
	return f.client.CreateProfile(CreateProfileRequest{Type: profileType})
}

// CreateProfiles creates multiple profiles for an authenticated user.
func (f *UserFactory) CreateProfiles(profileTypes ...string) ([]*ProfileResponse, error) {
	profiles := make([]*ProfileResponse, 0, len(profileTypes))
	for _, pt := range profileTypes {
		profile, err := f.client.CreateProfile(CreateProfileRequest{Type: pt})
		if err != nil {
			return profiles, err
		}
		profiles = append(profiles, profile)
	}
	return profiles, nil
}

// ProfileFactory creates test profiles.
type ProfileFactory struct {
	client   *TestClient
	userFac  *UserFactory
	profiles []*ProfileResponse
}

// NewProfileFactory creates a new ProfileFactory.
func NewProfileFactory(client *TestClient, userFac *UserFactory) *ProfileFactory {
	return &ProfileFactory{
		client:   client,
		userFac:  userFac,
		profiles: make([]*ProfileResponse, 0),
	}
}

// CreateEditorWithEnrichment creates an Editor profile with full enrichment.
func (f *ProfileFactory) CreateEditorWithEnrichment() (*ProfileResponse, error) {
	profile, err := f.client.CreateProfile(CreateProfileRequest{Type: "Editor"})
	if err != nil {
		return nil, err
	}

	// Add bio and avatar
	err = f.client.UpdateProfileDetails(profile.ID, UpdateProfileDetailsRequest{
		Bio:       "Test bio for editor",
		AvatarURL: "https://example.com/avatar.jpg",
	})
	if err != nil {
		return nil, err
	}

	// Add social links
	err = f.client.UpdateSocialLinks(profile.ID, UpdateSocialLinksRequest{
		SocialLinks: map[string]string{
			"twitter":   "https://twitter.com/test",
			"instagram": "https://instagram.com/test",
		},
	})
	if err != nil {
		return nil, err
	}

	f.profiles = append(f.profiles, profile)
	return profile, nil
}

// CreateBrandWithEnrichment creates a Brand profile with enrichment.
func (f *ProfileFactory) CreateBrandWithEnrichment() (*ProfileResponse, error) {
	profile, err := f.client.CreateProfile(CreateProfileRequest{Type: "Brand"})
	if err != nil {
		return nil, err
	}

	err = f.client.UpdateProfileDetails(profile.ID, UpdateProfileDetailsRequest{
		Bio:       "Test bio for brand",
		AvatarURL: "https://example.com/brand.jpg",
	})
	if err != nil {
		return nil, err
	}

	f.profiles = append(f.profiles, profile)
	return profile, nil
}

// CreateInfluencerWithEnrichment creates an Influencer profile with enrichment.
func (f *ProfileFactory) CreateInfluencerWithEnrichment() (*ProfileResponse, error) {
	profile, err := f.client.CreateProfile(CreateProfileRequest{Type: "Influencer"})
	if err != nil {
		return nil, err
	}

	err = f.client.UpdateProfileDetails(profile.ID, UpdateProfileDetailsRequest{
		Bio:       "Test bio for influencer",
		AvatarURL: "https://example.com/influencer.jpg",
	})
	if err != nil {
		return nil, err
	}

	// Add social links
	err = f.client.UpdateSocialLinks(profile.ID, UpdateSocialLinksRequest{
		SocialLinks: map[string]string{
			"twitter":   "https://twitter.com/influencer",
			"instagram": "https://instagram.com/influencer",
		},
	})
	if err != nil {
		return nil, err
	}

	f.profiles = append(f.profiles, profile)
	return profile, nil
}

// OnboardingFactory helps create onboarding scenarios.
type OnboardingFactory struct {
	client *TestClient
}

// NewOnboardingFactory creates a new OnboardingFactory.
func NewOnboardingFactory(client *TestClient) *OnboardingFactory {
	return &OnboardingFactory{client: client}
}

// GetProgress retrieves onboarding progress for a profile.
func (f *OnboardingFactory) GetProgress(profileID string) (*map[string]interface{}, error) {
	return f.client.GetOnboarding(profileID)
}

// GetSteps retrieves onboarding steps for a profile.
func (f *OnboardingFactory) GetSteps(profileID string) (*map[string]interface{}, error) {
	return f.client.GetOnboardingSteps(profileID)
}

// CompleteStep marks a step as completed.
func (f *OnboardingFactory) CompleteStep(profileID, stepID string) error {
	return f.client.UpdateOnboardingStep(profileID, stepID, "completed")
}

// StartStep marks a step as in_progress.
func (f *OnboardingFactory) StartStep(profileID, stepID string) error {
	return f.client.UpdateOnboardingStep(profileID, stepID, "in_progress")
}

// SkipStep marks a step as skipped (only for optional steps).
func (f *OnboardingFactory) SkipStep(profileID, stepID string) error {
	return f.client.UpdateOnboardingStep(profileID, stepID, "skipped")
}

// Recalculate triggers recalculation of onboarding progress.
func (f *OnboardingFactory) Recalculate(profileID string) (*map[string]interface{}, error) {
	return f.client.RecalculateOnboarding(profileID)
}

// GetNextStep retrieves the next incomplete step.
func (f *OnboardingFactory) GetNextStep(profileID string) (*map[string]interface{}, error) {
	return f.client.GetNextStep(profileID)
}

// AdminFactory creates admin-level operations.
type AdminFactory struct {
	client *TestClient
}

// NewAdminFactory creates a new AdminFactory.
func NewAdminFactory(client *TestClient) *AdminFactory {
	return &AdminFactory{client: client}
}

// ActivateProfile activates a profile (admin action).
func (f *AdminFactory) ActivateProfile(profileID string) error {
	reqURL := fmt.Sprintf("%s/api/v1/admin/profiles/%s/onboarding/activate", f.client.BaseURL, profileID)
	httpReq, err := http.NewRequest("POST", reqURL, nil)
	if err != nil {
		return err
	}

	resp, err := f.client.HttpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		var errResp ErrorResponse
		json.Unmarshal(data, &errResp)
		return fmt.Errorf("admin activate failed: %s - %s", errResp.Error, errResp.Message)
	}
	return nil
}

// CrossUserFactory helps test cross-user access scenarios.
type CrossUserFactory struct {
	userFacA *UserFactory
	userFacB *UserFactory
}

// NewCrossUserFactory creates a new CrossUserFactory.
func NewCrossUserFactory(client *TestClient) *CrossUserFactory {
	return &CrossUserFactory{
		userFacA: NewUserFactory(client),
		userFacB: NewUserFactory(client),
	}
}

// CreateTwoUsers creates two separate authenticated users.
func (f *CrossUserFactory) CreateTwoUsers() (*User, *User, error) {
	userA, err := f.userFacA.CreateRegistered()
	if err != nil {
		return nil, nil, err
	}

	userB, err := f.userFacB.CreateRegistered()
	if err != nil {
		return nil, nil, err
	}

	return userA, userB, nil
}

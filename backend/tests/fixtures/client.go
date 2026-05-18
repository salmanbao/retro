package fixtures

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

// TestClient is an HTTP client for integration testing.
type TestClient struct {
	BaseURL    string
	HttpClient *http.Client
	jar        *cookiejar.Jar
	Token      string // Auth token from last login, used for Authorization header
}

// NewTestClient creates a new test client.
func NewTestClient(baseURL string) (*TestClient, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	return &TestClient{
		BaseURL: baseURL,
		HttpClient: &http.Client{
			Jar: jar,
		},
		jar: jar,
	}, nil
}

// RegisterRequest represents a registration request.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterResponse represents a registration response.
type RegisterResponse struct {
	Message   string `json:"message"`
	Email     string `json:"email"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

// VerifyRequest represents an email verification request.
type VerifyRequest struct {
	Token string `json:"token"`
}

// LoginRequest represents a login request.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents a login response.
type LoginResponse struct {
	Message   string `json:"message"`
	UserID    string `json:"user_id,omitempty"`
	Token     string `json:"token,omitempty"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

// ProfileResponse represents a profile response.
type ProfileResponse struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Type      string `json:"type"`
	CreatedAt string `json:"created_at"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// Register sends a registration request.
func (c *TestClient) Register(req RegisterRequest) (*RegisterResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.HttpClient.Post(c.BaseURL+"/api/v1/auth/register", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		var errResp ErrorResponse
		json.Unmarshal(data, &errResp)
		return nil, fmt.Errorf("registration failed: %s - %s", errResp.Error, errResp.Message)
	}

	var result RegisterResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// VerifyEmail sends an email verification request.
func (c *TestClient) VerifyEmail(token string) error {
	req := VerifyRequest{Token: token}
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := c.HttpClient.Post(c.BaseURL+"/api/v1/auth/verify-email", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		var errResp ErrorResponse
		json.Unmarshal(data, &errResp)
		return fmt.Errorf("verification failed: %s - %s", errResp.Error, errResp.Message)
	}
	return nil
}

// Login sends a login request.
func (c *TestClient) Login(req LoginRequest) (*LoginResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.HttpClient.Post(c.BaseURL+"/api/v1/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		json.Unmarshal(data, &errResp)
		return nil, fmt.Errorf("login failed: %s - %s", errResp.Error, errResp.Message)
	}

	var result LoginResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	// Store token for subsequent authenticated requests
	c.Token = result.Token

	return &result, nil
}

// GetMe returns the current authenticated user.
func (c *TestClient) GetMe() (*map[string]interface{}, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/api/v1/auth/me", nil)
	if err != nil {
		return nil, err
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get me failed with status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateProfileRequest represents a profile creation request.
type CreateProfileRequest struct {
	Type string `json:"type"`
}

// CreateProfile creates a new profile.
func (c *TestClient) CreateProfile(req CreateProfileRequest) (*ProfileResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.HttpClient.Post(c.BaseURL+"/api/v1/profiles", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		var errResp ErrorResponse
		json.Unmarshal(data, &errResp)
		return nil, fmt.Errorf("create profile failed: %s - %s", errResp.Error, errResp.Message)
	}

	var result ProfileResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetOnboarding retrieves onboarding progress for a profile.
func (c *TestClient) GetOnboarding(profileID string) (*map[string]interface{}, error) {
	resp, err := c.HttpClient.Get(c.BaseURL + "/api/v1/profiles/" + profileID + "/onboarding")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		var errResp ErrorResponse
		json.Unmarshal(data, &errResp)
		return nil, fmt.Errorf("get onboarding failed: %s - %s", errResp.Error, errResp.Message)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetOnboardingSteps retrieves onboarding steps for a profile.
func (c *TestClient) GetOnboardingSteps(profileID string) (*map[string]interface{}, error) {
	resp, err := c.HttpClient.Get(c.BaseURL + "/api/v1/profiles/" + profileID + "/onboarding/steps")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		var errResp ErrorResponse
		json.Unmarshal(data, &errResp)
		return nil, fmt.Errorf("get onboarding steps failed: %s - %s", errResp.Error, errResp.Message)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateOnboardingStepRequest represents a step update request.
type UpdateOnboardingStepRequest struct {
	Status string `json:"status"`
}

// UpdateOnboardingStep updates an onboarding step status.
func (c *TestClient) UpdateOnboardingStep(profileID, stepID string, status string) error {
	req := UpdateOnboardingStepRequest{Status: status}
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	reqURL := fmt.Sprintf("%s/api/v1/profiles/%s/onboarding/steps/%s", c.BaseURL, profileID, stepID)
	httpReq, err := http.NewRequest("PATCH", reqURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HttpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		var errResp ErrorResponse
		json.Unmarshal(data, &errResp)
		return fmt.Errorf("update step failed: %s - %s", errResp.Error, errResp.Message)
	}
	return nil
}

// UpdateProfileDetailsRequest represents a details update request.
type UpdateProfileDetailsRequest struct {
	Bio       string `json:"bio,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

// UpdateProfileDetails updates profile enrichment details.
func (c *TestClient) UpdateProfileDetails(profileID string, req UpdateProfileDetailsRequest) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	reqURL := fmt.Sprintf("%s/api/v1/profiles/%s/details", c.BaseURL, profileID)
	httpReq, err := http.NewRequest("PATCH", reqURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HttpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		var errResp ErrorResponse
		json.Unmarshal(data, &errResp)
		return fmt.Errorf("update details failed: %s - %s", errResp.Error, errResp.Message)
	}
	return nil
}

// UpdateSocialLinksRequest represents a social links update request.
type UpdateSocialLinksRequest struct {
	SocialLinks map[string]string `json:"social_links,omitempty"`
}

// UpdateSocialLinks updates profile social links.
func (c *TestClient) UpdateSocialLinks(profileID string, req UpdateSocialLinksRequest) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	reqURL := fmt.Sprintf("%s/api/v1/profiles/%s/details", c.BaseURL, profileID)
	httpReq, err := http.NewRequest("PATCH", reqURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HttpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		var errResp ErrorResponse
		json.Unmarshal(data, &errResp)
		return fmt.Errorf("update social links failed: %s - %s", errResp.Error, errResp.Message)
	}
	return nil
}

// GetProfileDetails retrieves profile details.
func (c *TestClient) GetProfileDetails(profileID string) (*map[string]interface{}, error) {
	resp, err := c.HttpClient.Get(c.BaseURL + "/api/v1/profiles/" + profileID + "/details")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		var errResp ErrorResponse
		json.Unmarshal(data, &errResp)
		return nil, fmt.Errorf("get profile details failed: %s - %s", errResp.Error, errResp.Message)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// RecalculateOnboarding triggers onboarding recalculation.
func (c *TestClient) RecalculateOnboarding(profileID string) (*map[string]interface{}, error) {
	reqURL := fmt.Sprintf("%s/api/v1/profiles/%s/onboarding/recalculate", c.BaseURL, profileID)
	httpReq, err := http.NewRequest("POST", reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HttpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		var errResp ErrorResponse
		json.Unmarshal(data, &errResp)
		return nil, fmt.Errorf("recalculate onboarding failed: %s - %s", errResp.Error, errResp.Message)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetNextStep retrieves the next incomplete onboarding step.
func (c *TestClient) GetNextStep(profileID string) (*map[string]interface{}, error) {
	resp, err := c.HttpClient.Get(c.BaseURL + "/api/v1/profiles/" + profileID + "/onboarding/next-step")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		var errResp ErrorResponse
		json.Unmarshal(data, &errResp)
		return nil, fmt.Errorf("get next step failed: %s - %s", errResp.Error, errResp.Message)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DoRequest performs a custom HTTP request and returns the response.
func (c *TestClient) DoRequest(method, path string, body interface{}) (*http.Response, []byte, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, nil, err
		}
		reqBody = bytes.NewReader(data)
	}

	reqURL := c.BaseURL + path
	httpReq, err := http.NewRequest(method, reqURL, reqBody)
	if err != nil {
		return nil, nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HttpClient.Do(httpReq)
	if err != nil {
		return nil, nil, err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		return resp, nil, err
	}
	resp.Body.Close()
	return resp, respBody, nil
}

// SetBaseURL changes the base URL (useful for switching between test server and real server).
func (c *TestClient) SetBaseURL(baseURL string) {
	c.BaseURL = baseURL
}

// GetCookies returns the cookies for the base URL.
func (c *TestClient) GetCookies() []*http.Cookie {
	return c.jar.Cookies(&url.URL{Scheme: "http", Host: c.BaseURL})
}

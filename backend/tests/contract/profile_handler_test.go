package contract

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/handler"
	"viralforge/backend/src/middleware"
	"viralforge/backend/src/service"
)

// mockProfileRepo implements repository.ProfileRepository for contract testing.
type mockProfileRepo struct {
	profiles map[uuid.UUID]*domain.Profile
	byUser   map[uuid.UUID][]*domain.Profile
}

func newMockProfileRepo() *mockProfileRepo {
	return &mockProfileRepo{
		profiles: make(map[uuid.UUID]*domain.Profile),
		byUser:   make(map[uuid.UUID][]*domain.Profile),
	}
}

func (r *mockProfileRepo) Create(ctx context.Context, profile *domain.Profile) error {
	r.profiles[profile.ID] = profile
	r.byUser[profile.UserID] = append(r.byUser[profile.UserID], profile)
	return nil
}

func (r *mockProfileRepo) ByID(ctx context.Context, id uuid.UUID) (*domain.Profile, error) {
	if p, ok := r.profiles[id]; ok {
		if p.DeletedAt != nil {
			return nil, domain.ErrProfileNotFound
		}
		return p, nil
	}
	return nil, domain.ErrProfileNotFound
}

func (r *mockProfileRepo) ByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Profile, error) {
	if profiles, ok := r.byUser[userID]; ok {
		var active []*domain.Profile
		for _, p := range profiles {
			if p.DeletedAt == nil {
				active = append(active, p)
			}
		}
		return active, nil
	}
	return []*domain.Profile{}, nil
}

func (r *mockProfileRepo) Update(ctx context.Context, profile *domain.Profile) error {
	if _, ok := r.profiles[profile.ID]; !ok {
		return domain.ErrProfileNotFound
	}
	r.profiles[profile.ID] = profile
	return nil
}

func (r *mockProfileRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if p, ok := r.profiles[id]; ok {
		now := &p.DeletedAt
		_ = now // Soft delete would set this
	}
	return nil
}

// mockProfileUserRepo implements repository.UserRepository for contract testing.
type mockProfileUserRepo struct {
	users map[uuid.UUID]*domain.User
}

func newMockProfileUserRepo() *mockProfileUserRepo {
	return &mockProfileUserRepo{
		users: make(map[uuid.UUID]*domain.User),
	}
}

func (r *mockProfileUserRepo) Create(ctx context.Context, user *domain.User) error {
	r.users[user.ID] = user
	return nil
}

func (r *mockProfileUserRepo) ByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if u, ok := r.users[id]; ok {
		return u, nil
	}
	return nil, domain.ErrUserNotFound
}

func (r *mockProfileUserRepo) ByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, domain.ErrUserNotFound
}

func (r *mockProfileUserRepo) Update(ctx context.Context, user *domain.User) error {
	r.users[user.ID] = user
	return nil
}

// setupTestProfileHandler creates a handler configured for testing.
func setupTestProfileHandler() (*handler.ProfileHandler, *mockProfileRepo, *mockProfileUserRepo) {
	profileRepo := newMockProfileRepo()
	userRepo := newMockProfileUserRepo()
	profileSvc := service.NewProfileService(profileRepo, userRepo)
	profileHandler := handler.NewProfileHandler(profileSvc)
	return profileHandler, profileRepo, userRepo
}

// TestContractCreateProfileEndpoint tests POST /api/v1/profiles contract.
func TestContractCreateProfileEndpoint(t *testing.T) {
	t.Run("201 Created - valid brand profile", func(t *testing.T) {
		h, _, userRepo := setupTestProfileHandler()

		userID := uuid.New()
		userRepo.Create(context.Background(), &domain.User{ID: userID, Email: "brand@example.com"})

		// Create context with authenticated user
		reqBody := handler.CreateProfileRequest{
			ProfileType: "brand",
			Name:        "Acme Brand",
			Details: map[string]interface{}{
				"company_name": "Acme Corporation",
				"size":         "100-500",
				"industry":     "Technology",
			},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/profiles", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Add user to context (simulate auth middleware)
		ctx := context.WithValue(req.Context(), middleware.UserContextKey, userRepo.users[userID])
		req = req.WithContext(ctx)

		h.Create(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp handler.CreateProfileResponse
		json.NewDecoder(w.Body).Decode(&resp)
		assert.Equal(t, "brand", resp.Type)
		assert.Equal(t, "Acme Brand", resp.Name)
		assert.NotEmpty(t, resp.ID)
	})

	t.Run("201 Created - valid editor profile", func(t *testing.T) {
		h, _, userRepo := setupTestProfileHandler()

		userID := uuid.New()
		userRepo.Create(context.Background(), &domain.User{ID: userID, Email: "editor@example.com"})

		reqBody := handler.CreateProfileRequest{
			ProfileType: "editor",
			Name:        "Editor Profile",
			Details: map[string]interface{}{
				"specializations": []interface{}{"video", "photo"},
				"portfolio_url":   "https://portfolio.example.com",
			},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/profiles", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		ctx := context.WithValue(req.Context(), middleware.UserContextKey, userRepo.users[userID])
		req = req.WithContext(ctx)

		h.Create(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp handler.CreateProfileResponse
		json.NewDecoder(w.Body).Decode(&resp)
		assert.Equal(t, "editor", resp.Type)
	})

	t.Run("201 Created - valid influencer profile", func(t *testing.T) {
		h, _, userRepo := setupTestProfileHandler()

		userID := uuid.New()
		userRepo.Create(context.Background(), &domain.User{ID: userID, Email: "influencer@example.com"})

		reqBody := handler.CreateProfileRequest{
			ProfileType: "influencer",
			Name:        "Influencer Profile",
			Details: map[string]interface{}{
				"platforms":       []interface{}{"instagram", "youtube"},
				"follower_counts": map[string]interface{}{"instagram": 50000},
			},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/profiles", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		ctx := context.WithValue(req.Context(), middleware.UserContextKey, userRepo.users[userID])
		req = req.WithContext(ctx)

		h.Create(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp handler.CreateProfileResponse
		json.NewDecoder(w.Body).Decode(&resp)
		assert.Equal(t, "influencer", resp.Type)
	})

	t.Run("401 Unauthorized - no user in context", func(t *testing.T) {
		h, _, _ := setupTestProfileHandler()

		reqBody := handler.CreateProfileRequest{
			ProfileType: "brand",
			Name:        "Test",
			Details:     map[string]interface{}{},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/profiles", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.Create(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("400 Bad Request - invalid JSON", func(t *testing.T) {
		h, _, _ := setupTestProfileHandler()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/profiles", bytes.NewReader([]byte("not json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Add mock user to context
		ctx := context.WithValue(req.Context(), middleware.UserContextKey, &domain.User{ID: uuid.New()})
		req = req.WithContext(ctx)

		h.Create(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("400 Bad Request - missing name", func(t *testing.T) {
		h, _, userRepo := setupTestProfileHandler()

		userID := uuid.New()
		userRepo.Create(context.Background(), &domain.User{ID: userID, Email: "test@example.com"})

		reqBody := map[string]interface{}{
			"profile_type": "brand",
			"details":      map[string]interface{}{},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/profiles", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		ctx := context.WithValue(req.Context(), middleware.UserContextKey, userRepo.users[userID])
		req = req.WithContext(ctx)

		h.Create(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("400 Bad Request - invalid profile type", func(t *testing.T) {
		h, _, userRepo := setupTestProfileHandler()

		userID := uuid.New()
		userRepo.Create(context.Background(), &domain.User{ID: userID, Email: "test@example.com"})

		reqBody := handler.CreateProfileRequest{
			ProfileType: "invalid_type",
			Name:        "Test Profile",
			Details:     map[string]interface{}{},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/profiles", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		ctx := context.WithValue(req.Context(), middleware.UserContextKey, userRepo.users[userID])
		req = req.WithContext(ctx)

		h.Create(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp handler.ErrorResponse
		json.NewDecoder(w.Body).Decode(&resp)
		assert.Equal(t, "invalid_profile_type", resp.Error)
	})

	t.Run("400 Bad Request - missing brand details", func(t *testing.T) {
		h, _, userRepo := setupTestProfileHandler()

		userID := uuid.New()
		userRepo.Create(context.Background(), &domain.User{ID: userID, Email: "brand@example.com"})

		reqBody := handler.CreateProfileRequest{
			ProfileType: "brand",
			Name:        "Incomplete Brand",
			Details:     map[string]interface{}{"size": "100"}, // missing company_name, industry
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/profiles", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		ctx := context.WithValue(req.Context(), middleware.UserContextKey, userRepo.users[userID])
		req = req.WithContext(ctx)

		h.Create(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// TestContractListProfilesEndpoint tests GET /api/v1/profiles contract.
func TestContractListProfilesEndpoint(t *testing.T) {
	t.Run("200 OK - empty list", func(t *testing.T) {
		h, _, userRepo := setupTestProfileHandler()

		userID := uuid.New()
		userRepo.Create(context.Background(), &domain.User{ID: userID, Email: "empty@example.com"})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/profiles", nil)
		w := httptest.NewRecorder()

		ctx := context.WithValue(req.Context(), middleware.UserContextKey, userRepo.users[userID])
		req = req.WithContext(ctx)

		h.List(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp []handler.ProfileResponse
		json.NewDecoder(w.Body).Decode(&resp)
		assert.Empty(t, resp)
	})

	t.Run("200 OK - with profiles", func(t *testing.T) {
		h, profileRepo, userRepo := setupTestProfileHandler()

		userID := uuid.New()
		userRepo.Create(context.Background(), &domain.User{ID: userID, Email: "multi@example.com"})

		// Create a profile directly via repo
		details, _ := json.Marshal(map[string]interface{}{"company_name": "Test"})
		profile := domain.NewProfile(userID, domain.ProfileTypeBrand, "Test Brand", details)
		profileRepo.Create(context.Background(), profile)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/profiles", nil)
		w := httptest.NewRecorder()

		ctx := context.WithValue(req.Context(), middleware.UserContextKey, userRepo.users[userID])
		req = req.WithContext(ctx)

		h.List(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp []handler.ProfileResponse
		json.NewDecoder(w.Body).Decode(&resp)
		assert.Len(t, resp, 1)
		assert.Equal(t, "Test Brand", resp[0].Name)
	})

	t.Run("401 Unauthorized - no user in context", func(t *testing.T) {
		h, _, _ := setupTestProfileHandler()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/profiles", nil)
		w := httptest.NewRecorder()

		h.List(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

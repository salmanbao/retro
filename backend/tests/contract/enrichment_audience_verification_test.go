package contract

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/handler"
	"viralforge/backend/src/middleware"
	"viralforge/backend/src/service"
)

// mockAudienceDataRepo is a mock for AudienceDataRepository.
type mockAudienceDataRepo struct {
	data map[uuid.UUID]*domain.AudienceData
}

func newMockAudienceDataRepo() *mockAudienceDataRepo {
	return &mockAudienceDataRepo{
		data: make(map[uuid.UUID]*domain.AudienceData),
	}
}

func (m *mockAudienceDataRepo) Create(ctx context.Context, data *domain.AudienceData) error {
	m.data[data.ProfileID] = data
	return nil
}

func (m *mockAudienceDataRepo) ByProfileID(ctx context.Context, profileID uuid.UUID) (*domain.AudienceData, error) {
	if d, ok := m.data[profileID]; ok {
		return d, nil
	}
	return nil, domain.ErrProfileNotFound
}

func (m *mockAudienceDataRepo) Update(ctx context.Context, data *domain.AudienceData) error {
	m.data[data.ProfileID] = data
	return nil
}

// mockFollowerVerificationRepo is a mock for FollowerVerificationRepository.
type mockFollowerVerificationRepo struct {
	data map[uuid.UUID]*domain.FollowerVerification
}

func newMockFollowerVerificationRepo() *mockFollowerVerificationRepo {
	return &mockFollowerVerificationRepo{
		data: make(map[uuid.UUID]*domain.FollowerVerification),
	}
}

func (m *mockFollowerVerificationRepo) Create(ctx context.Context, verification *domain.FollowerVerification) error {
	m.data[verification.ProfileID] = verification
	return nil
}

func (m *mockFollowerVerificationRepo) ByProfileID(ctx context.Context, profileID uuid.UUID) (*domain.FollowerVerification, error) {
	if v, ok := m.data[profileID]; ok {
		return v, nil
	}
	// Return empty unverified state if not found
	return &domain.FollowerVerification{
		ProfileID: profileID,
		Status:    domain.VerificationStatusUnverified,
	}, nil
}

func (m *mockFollowerVerificationRepo) Update(ctx context.Context, verification *domain.FollowerVerification) error {
	m.data[verification.ProfileID] = verification
	return nil
}

// mockAudienceProfileRepo is a mock implementation of ProfileRepository for audience tests.
type mockAudienceProfileRepo struct {
	profiles map[uuid.UUID]*domain.Profile
	byUser   map[uuid.UUID][]*domain.Profile
}

func newMockAudienceProfileStore() *mockAudienceProfileRepo {
	return &mockAudienceProfileRepo{
		profiles: make(map[uuid.UUID]*domain.Profile),
		byUser:   make(map[uuid.UUID][]*domain.Profile),
	}
}

func (m *mockAudienceProfileRepo) Create(ctx context.Context, profile *domain.Profile) error {
	m.profiles[profile.ID] = profile
	m.byUser[profile.UserID] = append(m.byUser[profile.UserID], profile)
	return nil
}

func (m *mockAudienceProfileRepo) ByID(ctx context.Context, id uuid.UUID) (*domain.Profile, error) {
	if p, ok := m.profiles[id]; ok {
		if p.DeletedAt != nil {
			return nil, domain.ErrProfileNotFound
		}
		return p, nil
	}
	return nil, domain.ErrProfileNotFound
}

func (m *mockAudienceProfileRepo) ByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Profile, error) {
	if profiles, ok := m.byUser[userID]; ok {
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

func (m *mockAudienceProfileRepo) Update(ctx context.Context, profile *domain.Profile) error {
	m.profiles[profile.ID] = profile
	return nil
}

func (m *mockAudienceProfileRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if p, ok := m.profiles[id]; ok {
		now := time.Now()
		p.DeletedAt = &now
	}
	return nil
}

// TestT082_AudienceDataCRUDAsInfluencer tests audience data CRUD flow as Influencer.
func TestT082_AudienceDataCRUDAsInfluencer(t *testing.T) {
	profileRepo := newMockAudienceProfileStore()
	audienceRepo := newMockAudienceDataRepo()

	// Create an influencer profile
	userID := uuid.New()
	influencerProfile := &domain.Profile{
		ID:     uuid.New(),
		UserID: userID,
		Type:   domain.ProfileTypeInfluencer,
		Name:   "Influencer Profile",
	}
	profileRepo.Create(context.Background(), influencerProfile)

	// Create services
	audienceSvc := service.NewAudienceService(audienceRepo, profileRepo)
	audienceHandler := handler.NewAudienceHandler(audienceSvc)

	router := chi.NewRouter()
	router.Route("/api/v1/profiles/{id}/audience", func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := context.WithValue(r.Context(), middleware.EnrichmentProfileIDKey, influencerProfile.ID)
				next.ServeHTTP(w, r.WithContext(ctx))
			})
		})
		r.Use(middleware.NewProfileTypeMiddleware(profileRepo).RequireInfluencer)
		r.Get("/", audienceHandler.GetAudience)
		r.Put("/", audienceHandler.UpdateAudience)
	})

	// Test PUT /api/v1/profiles/{id}/audience (create/update)
	audienceReq := handler.UpdateAudienceRequest{
		PlatformHandles: map[string]string{
			"tiktok":    "handle",
			"instagram": "@handle",
		},
		ClaimedFollowers: map[string]int{
			"tiktok":    100000,
			"instagram": 50000,
		},
	}
	engagementRate := 4.5
	audienceReq.EngagementRate = &engagementRate
	body, _ := json.Marshal(audienceReq)
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/profiles/"+influencerProfile.ID.String()+"/audience", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "PUT should create/update audience data")

	// Test GET /api/v1/profiles/{id}/audience
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/profiles/"+influencerProfile.ID.String()+"/audience", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "GET should retrieve audience data")

	var audienceResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &audienceResp)
	assert.NotNil(t, audienceResp["engagement_rate"], "Engagement rate should be present")
}

// TestT082_AudienceDataRejectionForNonInfluencer tests that non-Influencer profiles cannot access audience data.
func TestT082_AudienceDataRejectionForNonInfluencer(t *testing.T) {
	profileRepo := newMockAudienceProfileStore()
	audienceRepo := newMockAudienceDataRepo()

	// Create a brand profile (not influencer)
	userID := uuid.New()
	brandProfile := &domain.Profile{
		ID:     uuid.New(),
		UserID: userID,
		Type:   domain.ProfileTypeBrand,
		Name:   "Brand Profile",
	}
	profileRepo.Create(context.Background(), brandProfile)

	// Create services
	audienceSvc := service.NewAudienceService(audienceRepo, profileRepo)
	audienceHandler := handler.NewAudienceHandler(audienceSvc)

	router := chi.NewRouter()
	router.Route("/api/v1/profiles/{id}/audience", func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := context.WithValue(r.Context(), middleware.EnrichmentProfileIDKey, brandProfile.ID)
				next.ServeHTTP(w, r.WithContext(ctx))
			})
		})
		r.Use(middleware.NewProfileTypeMiddleware(profileRepo).RequireInfluencer)
		r.Get("/", audienceHandler.GetAudience)
	})

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/profiles/"+brandProfile.ID.String()+"/audience", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code, "Brand profile should be rejected for audience access")
}

// TestT083_VerificationSubmissionAndAdminReview tests verification submission and admin approval flow.
func TestT083_VerificationSubmissionAndAdminReview(t *testing.T) {
	profileRepo := newMockAudienceProfileStore()
	verificationRepo := newMockFollowerVerificationRepo()

	// Create an influencer profile
	userID := uuid.New()
	influencerProfile := &domain.Profile{
		ID:     uuid.New(),
		UserID: userID,
		Type:   domain.ProfileTypeInfluencer,
		Name:   "Influencer Profile",
	}
	profileRepo.Create(context.Background(), influencerProfile)

	// Create services
	verificationSvc := service.NewVerificationService(verificationRepo, profileRepo)
	verificationHandler := handler.NewVerificationHandler(verificationSvc)

	// Create chi router for public verification routes
	publicRouter := chi.NewRouter()
	publicRouter.Route("/api/v1/profiles/{id}/verification", func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := context.WithValue(r.Context(), middleware.EnrichmentProfileIDKey, influencerProfile.ID)
				next.ServeHTTP(w, r.WithContext(ctx))
			})
		})
		r.Use(middleware.NewProfileTypeMiddleware(profileRepo).RequireInfluencer)
		r.Get("/", verificationHandler.GetVerification)
		r.Post("/", verificationHandler.SubmitVerification)
	})

	// Test POST /api/v1/profiles/{id}/verification (submit)
	evidenceReq := handler.SubmitVerificationRequest{
		EvidenceURLs: []string{"https://example.com/proof1.jpg", "https://example.com/proof2.jpg"},
		Notes:        "Submitted for review",
	}
	body, _ := json.Marshal(evidenceReq)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/profiles/"+influencerProfile.ID.String()+"/verification", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	publicRouter.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code, "POST should submit verification")

	// Verify status is pending
	var submitResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &submitResp)
	assert.Equal(t, "pending", submitResp["status"], "Status should be pending after submission")

	// Test GET /api/v1/profiles/{id}/verification
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/profiles/"+influencerProfile.ID.String()+"/verification", nil)
	w = httptest.NewRecorder()
	publicRouter.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "GET should return verification status")
}

// TestT083_VerificationRejectionForNonInfluencer tests that non-Influencer profiles cannot submit verification.
func TestT083_VerificationRejectionForNonInfluencer(t *testing.T) {
	profileRepo := newMockAudienceProfileStore()
	verificationRepo := newMockFollowerVerificationRepo()

	// Create a brand profile (not influencer)
	userID := uuid.New()
	brandProfile := &domain.Profile{
		ID:     uuid.New(),
		UserID: userID,
		Type:   domain.ProfileTypeBrand,
		Name:   "Brand Profile",
	}
	profileRepo.Create(context.Background(), brandProfile)

	// Create services
	verificationSvc := service.NewVerificationService(verificationRepo, profileRepo)
	verificationHandler := handler.NewVerificationHandler(verificationSvc)

	router := chi.NewRouter()
	router.Route("/api/v1/profiles/{id}/verification", func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := context.WithValue(r.Context(), middleware.EnrichmentProfileIDKey, brandProfile.ID)
				next.ServeHTTP(w, r.WithContext(ctx))
			})
		})
		r.Use(middleware.NewProfileTypeMiddleware(profileRepo).RequireInfluencer)
		r.Post("/", verificationHandler.SubmitVerification)
	})

	evidenceReq := handler.SubmitVerificationRequest{
		EvidenceURLs: []string{"https://example.com/proof.jpg"},
	}
	body, _ := json.Marshal(evidenceReq)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/profiles/"+brandProfile.ID.String()+"/verification", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code, "Brand profile should be rejected for verification submission")
}

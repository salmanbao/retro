package contract

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/handler"
	"viralforge/backend/src/middleware"
	"viralforge/backend/src/service"
)

// mockSwitchSessionRepo implements session repository for contract testing.
type mockSwitchSessionRepo struct {
	sessions map[uuid.UUID]*domain.Session
}

func newMockSwitchSessionRepo() *mockSwitchSessionRepo {
	return &mockSwitchSessionRepo{
		sessions: make(map[uuid.UUID]*domain.Session),
	}
}

func (r *mockSwitchSessionRepo) Create(ctx context.Context, session *domain.Session) error {
	r.sessions[session.ID] = session
	return nil
}

func (r *mockSwitchSessionRepo) ByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	if s, ok := r.sessions[id]; ok {
		return s, nil
	}
	return nil, domain.ErrSessionNotFound
}

func (r *mockSwitchSessionRepo) ByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error) {
	for _, s := range r.sessions {
		if s.TokenHash == tokenHash {
			return s, nil
		}
	}
	return nil, domain.ErrSessionNotFound
}

func (r *mockSwitchSessionRepo) ByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Session, error) {
	var result []*domain.Session
	for _, s := range r.sessions {
		if s.UserID == userID {
			result = append(result, s)
		}
	}
	return result, nil
}

func (r *mockSwitchSessionRepo) Update(ctx context.Context, session *domain.Session) error {
	r.sessions[session.ID] = session
	return nil
}

func (r *mockSwitchSessionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(r.sessions, id)
	return nil
}

func (r *mockSwitchSessionRepo) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	for id, s := range r.sessions {
		if s.UserID == userID {
			delete(r.sessions, id)
		}
	}
	return nil
}

// mockSwitchProfileRepo implements profile repository for contract testing.
type mockSwitchProfileRepo struct {
	profiles map[uuid.UUID]*domain.Profile
}

func newMockSwitchProfileRepo() *mockSwitchProfileRepo {
	return &mockSwitchProfileRepo{
		profiles: make(map[uuid.UUID]*domain.Profile),
	}
}

func (r *mockSwitchProfileRepo) Create(ctx context.Context, profile *domain.Profile) error {
	r.profiles[profile.ID] = profile
	return nil
}

func (r *mockSwitchProfileRepo) ByID(ctx context.Context, id uuid.UUID) (*domain.Profile, error) {
	if p, ok := r.profiles[id]; ok {
		if p.DeletedAt != nil {
			return nil, domain.ErrProfileNotFound
		}
		return p, nil
	}
	return nil, domain.ErrProfileNotFound
}

func (r *mockSwitchProfileRepo) ByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Profile, error) {
	var result []*domain.Profile
	for _, p := range r.profiles {
		if p.UserID == userID && p.DeletedAt == nil {
			result = append(result, p)
		}
	}
	return result, nil
}

func (r *mockSwitchProfileRepo) Update(ctx context.Context, profile *domain.Profile) error {
	r.profiles[profile.ID] = profile
	return nil
}

func (r *mockSwitchProfileRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if p, ok := r.profiles[id]; ok {
		p.DeletedAt = nil // Will be set by service
	}
	return nil
}

// setupTestSessionHandler creates a handler configured for testing.
func setupTestSessionHandler() (*handler.SessionHandler, *mockSwitchSessionRepo, *mockSwitchProfileRepo) {
	sessionRepo := newMockSwitchSessionRepo()
	profileRepo := newMockSwitchProfileRepo()
	sessionSvc := service.NewSessionService(sessionRepo, profileRepo)
	sessionHandler := handler.NewSessionHandler(sessionSvc)
	return sessionHandler, sessionRepo, profileRepo
}

// TestContractSwitchActiveProfileEndpoint tests PATCH /api/v1/sessions/active contract.
func TestContractSwitchActiveProfileEndpoint(t *testing.T) {
	t.Run("200 OK - successful profile switch", func(t *testing.T) {
		h, sessionRepo, profileRepo := setupTestSessionHandler()

		userID := uuid.New()
		sessionID := uuid.New()
		profileID := uuid.New()

		// Create user
		user := &domain.User{ID: userID}

		// Create session
		session := &domain.Session{
			ID:        sessionID,
			UserID:    userID,
			TokenHash: "test-token",
		}
		sessionRepo.Create(context.Background(), session)

		// Create profile
		profile := &domain.Profile{
			ID:     profileID,
			UserID: userID,
			Type:   domain.ProfileTypeBrand,
			Name:   "Test Brand",
		}
		profileRepo.Create(context.Background(), profile)

		reqBody := handler.SwitchActiveProfileRequest{
			ProfileID: profileID.String(),
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/sessions/active", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
		ctx = context.WithValue(ctx, middleware.SessionContextKey, session)
		req = req.WithContext(ctx)

		h.SwitchActiveProfile(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp handler.SessionResponse
		json.NewDecoder(w.Body).Decode(&resp)
		assert.Equal(t, profileID.String(), resp.ActiveProfileID)
	})

	t.Run("401 Unauthorized - no user in context", func(t *testing.T) {
		h, _, _ := setupTestSessionHandler()

		reqBody := handler.SwitchActiveProfileRequest{
			ProfileID: uuid.New().String(),
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/sessions/active", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.SwitchActiveProfile(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("400 Bad Request - missing profile ID", func(t *testing.T) {
		h, _, _ := setupTestSessionHandler()

		userID := uuid.New()
		sessionID := uuid.New()

		user := &domain.User{ID: userID}
		session := &domain.Session{
			ID:     sessionID,
			UserID: userID,
		}

		reqBody := handler.SwitchActiveProfileRequest{
			ProfileID: "",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/sessions/active", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
		ctx = context.WithValue(ctx, middleware.SessionContextKey, session)
		req = req.WithContext(ctx)

		h.SwitchActiveProfile(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("400 Bad Request - invalid profile ID UUID", func(t *testing.T) {
		h, _, _ := setupTestSessionHandler()

		userID := uuid.New()
		sessionID := uuid.New()

		user := &domain.User{ID: userID}
		session := &domain.Session{
			ID:     sessionID,
			UserID: userID,
		}

		reqBody := handler.SwitchActiveProfileRequest{
			ProfileID: "not-a-valid-uuid",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/sessions/active", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
		ctx = context.WithValue(ctx, middleware.SessionContextKey, session)
		req = req.WithContext(ctx)

		h.SwitchActiveProfile(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("403 Forbidden - profile does not belong to user", func(t *testing.T) {
		h, sessionRepo, profileRepo := setupTestSessionHandler()

		userID := uuid.New()
		sessionID := uuid.New()
		profileID := uuid.New()

		user := &domain.User{ID: userID}
		session := &domain.Session{
			ID:        sessionID,
			UserID:    userID,
			TokenHash: "test-token",
		}
		sessionRepo.Create(context.Background(), session)

		// Profile owned by DIFFERENT user
		profile := &domain.Profile{
			ID:     profileID,
			UserID: uuid.New(), // Different user!
			Type:   domain.ProfileTypeBrand,
			Name:   "Other Brand",
		}
		profileRepo.Create(context.Background(), profile)

		reqBody := handler.SwitchActiveProfileRequest{
			ProfileID: profileID.String(),
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/sessions/active", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
		ctx = context.WithValue(ctx, middleware.SessionContextKey, session)
		req = req.WithContext(ctx)

		h.SwitchActiveProfile(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

// TestContractListSessionsEndpoint tests GET /api/v1/sessions contract.
func TestContractListSessionsEndpoint(t *testing.T) {
	t.Run("200 OK - returns user sessions", func(t *testing.T) {
		h, sessionRepo, _ := setupTestSessionHandler()

		userID := uuid.New()

		user := &domain.User{ID: userID}

		// Create two sessions
		session1 := &domain.Session{
			ID:        uuid.New(),
			UserID:    userID,
			TokenHash: "token1",
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		session2 := &domain.Session{
			ID:        uuid.New(),
			UserID:    userID,
			TokenHash: "token2",
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		sessionRepo.Create(context.Background(), session1)
		sessionRepo.Create(context.Background(), session2)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions", nil)
		w := httptest.NewRecorder()

		ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
		req = req.WithContext(ctx)

		h.ListSessions(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp []handler.SessionResponse
		json.NewDecoder(w.Body).Decode(&resp)
		assert.Len(t, resp, 2)
	})

	t.Run("401 Unauthorized - no user in context", func(t *testing.T) {
		h, _, _ := setupTestSessionHandler()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions", nil)
		w := httptest.NewRecorder()

		h.ListSessions(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("200 OK - empty list for user with no sessions", func(t *testing.T) {
		h, _, _ := setupTestSessionHandler()

		userID := uuid.New()
		user := &domain.User{ID: userID}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions", nil)
		w := httptest.NewRecorder()

		ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
		req = req.WithContext(ctx)

		h.ListSessions(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp []handler.SessionResponse
		json.NewDecoder(w.Body).Decode(&resp)
		assert.Empty(t, resp)
	})
}

// TestContractDeleteSessionEndpoint tests DELETE /api/v1/sessions/{id} contract.
func TestContractDeleteSessionEndpoint(t *testing.T) {
	t.Run("204 No Content - successful deletion", func(t *testing.T) {
		h, sessionRepo, _ := setupTestSessionHandler()

		userID := uuid.New()
		sessionID := uuid.New()

		user := &domain.User{ID: userID}
		session := &domain.Session{
			ID:        sessionID,
			UserID:    userID,
			TokenHash: "delete-token",
		}
		sessionRepo.Create(context.Background(), session)

		r := chi.NewRouter()
		r.Delete("/{id}", h.DeleteSession)

		req := httptest.NewRequest(http.MethodDelete, "/"+sessionID.String(), nil)
		w := httptest.NewRecorder()

		ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
		req = req.WithContext(ctx)

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("400 Bad Request - invalid UUID", func(t *testing.T) {
		h, _, _ := setupTestSessionHandler()

		userID := uuid.New()
		user := &domain.User{ID: userID}

		r := chi.NewRouter()
		r.Delete("/{id}", h.DeleteSession)

		req := httptest.NewRequest(http.MethodDelete, "/not-a-uuid", nil)
		w := httptest.NewRecorder()

		ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
		req = req.WithContext(ctx)

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("401 Unauthorized - no user in context", func(t *testing.T) {
		h, _, _ := setupTestSessionHandler()

		r := chi.NewRouter()
		r.Delete("/{id}", h.DeleteSession)

		req := httptest.NewRequest(http.MethodDelete, "/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()

		h.DeleteSession(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("404 Not Found - session not found", func(t *testing.T) {
		h, _, _ := setupTestSessionHandler()

		userID := uuid.New()
		user := &domain.User{ID: userID}

		r := chi.NewRouter()
		r.Delete("/{id}", h.DeleteSession)

		req := httptest.NewRequest(http.MethodDelete, "/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()

		ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
		req = req.WithContext(ctx)

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

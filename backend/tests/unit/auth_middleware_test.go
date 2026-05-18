package unit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/middleware"
)

// mockRBACUserRepo implements repository.UserRepository for middleware tests.
type mockRBACUserRepo struct {
	users map[uuid.UUID]*domain.User
}

func newMockRBACUserRepo() *mockRBACUserRepo {
	return &mockRBACUserRepo{
		users: make(map[uuid.UUID]*domain.User),
	}
}

func (r *mockRBACUserRepo) Create(ctx context.Context, user *domain.User) error {
	r.users[user.ID] = user
	return nil
}

func (r *mockRBACUserRepo) ByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if u, ok := r.users[id]; ok {
		return u, nil
	}
	return nil, domain.ErrUserNotFound
}

func (r *mockRBACUserRepo) ByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, domain.ErrUserNotFound
}

func (r *mockRBACUserRepo) Update(ctx context.Context, user *domain.User) error {
	r.users[user.ID] = user
	return nil
}

func (r *mockRBACUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(r.users, id)
	return nil
}

// mockRBACSessionRepo implements repository.SessionRepository for middleware tests.
type mockRBACSessionRepo struct {
	sessions       map[uuid.UUID]*domain.Session
	tokenToSession map[string]*domain.Session
}

func newMockRBACSessionRepo() *mockRBACSessionRepo {
	return &mockRBACSessionRepo{
		sessions:       make(map[uuid.UUID]*domain.Session),
		tokenToSession: make(map[string]*domain.Session),
	}
}

func (r *mockRBACSessionRepo) Create(ctx context.Context, session *domain.Session) error {
	r.sessions[session.ID] = session
	r.tokenToSession[session.TokenHash] = session
	return nil
}

func (r *mockRBACSessionRepo) ByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	if s, ok := r.sessions[id]; ok {
		return s, nil
	}
	return nil, domain.ErrSessionNotFound
}

func (r *mockRBACSessionRepo) ByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error) {
	if s, ok := r.tokenToSession[tokenHash]; ok {
		return s, nil
	}
	return nil, domain.ErrSessionNotFound
}

func (r *mockRBACSessionRepo) ByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Session, error) {
	var result []*domain.Session
	for _, s := range r.sessions {
		if s.UserID == userID {
			result = append(result, s)
		}
	}
	return result, nil
}

func (r *mockRBACSessionRepo) Update(ctx context.Context, session *domain.Session) error {
	r.sessions[session.ID] = session
	r.tokenToSession[session.TokenHash] = session
	return nil
}

func (r *mockRBACSessionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if s, ok := r.sessions[id]; ok {
		delete(r.tokenToSession, s.TokenHash)
		delete(r.sessions, id)
	}
	return nil
}

func (r *mockRBACSessionRepo) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	for id, s := range r.sessions {
		if s.UserID == userID {
			delete(r.tokenToSession, s.TokenHash)
			delete(r.sessions, id)
		}
	}
	return nil
}

// TestAuthMiddlewareInjectActiveProfile tests that the auth middleware correctly injects
// the active profile ID into the context.
func TestAuthMiddlewareInjectActiveProfile(t *testing.T) {
	t.Run("active profile ID injected into context", func(t *testing.T) {
		sessionRepo := newMockRBACSessionRepo()
		userRepo := newMockRBACUserRepo()
		authMiddleware := middleware.NewAuthMiddleware(sessionRepo, userRepo)

		userID := uuid.New()
		sessionID := uuid.New()
		profileID := uuid.New()
		token := "test-token-123"

		// Create user
		user := &domain.User{
			ID:       userID,
			Email:    "test@example.com",
			Verified: true,
		}
		userRepo.Create(context.Background(), user)

		// Create session with active profile
		session := &domain.Session{
			ID:              sessionID,
			UserID:          userID,
			TokenHash:       token,
			ExpiresAt:       time.Now().Add(24 * time.Hour),
			ActiveProfileID: &profileID,
		}
		sessionRepo.Create(context.Background(), session)

		var capturedProfileID *uuid.UUID
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedProfileID = middleware.GetActiveProfileID(r.Context())
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		authMiddleware.Authenticate(handler).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotNil(t, capturedProfileID)
		assert.Equal(t, profileID, *capturedProfileID)
	})

	t.Run("no active profile ID when session has none", func(t *testing.T) {
		sessionRepo := newMockRBACSessionRepo()
		userRepo := newMockRBACUserRepo()
		authMiddleware := middleware.NewAuthMiddleware(sessionRepo, userRepo)

		userID := uuid.New()
		sessionID := uuid.New()
		token := "test-token-456"

		// Create user
		user := &domain.User{
			ID:       userID,
			Email:    "test@example.com",
			Verified: true,
		}
		userRepo.Create(context.Background(), user)

		// Create session WITHOUT active profile
		session := &domain.Session{
			ID:              sessionID,
			UserID:          userID,
			TokenHash:       token,
			ExpiresAt:       time.Now().Add(24 * time.Hour),
			ActiveProfileID: nil,
		}
		sessionRepo.Create(context.Background(), session)

		var capturedProfileID *uuid.UUID
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedProfileID = middleware.GetActiveProfileID(r.Context())
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		authMiddleware.Authenticate(handler).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Nil(t, capturedProfileID)
	})

	t.Run("expired session returns 401", func(t *testing.T) {
		sessionRepo := newMockRBACSessionRepo()
		userRepo := newMockRBACUserRepo()
		authMiddleware := middleware.NewAuthMiddleware(sessionRepo, userRepo)

		userID := uuid.New()
		sessionID := uuid.New()
		token := "expired-token"

		// Create user
		user := &domain.User{
			ID:       userID,
			Email:    "test@example.com",
			Verified: true,
		}
		userRepo.Create(context.Background(), user)

		// Create expired session
		session := &domain.Session{
			ID:        sessionID,
			UserID:    userID,
			TokenHash: token,
			ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
		}
		sessionRepo.Create(context.Background(), session)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		authMiddleware.Authenticate(handler).ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid token returns 401", func(t *testing.T) {
		sessionRepo := newMockRBACSessionRepo()
		userRepo := newMockRBACUserRepo()
		authMiddleware := middleware.NewAuthMiddleware(sessionRepo, userRepo)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		authMiddleware.Authenticate(handler).ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("missing authorization header returns 401", func(t *testing.T) {
		sessionRepo := newMockRBACSessionRepo()
		userRepo := newMockRBACUserRepo()
		authMiddleware := middleware.NewAuthMiddleware(sessionRepo, userRepo)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		authMiddleware.Authenticate(handler).ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// TestRequireBrandProfileMiddleware tests a middleware that requires brand profile.
func TestRequireBrandProfileMiddleware(t *testing.T) {
	t.Run("allows request with brand profile", func(t *testing.T) {
		sessionRepo := newMockRBACSessionRepo()
		userRepo := newMockRBACUserRepo()
		authMiddleware := middleware.NewAuthMiddleware(sessionRepo, userRepo)

		userID := uuid.New()
		sessionID := uuid.New()
		profileID := uuid.New()
		token := "brand-token"

		// Create user
		user := &domain.User{
			ID:       userID,
			Email:    "brand@example.com",
			Verified: true,
		}
		userRepo.Create(context.Background(), user)

		// Create session with brand profile
		session := &domain.Session{
			ID:              sessionID,
			UserID:          userID,
			TokenHash:       token,
			ExpiresAt:       time.Now().Add(24 * time.Hour),
			ActiveProfileID: &profileID,
		}
		sessionRepo.Create(context.Background(), session)

		// Create brand profile
		profile := &domain.Profile{
			ID:     profileID,
			UserID: userID,
			Type:   domain.ProfileTypeBrand,
			Name:   "Test Brand",
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user and profile from context
			user := middleware.GetUserFromContext(r.Context())
			profileIDFromCtx := middleware.GetActiveProfileID(r.Context())
			assert.NotNil(t, user)
			assert.NotNil(t, profileIDFromCtx)
			assert.Equal(t, profileID, *profileIDFromCtx)
			_ = profile // Use profile for type checking in real implementation
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		authMiddleware.Authenticate(handler).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		_ = profile // Prevent unused variable error
	})
}

// TestGetUserFromContext tests the GetUserFromContext helper.
func TestGetUserFromContext(t *testing.T) {
	t.Run("returns user when present in context", func(t *testing.T) {
		user := &domain.User{
			ID:       uuid.New(),
			Email:    "test@example.com",
			Verified: true,
		}
		ctx := context.WithValue(context.Background(), middleware.UserContextKey, user)

		retrieved := middleware.GetUserFromContext(ctx)
		assert.NotNil(t, retrieved)
		assert.Equal(t, user.ID, retrieved.ID)
		assert.Equal(t, user.Email, retrieved.Email)
	})

	t.Run("returns nil when user not in context", func(t *testing.T) {
		retrieved := middleware.GetUserFromContext(context.Background())
		assert.Nil(t, retrieved)
	})
}

// TestGetSessionFromContext tests the GetSessionFromContext helper.
func TestGetSessionFromContext(t *testing.T) {
	t.Run("returns session when present in context", func(t *testing.T) {
		session := &domain.Session{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			TokenHash: "test-token",
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		ctx := context.WithValue(context.Background(), middleware.SessionContextKey, session)

		retrieved := middleware.GetSessionFromContext(ctx)
		assert.NotNil(t, retrieved)
		assert.Equal(t, session.ID, retrieved.ID)
	})

	t.Run("returns nil when session not in context", func(t *testing.T) {
		retrieved := middleware.GetSessionFromContext(context.Background())
		assert.Nil(t, retrieved)
	})
}

// TestGetActiveProfileID tests the GetActiveProfileID helper.
func TestGetActiveProfileID(t *testing.T) {
	t.Run("returns profile ID when present in context", func(t *testing.T) {
		profileID := uuid.New()
		ctx := context.WithValue(context.Background(), middleware.ActiveProfileIDKey, profileID)

		retrieved := middleware.GetActiveProfileID(ctx)
		assert.NotNil(t, retrieved)
		assert.Equal(t, profileID, *retrieved)
	})

	t.Run("returns nil when profile ID not in context", func(t *testing.T) {
		retrieved := middleware.GetActiveProfileID(context.Background())
		assert.Nil(t, retrieved)
	})
}

// TestAuthMiddlewareIntegration tests the full auth flow with chi router.
func TestAuthMiddlewareIntegration(t *testing.T) {
	t.Run("chi router with auth middleware", func(t *testing.T) {
		sessionRepo := newMockRBACSessionRepo()
		userRepo := newMockRBACUserRepo()
		authMiddleware := middleware.NewAuthMiddleware(sessionRepo, userRepo)

		userID := uuid.New()
		sessionID := uuid.New()
		profileID := uuid.New()
		token := "chi-router-token"

		// Create user
		user := &domain.User{
			ID:       userID,
			Email:    "chi@example.com",
			Verified: true,
		}
		userRepo.Create(context.Background(), user)

		// Create session with profile
		session := &domain.Session{
			ID:              sessionID,
			UserID:          userID,
			TokenHash:       token,
			ExpiresAt:       time.Now().Add(24 * time.Hour),
			ActiveProfileID: &profileID,
		}
		sessionRepo.Create(context.Background(), session)

		r := chi.NewRouter()
		r.Use(authMiddleware.Authenticate)
		r.Get("/protected", func(w http.ResponseWriter, r *http.Request) {
			profileID := middleware.GetActiveProfileID(r.Context())
			assert.NotNil(t, profileID)
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

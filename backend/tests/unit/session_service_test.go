package unit

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/service"
)

// mockSwitchProfileUserRepo implements repository.UserRepository for session tests.
type mockSwitchProfileUserRepo struct {
	users map[uuid.UUID]*domain.User
}

func newMockSwitchProfileUserRepo() *mockSwitchProfileUserRepo {
	return &mockSwitchProfileUserRepo{
		users: make(map[uuid.UUID]*domain.User),
	}
}

func (r *mockSwitchProfileUserRepo) Create(ctx context.Context, user *domain.User) error {
	r.users[user.ID] = user
	return nil
}

func (r *mockSwitchProfileUserRepo) ByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if u, ok := r.users[id]; ok {
		return u, nil
	}
	return nil, domain.ErrUserNotFound
}

func (r *mockSwitchProfileUserRepo) ByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, domain.ErrUserNotFound
}

func (r *mockSwitchProfileUserRepo) Update(ctx context.Context, user *domain.User) error {
	r.users[user.ID] = user
	return nil
}

func (r *mockSwitchProfileUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(r.users, id)
	return nil
}

// mockSwitchProfileSessionRepo implements repository.SessionRepository for session tests.
type mockSwitchProfileSessionRepo struct {
	sessions map[uuid.UUID]*domain.Session
}

func newMockSwitchProfileSessionRepo() *mockSwitchProfileSessionRepo {
	return &mockSwitchProfileSessionRepo{
		sessions: make(map[uuid.UUID]*domain.Session),
	}
}

func (r *mockSwitchProfileSessionRepo) Create(ctx context.Context, session *domain.Session) error {
	r.sessions[session.ID] = session
	return nil
}

func (r *mockSwitchProfileSessionRepo) ByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	if s, ok := r.sessions[id]; ok {
		return s, nil
	}
	return nil, domain.ErrSessionNotFound
}

func (r *mockSwitchProfileSessionRepo) ByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error) {
	for _, s := range r.sessions {
		return s, nil
	}
	return nil, domain.ErrSessionNotFound
}

func (r *mockSwitchProfileSessionRepo) ByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Session, error) {
	var result []*domain.Session
	for _, s := range r.sessions {
		if s.UserID == userID {
			result = append(result, s)
		}
	}
	return result, nil
}

func (r *mockSwitchProfileSessionRepo) Update(ctx context.Context, session *domain.Session) error {
	r.sessions[session.ID] = session
	return nil
}

func (r *mockSwitchProfileSessionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(r.sessions, id)
	return nil
}

func (r *mockSwitchProfileSessionRepo) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	for id, s := range r.sessions {
		if s.UserID == userID {
			delete(r.sessions, id)
		}
	}
	return nil
}

// mockSwitchProfileProfileRepo implements repository.ProfileRepository for session tests.
type mockSwitchProfileProfileRepo struct {
	profiles map[uuid.UUID]*domain.Profile
}

func newMockSwitchProfileProfileRepo() *mockSwitchProfileProfileRepo {
	return &mockSwitchProfileProfileRepo{
		profiles: make(map[uuid.UUID]*domain.Profile),
	}
}

func (r *mockSwitchProfileProfileRepo) Create(ctx context.Context, profile *domain.Profile) error {
	r.profiles[profile.ID] = profile
	return nil
}

func (r *mockSwitchProfileProfileRepo) ByID(ctx context.Context, id uuid.UUID) (*domain.Profile, error) {
	if p, ok := r.profiles[id]; ok {
		if p.DeletedAt != nil {
			return nil, domain.ErrProfileNotFound
		}
		return p, nil
	}
	return nil, domain.ErrProfileNotFound
}

func (r *mockSwitchProfileProfileRepo) ByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Profile, error) {
	var result []*domain.Profile
	for _, p := range r.profiles {
		if p.UserID == userID && p.DeletedAt == nil {
			result = append(result, p)
		}
	}
	return result, nil
}

func (r *mockSwitchProfileProfileRepo) Update(ctx context.Context, profile *domain.Profile) error {
	r.profiles[profile.ID] = profile
	return nil
}

func (r *mockSwitchProfileProfileRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if p, ok := r.profiles[id]; ok {
		now := time.Now()
		p.DeletedAt = &now
	}
	return nil
}

// TestSwitchActiveProfile tests the SwitchActiveProfile method of SessionService.
func TestSwitchActiveProfile(t *testing.T) {
	t.Run("successful profile switch", func(t *testing.T) {
		sessionRepo := newMockSwitchProfileSessionRepo()
		profileRepo := newMockSwitchProfileProfileRepo()
		svc := service.NewSessionService(sessionRepo, profileRepo)

		userID := uuid.New()
		sessionID := uuid.New()
		profileID := uuid.New()

		// Create a session
		session := &domain.Session{
			ID:        sessionID,
			UserID:    userID,
			TokenHash: "test-token",
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		sessionRepo.Create(context.Background(), session)

		// Create a profile
		profile := &domain.Profile{
			ID:     profileID,
			UserID: userID,
			Type:   domain.ProfileTypeBrand,
			Name:   "Test Brand",
		}
		profileRepo.Create(context.Background(), profile)

		// Switch active profile
		updatedSession, err := svc.SwitchActiveProfile(context.Background(), userID, sessionID, profileID)
		assert.NoError(t, err)
		assert.NotNil(t, updatedSession)
		assert.Equal(t, profileID, *updatedSession.ActiveProfileID)
	})

	t.Run("session not found", func(t *testing.T) {
		sessionRepo := newMockSwitchProfileSessionRepo()
		profileRepo := newMockSwitchProfileProfileRepo()
		svc := service.NewSessionService(sessionRepo, profileRepo)

		userID := uuid.New()
		sessionID := uuid.New()
		profileID := uuid.New()

		_, err := svc.SwitchActiveProfile(context.Background(), userID, sessionID, profileID)
		assert.Error(t, err)
		assert.Equal(t, domain.ErrSessionNotFound, err)
	})

	t.Run("profile not owned by user", func(t *testing.T) {
		sessionRepo := newMockSwitchProfileSessionRepo()
		profileRepo := newMockSwitchProfileProfileRepo()
		svc := service.NewSessionService(sessionRepo, profileRepo)

		userID := uuid.New()
		sessionID := uuid.New()
		profileID := uuid.New()

		// Create a session owned by user
		session := &domain.Session{
			ID:        sessionID,
			UserID:    userID,
			TokenHash: "test-token",
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		sessionRepo.Create(context.Background(), session)

		// Create a profile owned by DIFFERENT user
		profile := &domain.Profile{
			ID:     profileID,
			UserID: uuid.New(), // Different user
			Type:   domain.ProfileTypeBrand,
			Name:   "Test Brand",
		}
		profileRepo.Create(context.Background(), profile)

		_, err := svc.SwitchActiveProfile(context.Background(), userID, sessionID, profileID)
		assert.Error(t, err)
		assert.Equal(t, domain.ErrProfileNotOwned, err)
	})

	t.Run("session belongs to different user", func(t *testing.T) {
		sessionRepo := newMockSwitchProfileSessionRepo()
		profileRepo := newMockSwitchProfileProfileRepo()
		svc := service.NewSessionService(sessionRepo, profileRepo)

		userID := uuid.New()
		sessionID := uuid.New()
		profileID := uuid.New()

		// Create a session owned by DIFFERENT user
		session := &domain.Session{
			ID:        sessionID,
			UserID:    uuid.New(), // Different user
			TokenHash: "test-token",
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		sessionRepo.Create(context.Background(), session)

		// Create a profile owned by user
		profile := &domain.Profile{
			ID:     profileID,
			UserID: userID,
			Type:   domain.ProfileTypeBrand,
			Name:   "Test Brand",
		}
		profileRepo.Create(context.Background(), profile)

		_, err := svc.SwitchActiveProfile(context.Background(), userID, sessionID, profileID)
		assert.Error(t, err)
		assert.Equal(t, domain.ErrSessionNotFound, err)
	})
}

// TestListSessions tests the ListSessions method of SessionService.
func TestListSessions(t *testing.T) {
	t.Run("list active sessions only", func(t *testing.T) {
		sessionRepo := newMockSwitchProfileSessionRepo()
		profileRepo := newMockSwitchProfileProfileRepo()
		svc := service.NewSessionService(sessionRepo, profileRepo)

		userID := uuid.New()

		// Create active session
		activeSession := &domain.Session{
			ID:        uuid.New(),
			UserID:    userID,
			TokenHash: "active-token",
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		sessionRepo.Create(context.Background(), activeSession)

		// Create expired session
		expiredSession := &domain.Session{
			ID:        uuid.New(),
			UserID:    userID,
			TokenHash: "expired-token",
			ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
		}
		sessionRepo.Create(context.Background(), expiredSession)

		sessions, err := svc.ListSessions(context.Background(), userID)
		assert.NoError(t, err)
		assert.Len(t, sessions, 1)
		assert.Equal(t, activeSession.ID, sessions[0].ID)
	})

	t.Run("empty list for user with no sessions", func(t *testing.T) {
		sessionRepo := newMockSwitchProfileSessionRepo()
		profileRepo := newMockSwitchProfileProfileRepo()
		svc := service.NewSessionService(sessionRepo, profileRepo)

		sessions, err := svc.ListSessions(context.Background(), uuid.New())
		assert.NoError(t, err)
		assert.Empty(t, sessions)
	})
}

// TestRevokeSession tests the RevokeSession method of SessionService.
func TestRevokeSession(t *testing.T) {
	t.Run("successful revocation", func(t *testing.T) {
		sessionRepo := newMockSwitchProfileSessionRepo()
		profileRepo := newMockSwitchProfileProfileRepo()
		svc := service.NewSessionService(sessionRepo, profileRepo)

		userID := uuid.New()
		sessionID := uuid.New()

		session := &domain.Session{
			ID:        sessionID,
			UserID:    userID,
			TokenHash: "test-token",
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		sessionRepo.Create(context.Background(), session)

		err := svc.RevokeSession(context.Background(), userID, sessionID)
		assert.NoError(t, err)

		// Verify session is deleted
		_, err = sessionRepo.ByID(context.Background(), sessionID)
		assert.Error(t, err)
	})

	t.Run("session not found", func(t *testing.T) {
		sessionRepo := newMockSwitchProfileSessionRepo()
		profileRepo := newMockSwitchProfileProfileRepo()
		svc := service.NewSessionService(sessionRepo, profileRepo)

		err := svc.RevokeSession(context.Background(), uuid.New(), uuid.New())
		assert.Error(t, err)
	})

	t.Run("session belongs to different user", func(t *testing.T) {
		sessionRepo := newMockSwitchProfileSessionRepo()
		profileRepo := newMockSwitchProfileProfileRepo()
		svc := service.NewSessionService(sessionRepo, profileRepo)

		userID := uuid.New()
		sessionID := uuid.New()

		session := &domain.Session{
			ID:        sessionID,
			UserID:    uuid.New(), // Different user
			TokenHash: "test-token",
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		sessionRepo.Create(context.Background(), session)

		err := svc.RevokeSession(context.Background(), userID, sessionID)
		assert.Error(t, err)
		assert.Equal(t, domain.ErrSessionNotFound, err)
	})
}

// TestRBACCheckProfileAccess tests RBAC helper functions.
func TestRBACCheckProfileAccess(t *testing.T) {
	t.Run("brand profile access granted for brand requirement", func(t *testing.T) {
		profile := &domain.Profile{
			ID:   uuid.New(),
			Type: domain.ProfileTypeBrand,
			Name: "Test Brand",
		}
		result := service.CheckProfileAccess(profile, domain.ProfileTypeBrand)
		assert.True(t, result.Allowed)
		assert.Empty(t, result.Reason)
	})

	t.Run("editor profile access denied for brand requirement", func(t *testing.T) {
		profile := &domain.Profile{
			ID:   uuid.New(),
			Type: domain.ProfileTypeEditor,
			Name: "Test Editor",
		}
		result := service.CheckProfileAccess(profile, domain.ProfileTypeBrand)
		assert.False(t, result.Allowed)
		assert.Equal(t, "profile type does not match required type", result.Reason)
	})

	t.Run("nil profile access denied", func(t *testing.T) {
		result := service.CheckProfileAccess(nil, domain.ProfileTypeBrand)
		assert.False(t, result.Allowed)
		assert.Equal(t, "no active profile", result.Reason)
	})

	t.Run("RequireBrandProfile returns error for non-brand profile", func(t *testing.T) {
		profile := &domain.Profile{
			ID:   uuid.New(),
			Type: domain.ProfileTypeEditor,
			Name: "Test Editor",
		}
		err := service.RequireBrandProfile(profile)
		assert.Error(t, err)
		assert.Equal(t, domain.ErrUnauthorized, err)
	})

	t.Run("RequireBrandProfile returns nil for brand profile", func(t *testing.T) {
		profile := &domain.Profile{
			ID:   uuid.New(),
			Type: domain.ProfileTypeBrand,
			Name: "Test Brand",
		}
		err := service.RequireBrandProfile(profile)
		assert.NoError(t, err)
	})
}
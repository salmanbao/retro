package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/service"
)

// mockProfileRepoForSessionIntegration implements repository.ProfileRepository for integration tests.
type mockProfileRepoForSessionIntegration struct {
	profiles map[uuid.UUID]*domain.Profile
}

func newMockProfileRepoForSessionIntegration() *mockProfileRepoForSessionIntegration {
	return &mockProfileRepoForSessionIntegration{
		profiles: make(map[uuid.UUID]*domain.Profile),
	}
}

func (r *mockProfileRepoForSessionIntegration) Create(ctx context.Context, profile *domain.Profile) error {
	r.profiles[profile.ID] = profile
	return nil
}

func (r *mockProfileRepoForSessionIntegration) ByID(ctx context.Context, id uuid.UUID) (*domain.Profile, error) {
	if p, ok := r.profiles[id]; ok {
		if p.DeletedAt != nil {
			return nil, domain.ErrProfileNotFound
		}
		return p, nil
	}
	return nil, domain.ErrProfileNotFound
}

func (r *mockProfileRepoForSessionIntegration) ByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Profile, error) {
	var result []*domain.Profile
	for _, p := range r.profiles {
		if p.UserID == userID && p.DeletedAt == nil {
			result = append(result, p)
		}
	}
	return result, nil
}

func (r *mockProfileRepoForSessionIntegration) Update(ctx context.Context, profile *domain.Profile) error {
	r.profiles[profile.ID] = profile
	return nil
}

func (r *mockProfileRepoForSessionIntegration) Delete(ctx context.Context, id uuid.UUID) error {
	if p, ok := r.profiles[id]; ok {
		now := time.Now()
		p.DeletedAt = &now
	}
	return nil
}

// TestSessionServiceIntegration tests the SessionService with real collaborator interactions.
func TestSessionServiceIntegration(t *testing.T) {
	t.Run("SwitchActiveProfile - full flow with profile validation", func(t *testing.T) {
		sessionRepo := &mockSessionRepoForProfileSwitch{}
		profileRepo := newMockProfileRepoForSessionIntegration()
		svc := service.NewSessionService(sessionRepo, profileRepo)

		userID := uuid.New()
		sessionID := uuid.New()
		brandProfileID := uuid.New()
		editorProfileID := uuid.New()

		// Setup: Create user with two profiles (brand and editor)
		brandProfile := &domain.Profile{
			ID:     brandProfileID,
			UserID: userID,
			Type:   domain.ProfileTypeBrand,
			Name:   "Test Brand",
		}
		profileRepo.Create(context.Background(), brandProfile)

		editorProfile := &domain.Profile{
			ID:     editorProfileID,
			UserID: userID,
			Type:   domain.ProfileTypeEditor,
			Name:   "Test Editor",
		}
		profileRepo.Create(context.Background(), editorProfile)

		// Setup: Create session
		session := &domain.Session{
			ID:             sessionID,
			UserID:         userID,
			TokenHash:      "test-token",
			ExpiresAt:      time.Now().Add(24 * time.Hour),
			ActiveProfileID: nil,
		}
		sessionRepo.Create(context.Background(), session)

		// Test: Switch to brand profile
		updatedSession, err := svc.SwitchActiveProfile(context.Background(), userID, sessionID, brandProfileID)
		assert.NoError(t, err)
		assert.NotNil(t, updatedSession)
		assert.Equal(t, brandProfileID, *updatedSession.ActiveProfileID)

		// Test: Switch to editor profile
		updatedSession, err = svc.SwitchActiveProfile(context.Background(), userID, sessionID, editorProfileID)
		assert.NoError(t, err)
		assert.NotNil(t, updatedSession)
		assert.Equal(t, editorProfileID, *updatedSession.ActiveProfileID)

		// Verify session was updated in repo
		retrievedSession, err := sessionRepo.ByID(context.Background(), sessionID)
		assert.NoError(t, err)
		assert.Equal(t, editorProfileID, *retrievedSession.ActiveProfileID)
	})

	t.Run("ListSessions - filters expired sessions", func(t *testing.T) {
		sessionRepo := &mockSessionRepoForProfileSwitch{}
		profileRepo := newMockProfileRepoForSessionIntegration()
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
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}
		sessionRepo.Create(context.Background(), expiredSession)

		// Create session expiring in 1 hour (still active)
		nearExpirySession := &domain.Session{
			ID:        uuid.New(),
			UserID:    userID,
			TokenHash: "near-expiry-token",
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}
		sessionRepo.Create(context.Background(), nearExpirySession)

		sessions, err := svc.ListSessions(context.Background(), userID)
		assert.NoError(t, err)
		assert.Len(t, sessions, 2)

		// Verify both active sessions are returned
		sessionIDs := make(map[uuid.UUID]bool)
		for _, s := range sessions {
			sessionIDs[s.ID] = true
		}
		assert.True(t, sessionIDs[activeSession.ID])
		assert.True(t, sessionIDs[nearExpirySession.ID])
		assert.False(t, sessionIDs[expiredSession.ID])
	})

	t.Run("RevokeSession - removes session and prevents reuse", func(t *testing.T) {
		sessionRepo := &mockSessionRepoForProfileSwitch{}
		profileRepo := newMockProfileRepoForSessionIntegration()
		svc := service.NewSessionService(sessionRepo, profileRepo)

		userID := uuid.New()
		sessionID := uuid.New()

		session := &domain.Session{
			ID:        sessionID,
			UserID:    userID,
			TokenHash: "revoke-token",
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		sessionRepo.Create(context.Background(), session)

		// Revoke the session
		err := svc.RevokeSession(context.Background(), userID, sessionID)
		assert.NoError(t, err)

		// Verify session no longer exists
		_, err = sessionRepo.ByID(context.Background(), sessionID)
		assert.Error(t, err)
		assert.Equal(t, domain.ErrSessionNotFound, err)

		// Attempting to revoke again should fail
		err = svc.RevokeSession(context.Background(), userID, sessionID)
		assert.Error(t, err)
	})

	t.Run("GetSessionWithProfile - returns session with profile", func(t *testing.T) {
		sessionRepo := &mockSessionRepoForProfileSwitch{}
		profileRepo := newMockProfileRepoForSessionIntegration()
		svc := service.NewSessionService(sessionRepo, profileRepo)

		userID := uuid.New()
		sessionID := uuid.New()
		profileID := uuid.New()

		// Create profile
		profile := &domain.Profile{
			ID:     profileID,
			UserID: userID,
			Type:   domain.ProfileTypeBrand,
			Name:   "Test Brand",
		}
		profileRepo.Create(context.Background(), profile)

		// Create session with active profile
		session := &domain.Session{
			ID:             sessionID,
			UserID:         userID,
			TokenHash:      "profile-token",
			ExpiresAt:      time.Now().Add(24 * time.Hour),
			ActiveProfileID: &profileID,
		}
		sessionRepo.Create(context.Background(), session)

		retrievedSession, retrievedProfile, err := svc.GetSessionWithProfile(context.Background(), userID, sessionID)
		assert.NoError(t, err)
		assert.NotNil(t, retrievedSession)
		assert.NotNil(t, retrievedProfile)
		assert.Equal(t, sessionID, retrievedSession.ID)
		assert.Equal(t, profileID, retrievedProfile.ID)
	})

	t.Run("GetSessionWithProfile - returns session without profile when none active", func(t *testing.T) {
		sessionRepo := &mockSessionRepoForProfileSwitch{}
		profileRepo := newMockProfileRepoForSessionIntegration()
		svc := service.NewSessionService(sessionRepo, profileRepo)

		userID := uuid.New()
		sessionID := uuid.New()

		// Create session without active profile
		session := &domain.Session{
			ID:             sessionID,
			UserID:         userID,
			TokenHash:      "no-profile-token",
			ExpiresAt:      time.Now().Add(24 * time.Hour),
			ActiveProfileID: nil,
		}
		sessionRepo.Create(context.Background(), session)

		retrievedSession, retrievedProfile, err := svc.GetSessionWithProfile(context.Background(), userID, sessionID)
		assert.NoError(t, err)
		assert.NotNil(t, retrievedSession)
		assert.Nil(t, retrievedProfile)
	})

	t.Run("SwitchActiveProfile - rejects profile from different user", func(t *testing.T) {
		sessionRepo := &mockSessionRepoForProfileSwitch{}
		profileRepo := newMockProfileRepoForSessionIntegration()
		svc := service.NewSessionService(sessionRepo, profileRepo)

		userID := uuid.New()
		sessionID := uuid.New()
		otherUserProfileID := uuid.New()

		// Create session for user
		session := &domain.Session{
			ID:        sessionID,
			UserID:    userID,
			TokenHash: "other-user-token",
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		sessionRepo.Create(context.Background(), session)

		// Create profile for DIFFERENT user
		otherUserProfile := &domain.Profile{
			ID:     otherUserProfileID,
			UserID: uuid.New(), // Different user!
			Type:   domain.ProfileTypeBrand,
			Name:   "Other User Brand",
		}
		profileRepo.Create(context.Background(), otherUserProfile)

		// Attempt to switch to other user's profile
		_, err := svc.SwitchActiveProfile(context.Background(), userID, sessionID, otherUserProfileID)
		assert.Error(t, err)
		assert.Equal(t, domain.ErrProfileNotOwned, err)
	})
}

// mockSessionRepoForProfileSwitch simulates the session repository for integration tests.
type mockSessionRepoForProfileSwitch struct {
	sessions map[uuid.UUID]*domain.Session
}

func (r *mockSessionRepoForProfileSwitch) Create(ctx context.Context, session *domain.Session) error {
	if r.sessions == nil {
		r.sessions = make(map[uuid.UUID]*domain.Session)
	}
	r.sessions[session.ID] = session
	return nil
}

func (r *mockSessionRepoForProfileSwitch) ByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	if s, ok := r.sessions[id]; ok {
		return s, nil
	}
	return nil, domain.ErrSessionNotFound
}

func (r *mockSessionRepoForProfileSwitch) ByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error) {
	for _, s := range r.sessions {
		if s.TokenHash == tokenHash {
			return s, nil
		}
	}
	return nil, domain.ErrSessionNotFound
}

func (r *mockSessionRepoForProfileSwitch) ByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Session, error) {
	var result []*domain.Session
	for _, s := range r.sessions {
		if s.UserID == userID {
			result = append(result, s)
		}
	}
	return result, nil
}

func (r *mockSessionRepoForProfileSwitch) Update(ctx context.Context, session *domain.Session) error {
	r.sessions[session.ID] = session
	return nil
}

func (r *mockSessionRepoForProfileSwitch) Delete(ctx context.Context, id uuid.UUID) error {
	delete(r.sessions, id)
	return nil
}

func (r *mockSessionRepoForProfileSwitch) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	for id, s := range r.sessions {
		if s.UserID == userID {
			delete(r.sessions, id)
		}
	}
	return nil
}
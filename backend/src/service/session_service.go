package service

import (
	"context"

	"github.com/google/uuid"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/repository"
)

// SessionService handles session-related business logic.
type SessionService struct {
	sessionRepo repository.SessionRepository
	profileRepo repository.ProfileRepository
}

// NewSessionService creates a new SessionService.
func NewSessionService(sessionRepo repository.SessionRepository, profileRepo repository.ProfileRepository) *SessionService {
	return &SessionService{
		sessionRepo: sessionRepo,
		profileRepo: profileRepo,
	}
}

// SwitchActiveProfile switches the active profile for a session.
func (s *SessionService) SwitchActiveProfile(ctx context.Context, userID, sessionID, profileID uuid.UUID) (*domain.Session, error) {
	// Get the session
	session, err := s.sessionRepo.ByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Verify session belongs to user
	if session.UserID != userID {
		return nil, domain.ErrSessionNotFound
	}

	// Verify profile belongs to user
	profile, err := s.profileRepo.ByID(ctx, profileID)
	if err != nil {
		return nil, err
	}
	if profile.UserID != userID {
		return nil, domain.ErrProfileNotOwned
	}

	// Set active profile
	session.SetActiveProfile(profileID)

	// Persist
	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

// ListSessions lists all sessions for a user.
func (s *SessionService) ListSessions(ctx context.Context, userID uuid.UUID) ([]*domain.Session, error) {
	sessions, err := s.sessionRepo.ByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Filter out expired sessions
	var active []*domain.Session
	for _, sess := range sessions {
		if !sess.IsExpired() {
			active = append(active, sess)
		}
	}
	return active, nil
}

// RevokeSession revokes a specific session.
func (s *SessionService) RevokeSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	session, err := s.sessionRepo.ByID(ctx, sessionID)
	if err != nil {
		return err
	}

	// Verify session belongs to user
	if session.UserID != userID {
		return domain.ErrSessionNotFound
	}

	return s.sessionRepo.Delete(ctx, sessionID)
}

// GetSessionWithProfile retrieves a session with its associated profile.
func (s *SessionService) GetSessionWithProfile(ctx context.Context, userID, sessionID uuid.UUID) (*domain.Session, *domain.Profile, error) {
	session, err := s.sessionRepo.ByID(ctx, sessionID)
	if err != nil {
		return nil, nil, err
	}

	if session.UserID != userID {
		return nil, nil, domain.ErrSessionNotFound
	}

	if session.ActiveProfileID == nil {
		return session, nil, nil
	}

	profile, err := s.profileRepo.ByID(ctx, *session.ActiveProfileID)
	if err != nil {
		return session, nil, err
	}

	return session, profile, nil
}

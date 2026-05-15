package domain

import (
	"time"

	"github.com/google/uuid"
)

// Session represents an authenticated user connection.
type Session struct {
	ID             uuid.UUID  `json:"id"`
	UserID         uuid.UUID  `json:"user_id"`
	ActiveProfileID *uuid.UUID `json:"active_profile_id,omitempty"`
	TokenHash      string     `json:"-"`
	UserAgent      string     `json:"user_agent,omitempty"`
	IPAddress      string     `json:"ip_address,omitempty"`
	ExpiresAt      time.Time  `json:"expires_at"`
	CreatedAt      time.Time  `json:"created_at"`
}

// NewSession creates a new session for a user.
func NewSession(userID uuid.UUID, tokenHash, userAgent, ipAddress string, expiresAt time.Time) *Session {
	return &Session{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		UserAgent: userAgent,
		IPAddress: ipAddress,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
}

// IsExpired returns true if the session has expired.
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// SetActiveProfile sets the active profile for this session.
func (s *Session) SetActiveProfile(profileID uuid.UUID) {
	s.ActiveProfileID = &profileID
}

// ClearActiveProfile clears the active profile.
func (s *Session) ClearActiveProfile() {
	s.ActiveProfileID = nil
}
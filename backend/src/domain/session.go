package domain

import (
	"time"

	"github.com/google/uuid"
)

// Session represents an authenticated user connection.
type Session struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID          uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"`
	ActiveProfileID *uuid.UUID `gorm:"type:uuid" json:"active_profile_id,omitempty"`
	TokenHash       string     `gorm:"type:varchar(255);index;not null" json:"-"`
	CSRFToken       string     `gorm:"type:varchar(255)" json:"csrf_token,omitempty"`
	UserAgent       string     `gorm:"type:varchar(512)" json:"user_agent,omitempty"`
	IPAddress       string     `gorm:"type:varchar(45)" json:"ip_address,omitempty"`
	ExpiresAt       time.Time  `gorm:"type:timestamptz;not null" json:"expires_at"`
	CreatedAt       time.Time  `gorm:"type:timestamptz;autoCreateTime" json:"created_at"`
}

// TableName sets the table name for Session.
func (Session) TableName() string { return "sessions" }

// SetCSRFToken sets the CSRF token for this session.
func (s *Session) SetCSRFToken(token string) {
	s.CSRFToken = token
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

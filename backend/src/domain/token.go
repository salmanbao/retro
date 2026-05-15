package domain

import (
	"time"

	"github.com/google/uuid"
)

// TokenType represents the type of one-time auth token.
type TokenType string

const (
	TokenTypeVerification  TokenType = "verification"
	TokenTypePasswordReset TokenType = "password_reset"
)

// AuthToken represents a one-time token for email verification or password reset.
type AuthToken struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	TokenType TokenType  `json:"token_type"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// NewAuthToken creates a new one-time auth token.
func NewAuthToken(userID uuid.UUID, tokenType TokenType, tokenHash string, expiresAt time.Time) *AuthToken {
	return &AuthToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenType: tokenType,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		UsedAt:    nil,
		CreatedAt: time.Now(),
	}
}

// IsExpired returns true if the token has expired.
func (t *AuthToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsUsed returns true if the token has been consumed.
func (t *AuthToken) IsUsed() bool {
	return t.UsedAt != nil
}

// MarkUsed marks the token as consumed.
func (t *AuthToken) MarkUsed() {
	now := time.Now()
	t.UsedAt = &now
}

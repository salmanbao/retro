package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents a registered user account.
type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Verified     bool      `json:"verified"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// NewUser creates a new unverified user with a hashed password.
func NewUser(email, passwordHash string) *User {
	now := time.Now()
	return &User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: passwordHash,
		Verified:     false,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// Verify marks the user as email-verified.
func (u *User) Verify() {
	u.Verified = true
	u.UpdatedAt = time.Now()
}
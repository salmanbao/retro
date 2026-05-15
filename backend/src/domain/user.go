package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents a registered user account.
type User struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email               string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash        string     `gorm:"type:varchar(255);not null" json:"-"`
	Verified            bool       `gorm:"default:false" json:"verified"`
	FailedLoginAttempts int        `gorm:"default:0" json:"failed_login_attempts"`
	LockedUntil         *time.Time `gorm:"type:timestamptz" json:"locked_until,omitempty"`
	CreatedAt           time.Time  `gorm:"type:timestamptz;autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time  `gorm:"type:timestamptz;autoUpdateTime" json:"updated_at"`
}

// TableName sets the table name for User.
func (User) TableName() string { return "users" }

// IsLocked returns true if the account is currently locked due to failed login attempts.
func (u *User) IsLocked() bool {
	if u.LockedUntil == nil {
		return false
	}
	return time.Now().Before(*u.LockedUntil)
}

// Lock locks the account until the specified time.
func (u *User) Lock(until time.Time) {
	u.LockedUntil = &until
	u.UpdatedAt = time.Now()
}

// ResetFailedAttempts resets the failed login attempt counter.
func (u *User) ResetFailedAttempts() {
	u.FailedLoginAttempts = 0
	u.LockedUntil = nil
	u.UpdatedAt = time.Now()
}

// IncrementFailedAttempts increments the failed login counter and locks if threshold reached.
func (u *User) IncrementFailedAttempts(lockDuration time.Duration) {
	u.FailedLoginAttempts++
	if u.FailedLoginAttempts >= 5 {
		u.Lock(time.Now().Add(lockDuration))
	} else {
		u.UpdatedAt = time.Now()
	}
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
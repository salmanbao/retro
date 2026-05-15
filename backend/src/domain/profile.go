package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ProfileType represents the type of role profile.
type ProfileType string

const (
	ProfileTypeBrand      ProfileType = "brand"
	ProfileTypeEditor     ProfileType = "editor"
	ProfileTypeInfluencer ProfileType = "influencer"
)

// Profile represents a user's persona within the marketplace.
type Profile struct {
	ID        uuid.UUID       `json:"id"`
	UserID    uuid.UUID       `json:"user_id"`
	Type      ProfileType     `json:"profile_type"`
	Name      string          `json:"name"`
	Details   json.RawMessage `json:"details"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt *time.Time      `json:"deleted_at,omitempty"`
}

// NewProfile creates a new profile for a user.
func NewProfile(userID uuid.UUID, profileType ProfileType, name string, details json.RawMessage) *Profile {
	now := time.Now()
	return &Profile{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      profileType,
		Name:      name,
		Details:   details,
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: nil,
	}
}

// IsDeleted returns true if the profile has been soft-deleted.
func (p *Profile) IsDeleted() bool {
	return p.DeletedAt != nil
}

// SoftDelete marks the profile as deleted.
func (p *Profile) SoftDelete() {
	now := time.Now()
	p.DeletedAt = &now
	p.UpdatedAt = now
}

// Update updates the profile name and details.
func (p *Profile) Update(name string, details json.RawMessage) {
	p.Name = name
	p.Details = details
	p.UpdatedAt = time.Now()
}

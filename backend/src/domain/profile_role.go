package domain

import (
	"time"

	"github.com/google/uuid"
)

// ProfileRole represents the many-to-many relationship between profiles and roles.
// A profile can have multiple roles, and a role can be assigned to multiple profiles.
// Role assignments include timestamps for audit purposes.
type ProfileRole struct {
	ProfileID uuid.UUID `gorm:"type:uuid;primaryKey" json:"profile_id"`
	RoleID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"role_id"`
	CreatedAt time.Time `gorm:"type:timestamptz;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamptz;autoUpdateTime" json:"updated_at"`
}

// TableName sets the table name for ProfileRole.
func (ProfileRole) TableName() string {
	return "profile_roles"
}

// NewProfileRole creates a new ProfileRole linking a profile to a role.
func NewProfileRole(profileID, roleID uuid.UUID) *ProfileRole {
	return &ProfileRole{
		ProfileID: profileID,
		RoleID:    roleID,
	}
}
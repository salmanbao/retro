package domain

import (
	"time"

	"github.com/google/uuid"
)

// Role represents a named collection of permissions with optional parent for inheritance.
// Child roles inherit all permissions from their parent role.
type Role struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string     `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	Description string     `gorm:"type:text;not null" json:"description"`
	ParentID    *uuid.UUID `gorm:"type:uuid;index" json:"parent_id,omitempty"`
	CreatedAt   time.Time  `gorm:"type:timestamptz;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"type:timestamptz;autoUpdateTime" json:"updated_at"`
}

// TableName sets the table name for Role.
func (Role) TableName() string {
	return "roles"
}

// NewRole creates a new Role with the given name, description, and optional parent ID.
func NewRole(name, description string, parentID *uuid.UUID) *Role {
	return &Role{
		Name:        name,
		Description: description,
		ParentID:    parentID,
	}
}

// HasParent returns true if the role has a parent role.
func (r *Role) HasParent() bool {
	return r.ParentID != nil
}

// IsRoot returns true if the role is a root role (no parent).
func (r *Role) IsRoot() bool {
	return r.ParentID == nil
}
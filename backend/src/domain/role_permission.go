package domain

import (
	"time"

	"github.com/google/uuid"
)

// RolePermission represents the many-to-many relationship between roles and permissions.
// A role has permissions, and a permission can be assigned to multiple roles.
type RolePermission struct {
	RoleID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"role_id"`
	PermissionKey  string    `gorm:"type:varchar(100);primaryKey" json:"permission_key"`
	CreatedAt      time.Time `gorm:"type:timestamptz;autoCreateTime" json:"created_at"`
}

// TableName sets the table name for RolePermission.
func (RolePermission) TableName() string {
	return "role_permissions"
}

// NewRolePermission creates a new RolePermission linking a role to a permission.
func NewRolePermission(roleID uuid.UUID, permissionKey string) *RolePermission {
	return &RolePermission{
		RoleID:        roleID,
		PermissionKey: permissionKey,
	}
}
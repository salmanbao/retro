package domain

import (
	"time"
)

// PermissionDomain represents the domain/category of a permission.
type PermissionDomain string

const (
	DomainBrand      PermissionDomain = "Brand"
	DomainEditor     PermissionDomain = "Editor"
	DomainInfluencer PermissionDomain = "Influencer"
	DomainPlatform   PermissionDomain = "Platform"
)

// Permission represents a system capability in dot-notation format.
// Permission keys must be unique and follow the format "resource.action".
type Permission struct {
	Key         string           `gorm:"primaryKey;type:varchar(100)" json:"key"`
	Description string           `gorm:"type:text;not null" json:"description"`
	Domain      PermissionDomain `gorm:"type:varchar(50);not null" json:"domain"`
	CreatedAt   time.Time        `gorm:"type:timestamptz;autoCreateTime" json:"created_at"`
}

// TableName sets the table name for Permission.
func (Permission) TableName() string {
	return "permissions"
}

// NewPermission creates a new Permission with the given key, description, and domain.
func NewPermission(key, description string, domain PermissionDomain) *Permission {
	return &Permission{
		Key:         key,
		Description: description,
		Domain:      domain,
	}
}

// IsValid checks if the permission key follows dot-notation format.
func (p *Permission) IsValid() bool {
	// Format: resource.action (lowercase letters only)
	if len(p.Key) < 3 || len(p.Key) > 100 {
		return false
	}
	for i := 0; i < len(p.Key); i++ {
		c := p.Key[i]
		if c == '.' {
			continue
		}
		if c < 'a' || c > 'z' {
			return false
		}
	}
	// Must contain exactly one dot
	dotCount := 0
	for _, c := range p.Key {
		if c == '.' {
			dotCount++
		}
	}
	return dotCount == 1
}

// IsWildcard returns true if the permission key ends with ".*" (wildcard permission).
func (p *Permission) IsWildcard() bool {
	return len(p.Key) >= 2 && p.Key[len(p.Key)-2:] == ".*"
}

// Matches checks if this wildcard permission matches the given permission key.
// Only valid for wildcard permissions.
func (p *Permission) Matches(key string) bool {
	if !p.IsWildcard() {
		return p.Key == key
	}
	// "campaign.*" matches "campaign.create", "campaign.update", etc.
	prefix := p.Key[:len(p.Key)-2] // Remove ".*"
	return len(key) > len(prefix) && key[:len(prefix)] == prefix && key[len(prefix)] == '.'
}

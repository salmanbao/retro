package unit

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"viralforge/backend/src/domain"
)

// TestT057_PermissionEntityValidation tests permission entity validation for key format and domain.
// Note: The IsValid() method enforces dot-notation format (exactly one dot), so wildcards like
// "campaign.*" are technically invalid per this strict validation. The service layer handles
// wildcards as a special case, but the domain entity enforces the standard format.
func TestT057_PermissionEntityValidation(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		description string
		domain      domain.PermissionDomain
		wantValid   bool
	}{
		{
			name:        "valid permission with dot notation",
			key:         "campaign.create",
			description: "Create campaigns",
			domain:      domain.DomainBrand,
			wantValid:   true,
		},
		{
			name:        "wildcard campaign.* is invalid per IsValid (enforces dot-notation)",
			key:         "campaign.*",
			description: "All campaign permissions",
			domain:      domain.DomainBrand,
			wantValid:   false,
		},
		{
			name:        "star is invalid per IsValid (enforces dot-notation)",
			key:         "*",
			description: "All permissions",
			domain:      domain.DomainPlatform,
			wantValid:   false,
		},
		{
			name:        "invalid - too short",
			key:         "ab",
			description: "Too short",
			domain:      domain.DomainBrand,
			wantValid:   false,
		},
		{
			name:        "invalid - no dot",
			key:         "campaigncreate",
			description: "Missing dot",
			domain:      domain.DomainBrand,
			wantValid:   false,
		},
		{
			name:        "invalid - multiple dots",
			key:         "campaign.create.edit",
			description: "Too many dots",
			domain:      domain.DomainBrand,
			wantValid:   false,
		},
		{
			name:        "invalid - uppercase",
			key:         "Campaign.Create",
			description: "Uppercase letters",
			domain:      domain.DomainBrand,
			wantValid:   false,
		},
		{
			name:        "invalid - numbers",
			key:         "campaign.123",
			description: "Numbers not allowed",
			domain:      domain.DomainBrand,
			wantValid:   false,
		},
		{
			name:        "invalid - starts with dot",
			key:         ".campaign.create",
			description: "Starts with dot",
			domain:      domain.DomainBrand,
			wantValid:   false,
		},
		{
			name:        "invalid - ends with dot",
			key:         "campaign.create.",
			description: "Ends with dot",
			domain:      domain.DomainBrand,
			wantValid:   false,
		},
		{
			name:        "invalid - special chars",
			key:         "campaign@create",
			description: "Special characters",
			domain:      domain.DomainBrand,
			wantValid:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perm := domain.NewPermission(tt.key, tt.description, tt.domain)
			assert.Equal(t, tt.wantValid, perm.IsValid(), "IsValid() for key %s", tt.key)
		})
	}
}

// TestT057_PermissionWildcard tests wildcard permission detection.
func TestT057_PermissionWildcard(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		wantWild bool
	}{
		{"wildcard campaign", "campaign.*", true},
		{"wildcard analytics", "analytics.*", true},
		{"concrete campaign.create", "campaign.create", false},
		{"concrete submission.review", "submission.review", false},
		{"star is NOT detected as wildcard by IsWildcard (special case)", "*", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perm := domain.NewPermission(tt.key, "test", domain.DomainBrand)
			assert.Equal(t, tt.wantWild, perm.IsWildcard(), "IsWildcard() for key %s", tt.key)
		})
	}
}

// TestT057_PermissionMatches tests wildcard matching logic.
// Note: The Matches() method handles wildcards ending with ".*". The "*" permission
// is a special case handled at the service layer, not through the domain Matches method.
func TestT057_PermissionMatches(t *testing.T) {
	tests := []struct {
		name      string
		wildcard  string
		concrete  string
		wantMatch bool
	}{
		{"campaign wildcard matches create", "campaign.*", "campaign.create", true},
		{"campaign wildcard matches update", "campaign.*", "campaign.update", true},
		{"campaign wildcard matches delete", "campaign.*", "campaign.delete", true},
		{"campaign wildcard does not match submission", "campaign.*", "submission.review", false},
		{"campaign wildcard does not match analytics", "campaign.*", "analytics.view", false},
		{"star is special case not handled by Matches", "*", "anything.at.all", false},
		{"concrete exact match", "campaign.create", "campaign.create", true},
		{"concrete no match", "campaign.create", "campaign.update", false},
		{"nested wildcard matches sub", "campaign.*", "campaign.sub.nested", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perm := domain.NewPermission(tt.wildcard, "test", domain.DomainBrand)
			assert.Equal(t, tt.wantMatch, perm.Matches(tt.concrete), "Matches(%s) for wildcard %s", tt.concrete, tt.wildcard)
		})
	}
}

// TestT058_RoleEntityValidation tests role entity validation for name uniqueness and parent validation.
func TestT058_RoleEntityValidation(t *testing.T) {
	t.Run("HasParent returns correct value", func(t *testing.T) {
		roleWithParent := &domain.Role{
			ID:       uuid.New(),
			Name:     "child_role",
			ParentID: func() *uuid.UUID { id := uuid.New(); return &id }(),
		}
		assert.True(t, roleWithParent.HasParent())

		roleWithoutParent := &domain.Role{
			ID:       uuid.New(),
			Name:     "root_role",
			ParentID: nil,
		}
		assert.False(t, roleWithoutParent.HasParent())
	})

	t.Run("IsRoot returns correct value", func(t *testing.T) {
		rootRole := &domain.Role{
			ID:       uuid.New(),
			Name:     "root_role",
			ParentID: nil,
		}
		assert.True(t, rootRole.IsRoot())

		childRole := &domain.Role{
			ID:       uuid.New(),
			Name:     "child_role",
			ParentID: func() *uuid.UUID { id := uuid.New(); return &id }(),
		}
		assert.False(t, childRole.IsRoot())
	})

	t.Run("NewRole creates role with correct fields", func(t *testing.T) {
		name := "test_role"
		description := "Test role description"
		parentID := uuid.New()

		role := domain.NewRole(name, description, &parentID)

		assert.Equal(t, name, role.Name)
		assert.Equal(t, description, role.Description)
		require.NotNil(t, role.ParentID)
		assert.Equal(t, parentID, *role.ParentID)
	})

	t.Run("NewRole with nil parent creates root role", func(t *testing.T) {
		role := domain.NewRole("root_role", "Root role", nil)
		assert.Nil(t, role.ParentID)
		assert.True(t, role.IsRoot())
		assert.False(t, role.HasParent())
	})
}

// TestT058_RoleParentValidation tests parent validation for roles.
func TestT058_RoleParentValidation(t *testing.T) {
	t.Run("circular parent detection is not role's responsibility", func(t *testing.T) {
		// The circular detection is done at the service level, not the domain level
		// The role entity itself doesn't validate circular references - that's
		// handled by the AuthorizationService when creating/updating roles
		parentID := uuid.New()
		childID := uuid.New()

		parent := domain.NewRole("parent_role", "Parent", nil)
		parent.ID = parentID

		child := domain.NewRole("child_role", "Child", &parentID)
		child.ID = childID

		// Both should be valid entities individually
		assert.NotNil(t, child.ParentID)
		assert.Equal(t, parentID, *child.ParentID)
		assert.True(t, child.HasParent())
		assert.True(t, parent.IsRoot())
	})

	t.Run("parent ID can be nil", func(t *testing.T) {
		role := domain.NewRole("orphan_role", "No parent", nil)
		assert.Nil(t, role.ParentID)
		assert.False(t, role.HasParent())
		assert.True(t, role.IsRoot())
	})
}

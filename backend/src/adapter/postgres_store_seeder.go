package adapter

import (
	"context"

	"github.com/google/uuid"
	"viralforge/backend/src/domain"
)

// SeedData seeds the database with initial permissions and roles.
func (s *PostgresStore) SeedData(ctx context.Context) error {
	// Seed permissions
	if err := s.seedPermissions(ctx); err != nil {
		return err
	}

	// Seed roles
	if err := s.seedRoles(ctx); err != nil {
		return err
	}

	// Seed role-permission mappings
	if err := s.seedRolePermissions(ctx); err != nil {
		return err
	}

	return nil
}

func (s *PostgresStore) seedPermissions(ctx context.Context) error {
	permissions := []struct {
		key         string
		description string
		domain      domain.PermissionDomain
	}{
		// Brand permissions
		{"campaign.create", "Create new campaigns", domain.DomainBrand},
		{"campaign.update", "Modify existing campaigns", domain.DomainBrand},
		{"campaign.delete", "Delete campaigns", domain.DomainBrand},
		{"campaign.view", "View campaign details", domain.DomainBrand},
		// Editor permissions
		{"submission.review", "Review submissions", domain.DomainEditor},
		{"submission.approve", "Approve submissions", domain.DomainEditor},
		{"submission.reject", "Reject submissions", domain.DomainEditor},
		// Influencer permissions
		{"offer.claim", "Claim offers", domain.DomainInfluencer},
		{"offer.view", "View available offers", domain.DomainInfluencer},
		{"wallet.view", "View wallet balance", domain.DomainInfluencer},
		{"analytics.view", "View performance analytics", domain.DomainInfluencer},
		// Cross-domain permissions
		{"profile.manage", "Manage profile settings", domain.DomainBrand},
		// Platform permissions
		{"role.assign", "Assign roles to profiles", domain.DomainPlatform},
		{"role.revoke", "Remove roles from profiles", domain.DomainPlatform},
		{"role.create", "Create new roles", domain.DomainPlatform},
		{"role.delete", "Delete roles", domain.DomainPlatform},
		{"permission.assign", "Assign permissions to roles", domain.DomainPlatform},
		{"permission.revoke", "Remove permissions from roles", domain.DomainPlatform},
		// Wildcard permissions
		{"campaign.*", "All campaign permissions", domain.DomainBrand},
		{"offer.*", "All offer permissions", domain.DomainInfluencer},
		{"submission.*", "All submission permissions", domain.DomainEditor},
	}

	for _, p := range permissions {
		perm := domain.NewPermission(p.key, p.description, p.domain)
		// Use upsert logic - skip if exists
		existing, _ := s.PermissionRepository().ByKey(ctx, p.key)
		if existing != nil {
			continue
		}
		if err := s.PermissionRepository().Create(ctx, perm); err != nil {
			return err
		}
	}

	return nil
}

func (s *PostgresStore) seedRoles(ctx context.Context) error {
	roles := []struct {
		name        string
		description string
		parentName  *string
	}{
		// Brand roles
		{"brand_owner", "Full Brand access", nil},
		{"brand_admin", "Brand administration", strPtr("brand_owner")},
		{"brand_marketer", "Brand marketing", strPtr("brand_admin")},
		// Editor roles
		{"editor_owner", "Full Editor access", nil},
		{"editor_senior", "Senior Editor", strPtr("editor_owner")},
		{"editor_junior", "Junior Editor", strPtr("editor_senior")},
		// Influencer roles
		{"influencer_owner", "Full Influencer access", nil},
		{"influencer_partner", "Influencer partner", strPtr("influencer_owner")},
		// Platform role
		{"platform_admin", "Platform administration", nil},
	}

	createdRoles := make(map[string]*domain.Role)

	for _, r := range roles {
		existing, _ := s.RoleRepository().ByName(ctx, r.name)
		if existing != nil {
			createdRoles[r.name] = existing
			continue
		}

		role := &domain.Role{
			ID:          uuid.New(),
			Name:        r.name,
			Description: r.description,
		}

		if r.parentName != nil {
			parent, ok := createdRoles[*r.parentName]
			if !ok {
				parent, _ = s.RoleRepository().ByName(ctx, *r.parentName)
			}
			if parent != nil {
				role.ParentID = &parent.ID
			}
		}

		if err := s.RoleRepository().Create(ctx, role); err != nil {
			return err
		}
		createdRoles[r.name] = role
	}

	return nil
}

func (s *PostgresStore) seedRolePermissions(ctx context.Context) error {
	rolePermissions := map[string][]string{
		"brand_owner": {
			"campaign.*", "analytics.view", "offer.view", "profile.manage",
		},
		"brand_admin": {
			"campaign.create", "campaign.update", "analytics.view", "profile.manage",
		},
		"brand_marketer": {
			"campaign.create", "campaign.view", "profile.manage",
		},
		"editor_owner": {
			"submission.*", "profile.manage",
		},
		"editor_senior": {
			"submission.review", "submission.approve", "profile.manage",
		},
		"editor_junior": {
			"submission.review", "profile.manage",
		},
		"influencer_owner": {
			"offer.*", "wallet.view", "analytics.view", "profile.manage",
		},
		"influencer_partner": {
			"offer.claim", "offer.view", "profile.manage",
		},
		"platform_admin": {
			"*",
		},
	}

	for roleName, perms := range rolePermissions {
		role, err := s.RoleRepository().ByName(ctx, roleName)
		if err != nil {
			continue
		}

		for _, permKey := range perms {
			perm, err := s.PermissionRepository().ByKey(ctx, permKey)
			if err != nil {
				continue
			}

			rp := &domain.RolePermission{
				RoleID:         role.ID,
				PermissionKey:  perm.Key,
			}

			// Check if already exists
			existing, _ := s.rolePermissionRepoByRoleAndPerm(ctx, role.ID, perm.Key)
			if existing != nil {
				continue
			}

			if err := s.RolePermissionRepository().Create(ctx, rp); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *PostgresStore) rolePermissionRepoByRoleAndPerm(ctx context.Context, roleID uuid.UUID, permKey string) (*domain.RolePermission, error) {
	var rp domain.RolePermission
	if err := s.db.WithContext(ctx).Where("role_id = ? AND permission_key = ?", roleID, permKey).First(&rp).Error; err != nil {
		return nil, err
	}
	return &rp, nil
}

func strPtr(s string) *string {
	return &s
}
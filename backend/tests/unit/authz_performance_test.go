package unit

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/service"
)

// BenchmarkT059_HasPermissionResolution measures hasPermission resolution time
// for a profile with 5 roles (3-level hierarchy). Target: < 50ms
func BenchmarkT059_HasPermissionResolution(b *testing.B) {
	ctx := context.Background()

	// Create 5 roles in a hierarchy: role0 -> role1 -> role2 -> role3 -> role4
	roleIDs := make([]uuid.UUID, 5)
	for i := 0; i < 5; i++ {
		roleIDs[i] = uuid.New()
	}

	permRepo := &mockPermissionRepository{permissions: make(map[string]*domain.Permission)}
	roleRepo := &mockRoleRepository{roles: make(map[uuid.UUID]*domain.Role)}
	rolePermRepo := &mockRolePermissionRepository{rolePerms: make(map[uuid.UUID][]*domain.RolePermission)}
	profileRoleRepo := &mockProfileRoleRepository{profileRoles: make(map[uuid.UUID][]*domain.ProfileRole)}

	// Create role hierarchy and permissions
	for i := 0; i < 5; i++ {
		var parentID *uuid.UUID
		if i > 0 {
			parentID = &roleIDs[i-1]
		}
		role := &domain.Role{ID: roleIDs[i], Name: "role", Description: "test role", ParentID: parentID}
		require.NoError(b, roleRepo.Create(ctx, role))
		// Give each role a unique permission
		require.NoError(b, rolePermRepo.Create(ctx, &domain.RolePermission{
			RoleID:        roleIDs[i],
			PermissionKey: "permission." + string(rune('a'+i)),
		}))
	}

	// Assign role4 (deepest) to profile - inherits all 5 permissions
	profileID := uuid.New()
	profileRoleRepo.profileRoles[profileID] = []*domain.ProfileRole{{ProfileID: profileID, RoleID: roleIDs[4]}}

	svc := service.NewAuthorizationService(permRepo, roleRepo, rolePermRepo, profileRoleRepo, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.HasPermission(ctx, profileID, "permission.e")
	}
}

// BenchmarkT060_MiddlewareLatency measures requirePermission middleware latency.
// Target: < 10ms per request.
func BenchmarkT060_MiddlewareLatency(b *testing.B) {
	ctx := context.Background()

	profileID := uuid.New()
	roleID := uuid.New()

	permRepo := &mockPermissionRepository{permissions: make(map[string]*domain.Permission)}
	roleRepo := &mockRoleRepository{roles: make(map[uuid.UUID]*domain.Role)}
	rolePermRepo := &mockRolePermissionRepository{rolePerms: make(map[uuid.UUID][]*domain.RolePermission)}
	profileRoleRepo := &mockProfileRoleRepository{profileRoles: make(map[uuid.UUID][]*domain.ProfileRole)}

	// Create role with campaign.create permission
	role := &domain.Role{ID: roleID, Name: "brand_admin", Description: "Brand admin"}
	require.NoError(b, roleRepo.Create(ctx, role))
	require.NoError(b, rolePermRepo.Create(ctx, &domain.RolePermission{
		RoleID:        roleID,
		PermissionKey: "campaign.create",
	}))
	profileRoleRepo.profileRoles[profileID] = []*domain.ProfileRole{{ProfileID: profileID, RoleID: roleID}}

	svc := service.NewAuthorizationService(permRepo, roleRepo, rolePermRepo, profileRoleRepo, nil)

	// Measure HasPermission call latency (middleware uses this internally)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.HasPermission(ctx, profileID, "campaign.create")
	}
}

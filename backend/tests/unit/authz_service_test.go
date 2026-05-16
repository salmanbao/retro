package unit

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/service"
)

// mockPermissionRepository is a mock implementation of PermissionRepository.
type mockPermissionRepository struct {
	permissions map[string]*domain.Permission
}

func (m *mockPermissionRepository) Create(ctx context.Context, permission *domain.Permission) error {
	m.permissions[permission.Key] = permission
	return nil
}
func (m *mockPermissionRepository) ByKey(ctx context.Context, key string) (*domain.Permission, error) {
	if p, ok := m.permissions[key]; ok {
		return p, nil
	}
	return nil, domain.ErrPermissionNotFound
}
func (m *mockPermissionRepository) ListAll(ctx context.Context) ([]*domain.Permission, error) {
	perms := make([]*domain.Permission, 0, len(m.permissions))
	for _, p := range m.permissions {
		perms = append(perms, p)
	}
	return perms, nil
}

// mockRoleRepository is a mock implementation of RoleRepository.
type mockRoleRepository struct {
	roles map[uuid.UUID]*domain.Role
}

func (m *mockRoleRepository) Create(ctx context.Context, role *domain.Role) error {
	m.roles[role.ID] = role
	return nil
}
func (m *mockRoleRepository) ByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	if r, ok := m.roles[id]; ok {
		return r, nil
	}
	return nil, domain.ErrRoleNotFound
}
func (m *mockRoleRepository) ByName(ctx context.Context, name string) (*domain.Role, error) {
	for _, r := range m.roles {
		if r.Name == name {
			return r, nil
		}
	}
	return nil, domain.ErrRoleNotFound
}
func (m *mockRoleRepository) ListAll(ctx context.Context) ([]*domain.Role, error) {
	roles := make([]*domain.Role, 0, len(m.roles))
	for _, r := range m.roles {
		roles = append(roles, r)
	}
	return roles, nil
}
func (m *mockRoleRepository) Update(ctx context.Context, role *domain.Role) error {
	m.roles[role.ID] = role
	return nil
}
func (m *mockRoleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.roles, id)
	return nil
}

// mockRolePermissionRepository is a mock implementation of RolePermissionRepository.
type mockRolePermissionRepository struct {
	rolePerms map[uuid.UUID][]*domain.RolePermission
}

func (m *mockRolePermissionRepository) Create(ctx context.Context, rp *domain.RolePermission) error {
	m.rolePerms[rp.RoleID] = append(m.rolePerms[rp.RoleID], rp)
	return nil
}
func (m *mockRolePermissionRepository) ByRoleID(ctx context.Context, roleID uuid.UUID) ([]*domain.RolePermission, error) {
	return m.rolePerms[roleID], nil
}
func (m *mockRolePermissionRepository) ByPermissionKey(ctx context.Context, permissionKey string) ([]*domain.RolePermission, error) {
	var result []*domain.RolePermission
	for _, rps := range m.rolePerms {
		for _, rp := range rps {
			if rp.PermissionKey == permissionKey {
				result = append(result, rp)
			}
		}
	}
	return result, nil
}
func (m *mockRolePermissionRepository) Delete(ctx context.Context, roleID uuid.UUID, permissionKey string) error {
	rps := m.rolePerms[roleID]
	for i, rp := range rps {
		if rp.PermissionKey == permissionKey {
			m.rolePerms[roleID] = append(rps[:i], rps[i+1:]...)
			break
		}
	}
	return nil
}
func (m *mockRolePermissionRepository) DeleteByRoleID(ctx context.Context, roleID uuid.UUID) error {
	delete(m.rolePerms, roleID)
	return nil
}

// mockProfileRoleRepository is a mock implementation of ProfileRoleRepository.
type mockProfileRoleRepository struct {
	profileRoles map[uuid.UUID][]*domain.ProfileRole
}

func (m *mockProfileRoleRepository) Create(ctx context.Context, pr *domain.ProfileRole) error {
	m.profileRoles[pr.ProfileID] = append(m.profileRoles[pr.ProfileID], pr)
	return nil
}
func (m *mockProfileRoleRepository) ByProfileID(ctx context.Context, profileID uuid.UUID) ([]*domain.ProfileRole, error) {
	return m.profileRoles[profileID], nil
}
func (m *mockProfileRoleRepository) ByRoleID(ctx context.Context, roleID uuid.UUID) ([]*domain.ProfileRole, error) {
	var result []*domain.ProfileRole
	for _, prs := range m.profileRoles {
		for _, pr := range prs {
			if pr.RoleID == roleID {
				result = append(result, pr)
			}
		}
	}
	return result, nil
}
func (m *mockProfileRoleRepository) Delete(ctx context.Context, profileID uuid.UUID, roleID uuid.UUID) error {
	prs := m.profileRoles[profileID]
	for i, pr := range prs {
		if pr.RoleID == roleID {
			m.profileRoles[profileID] = append(prs[:i], prs[i+1:]...)
			break
		}
	}
	return nil
}
func (m *mockProfileRoleRepository) CountByProfileID(ctx context.Context, profileID uuid.UUID) (int64, error) {
	return int64(len(m.profileRoles[profileID])), nil
}
func (m *mockProfileRoleRepository) DeleteByRoleID(ctx context.Context, roleID uuid.UUID) error {
	for profileID, prs := range m.profileRoles {
		for i, pr := range prs {
			if pr.RoleID == roleID {
				m.profileRoles[profileID] = append(prs[:i], prs[i+1:]...)
			}
		}
	}
	return nil
}

func setupAuthzService(t *testing.T) (*service.AuthorizationService, *mockPermissionRepository, *mockRoleRepository, *mockRolePermissionRepository, *mockProfileRoleRepository) {
	permRepo := &mockPermissionRepository{permissions: make(map[string]*domain.Permission)}
	roleRepo := &mockRoleRepository{roles: make(map[uuid.UUID]*domain.Role)}
	rolePermRepo := &mockRolePermissionRepository{rolePerms: make(map[uuid.UUID][]*domain.RolePermission)}
	profileRoleRepo := &mockProfileRoleRepository{profileRoles: make(map[uuid.UUID][]*domain.ProfileRole)}

	svc := service.NewAuthorizationService(permRepo, roleRepo, rolePermRepo, profileRoleRepo, nil)
	return svc, permRepo, roleRepo, rolePermRepo, profileRoleRepo
}

// TestT012_WildcardPermissionMatching tests wildcard permission matching.
func TestT012_WildcardPermissionMatching(t *testing.T) {
	svc, permRepo, _, _, _ := setupAuthzService(t)
	ctx := context.Background()

	// Create wildcard permission
	wildcardPerm := domain.NewPermission("campaign.*", "All campaign permissions", domain.DomainBrand)
	require.NoError(t, permRepo.Create(ctx, wildcardPerm))

	// Create concrete permissions
	concretePerm := domain.NewPermission("campaign.create", "Create campaigns", domain.DomainBrand)
	require.NoError(t, permRepo.Create(ctx, concretePerm))

	// Test wildcard matching
	assert.True(t, svc.PermissionMatchesWildcard("campaign.*", "campaign.create"))
	assert.True(t, svc.PermissionMatchesWildcard("campaign.*", "campaign.update"))
	assert.True(t, svc.PermissionMatchesWildcard("campaign.*", "campaign.delete"))
	assert.False(t, svc.PermissionMatchesWildcard("campaign.*", "submission.review"))
	assert.False(t, svc.PermissionMatchesWildcard("campaign.*", "analytics.view"))

	// Test non-wildcard returns false
	assert.False(t, svc.PermissionMatchesWildcard("campaign.create", "campaign.update"))
}

// TestT013_RoleHierarchyTraversal tests role hierarchy traversal.
func TestT013_RoleHierarchyTraversal(t *testing.T) {
	svc, permRepo, roleRepo, rolePermRepo, profileRoleRepo := setupAuthzService(t)
	ctx := context.Background()

	// Create brand_owner role
	brandOwnerID := uuid.New()
	brandOwner := &domain.Role{
		ID:          brandOwnerID,
		Name:        "brand_owner",
		Description: "Full Brand access",
	}
	require.NoError(t, roleRepo.Create(ctx, brandOwner))

	// Create brand_admin role (child of brand_owner)
	brandAdminID := uuid.New()
	brandAdmin := &domain.Role{
		ID:          brandAdminID,
		Name:        "brand_admin",
		Description: "Brand administration",
		ParentID:    &brandOwnerID,
	}
	require.NoError(t, roleRepo.Create(ctx, brandAdmin))

	// Create brand_marketer role (child of brand_admin)
	brandMarketerID := uuid.New()
	brandMarketer := &domain.Role{
		ID:          brandMarketerID,
		Name:        "brand_marketer",
		Description: "Brand marketing",
		ParentID:    &brandAdminID,
	}
	require.NoError(t, roleRepo.Create(ctx, brandMarketer))

	// Assign permissions to brand_owner
	ownerPerm := domain.NewPermission("campaign.*", "All campaign permissions", domain.DomainBrand)
	require.NoError(t, permRepo.Create(ctx, ownerPerm))
	require.NoError(t, rolePermRepo.Create(ctx, &domain.RolePermission{
		RoleID:        brandOwnerID,
		PermissionKey: "campaign.*",
	}))

	// Assign direct permission to brand_admin
	adminPerm := domain.NewPermission("campaign.create", "Create campaigns", domain.DomainBrand)
	require.NoError(t, permRepo.Create(ctx, adminPerm))
	require.NoError(t, rolePermRepo.Create(ctx, &domain.RolePermission{
		RoleID:        brandAdminID,
		PermissionKey: "campaign.create",
	}))

	// Create profile and assign brand_marketer role
	profileID := uuid.New()
	profileRole := domain.NewProfileRole(profileID, brandMarketerID)
	require.NoError(t, profileRoleRepo.Create(ctx, profileRole))

	// Test HasPermission with brand_marketer (should have both direct and inherited)
	hasCampaignCreate, err := svc.HasPermission(ctx, profileID, "campaign.create")
	require.NoError(t, err)
	assert.True(t, hasCampaignCreate)

	// brand_marketer should also inherit from brand_admin
	hasCampaignUpdate, err := svc.HasPermission(ctx, profileID, "campaign.update")
	require.NoError(t, err)
	assert.True(t, hasCampaignUpdate) // Inherited via campaign.* from brand_owner

	// Test GetEffectivePermissions
	perms, err := svc.GetEffectivePermissions(ctx, profileID)
	require.NoError(t, err)
	assert.Contains(t, perms, "campaign.create")
	assert.Contains(t, perms, "campaign.*")
}

// TestT014_EffectivePermissionsWithWildcard tests that wildcard permissions match concrete permissions.
func TestT014_EffectivePermissionsWithWildcard(t *testing.T) {
	svc, permRepo, roleRepo, rolePermRepo, profileRoleRepo := setupAuthzService(t)
	ctx := context.Background()

	// Create platform_admin role with wildcard
	platformAdminID := uuid.New()
	platformAdmin := &domain.Role{
		ID:          platformAdminID,
		Name:        "platform_admin",
		Description: "Platform administration",
	}
	require.NoError(t, roleRepo.Create(ctx, platformAdmin))

	// Assign * permission to platform_admin
	wildcardPerm := domain.NewPermission("*", "All permissions", domain.DomainPlatform)
	require.NoError(t, permRepo.Create(ctx, wildcardPerm))
	require.NoError(t, rolePermRepo.Create(ctx, &domain.RolePermission{
		RoleID:        platformAdminID,
		PermissionKey: "*",
	}))

	// Create profile and assign platform_admin role
	profileID := uuid.New()
	profileRole := domain.NewProfileRole(profileID, platformAdminID)
	require.NoError(t, profileRoleRepo.Create(ctx, profileRole))

	// platform_admin with * should have all permissions
	hasCampaignCreate, err := svc.HasPermission(ctx, profileID, "campaign.create")
	require.NoError(t, err)
	assert.True(t, hasCampaignCreate)

	hasSubmissionReview, err := svc.HasPermission(ctx, profileID, "submission.review")
	require.NoError(t, err)
	assert.True(t, hasSubmissionReview)

	hasOfferClaim, err := svc.HasPermission(ctx, profileID, "offer.claim")
	require.NoError(t, err)
	assert.True(t, hasOfferClaim)
}

// TestT015_NonExistentPermission tests that non-existent permissions return false.
func TestT015_NonExistentPermission(t *testing.T) {
	svc, _, _, _, profileRoleRepo := setupAuthzService(t)
	ctx := context.Background()

	// Create profile with some role
	profileID := uuid.New()
	roleID := uuid.New()
	profileRole := domain.NewProfileRole(profileID, roleID)
	require.NoError(t, profileRoleRepo.Create(ctx, profileRole))

	// Non-existent permission should return false (FR-AUTH-013)
	hasFakePermission, err := svc.HasPermission(ctx, profileID, "fake.permission")
	require.NoError(t, err)
	assert.False(t, hasFakePermission)
}

// TestT041_ThreeLevelInheritance tests 3-level role hierarchy resolution.
func TestT041_ThreeLevelInheritance(t *testing.T) {
	svc, permRepo, roleRepo, rolePermRepo, profileRoleRepo := setupAuthzService(t)
	ctx := context.Background()

	// Create brand_owner (level 0)
	brandOwnerID := uuid.New()
	brandOwner := &domain.Role{
		ID:          brandOwnerID,
		Name:        "brand_owner",
		Description: "Full Brand access",
	}
	require.NoError(t, roleRepo.Create(ctx, brandOwner))

	// Assign permission to brand_owner
	ownerPerm := domain.NewPermission("campaign.*", "All campaign permissions", domain.DomainBrand)
	require.NoError(t, permRepo.Create(ctx, ownerPerm))
	require.NoError(t, rolePermRepo.Create(ctx, &domain.RolePermission{
		RoleID:        brandOwnerID,
		PermissionKey: "campaign.*",
	}))

	// Create brand_admin (child of brand_owner, level 1)
	brandAdminID := uuid.New()
	brandAdmin := &domain.Role{
		ID:          brandAdminID,
		Name:        "brand_admin",
		Description: "Brand administration",
		ParentID:    &brandOwnerID,
	}
	require.NoError(t, roleRepo.Create(ctx, brandAdmin))

	// Assign direct permission to brand_admin
	adminPerm := domain.NewPermission("campaign.create", "Create campaigns", domain.DomainBrand)
	require.NoError(t, permRepo.Create(ctx, adminPerm))
	require.NoError(t, rolePermRepo.Create(ctx, &domain.RolePermission{
		RoleID:        brandAdminID,
		PermissionKey: "campaign.create",
	}))

	// Create brand_marketer (child of brand_admin, level 2)
	brandMarketerID := uuid.New()
	brandMarketer := &domain.Role{
		ID:          brandMarketerID,
		Name:        "brand_marketer",
		Description: "Brand marketing",
		ParentID:    &brandAdminID,
	}
	require.NoError(t, roleRepo.Create(ctx, brandMarketer))

	// Create profile and assign brand_marketer
	profileID := uuid.New()
	profileRole := domain.NewProfileRole(profileID, brandMarketerID)
	require.NoError(t, profileRoleRepo.Create(ctx, profileRole))

	// brand_marketer should have:
	// - campaign.create (direct)
	// - campaign.* (inherited from brand_admin which inherits from brand_owner)
	hasCampaignCreate, err := svc.HasPermission(ctx, profileID, "campaign.create")
	require.NoError(t, err)
	assert.True(t, hasCampaignCreate)

	hasCampaignUpdate, err := svc.HasPermission(ctx, profileID, "campaign.update")
	require.NoError(t, err)
	assert.True(t, hasCampaignUpdate) // inherited via campaign.* from brand_owner

	// Get inherited roles
	inherited, err := svc.GetInheritedRoleIDs(ctx, brandMarketerID)
	require.NoError(t, err)
	assert.Len(t, inherited, 2)
	assert.Contains(t, inherited, brandAdminID)
	assert.Contains(t, inherited, brandOwnerID)
}

// TestT042_CircularInheritancePrevention tests that circular inheritance is rejected.
func TestT042_CircularInheritancePrevention(t *testing.T) {
	svc, _, roleRepo, _, _ := setupAuthzService(t)
	ctx := context.Background()

	// Create role_a
	roleAID := uuid.New()
	roleA := &domain.Role{
		ID:          roleAID,
		Name:        "role_a",
		Description: "Role A",
	}
	require.NoError(t, roleRepo.Create(ctx, roleA))

	// Create role_b with parent role_a
	roleBID := uuid.New()
	roleB := &domain.Role{
		ID:          roleBID,
		Name:        "role_b",
		Description: "Role B",
		ParentID:    &roleAID,
	}
	require.NoError(t, roleRepo.Create(ctx, roleB))

	// Try to update role_a to have role_b as parent (would create cycle: role_a -> role_b -> role_a)
	_, err := svc.UpdateRole(ctx, roleAID, "role_a", "Role A", &roleBID)
	require.Error(t, err)
	assert.Equal(t, domain.ErrCircularInheritance, err)
}

// TestT043_CascadeDelete tests that deleting a role cascades to remove assignments.
func TestT043_CascadeDelete(t *testing.T) {
	svc, permRepo, roleRepo, rolePermRepo, profileRoleRepo := setupAuthzService(t)
	ctx := context.Background()

	// Create role with permission
	roleID := uuid.New()
	role := &domain.Role{
		ID:          roleID,
		Name:        "test_role",
		Description: "Test role",
	}
	require.NoError(t, roleRepo.Create(ctx, role))

	perm := domain.NewPermission("test.permission", "Test permission", domain.DomainBrand)
	require.NoError(t, permRepo.Create(ctx, perm))
	require.NoError(t, rolePermRepo.Create(ctx, &domain.RolePermission{
		RoleID:        roleID,
		PermissionKey: "test.permission",
	}))

	// Assign role to profile
	profileID := uuid.New()
	profileRole := domain.NewProfileRole(profileID, roleID)
	require.NoError(t, profileRoleRepo.Create(ctx, profileRole))

	// Verify assignment exists
	count, err := profileRoleRepo.CountByProfileID(ctx, profileID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Delete the role
	err = svc.DeleteRole(ctx, roleID)
	require.NoError(t, err)

	// Verify role is deleted
	_, err = roleRepo.ByID(ctx, roleID)
	assert.Error(t, err)

	// Verify ProfileRole entries are cascade deleted
	count, err = profileRoleRepo.CountByProfileID(ctx, profileID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

// TestT048_PreserveChildPermissions tests that deleting a parent doesn't remove child permissions.
// When a parent role is deleted, child roles keep their direct permissions but may have
// their ParentID pointer become dangling. For this test, we verify the child role's
// direct permissions are preserved (not cascade deleted).
func TestT048_PreserveChildPermissions(t *testing.T) {
	svc, permRepo, roleRepo, rolePermRepo, _ := setupAuthzService(t)
	ctx := context.Background()

	// Create parent role
	parentID := uuid.New()
	parent := &domain.Role{
		ID:          parentID,
		Name:        "parent",
		Description: "Parent role",
	}
	require.NoError(t, roleRepo.Create(ctx, parent))

	// Create child role with parent
	childID := uuid.New()
	child := &domain.Role{
		ID:          childID,
		Name:        "child",
		Description: "Child role",
		ParentID:    &parentID,
	}
	require.NoError(t, roleRepo.Create(ctx, child))

	// Assign permission directly to child
	childPerm := domain.NewPermission("child.permission", "Child permission", domain.DomainBrand)
	require.NoError(t, permRepo.Create(ctx, childPerm))
	require.NoError(t, rolePermRepo.Create(ctx, &domain.RolePermission{
		RoleID:        childID,
		PermissionKey: "child.permission",
	}))

	// Verify child has the permission
	childPerms, err := rolePermRepo.ByRoleID(ctx, childID)
	require.NoError(t, err)
	assert.Len(t, childPerms, 1)
	assert.Equal(t, "child.permission", childPerms[0].PermissionKey)

	// Delete parent
	err = svc.DeleteRole(ctx, parentID)
	require.NoError(t, err)

	// Verify parent is deleted
	_, err = roleRepo.ByID(ctx, parentID)
	assert.Error(t, err)

	// Verify child role still exists
	childAfter, err := roleRepo.ByID(ctx, childID)
	require.NoError(t, err)
	assert.NotNil(t, childAfter)

	// Verify child's direct permissions are preserved (not deleted)
	childPermsAfter, err := rolePermRepo.ByRoleID(ctx, childID)
	require.NoError(t, err)
	assert.Len(t, childPermsAfter, 1)
	assert.Equal(t, "child.permission", childPermsAfter[0].PermissionKey)
}
package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/repository"
	"viralforge/backend/src/service"
)

// mockPermissionRepo is a mock implementation of PermissionRepository.
type mockPermissionRepo struct {
	permissions map[string]*domain.Permission
}

func (m *mockPermissionRepo) Create(ctx context.Context, permission *domain.Permission) error {
	m.permissions[permission.Key] = permission
	return nil
}
func (m *mockPermissionRepo) ByKey(ctx context.Context, key string) (*domain.Permission, error) {
	if p, ok := m.permissions[key]; ok {
		return p, nil
	}
	return nil, domain.ErrPermissionNotFound
}
func (m *mockPermissionRepo) ListAll(ctx context.Context) ([]*domain.Permission, error) {
	perms := make([]*domain.Permission, 0, len(m.permissions))
	for _, p := range m.permissions {
		perms = append(perms, p)
	}
	return perms, nil
}

// mockRoleRepo is a mock implementation of RoleRepository.
type mockRoleRepo struct {
	roles map[uuid.UUID]*domain.Role
}

func (m *mockRoleRepo) Create(ctx context.Context, role *domain.Role) error {
	m.roles[role.ID] = role
	return nil
}
func (m *mockRoleRepo) ByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	if r, ok := m.roles[id]; ok {
		return r, nil
	}
	return nil, domain.ErrRoleNotFound
}
func (m *mockRoleRepo) ByName(ctx context.Context, name string) (*domain.Role, error) {
	for _, r := range m.roles {
		if r.Name == name {
			return r, nil
		}
	}
	return nil, domain.ErrRoleNotFound
}
func (m *mockRoleRepo) ListAll(ctx context.Context) ([]*domain.Role, error) {
	roles := make([]*domain.Role, 0, len(m.roles))
	for _, r := range m.roles {
		roles = append(roles, r)
	}
	return roles, nil
}
func (m *mockRoleRepo) Update(ctx context.Context, role *domain.Role) error {
	m.roles[role.ID] = role
	return nil
}
func (m *mockRoleRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.roles, id)
	return nil
}

// mockRolePermRepo is a mock implementation of RolePermissionRepository.
type mockRolePermRepo struct {
	rolePerms map[uuid.UUID][]*domain.RolePermission
}

func (m *mockRolePermRepo) Create(ctx context.Context, rp *domain.RolePermission) error {
	m.rolePerms[rp.RoleID] = append(m.rolePerms[rp.RoleID], rp)
	return nil
}
func (m *mockRolePermRepo) ByRoleID(ctx context.Context, roleID uuid.UUID) ([]*domain.RolePermission, error) {
	return m.rolePerms[roleID], nil
}
func (m *mockRolePermRepo) ByPermissionKey(ctx context.Context, permissionKey string) ([]*domain.RolePermission, error) {
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
func (m *mockRolePermRepo) Delete(ctx context.Context, roleID uuid.UUID, permissionKey string) error {
	rps := m.rolePerms[roleID]
	for i, rp := range rps {
		if rp.PermissionKey == permissionKey {
			m.rolePerms[roleID] = append(rps[:i], rps[i+1:]...)
			break
		}
	}
	return nil
}
func (m *mockRolePermRepo) DeleteByRoleID(ctx context.Context, roleID uuid.UUID) error {
	delete(m.rolePerms, roleID)
	return nil
}

// mockProfileRoleRepo is a mock implementation of ProfileRoleRepository.
type mockProfileRoleRepo struct {
	profileRoles map[uuid.UUID][]*domain.ProfileRole
}

func (m *mockProfileRoleRepo) Create(ctx context.Context, pr *domain.ProfileRole) error {
	m.profileRoles[pr.ProfileID] = append(m.profileRoles[pr.ProfileID], pr)
	return nil
}
func (m *mockProfileRoleRepo) ByProfileID(ctx context.Context, profileID uuid.UUID) ([]*domain.ProfileRole, error) {
	return m.profileRoles[profileID], nil
}
func (m *mockProfileRoleRepo) ByRoleID(ctx context.Context, roleID uuid.UUID) ([]*domain.ProfileRole, error) {
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
func (m *mockProfileRoleRepo) Delete(ctx context.Context, profileID uuid.UUID, roleID uuid.UUID) error {
	prs := m.profileRoles[profileID]
	for i, pr := range prs {
		if pr.RoleID == roleID {
			m.profileRoles[profileID] = append(prs[:i], prs[i+1:]...)
			break
		}
	}
	return nil
}
func (m *mockProfileRoleRepo) CountByProfileID(ctx context.Context, profileID uuid.UUID) (int64, error) {
	return int64(len(m.profileRoles[profileID])), nil
}
func (m *mockProfileRoleRepo) DeleteByRoleID(ctx context.Context, roleID uuid.UUID) error {
	for profileID, prs := range m.profileRoles {
		for i, pr := range prs {
			if pr.RoleID == roleID {
				m.profileRoles[profileID] = append(prs[:i], prs[i+1:]...)
			}
		}
	}
	return nil
}

func setupUS2Service(t *testing.T) (*service.AuthorizationService, *mockProfileRoleRepo) {
	permRepo := &mockPermissionRepo{permissions: make(map[string]*domain.Permission)}
	roleRepo := &mockRoleRepo{roles: make(map[uuid.UUID]*domain.Role)}
	rolePermRepo := &mockRolePermRepo{rolePerms: make(map[uuid.UUID][]*domain.RolePermission)}
	profileRoleRepo := &mockProfileRoleRepo{profileRoles: make(map[uuid.UUID][]*domain.ProfileRole)}

	svc := service.NewAuthorizationService(permRepo, roleRepo, rolePermRepo, profileRoleRepo, nil)
	return svc, profileRoleRepo
}

// TestT022_AssignRole tests role assignment.
func TestT022_AssignRole(t *testing.T) {
	svc, profileRoleRepo := setupUS2Service(t)
	ctx := context.Background()

	profileID := uuid.New()
	roleID := uuid.New()

	// Verify profile has no roles initially
	count, err := profileRoleRepo.CountByProfileID(ctx, profileID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Assign role
	err = svc.AssignRole(ctx, profileID, roleID)
	require.NoError(t, err)

	// Verify role was assigned
	count, err = profileRoleRepo.CountByProfileID(ctx, profileID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

// TestT022_AssignRole_MaxRolesExceeded tests that AssignRole fails when max roles reached.
func TestT022_AssignRole_MaxRolesExceeded(t *testing.T) {
	svc, _ := setupUS2Service(t)
	ctx := context.Background()

	profileID := uuid.New()

	// Assign 10 roles
	for i := 0; i < 10; i++ {
		roleID := uuid.New()
		err := svc.AssignRole(ctx, profileID, roleID)
		require.NoError(t, err)
	}

	// 11th role should fail with ErrMaxRolesExceeded
	roleID := uuid.New()
	err := svc.AssignRole(ctx, profileID, roleID)
	require.Error(t, err)
	assert.Equal(t, domain.ErrMaxRolesExceeded, err)
}

// TestT022_AssignRole_DuplicateRole tests that assigning the same role twice fails.
func TestT022_AssignRole_DuplicateRole(t *testing.T) {
	svc, _ := setupUS2Service(t)
	ctx := context.Background()

	profileID := uuid.New()
	roleID := uuid.New()

	// Assign role first time
	err := svc.AssignRole(ctx, profileID, roleID)
	require.NoError(t, err)

	// Assign same role again should fail
	err = svc.AssignRole(ctx, profileID, roleID)
	require.Error(t, err)
	assert.Equal(t, domain.ErrRoleAlreadyAssigned, err)
}

// TestT023_RevokeRole tests role removal.
func TestT023_RevokeRole(t *testing.T) {
	svc, profileRoleRepo := setupUS2Service(t)
	ctx := context.Background()

	profileID := uuid.New()
	roleID := uuid.New()

	// Assign role
	err := svc.AssignRole(ctx, profileID, roleID)
	require.NoError(t, err)

	// Verify role assigned
	count, err := profileRoleRepo.CountByProfileID(ctx, profileID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Revoke role
	err = svc.RevokeRole(ctx, profileID, roleID)
	require.NoError(t, err)

	// Verify role removed
	count, err = profileRoleRepo.CountByProfileID(ctx, profileID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

// TestT023_RevokeRole_NotAssigned tests that revoking a non-assigned role fails.
func TestT023_RevokeRole_NotAssigned(t *testing.T) {
	svc, _ := setupUS2Service(t)
	ctx := context.Background()

	profileID := uuid.New()
	roleID := uuid.New()

	// Try to revoke a role that was never assigned
	err := svc.RevokeRole(ctx, profileID, roleID)
	require.Error(t, err)
	assert.Equal(t, domain.ErrRoleNotAssigned, err)
}

// TestT023_RevokeRole_CascadeNotTriggered tests that revoking a role doesn't cascade.
// When a role is revoked from a profile, only that assignment is removed.
// Any permissions the profile had through other roles remain.
// TestT023_RevokeRole_CascadeNotTriggered tests that revoking a role doesn't cascade.
// When a role is revoked from a profile, only that assignment is removed.
// Any permissions the profile had through other roles remain.
func TestT023_RevokeRole_CascadeNotTriggered(t *testing.T) {
	svc, profileRoleRepo := setupUS2Service(t)
	ctx := context.Background()

	profileID := uuid.New()
	roleID1 := uuid.New()
	roleID2 := uuid.New()

	// Assign two roles
	err := svc.AssignRole(ctx, profileID, roleID1)
	require.NoError(t, err)

	err = svc.AssignRole(ctx, profileID, roleID2)
	require.NoError(t, err)

	// Revoke one role
	err = svc.RevokeRole(ctx, profileID, roleID1)
	require.NoError(t, err)

	// The other role should still be assigned
	profileRoles, err := profileRoleRepo.ByProfileID(ctx, profileID)
	require.NoError(t, err)
	assert.Len(t, profileRoles, 1)
	assert.Equal(t, roleID2, profileRoles[0].RoleID)
}

// Compile-time interface check
var _ repository.ProfileRoleRepository = (*mockProfileRoleRepo)(nil)

package contract

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/handler"
	"viralforge/backend/src/service"
)

// mockUS4PermissionRepo is a mock implementation of PermissionRepository.
type mockUS4PermissionRepo struct {
	permissions map[string]*domain.Permission
}

func (m *mockUS4PermissionRepo) Create(ctx context.Context, permission *domain.Permission) error {
	m.permissions[permission.Key] = permission
	return nil
}
func (m *mockUS4PermissionRepo) ByKey(ctx context.Context, key string) (*domain.Permission, error) {
	if p, ok := m.permissions[key]; ok {
		return p, nil
	}
	return nil, domain.ErrPermissionNotFound
}
func (m *mockUS4PermissionRepo) ListAll(ctx context.Context) ([]*domain.Permission, error) {
	perms := make([]*domain.Permission, 0, len(m.permissions))
	for _, p := range m.permissions {
		perms = append(perms, p)
	}
	return perms, nil
}

// mockUS4RoleRepo is a mock implementation of RoleRepository.
type mockUS4RoleRepo struct {
	roles map[uuid.UUID]*domain.Role
}

func (m *mockUS4RoleRepo) Create(ctx context.Context, role *domain.Role) error {
	m.roles[role.ID] = role
	return nil
}
func (m *mockUS4RoleRepo) ByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	if r, ok := m.roles[id]; ok {
		return r, nil
	}
	return nil, domain.ErrRoleNotFound
}
func (m *mockUS4RoleRepo) ByName(ctx context.Context, name string) (*domain.Role, error) {
	for _, r := range m.roles {
		if r.Name == name {
			return r, nil
		}
	}
	return nil, domain.ErrRoleNotFound
}
func (m *mockUS4RoleRepo) ListAll(ctx context.Context) ([]*domain.Role, error) {
	roles := make([]*domain.Role, 0, len(m.roles))
	for _, r := range m.roles {
		roles = append(roles, r)
	}
	return roles, nil
}
func (m *mockUS4RoleRepo) Update(ctx context.Context, role *domain.Role) error {
	m.roles[role.ID] = role
	return nil
}
func (m *mockUS4RoleRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.roles, id)
	return nil
}

// mockUS4RolePermRepo is a mock implementation of RolePermissionRepository.
type mockUS4RolePermRepo struct {
	rolePerms map[uuid.UUID][]*domain.RolePermission
}

func (m *mockUS4RolePermRepo) Create(ctx context.Context, rp *domain.RolePermission) error {
	m.rolePerms[rp.RoleID] = append(m.rolePerms[rp.RoleID], rp)
	return nil
}
func (m *mockUS4RolePermRepo) ByRoleID(ctx context.Context, roleID uuid.UUID) ([]*domain.RolePermission, error) {
	return m.rolePerms[roleID], nil
}
func (m *mockUS4RolePermRepo) ByPermissionKey(ctx context.Context, permissionKey string) ([]*domain.RolePermission, error) {
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
func (m *mockUS4RolePermRepo) Delete(ctx context.Context, roleID uuid.UUID, permissionKey string) error {
	rps := m.rolePerms[roleID]
	for i, rp := range rps {
		if rp.PermissionKey == permissionKey {
			m.rolePerms[roleID] = append(rps[:i], rps[i+1:]...)
			break
		}
	}
	return nil
}
func (m *mockUS4RolePermRepo) DeleteByRoleID(ctx context.Context, roleID uuid.UUID) error {
	delete(m.rolePerms, roleID)
	return nil
}

// mockUS4ProfileRoleRepo is a mock implementation of ProfileRoleRepository.
type mockUS4ProfileRoleRepo struct {
	profileRoles map[uuid.UUID][]*domain.ProfileRole
}

func (m *mockUS4ProfileRoleRepo) Create(ctx context.Context, pr *domain.ProfileRole) error {
	m.profileRoles[pr.ProfileID] = append(m.profileRoles[pr.ProfileID], pr)
	return nil
}
func (m *mockUS4ProfileRoleRepo) ByProfileID(ctx context.Context, profileID uuid.UUID) ([]*domain.ProfileRole, error) {
	return m.profileRoles[profileID], nil
}
func (m *mockUS4ProfileRoleRepo) ByRoleID(ctx context.Context, roleID uuid.UUID) ([]*domain.ProfileRole, error) {
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
func (m *mockUS4ProfileRoleRepo) Delete(ctx context.Context, profileID uuid.UUID, roleID uuid.UUID) error {
	prs := m.profileRoles[profileID]
	for i, pr := range prs {
		if pr.RoleID == roleID {
			m.profileRoles[profileID] = append(prs[:i], prs[i+1:]...)
			break
		}
	}
	return nil
}
func (m *mockUS4ProfileRoleRepo) CountByProfileID(ctx context.Context, profileID uuid.UUID) (int64, error) {
	return int64(len(m.profileRoles[profileID])), nil
}
func (m *mockUS4ProfileRoleRepo) DeleteByRoleID(ctx context.Context, roleID uuid.UUID) error {
	for profileID, prs := range m.profileRoles {
		for i, pr := range prs {
			if pr.RoleID == roleID {
				m.profileRoles[profileID] = append(prs[:i], prs[i+1:]...)
			}
		}
	}
	return nil
}

func setupUS4Handler() (*handler.AuthzHandler, *mockUS4RoleRepo, *mockUS4RolePermRepo) {
	permRepo := &mockUS4PermissionRepo{permissions: make(map[string]*domain.Permission)}
	roleRepo := &mockUS4RoleRepo{roles: make(map[uuid.UUID]*domain.Role)}
	rolePermRepo := &mockUS4RolePermRepo{rolePerms: make(map[uuid.UUID][]*domain.RolePermission)}
	profileRoleRepo := &mockUS4ProfileRoleRepo{profileRoles: make(map[uuid.UUID][]*domain.ProfileRole)}

	svc := service.NewAuthorizationService(permRepo, roleRepo, rolePermRepo, profileRoleRepo)
	h := handler.NewAuthzHandler(svc)
	return h, roleRepo, rolePermRepo
}

// TestT038_GET_Roles_PlatformAdminWithWildcard tests that GET /api/v1/roles returns platform_admin with * permission.
func TestT038_GET_Roles_PlatformAdminWithWildcard(t *testing.T) {
	h, roleRepo, rolePermRepo := setupUS4Handler()
	ctx := context.Background()

	router := chi.NewRouter()
	h.RegisterRoutes(router)

	// Create platform_admin role
	platformAdminID := uuid.New()
	platformAdmin := &domain.Role{
		ID:          platformAdminID,
		Name:        "platform_admin",
		Description: "Platform administration",
	}
	require.NoError(t, roleRepo.Create(ctx, platformAdmin))

	// Assign * permission to platform_admin
	require.NoError(t, rolePermRepo.Create(ctx, &domain.RolePermission{
		RoleID:        platformAdminID,
		PermissionKey: "*",
	}))

	// Create another role
	brandOwnerID := uuid.New()
	brandOwner := &domain.Role{
		ID:          brandOwnerID,
		Name:        "brand_owner",
		Description: "Full Brand access",
	}
	require.NoError(t, roleRepo.Create(ctx, brandOwner))

	// Assign specific permission to brand_owner
	require.NoError(t, rolePermRepo.Create(ctx, &domain.RolePermission{
		RoleID:        brandOwnerID,
		PermissionKey: "campaign.*",
	}))

	// GET /api/v1/roles
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/roles", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var response handler.ListRolesResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))

	// Find platform_admin in response
	var platformAdminRole *handler.RoleResponse
	for i := range response.Roles {
		if response.Roles[i].Name == "platform_admin" {
			platformAdminRole = &response.Roles[i]
			break
		}
	}
	require.NotNil(t, platformAdminRole, "platform_admin role should be in response")

	// Verify platform_admin has * permission
	assert.Contains(t, platformAdminRole.Permissions, "*", "platform_admin should have * permission")

	// Verify brand_owner has campaign.* permission
	var brandOwnerRole *handler.RoleResponse
	for i := range response.Roles {
		if response.Roles[i].Name == "brand_owner" {
			brandOwnerRole = &response.Roles[i]
			break
		}
	}
	require.NotNil(t, brandOwnerRole, "brand_owner role should be in response")
	assert.Contains(t, brandOwnerRole.Permissions, "campaign.*", "brand_owner should have campaign.* permission")
}
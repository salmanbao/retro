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

// mockUS7PermissionRepo is a mock implementation of PermissionRepository.
type mockUS7PermissionRepo struct {
	permissions map[string]*domain.Permission
}

func (m *mockUS7PermissionRepo) Create(ctx context.Context, permission *domain.Permission) error {
	m.permissions[permission.Key] = permission
	return nil
}
func (m *mockUS7PermissionRepo) ByKey(ctx context.Context, key string) (*domain.Permission, error) {
	if p, ok := m.permissions[key]; ok {
		return p, nil
	}
	return nil, domain.ErrPermissionNotFound
}
func (m *mockUS7PermissionRepo) ListAll(ctx context.Context) ([]*domain.Permission, error) {
	perms := make([]*domain.Permission, 0, len(m.permissions))
	for _, p := range m.permissions {
		perms = append(perms, p)
	}
	return perms, nil
}

// mockUS7RoleRepo is a mock implementation of RoleRepository.
type mockUS7RoleRepo struct {
	roles map[uuid.UUID]*domain.Role
}

func (m *mockUS7RoleRepo) Create(ctx context.Context, role *domain.Role) error {
	m.roles[role.ID] = role
	return nil
}
func (m *mockUS7RoleRepo) ByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	if r, ok := m.roles[id]; ok {
		return r, nil
	}
	return nil, domain.ErrRoleNotFound
}
func (m *mockUS7RoleRepo) ByName(ctx context.Context, name string) (*domain.Role, error) {
	for _, r := range m.roles {
		if r.Name == name {
			return r, nil
		}
	}
	return nil, domain.ErrRoleNotFound
}
func (m *mockUS7RoleRepo) ListAll(ctx context.Context) ([]*domain.Role, error) {
	roles := make([]*domain.Role, 0, len(m.roles))
	for _, r := range m.roles {
		roles = append(roles, r)
	}
	return roles, nil
}
func (m *mockUS7RoleRepo) Update(ctx context.Context, role *domain.Role) error {
	m.roles[role.ID] = role
	return nil
}
func (m *mockUS7RoleRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.roles, id)
	return nil
}

// mockUS7RolePermRepo is a mock implementation of RolePermissionRepository.
type mockUS7RolePermRepo struct {
	rolePerms map[uuid.UUID][]*domain.RolePermission
}

func (m *mockUS7RolePermRepo) Create(ctx context.Context, rp *domain.RolePermission) error {
	m.rolePerms[rp.RoleID] = append(m.rolePerms[rp.RoleID], rp)
	return nil
}
func (m *mockUS7RolePermRepo) ByRoleID(ctx context.Context, roleID uuid.UUID) ([]*domain.RolePermission, error) {
	return m.rolePerms[roleID], nil
}
func (m *mockUS7RolePermRepo) ByPermissionKey(ctx context.Context, permissionKey string) ([]*domain.RolePermission, error) {
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
func (m *mockUS7RolePermRepo) Delete(ctx context.Context, roleID uuid.UUID, permissionKey string) error {
	rps := m.rolePerms[roleID]
	for i, rp := range rps {
		if rp.PermissionKey == permissionKey {
			m.rolePerms[roleID] = append(rps[:i], rps[i+1:]...)
			break
		}
	}
	return nil
}
func (m *mockUS7RolePermRepo) DeleteByRoleID(ctx context.Context, roleID uuid.UUID) error {
	delete(m.rolePerms, roleID)
	return nil
}

// mockUS7ProfileRoleRepo is a mock implementation of ProfileRoleRepository.
type mockUS7ProfileRoleRepo struct {
	profileRoles map[uuid.UUID][]*domain.ProfileRole
}

func (m *mockUS7ProfileRoleRepo) Create(ctx context.Context, pr *domain.ProfileRole) error {
	m.profileRoles[pr.ProfileID] = append(m.profileRoles[pr.ProfileID], pr)
	return nil
}
func (m *mockUS7ProfileRoleRepo) ByProfileID(ctx context.Context, profileID uuid.UUID) ([]*domain.ProfileRole, error) {
	return m.profileRoles[profileID], nil
}
func (m *mockUS7ProfileRoleRepo) ByRoleID(ctx context.Context, roleID uuid.UUID) ([]*domain.ProfileRole, error) {
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
func (m *mockUS7ProfileRoleRepo) Delete(ctx context.Context, profileID uuid.UUID, roleID uuid.UUID) error {
	prs := m.profileRoles[profileID]
	for i, pr := range prs {
		if pr.RoleID == roleID {
			m.profileRoles[profileID] = append(prs[:i], prs[i+1:]...)
			break
		}
	}
	return nil
}
func (m *mockUS7ProfileRoleRepo) CountByProfileID(ctx context.Context, profileID uuid.UUID) (int64, error) {
	return int64(len(m.profileRoles[profileID])), nil
}
func (m *mockUS7ProfileRoleRepo) DeleteByRoleID(ctx context.Context, roleID uuid.UUID) error {
	for profileID, prs := range m.profileRoles {
		for i, pr := range prs {
			if pr.RoleID == roleID {
				m.profileRoles[profileID] = append(prs[:i], prs[i+1:]...)
			}
		}
	}
	return nil
}

func setupUS7Handler() (*handler.AuthzHandler, *mockUS7RoleRepo, *mockUS7RolePermRepo) {
	permRepo := &mockUS7PermissionRepo{permissions: make(map[string]*domain.Permission)}
	roleRepo := &mockUS7RoleRepo{roles: make(map[uuid.UUID]*domain.Role)}
	rolePermRepo := &mockUS7RolePermRepo{rolePerms: make(map[uuid.UUID][]*domain.RolePermission)}
	profileRoleRepo := &mockUS7ProfileRoleRepo{profileRoles: make(map[uuid.UUID][]*domain.ProfileRole)}

	svc := service.NewAuthorizationService(permRepo, roleRepo, rolePermRepo, profileRoleRepo)
	h := handler.NewAuthzHandler(svc)
	return h, roleRepo, rolePermRepo
}

// TestT054_GET_Roles_ResponseFormat tests that GET /api/v1/roles returns the correct
// response format per contracts/authz-api.md.
func TestT054_GET_Roles_ResponseFormat(t *testing.T) {
	h, roleRepo, rolePermRepo := setupUS7Handler()
	ctx := context.Background()

	router := chi.NewRouter()
	h.RegisterRoutes(router)

	// Create brand_owner role
	brandOwnerID := uuid.New()
	brandOwner := &domain.Role{
		ID:          brandOwnerID,
		Name:        "brand_owner",
		Description: "Full Brand access",
	}
	require.NoError(t, roleRepo.Create(ctx, brandOwner))

	// Assign permissions to brand_owner
	require.NoError(t, rolePermRepo.Create(ctx, &domain.RolePermission{
		RoleID:        brandOwnerID,
		PermissionKey: "campaign.*",
	}))
	require.NoError(t, rolePermRepo.Create(ctx, &domain.RolePermission{
		RoleID:        brandOwnerID,
		PermissionKey: "analytics.view",
	}))

	// GET /api/v1/roles
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/roles", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var response handler.ListRolesResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))

	// Verify response structure per contracts/authz-api.md
	assert.NotEmpty(t, response.Roles, "roles should not be empty")

	// Find brand_owner in response
	var brandOwnerRole *handler.RoleResponse
	for i := range response.Roles {
		if response.Roles[i].Name == "brand_owner" {
			brandOwnerRole = &response.Roles[i]
			break
		}
	}
	require.NotNil(t, brandOwnerRole, "brand_owner role should be in response")

	// Verify response fields per contract
	assert.Equal(t, brandOwnerID.String(), brandOwnerRole.ID, "id should match")
	assert.Equal(t, "brand_owner", brandOwnerRole.Name, "name should match")
	assert.Equal(t, "Full Brand access", brandOwnerRole.Description, "description should match")
	assert.Nil(t, brandOwnerRole.ParentID, "brand_owner has no parent (root role)")

	// Verify permissions array
	assert.NotEmpty(t, brandOwnerRole.Permissions, "permissions should not be empty")
	assert.Contains(t, brandOwnerRole.Permissions, "campaign.*", "should have campaign.* permission")
	assert.Contains(t, brandOwnerRole.Permissions, "analytics.view", "should have analytics.view permission")
}

// TestT055_GET_Roles_ParentIDAndPermissions tests that GET /api/v1/roles correctly
// returns parent_id and permissions array for roles in a hierarchy.
func TestT055_GET_Roles_ParentIDAndPermissions(t *testing.T) {
	h, roleRepo, rolePermRepo := setupUS7Handler()
	ctx := context.Background()

	router := chi.NewRouter()
	h.RegisterRoutes(router)

	// Create brand_owner (parent role)
	brandOwnerID := uuid.New()
	brandOwner := &domain.Role{
		ID:          brandOwnerID,
		Name:        "brand_owner",
		Description: "Full Brand access",
	}
	require.NoError(t, roleRepo.Create(ctx, brandOwner))

	// Assign permissions to brand_owner
	require.NoError(t, rolePermRepo.Create(ctx, &domain.RolePermission{
		RoleID:        brandOwnerID,
		PermissionKey: "campaign.*",
	}))

	// Create brand_admin (child of brand_owner)
	brandAdminID := uuid.New()
	brandAdmin := &domain.Role{
		ID:          brandAdminID,
		Name:        "brand_admin",
		Description: "Brand administration",
		ParentID:    &brandOwnerID,
	}
	require.NoError(t, roleRepo.Create(ctx, brandAdmin))

	// Assign permissions to brand_admin
	require.NoError(t, rolePermRepo.Create(ctx, &domain.RolePermission{
		RoleID:        brandAdminID,
		PermissionKey: "campaign.create",
	}))
	require.NoError(t, rolePermRepo.Create(ctx, &domain.RolePermission{
		RoleID:        brandAdminID,
		PermissionKey: "campaign.update",
	}))

	// Create brand_marketer (child of brand_admin)
	brandMarketerID := uuid.New()
	brandMarketer := &domain.Role{
		ID:          brandMarketerID,
		Name:        "brand_marketer",
		Description: "Brand marketing",
		ParentID:    &brandAdminID,
	}
	require.NoError(t, roleRepo.Create(ctx, brandMarketer))

	// Assign permissions to brand_marketer
	require.NoError(t, rolePermRepo.Create(ctx, &domain.RolePermission{
		RoleID:        brandMarketerID,
		PermissionKey: "submission.review",
	}))

	// GET /api/v1/roles
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/roles", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var response handler.ListRolesResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))

	// Find brand_admin in response
	var brandAdminRole *handler.RoleResponse
	var brandMarketerRole *handler.RoleResponse
	var brandOwnerRole *handler.RoleResponse

	for i := range response.Roles {
		switch response.Roles[i].Name {
		case "brand_admin":
			brandAdminRole = &response.Roles[i]
		case "brand_marketer":
			brandMarketerRole = &response.Roles[i]
		case "brand_owner":
			brandOwnerRole = &response.Roles[i]
		}
	}

	require.NotNil(t, brandOwnerRole, "brand_owner role should be in response")
	require.NotNil(t, brandAdminRole, "brand_admin role should be in response")
	require.NotNil(t, brandMarketerRole, "brand_marketer role should be in response")

	// Verify brand_owner has no parent_id
	assert.Nil(t, brandOwnerRole.ParentID, "brand_owner should have nil parent_id")

	// Verify brand_admin has parent_id pointing to brand_owner
	require.NotNil(t, brandAdminRole.ParentID, "brand_admin should have parent_id")
	assert.Equal(t, brandOwnerID.String(), *brandAdminRole.ParentID, "parent_id should be brand_owner id")

	// Verify brand_marketer has parent_id pointing to brand_admin
	require.NotNil(t, brandMarketerRole.ParentID, "brand_marketer should have parent_id")
	assert.Equal(t, brandAdminID.String(), *brandMarketerRole.ParentID, "parent_id should be brand_admin id")

	// Verify permissions are correctly returned for each role
	assert.Contains(t, brandAdminRole.Permissions, "campaign.create", "brand_admin should have campaign.create")
	assert.Contains(t, brandAdminRole.Permissions, "campaign.update", "brand_admin should have campaign.update")
	assert.NotContains(t, brandAdminRole.Permissions, "submission.review", "brand_admin should NOT have submission.review")

	assert.Contains(t, brandMarketerRole.Permissions, "submission.review", "brand_marketer should have submission.review")
	assert.NotContains(t, brandMarketerRole.Permissions, "campaign.create", "brand_marketer should NOT have campaign.create")
}
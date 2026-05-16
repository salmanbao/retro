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
	"viralforge/backend/src/middleware"
	"viralforge/backend/src/service"
)

// mockUS6PermissionRepo is a mock implementation of PermissionRepository.
type mockUS6PermissionRepo struct {
	permissions map[string]*domain.Permission
}

func (m *mockUS6PermissionRepo) Create(ctx context.Context, permission *domain.Permission) error {
	m.permissions[permission.Key] = permission
	return nil
}
func (m *mockUS6PermissionRepo) ByKey(ctx context.Context, key string) (*domain.Permission, error) {
	if p, ok := m.permissions[key]; ok {
		return p, nil
	}
	return nil, domain.ErrPermissionNotFound
}
func (m *mockUS6PermissionRepo) ListAll(ctx context.Context) ([]*domain.Permission, error) {
	perms := make([]*domain.Permission, 0, len(m.permissions))
	for _, p := range m.permissions {
		perms = append(perms, p)
	}
	return perms, nil
}

// mockUS6RoleRepo is a mock implementation of RoleRepository.
type mockUS6RoleRepo struct {
	roles map[uuid.UUID]*domain.Role
}

func (m *mockUS6RoleRepo) Create(ctx context.Context, role *domain.Role) error {
	m.roles[role.ID] = role
	return nil
}
func (m *mockUS6RoleRepo) ByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	if r, ok := m.roles[id]; ok {
		return r, nil
	}
	return nil, domain.ErrRoleNotFound
}
func (m *mockUS6RoleRepo) ByName(ctx context.Context, name string) (*domain.Role, error) {
	for _, r := range m.roles {
		if r.Name == name {
			return r, nil
		}
	}
	return nil, domain.ErrRoleNotFound
}
func (m *mockUS6RoleRepo) ListAll(ctx context.Context) ([]*domain.Role, error) {
	roles := make([]*domain.Role, 0, len(m.roles))
	for _, r := range m.roles {
		roles = append(roles, r)
	}
	return roles, nil
}
func (m *mockUS6RoleRepo) Update(ctx context.Context, role *domain.Role) error {
	m.roles[role.ID] = role
	return nil
}
func (m *mockUS6RoleRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.roles, id)
	return nil
}

// mockUS6RolePermRepo is a mock implementation of RolePermissionRepository.
type mockUS6RolePermRepo struct {
	rolePerms map[uuid.UUID][]*domain.RolePermission
}

func (m *mockUS6RolePermRepo) Create(ctx context.Context, rp *domain.RolePermission) error {
	m.rolePerms[rp.RoleID] = append(m.rolePerms[rp.RoleID], rp)
	return nil
}
func (m *mockUS6RolePermRepo) ByRoleID(ctx context.Context, roleID uuid.UUID) ([]*domain.RolePermission, error) {
	return m.rolePerms[roleID], nil
}
func (m *mockUS6RolePermRepo) ByPermissionKey(ctx context.Context, permissionKey string) ([]*domain.RolePermission, error) {
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
func (m *mockUS6RolePermRepo) Delete(ctx context.Context, roleID uuid.UUID, permissionKey string) error {
	rps := m.rolePerms[roleID]
	for i, rp := range rps {
		if rp.PermissionKey == permissionKey {
			m.rolePerms[roleID] = append(rps[:i], rps[i+1:]...)
			break
		}
	}
	return nil
}
func (m *mockUS6RolePermRepo) DeleteByRoleID(ctx context.Context, roleID uuid.UUID) error {
	delete(m.rolePerms, roleID)
	return nil
}

// mockUS6ProfileRoleRepo is a mock implementation of ProfileRoleRepository.
type mockUS6ProfileRoleRepo struct {
	profileRoles map[uuid.UUID][]*domain.ProfileRole
}

func (m *mockUS6ProfileRoleRepo) Create(ctx context.Context, pr *domain.ProfileRole) error {
	m.profileRoles[pr.ProfileID] = append(m.profileRoles[pr.ProfileID], pr)
	return nil
}
func (m *mockUS6ProfileRoleRepo) ByProfileID(ctx context.Context, profileID uuid.UUID) ([]*domain.ProfileRole, error) {
	return m.profileRoles[profileID], nil
}
func (m *mockUS6ProfileRoleRepo) ByRoleID(ctx context.Context, roleID uuid.UUID) ([]*domain.ProfileRole, error) {
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
func (m *mockUS6ProfileRoleRepo) Delete(ctx context.Context, profileID uuid.UUID, roleID uuid.UUID) error {
	prs := m.profileRoles[profileID]
	for i, pr := range prs {
		if pr.RoleID == roleID {
			m.profileRoles[profileID] = append(prs[:i], prs[i+1:]...)
			break
		}
	}
	return nil
}
func (m *mockUS6ProfileRoleRepo) CountByProfileID(ctx context.Context, profileID uuid.UUID) (int64, error) {
	return int64(len(m.profileRoles[profileID])), nil
}
func (m *mockUS6ProfileRoleRepo) DeleteByRoleID(ctx context.Context, roleID uuid.UUID) error {
	for profileID, prs := range m.profileRoles {
		for i, pr := range prs {
			if pr.RoleID == roleID {
				m.profileRoles[profileID] = append(prs[:i], prs[i+1:]...)
			}
		}
	}
	return nil
}

func setupUS6Handler() (*handler.AuthzHandler, *mockUS6RoleRepo, *mockUS6RolePermRepo) {
	permRepo := &mockUS6PermissionRepo{permissions: make(map[string]*domain.Permission)}
	roleRepo := &mockUS6RoleRepo{roles: make(map[uuid.UUID]*domain.Role)}
	rolePermRepo := &mockUS6RolePermRepo{rolePerms: make(map[uuid.UUID][]*domain.RolePermission)}
	profileRoleRepo := &mockUS6ProfileRoleRepo{profileRoles: make(map[uuid.UUID][]*domain.ProfileRole)}

	svc := service.NewAuthorizationService(permRepo, roleRepo, rolePermRepo, profileRoleRepo, nil)
	h := handler.NewAuthzHandler(svc)
	return h, roleRepo, rolePermRepo
}

func setupUS6HandlerFull() (*handler.AuthzHandler, *mockUS6RoleRepo, *mockUS6RolePermRepo, *mockUS6PermissionRepo) {
	permRepo := &mockUS6PermissionRepo{permissions: make(map[string]*domain.Permission)}
	roleRepo := &mockUS6RoleRepo{roles: make(map[uuid.UUID]*domain.Role)}
	rolePermRepo := &mockUS6RolePermRepo{rolePerms: make(map[uuid.UUID][]*domain.RolePermission)}
	profileRoleRepo := &mockUS6ProfileRoleRepo{profileRoles: make(map[uuid.UUID][]*domain.ProfileRole)}

	svc := service.NewAuthorizationService(permRepo, roleRepo, rolePermRepo, profileRoleRepo, nil)
	h := handler.NewAuthzHandler(svc)
	return h, roleRepo, rolePermRepo, permRepo
}

// TestT049_GET_ProfilePermissions_ResponseFormat tests that GET /api/v1/profiles/{id}/permissions
// returns the correct response format per contracts/authz-api.md.
func TestT049_GET_ProfilePermissions_ResponseFormat(t *testing.T) {
	h, roleRepo, rolePermRepo, permRepo := setupUS6HandlerFull()
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
		PermissionKey: "campaign.create",
	}))
	require.NoError(t, rolePermRepo.Create(ctx, &domain.RolePermission{
		RoleID:        brandOwnerID,
		PermissionKey: "campaign.view",
	}))

	// Create brand_admin role (child of brand_owner)
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
		PermissionKey: "campaign.update",
	}))

	// Create profile and assign brand_admin role
	profileID := uuid.New()
	profileRole := domain.NewProfileRole(profileID, brandAdminID)

	// Override profile roles with our test data
	profileRoleRepo := &mockUS6ProfileRoleRepo{profileRoles: make(map[uuid.UUID][]*domain.ProfileRole)}
	profileRoleRepo.profileRoles[profileID] = []*domain.ProfileRole{profileRole}

	// Create a new service with the profile role repo (recreate to use our repos)
	svc := service.NewAuthorizationService(permRepo, roleRepo, rolePermRepo, profileRoleRepo, nil)
	h = handler.NewAuthzHandler(svc)

	router = chi.NewRouter()
	h.RegisterRoutes(router)

	// Add profile.view permission to the caller profile so permission check passes (FR-AUTH-010)
	// The caller is the same profileID, so we add the permission to brand_admin role
	rolePermRepo.Create(ctx, &domain.RolePermission{
		RoleID:        brandAdminID,
		PermissionKey: "profile.view",
	})

	// GET /api/v1/profiles/{id}/permissions
	// Set up context with active profile so permission check passes (FR-AUTH-010)
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/profiles/"+profileID.String()+"/permissions", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.ActiveProfileIDKey, profileID))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var response handler.GetEffectivePermissionsResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))

	// Verify response structure per contracts/authz-api.md
	assert.Equal(t, profileID.String(), response.ProfileID, "profile_id should match")

	// Verify permissions array is present and contains expected permissions
	assert.NotEmpty(t, response.Permissions, "permissions should not be empty")
	assert.Contains(t, response.Permissions, "campaign.update", "should contain direct permission")
	assert.Contains(t, response.Permissions, "campaign.create", "should contain inherited permission")
	assert.Contains(t, response.Permissions, "campaign.view", "should contain inherited permission")

	// Verify roles array is present
	assert.NotEmpty(t, response.Roles, "roles should not be empty")

	// Find brand_admin role in response
	var brandAdminRole *handler.RoleInfo
	for i := range response.Roles {
		if response.Roles[i].Name == "brand_admin" {
			brandAdminRole = &response.Roles[i]
			break
		}
	}
	require.NotNil(t, brandAdminRole, "brand_admin role should be in response")
	assert.Equal(t, brandAdminID.String(), brandAdminRole.ID)
	// brand_admin has a parent (brand_owner) in the role hierarchy
	assert.NotNil(t, brandAdminRole.InheritedFrom, "role with parent should have inherited_from")
	assert.Equal(t, "brand_owner", *brandAdminRole.InheritedFrom, "inherited_from should be brand_owner")
}

// TestT050_GET_ProfilePermissions_RoleInheritance tests that the response correctly shows
// role inheritance with inherited_from field populated.
func TestT050_GET_ProfilePermissions_RoleInheritance(t *testing.T) {
	h, roleRepo, rolePermRepo, permRepo := setupUS6HandlerFull()
	ctx := context.Background()

	router := chi.NewRouter()
	h.RegisterRoutes(router)

	// Create brand_owner (level 0)
	brandOwnerID := uuid.New()
	brandOwner := &domain.Role{
		ID:          brandOwnerID,
		Name:        "brand_owner",
		Description: "Full Brand access",
	}
	require.NoError(t, roleRepo.Create(ctx, brandOwner))

	// Assign permission to brand_owner
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

	// Assign permission to brand_admin
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

	// Create profile and assign brand_marketer role
	profileID := uuid.New()
	profileRole := domain.NewProfileRole(profileID, brandMarketerID)

	// Set up permissions for all roles
	rolePermRepoActual := &mockUS6RolePermRepo{rolePerms: make(map[uuid.UUID][]*domain.RolePermission)}

	// Add all role permissions
	rolePermRepoActual.rolePerms[brandOwnerID] = []*domain.RolePermission{
		{PermissionKey: "campaign.*"},
	}
	rolePermRepoActual.rolePerms[brandAdminID] = []*domain.RolePermission{
		{PermissionKey: "campaign.create"},
	}
	rolePermRepoActual.rolePerms[brandMarketerID] = []*domain.RolePermission{}

	profileRoleRepo := &mockUS6ProfileRoleRepo{profileRoles: make(map[uuid.UUID][]*domain.ProfileRole)}
	profileRoleRepo.profileRoles[profileID] = []*domain.ProfileRole{profileRole}

	svc := service.NewAuthorizationService(permRepo, roleRepo, rolePermRepoActual, profileRoleRepo, nil)
	h = handler.NewAuthzHandler(svc)

	router = chi.NewRouter()
	h.RegisterRoutes(router)

	// Add profile.view permission to brand_marketer so permission check passes (FR-AUTH-010)
	rolePermRepoActual.Create(ctx, &domain.RolePermission{
		RoleID:        brandMarketerID,
		PermissionKey: "profile.view",
	})

	// GET /api/v1/profiles/{id}/permissions
	// Set up context with active profile so permission check passes (FR-AUTH-010)
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/profiles/"+profileID.String()+"/permissions", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.ActiveProfileIDKey, profileID))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var response handler.GetEffectivePermissionsResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))

	// Verify that brand_marketer role shows inherited_from brand_admin
	var brandMarketerRole *handler.RoleInfo
	for i := range response.Roles {
		if response.Roles[i].Name == "brand_marketer" {
			brandMarketerRole = &response.Roles[i]
			break
		}
	}
	require.NotNil(t, brandMarketerRole, "brand_marketer role should be in response")
	assert.Equal(t, brandMarketerID.String(), brandMarketerRole.ID)
	assert.NotNil(t, brandMarketerRole.InheritedFrom, "brand_marketer should have inherited_from")
	assert.Equal(t, "brand_admin", *brandMarketerRole.InheritedFrom, "inherited_from should be brand_admin")

	// Verify that campaign.* is in permissions (from brand_owner via inheritance)
	assert.Contains(t, response.Permissions, "campaign.*", "should have inherited wildcard permission")

	// Verify that campaign.create is in permissions (direct from brand_admin)
	assert.Contains(t, response.Permissions, "campaign.create", "should have direct permission")
}
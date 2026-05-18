package contract

import (
	"bytes"
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

// mockAuthzPermissionRepo is a mock implementation of PermissionRepository.
type mockAuthzPermissionRepo struct {
	permissions map[string]*domain.Permission
}

func (m *mockAuthzPermissionRepo) Create(ctx context.Context, permission *domain.Permission) error {
	m.permissions[permission.Key] = permission
	return nil
}
func (m *mockAuthzPermissionRepo) ByKey(ctx context.Context, key string) (*domain.Permission, error) {
	if p, ok := m.permissions[key]; ok {
		return p, nil
	}
	return nil, domain.ErrPermissionNotFound
}
func (m *mockAuthzPermissionRepo) ListAll(ctx context.Context) ([]*domain.Permission, error) {
	perms := make([]*domain.Permission, 0, len(m.permissions))
	for _, p := range m.permissions {
		perms = append(perms, p)
	}
	return perms, nil
}

// mockAuthzRoleRepo is a mock implementation of RoleRepository.
type mockAuthzRoleRepo struct {
	roles map[uuid.UUID]*domain.Role
}

func (m *mockAuthzRoleRepo) Create(ctx context.Context, role *domain.Role) error {
	m.roles[role.ID] = role
	return nil
}
func (m *mockAuthzRoleRepo) ByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	if r, ok := m.roles[id]; ok {
		return r, nil
	}
	return nil, domain.ErrRoleNotFound
}
func (m *mockAuthzRoleRepo) ByName(ctx context.Context, name string) (*domain.Role, error) {
	for _, r := range m.roles {
		if r.Name == name {
			return r, nil
		}
	}
	return nil, domain.ErrRoleNotFound
}
func (m *mockAuthzRoleRepo) ListAll(ctx context.Context) ([]*domain.Role, error) {
	roles := make([]*domain.Role, 0, len(m.roles))
	for _, r := range m.roles {
		roles = append(roles, r)
	}
	return roles, nil
}
func (m *mockAuthzRoleRepo) Update(ctx context.Context, role *domain.Role) error {
	m.roles[role.ID] = role
	return nil
}
func (m *mockAuthzRoleRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.roles, id)
	return nil
}

// mockAuthzRolePermRepo is a mock implementation of RolePermissionRepository.
type mockAuthzRolePermRepo struct {
	rolePerms map[uuid.UUID][]*domain.RolePermission
}

func (m *mockAuthzRolePermRepo) Create(ctx context.Context, rp *domain.RolePermission) error {
	m.rolePerms[rp.RoleID] = append(m.rolePerms[rp.RoleID], rp)
	return nil
}
func (m *mockAuthzRolePermRepo) ByRoleID(ctx context.Context, roleID uuid.UUID) ([]*domain.RolePermission, error) {
	return m.rolePerms[roleID], nil
}
func (m *mockAuthzRolePermRepo) ByPermissionKey(ctx context.Context, permissionKey string) ([]*domain.RolePermission, error) {
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
func (m *mockAuthzRolePermRepo) Delete(ctx context.Context, roleID uuid.UUID, permissionKey string) error {
	rps := m.rolePerms[roleID]
	for i, rp := range rps {
		if rp.PermissionKey == permissionKey {
			m.rolePerms[roleID] = append(rps[:i], rps[i+1:]...)
			break
		}
	}
	return nil
}
func (m *mockAuthzRolePermRepo) DeleteByRoleID(ctx context.Context, roleID uuid.UUID) error {
	delete(m.rolePerms, roleID)
	return nil
}

// mockAuthzProfileRoleRepo is a mock implementation of ProfileRoleRepository.
type mockAuthzProfileRoleRepo struct {
	profileRoles map[uuid.UUID][]*domain.ProfileRole
}

func (m *mockAuthzProfileRoleRepo) Create(ctx context.Context, pr *domain.ProfileRole) error {
	m.profileRoles[pr.ProfileID] = append(m.profileRoles[pr.ProfileID], pr)
	return nil
}
func (m *mockAuthzProfileRoleRepo) ByProfileID(ctx context.Context, profileID uuid.UUID) ([]*domain.ProfileRole, error) {
	return m.profileRoles[profileID], nil
}
func (m *mockAuthzProfileRoleRepo) ByRoleID(ctx context.Context, roleID uuid.UUID) ([]*domain.ProfileRole, error) {
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
func (m *mockAuthzProfileRoleRepo) Delete(ctx context.Context, profileID uuid.UUID, roleID uuid.UUID) error {
	prs := m.profileRoles[profileID]
	for i, pr := range prs {
		if pr.RoleID == roleID {
			m.profileRoles[profileID] = append(prs[:i], prs[i+1:]...)
			break
		}
	}
	return nil
}
func (m *mockAuthzProfileRoleRepo) CountByProfileID(ctx context.Context, profileID uuid.UUID) (int64, error) {
	return int64(len(m.profileRoles[profileID])), nil
}
func (m *mockAuthzProfileRoleRepo) DeleteByRoleID(ctx context.Context, roleID uuid.UUID) error {
	for profileID, prs := range m.profileRoles {
		for i, pr := range prs {
			if pr.RoleID == roleID {
				m.profileRoles[profileID] = append(prs[:i], prs[i+1:]...)
			}
		}
	}
	return nil
}

func setupAuthzHandler() (*handler.AuthzHandler, *mockAuthzProfileRoleRepo) {
	permRepo := &mockAuthzPermissionRepo{permissions: make(map[string]*domain.Permission)}
	roleRepo := &mockAuthzRoleRepo{roles: make(map[uuid.UUID]*domain.Role)}
	rolePermRepo := &mockAuthzRolePermRepo{rolePerms: make(map[uuid.UUID][]*domain.RolePermission)}
	profileRoleRepo := &mockAuthzProfileRoleRepo{profileRoles: make(map[uuid.UUID][]*domain.ProfileRole)}

	svc := service.NewAuthorizationService(permRepo, roleRepo, rolePermRepo, profileRoleRepo, nil)
	h := handler.NewAuthzHandler(svc)
	return h, profileRoleRepo
}

// TestT024_POST_AssignRole tests POST /api/v1/profiles/{id}/roles.
func TestT024_POST_AssignRole(t *testing.T) {
	h, profileRoleRepo := setupAuthzHandler()
	ctx := context.Background()

	router := chi.NewRouter()
	h.RegisterRoutes(router)

	profileID := uuid.New()
	roleID := uuid.New()

	// Create request body
	reqBody := handler.AssignRoleRequest{RoleID: roleID.String()}
	body, _ := json.Marshal(reqBody)

	// Create request
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/profiles/"+profileID.String()+"/roles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert 201 Created
	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify role was assigned
	profileRoles, err := profileRoleRepo.ByProfileID(ctx, profileID)
	require.NoError(t, err)
	assert.Len(t, profileRoles, 1)
	assert.Equal(t, roleID, profileRoles[0].RoleID)
}

// TestT024_POST_AssignRole_InvalidProfileID tests POST with invalid profile ID.
func TestT024_POST_AssignRole_InvalidProfileID(t *testing.T) {
	h, _ := setupAuthzHandler()

	router := chi.NewRouter()
	h.RegisterRoutes(router)

	roleID := uuid.New()
	reqBody := handler.AssignRoleRequest{RoleID: roleID.String()}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/profiles/invalid-uuid/roles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestT024_POST_AssignRole_InvalidRoleID tests POST with invalid role ID.
func TestT024_POST_AssignRole_InvalidRoleID(t *testing.T) {
	h, _ := setupAuthzHandler()

	router := chi.NewRouter()
	h.RegisterRoutes(router)

	profileID := uuid.New()
	reqBody := handler.AssignRoleRequest{RoleID: "invalid-uuid"}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/profiles/"+profileID.String()+"/roles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestT024_DELETE_RevokeRole tests DELETE /api/v1/profiles/{id}/roles/{roleId}.
func TestT024_DELETE_RevokeRole(t *testing.T) {
	h, profileRoleRepo := setupAuthzHandler()
	ctx := context.Background()

	router := chi.NewRouter()
	h.RegisterRoutes(router)

	profileID := uuid.New()
	roleID := uuid.New()

	// Pre-assign the role
	profileRoleRepo.Create(ctx, domain.NewProfileRole(profileID, roleID))

	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/profiles/"+profileID.String()+"/roles/"+roleID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify role was revoked
	profileRoles, err := profileRoleRepo.ByProfileID(ctx, profileID)
	require.NoError(t, err)
	assert.Len(t, profileRoles, 0)
}

// TestT024_DELETE_RevokeRole_InvalidProfileID tests DELETE with invalid profile ID.
func TestT024_DELETE_RevokeRole_InvalidProfileID(t *testing.T) {
	h, _ := setupAuthzHandler()

	router := chi.NewRouter()
	h.RegisterRoutes(router)

	roleID := uuid.New()
	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/profiles/invalid-uuid/roles/"+roleID.String(), nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestT024_DELETE_RevokeRole_InvalidRoleID tests DELETE with invalid role ID.
func TestT024_DELETE_RevokeRole_InvalidRoleID(t *testing.T) {
	h, _ := setupAuthzHandler()

	router := chi.NewRouter()
	h.RegisterRoutes(router)

	profileID := uuid.New()
	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/profiles/"+profileID.String()+"/roles/invalid-uuid", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

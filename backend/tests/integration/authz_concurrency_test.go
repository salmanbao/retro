package integration

import (
	"context"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/repository"
	"viralforge/backend/src/service"
)

// --- Thread-safe mock repositories ---

type mockConcurrencyPermissionRepo struct {
	permissions map[string]*domain.Permission
	mu          sync.RWMutex
}

func (m *mockConcurrencyPermissionRepo) Create(ctx context.Context, permission *domain.Permission) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.permissions[permission.Key] = permission
	return nil
}
func (m *mockConcurrencyPermissionRepo) ByKey(ctx context.Context, key string) (*domain.Permission, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if p, ok := m.permissions[key]; ok {
		return p, nil
	}
	return nil, domain.ErrPermissionNotFound
}
func (m *mockConcurrencyPermissionRepo) ListAll(ctx context.Context) ([]*domain.Permission, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	perms := make([]*domain.Permission, 0, len(m.permissions))
	for _, p := range m.permissions {
		perms = append(perms, p)
	}
	return perms, nil
}

type mockConcurrencyRoleRepo struct {
	roles map[uuid.UUID]*domain.Role
	mu    sync.RWMutex
}

func (m *mockConcurrencyRoleRepo) Create(ctx context.Context, role *domain.Role) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.roles[role.ID] = role
	return nil
}
func (m *mockConcurrencyRoleRepo) ByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if r, ok := m.roles[id]; ok {
		return r, nil
	}
	return nil, domain.ErrRoleNotFound
}
func (m *mockConcurrencyRoleRepo) ByName(ctx context.Context, name string) (*domain.Role, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, r := range m.roles {
		if r.Name == name {
			return r, nil
		}
	}
	return nil, domain.ErrRoleNotFound
}
func (m *mockConcurrencyRoleRepo) ListAll(ctx context.Context) ([]*domain.Role, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	roles := make([]*domain.Role, 0, len(m.roles))
	for _, r := range m.roles {
		roles = append(roles, r)
	}
	return roles, nil
}
func (m *mockConcurrencyRoleRepo) Update(ctx context.Context, role *domain.Role) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.roles[role.ID] = role
	return nil
}
func (m *mockConcurrencyRoleRepo) Delete(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.roles, id)
	return nil
}

type mockConcurrencyRolePermRepo struct {
	rolePerms map[uuid.UUID][]*domain.RolePermission
	mu        sync.RWMutex
}

func (m *mockConcurrencyRolePermRepo) Create(ctx context.Context, rp *domain.RolePermission) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rolePerms[rp.RoleID] = append(m.rolePerms[rp.RoleID], rp)
	return nil
}
func (m *mockConcurrencyRolePermRepo) ByRoleID(ctx context.Context, roleID uuid.UUID) ([]*domain.RolePermission, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.rolePerms[roleID], nil
}
func (m *mockConcurrencyRolePermRepo) ByPermissionKey(ctx context.Context, permissionKey string) ([]*domain.RolePermission, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
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
func (m *mockConcurrencyRolePermRepo) Delete(ctx context.Context, roleID uuid.UUID, permissionKey string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	rps := m.rolePerms[roleID]
	for i, rp := range rps {
		if rp.PermissionKey == permissionKey {
			m.rolePerms[roleID] = append(rps[:i], rps[i+1:]...)
			break
		}
	}
	return nil
}
func (m *mockConcurrencyRolePermRepo) DeleteByRoleID(ctx context.Context, roleID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.rolePerms, roleID)
	return nil
}

type mockConcurrencyProfileRoleRepo struct {
	profileRoles map[uuid.UUID][]*domain.ProfileRole
	mu           sync.RWMutex
}

func (m *mockConcurrencyProfileRoleRepo) Create(ctx context.Context, pr *domain.ProfileRole) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.profileRoles[pr.ProfileID] = append(m.profileRoles[pr.ProfileID], pr)
	return nil
}
func (m *mockConcurrencyProfileRoleRepo) ByProfileID(ctx context.Context, profileID uuid.UUID) ([]*domain.ProfileRole, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.profileRoles[profileID], nil
}
func (m *mockConcurrencyProfileRoleRepo) ByRoleID(ctx context.Context, roleID uuid.UUID) ([]*domain.ProfileRole, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
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
func (m *mockConcurrencyProfileRoleRepo) Delete(ctx context.Context, profileID uuid.UUID, roleID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	prs := m.profileRoles[profileID]
	for i, pr := range prs {
		if pr.RoleID == roleID {
			m.profileRoles[profileID] = append(prs[:i], prs[i+1:]...)
			break
		}
	}
	return nil
}
func (m *mockConcurrencyProfileRoleRepo) CountByProfileID(ctx context.Context, profileID uuid.UUID) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return int64(len(m.profileRoles[profileID])), nil
}
func (m *mockConcurrencyProfileRoleRepo) DeleteByRoleID(ctx context.Context, roleID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for profileID, prs := range m.profileRoles {
		for i, pr := range prs {
			if pr.RoleID == roleID {
				m.profileRoles[profileID] = append(prs[:i], prs[i+1:]...)
			}
		}
	}
	return nil
}

// --- Compile-time interface checks ---
var _ repository.PermissionRepository = (*mockConcurrencyPermissionRepo)(nil)
var _ repository.RoleRepository = (*mockConcurrencyRoleRepo)(nil)
var _ repository.RolePermissionRepository = (*mockConcurrencyRolePermRepo)(nil)
var _ repository.ProfileRoleRepository = (*mockConcurrencyProfileRoleRepo)(nil)

// TestT064_ConcurrentAuthorizationRequests verifies no permission leakage across
// concurrent authorization requests (SC-AUTH-006).
func TestT064_ConcurrentAuthorizationRequests(t *testing.T) {
	ctx := context.Background()

	// Create separate profile namespaces to test isolation
	profileAID := uuid.New()
	profileBID := uuid.New()

	// Create separate roles for each profile
	roleAID := uuid.New()
	roleBID := uuid.New()

	// Mock repositories
	permRepo := &mockConcurrencyPermissionRepo{permissions: make(map[string]*domain.Permission)}
	roleRepo := &mockConcurrencyRoleRepo{roles: make(map[uuid.UUID]*domain.Role)}
	rolePermRepo := &mockConcurrencyRolePermRepo{rolePerms: make(map[uuid.UUID][]*domain.RolePermission)}
	profileRoleRepo := &mockConcurrencyProfileRoleRepo{profileRoles: make(map[uuid.UUID][]*domain.ProfileRole)}

	// Create role A with campaign.create only
	roleA := &domain.Role{ID: roleAID, Name: "role_a", Description: "Role A"}
	require.NoError(t, roleRepo.Create(ctx, roleA))
	require.NoError(t, rolePermRepo.Create(ctx, &domain.RolePermission{
		RoleID:        roleAID,
		PermissionKey: "campaign.create",
	}))

	// Create role B with submission.review only
	roleB := &domain.Role{ID: roleBID, Name: "role_b", Description: "Role B"}
	require.NoError(t, roleRepo.Create(ctx, roleB))
	require.NoError(t, rolePermRepo.Create(ctx, &domain.RolePermission{
		RoleID:        roleBID,
		PermissionKey: "submission.review",
	}))

	// Assign role A to profile A, role B to profile B
	profileRoleRepo.profileRoles[profileAID] = []*domain.ProfileRole{{ProfileID: profileAID, RoleID: roleAID}}
	profileRoleRepo.profileRoles[profileBID] = []*domain.ProfileRole{{ProfileID: profileBID, RoleID: roleBID}}

	// Create service using real service constructor
	svc := service.NewAuthorizationService(permRepo, roleRepo, rolePermRepo, profileRoleRepo)

	// Test concurrent permission checks
	const numGoroutines = 100
	resultsA := make([]bool, numGoroutines)
	resultsB := make([]bool, numGoroutines)
	var wg sync.WaitGroup

	// Launch concurrent checks for profile A (should only have campaign.create)
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			hasCampaignCreate, _ := svc.HasPermission(ctx, profileAID, "campaign.create")
			hasSubmissionReview, _ := svc.HasPermission(ctx, profileAID, "submission.review")
			resultsA[idx] = hasCampaignCreate && !hasSubmissionReview
		}(i)
	}

	// Launch concurrent checks for profile B (should only have submission.review)
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			hasCampaignCreate, _ := svc.HasPermission(ctx, profileBID, "campaign.create")
			hasSubmissionReview, _ := svc.HasPermission(ctx, profileBID, "submission.review")
			resultsB[idx] = !hasCampaignCreate && hasSubmissionReview
		}(i)
	}

	wg.Wait()

	// Verify all results for profile A
	for i, result := range resultsA {
		assert.True(t, result, "Profile A goroutine %d should have campaign.create but NOT submission.review", i)
	}

	// Verify all results for profile B
	for i, result := range resultsB {
		assert.True(t, result, "Profile B goroutine %d should have submission.review but NOT campaign.create", i)
	}
}
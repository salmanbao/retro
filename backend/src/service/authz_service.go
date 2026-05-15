package service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"viralforge/backend/src/domain"
	"viralforge/backend/src/repository"
)

// AuthorizationService handles permission resolution and role-based access control.
type AuthorizationService struct {
	permissionRepo  repository.PermissionRepository
	roleRepo        repository.RoleRepository
	rolePermRepo    repository.RolePermissionRepository
	profileRoleRepo repository.ProfileRoleRepository
}

// NewAuthorizationService creates a new AuthorizationService.
func NewAuthorizationService(
	permissionRepo repository.PermissionRepository,
	roleRepo repository.RoleRepository,
	rolePermRepo repository.RolePermissionRepository,
	profileRoleRepo repository.ProfileRoleRepository,
) *AuthorizationService {
	return &AuthorizationService{
		permissionRepo:  permissionRepo,
		roleRepo:        roleRepo,
		rolePermRepo:    rolePermRepo,
		profileRoleRepo: profileRoleRepo,
	}
}

// HasPermission checks if a profile has the specified permission.
// Returns true if the profile has the permission (directly or via role inheritance).
// Returns false for non-existent permission strings (FR-AUTH-013).
func (s *AuthorizationService) HasPermission(ctx context.Context, profileID uuid.UUID, permission string) (bool, error) {
	// Get all roles assigned to this profile
	profileRoles, err := s.profileRoleRepo.ByProfileID(ctx, profileID)
	if err != nil {
		return false, err
	}

	// Collect all effective permissions for this profile
	effectivePerms, err := s.collectEffectivePermissions(ctx, profileRoles)
	if err != nil {
		return false, err
	}

	// Check exact match
	if effectivePerms[permission] {
		return true, nil
	}

	// Check wildcard matches
	for perm := range effectivePerms {
		if s.permissionMatchesWildcard(perm, permission) {
			return true, nil
		}
	}

	// Special case: platform_admin with "*" has all permissions
	if effectivePerms["*"] {
		return true, nil
	}

	return false, nil
}

// GetEffectivePermissions returns all permissions (direct + inherited) for a profile.
func (s *AuthorizationService) GetEffectivePermissions(ctx context.Context, profileID uuid.UUID) ([]string, error) {
	profileRoles, err := s.profileRoleRepo.ByProfileID(ctx, profileID)
	if err != nil {
		return nil, err
	}

	effectivePerms, err := s.collectEffectivePermissions(ctx, profileRoles)
	if err != nil {
		return nil, err
	}

	// Convert map to sorted slice
	perms := make([]string, 0, len(effectivePerms))
	for perm := range effectivePerms {
		perms = append(perms, perm)
	}
	return perms, nil
}

// collectEffectivePermissions collects all permissions from profile roles including inheritance.
func (s *AuthorizationService) collectEffectivePermissions(ctx context.Context, profileRoles []*domain.ProfileRole) (map[string]bool, error) {
	effective := make(map[string]bool)

	for _, pr := range profileRoles {
		// Get direct permissions for this role
		rolePerms, err := s.rolePermRepo.ByRoleID(ctx, pr.RoleID)
		if err != nil {
			return nil, err
		}

		for _, rp := range rolePerms {
			effective[rp.PermissionKey] = true
		}

		// Get inherited permissions from role hierarchy (up to 3 levels)
		if err := s.collectInheritedPermissions(ctx, pr.RoleID, effective, 0); err != nil {
			return nil, err
		}
	}

	return effective, nil
}

// collectInheritedPermissions recursively collects permissions from parent roles.
func (s *AuthorizationService) collectInheritedPermissions(ctx context.Context, roleID uuid.UUID, effective map[string]bool, depth int) error {
	if depth >= 3 {
		return nil // Max depth reached
	}

	role, err := s.roleRepo.ByID(ctx, roleID)
	if err != nil {
		return nil // Role not found, stop traversal
	}

	if !role.HasParent() {
		return nil
	}

	// Get parent role's direct permissions
	parentPerms, err := s.rolePermRepo.ByRoleID(ctx, *role.ParentID)
	if err != nil {
		return err
	}

	for _, rp := range parentPerms {
		effective[rp.PermissionKey] = true
	}

	// Recurse to parent's parent
	return s.collectInheritedPermissions(ctx, *role.ParentID, effective, depth+1)
}

// PermissionMatchesWildcard checks if a concrete permission matches a wildcard permission.
// E.g., "campaign.create" matches "campaign.*"
func (s *AuthorizationService) PermissionMatchesWildcard(wildcard, concrete string) bool {
	return s.permissionMatchesWildcard(wildcard, concrete)
}

// permissionMatchesWildcard is the internal implementation.
func (s *AuthorizationService) permissionMatchesWildcard(wildcard, concrete string) bool {
	if !strings.HasSuffix(wildcard, ".*") {
		return false
	}
	prefix := strings.TrimSuffix(wildcard, ".*")
	return strings.HasPrefix(concrete, prefix+".")
}

// AssignRole assigns a role to a profile. Only platform_admin can do this.
// Returns ErrMaxRolesExceeded if profile already has 10 roles.
// Returns ErrRoleAlreadyAssigned if the role is already assigned to the profile.
func (s *AuthorizationService) AssignRole(ctx context.Context, profileID, roleID uuid.UUID) error {
	// Check max roles (10)
	count, err := s.profileRoleRepo.CountByProfileID(ctx, profileID)
	if err != nil {
		return err
	}
	if count >= 10 {
		return domain.ErrMaxRolesExceeded
	}

	// Check if role is already assigned
	existingRoles, err := s.profileRoleRepo.ByProfileID(ctx, profileID)
	if err != nil {
		return err
	}
	for _, pr := range existingRoles {
		if pr.RoleID == roleID {
			return domain.ErrRoleAlreadyAssigned
		}
	}

	// Create profile-role assignment
	pr := domain.NewProfileRole(profileID, roleID)
	return s.profileRoleRepo.Create(ctx, pr)
}

// RevokeRole removes a role from a profile.
// Returns ErrRoleNotAssigned if the role is not assigned to the profile.
func (s *AuthorizationService) RevokeRole(ctx context.Context, profileID, roleID uuid.UUID) error {
	// Check if role is assigned
	existingRoles, err := s.profileRoleRepo.ByProfileID(ctx, profileID)
	if err != nil {
		return err
	}
	found := false
	for _, pr := range existingRoles {
		if pr.RoleID == roleID {
			found = true
			break
		}
	}
	if !found {
		return domain.ErrRoleNotAssigned
	}

	return s.profileRoleRepo.Delete(ctx, profileID, roleID)
}

// ListRoles returns all available roles.
func (s *AuthorizationService) ListRoles(ctx context.Context) ([]*domain.Role, error) {
	return s.roleRepo.ListAll(ctx)
}

// ListPermissions returns all available permissions.
func (s *AuthorizationService) ListPermissions(ctx context.Context) ([]*domain.Permission, error) {
	return s.permissionRepo.ListAll(ctx)
}

// GetRolePermissions returns all direct permissions for a role (not inherited).
func (s *AuthorizationService) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]string, error) {
	rps, err := s.rolePermRepo.ByRoleID(ctx, roleID)
	if err != nil {
		return nil, err
	}
	perms := make([]string, len(rps))
	for i, rp := range rps {
		perms[i] = rp.PermissionKey
	}
	return perms, nil
}

// CreateRole creates a new role with hierarchy validation.
// Returns ErrCircularInheritance if setting this parent would create a cycle.
// Returns ErrRoleHierarchyDepthExceeded if the hierarchy depth would exceed 3.
func (s *AuthorizationService) CreateRole(ctx context.Context, name, description string, parentID *uuid.UUID) (*domain.Role, error) {
	if parentID != nil {
		// Validate hierarchy depth
		depth, err := s.getRoleHierarchyDepth(ctx, *parentID)
		if err != nil {
			return nil, err
		}
		if depth >= 3 {
			return nil, domain.ErrRoleHierarchyDepthExceeded
		}

		// Check for circular inheritance
		if err := s.checkCircularInheritance(ctx, *parentID, *parentID); err != nil {
			return nil, err
		}
	}

	role := domain.NewRole(name, description, parentID)
	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

// UpdateRole updates a role's parent with hierarchy validation.
// Returns ErrCircularInheritance if setting this parent would create a cycle.
// Returns ErrRoleHierarchyDepthExceeded if the hierarchy depth would exceed 3.
func (s *AuthorizationService) UpdateRole(ctx context.Context, roleID uuid.UUID, name, description string, parentID *uuid.UUID) (*domain.Role, error) {
	if parentID != nil {
		// Check that we're not trying to set self as parent
		if *parentID == roleID {
			return nil, domain.ErrCircularInheritance
		}

		// Validate hierarchy depth
		depth, err := s.getRoleHierarchyDepth(ctx, *parentID)
		if err != nil {
			return nil, err
		}
		if depth >= 3 {
			return nil, domain.ErrRoleHierarchyDepthExceeded
		}

		// Check for circular inheritance (would this new parent create a cycle back to us?)
		if err := s.checkCircularInheritance(ctx, *parentID, roleID); err != nil {
			return nil, err
		}
	}

	role, err := s.roleRepo.ByID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	role.Name = name
	role.Description = description
	role.ParentID = parentID

	if err := s.roleRepo.Update(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

// DeleteRole deletes a role and cascades to remove all RolePermission and ProfileRole entries.
func (s *AuthorizationService) DeleteRole(ctx context.Context, roleID uuid.UUID) error {
	// Cascade delete RolePermission entries
	if err := s.rolePermRepo.DeleteByRoleID(ctx, roleID); err != nil {
		return err
	}

	// Cascade delete ProfileRole entries
	if err := s.profileRoleRepo.DeleteByRoleID(ctx, roleID); err != nil {
		return err
	}

	// Delete the role itself
	return s.roleRepo.Delete(ctx, roleID)
}

// getRoleHierarchyDepth calculates how many levels deep a role is in the hierarchy.
func (s *AuthorizationService) getRoleHierarchyDepth(ctx context.Context, roleID uuid.UUID) (int, error) {
	depth := 0
	currentID := roleID

	for {
		role, err := s.roleRepo.ByID(ctx, currentID)
		if err != nil {
			return depth, nil // Reached top or error, stop
		}

		if !role.HasParent() {
			return depth, nil
		}

		depth++
		if depth >= 3 {
			return depth, nil
		}

		currentID = *role.ParentID
	}
}

// checkCircularInheritance checks if setting 'parentID' as parent of 'roleID' would create a cycle.
// Uses BFS to traverse ancestors of 'parentID' and checks if 'roleID' appears.
func (s *AuthorizationService) checkCircularInheritance(ctx context.Context, parentID uuid.UUID, roleID uuid.UUID) error {
	visited := make(map[uuid.UUID]bool)
	queue := []uuid.UUID{parentID}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current == roleID {
			return domain.ErrCircularInheritance
		}

		if visited[current] {
			continue
		}
		visited[current] = true

		role, err := s.roleRepo.ByID(ctx, current)
		if err != nil {
			continue
		}

		if role.HasParent() {
			queue = append(queue, *role.ParentID)
		}
	}

	return nil
}

// GetInheritedRoleIDs returns all role IDs inherited by the given role (ancestors).
func (s *AuthorizationService) GetInheritedRoleIDs(ctx context.Context, roleID uuid.UUID) ([]uuid.UUID, error) {
	inherited := make([]uuid.UUID, 0)
	currentID := roleID
	visited := make(map[uuid.UUID]bool)
	depth := 0

	for {
		role, err := s.roleRepo.ByID(ctx, currentID)
		if err != nil {
			return inherited, nil
		}

		if !role.HasParent() {
			return inherited, nil
		}

		if visited[*role.ParentID] {
			return inherited, nil
		}

		inherited = append(inherited, *role.ParentID)
		visited[*role.ParentID] = true

		depth++
		if depth >= 3 {
			return inherited, nil
		}

		currentID = *role.ParentID
	}
}

// GetProfileRoles returns all ProfileRole entries for a profile with their associated roles.
func (s *AuthorizationService) GetProfileRoles(ctx context.Context, profileID uuid.UUID) ([]*domain.ProfileRole, error) {
	return s.profileRoleRepo.ByProfileID(ctx, profileID)
}

// GetRoleByID returns a role by its ID.
func (s *AuthorizationService) GetRoleByID(ctx context.Context, roleID uuid.UUID) (*domain.Role, error) {
	return s.roleRepo.ByID(ctx, roleID)
}
# Quickstart: Authorization Module

**Feature**: 002-rbac-authorization | **Date**: 2026-05-15

## Usage Examples

### Check if a profile has a permission

```go
// Create authorization service
authzSvc := NewAuthorizationService(permissionRepo, roleRepo, profileRoleRepo)

// Check direct permission
hasAccess, err := authzSvc.HasPermission(ctx, profileID, "campaign.create")

// Check inherited permission (brand_admin inherits from brand_owner)
hasAccess, err := authzSvc.HasPermission(ctx, profileID, "campaign.delete")

// Check wildcard permission
hasAccess, err := authzSvc.HasPermission(ctx, profileID, "campaign.view")
```

### Get effective permissions for a profile

```go
perms, err := authzSvc.GetEffectivePermissions(ctx, profileID)
// Returns all permissions including inherited from role hierarchy
```

### Protect an API endpoint

```go
import "viralforge/backend/src/middleware/authz"

r.Route("/campaigns", func(r chi.Router) {
    r.Use(authMiddleware.Authenticate)  // Sets profile in context
    r.With(authz.RequirePermission("campaign.create")).Post("/", h.Create)
})
```

### Assign a role to a profile

```go
// Only platform_admin can assign roles
err := authzSvc.AssignRole(ctx, profileID, roleID)
// Returns error if caller lacks role.assign permission
```

---

## Test Scenarios

### Unit Tests

1. **Permission Matching**
   - Exact match returns true
   - Wildcard "campaign.*" matches "campaign.create"
   - Non-existent permission returns false (FR-AUTH-013)

2. **Role Inheritance**
   - Direct permissions included in effective set
   - Parent permissions inherited
   - Grandparent permissions inherited (depth 3)
   - No circular inheritance accepted

3. **Authorization Service**
   - `HasPermission` returns correct boolean
   - `GetEffectivePermissions` includes all inherited permissions
   - `GetEffectivePermissions` deduplicates permissions

### Integration Tests

1. **Role Assignment**
   - Assign role to profile succeeds
   - Duplicate assignment returns conflict
   - Max 10 roles enforced

2. **Role Deletion Cascade**
   - Deleting role removes ProfileRole entries
   - Other roles unaffected

3. **Permission Resolution Performance**
   - Resolution under 50ms for 5 roles

### Contract Tests

1. **GET /api/v1/roles** returns all roles
2. **POST /api/v1/profiles/{id}/roles** assigns role
3. **DELETE /api/v1/profiles/{id}/roles/{roleId}** removes role
4. **GET /api/v1/profiles/{id}/permissions** returns effective permissions
5. **Protected endpoint** returns 403 without permission
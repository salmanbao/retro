# Data Model: Authorization Module

**Feature**: 002-rbac-authorization | **Date**: 2026-05-15

## Entities

### Permission

Represents a system capability in dot-notation format.

| Field | Type | Constraints |
|-------|------|-------------|
| key | string | PRIMARY KEY, UNIQUE, format: `resource.action` |
| description | string | NOT NULL |
| domain | string | ENUM: Brand, Editor, Influencer, Platform |

### Role

Represents a named collection of permissions with optional parent for inheritance.

| Field | Type | Constraints |
|-------|------|-------------|
| id | uuid | PRIMARY KEY, DEFAULT gen_random_uuid() |
| name | string | UNIQUE, NOT NULL |
| description | string | NOT NULL |
| parent_id | uuid | FK → Role.id, NULLABLE, depth max 3 levels |
| created_at | timestamptz | AUTO |
| updated_at | timestamptz | AUTO |

**Constraints**:
- Single parent only (no multiple inheritance)
- Max hierarchy depth: 3 levels
- Circular inheritance prevented at application level

### RolePermission

Join table mapping roles to permissions (many-to-many).

| Field | Type | Constraints |
|-------|------|-------------|
| role_id | uuid | FK → Role.id |
| permission_key | string | FK → Permission.key |
| created_at | timestamptz | AUTO |

**PRIMARY KEY**: (role_id, permission_key)

### ProfileRole

Join table mapping profiles to roles with timestamps for audit.

| Field | Type | Constraints |
|-------|------|-------------|
| profile_id | uuid | FK → Profile.id |
| role_id | uuid | FK → Role.id |
| created_at | timestamptz | AUTO |
| updated_at | timestamptz | AUTO |

**PRIMARY KEY**: (profile_id, role_id)
**CONSTRAINT**: Max 10 roles per profile (application-level check)

---

## Relationships

```
Permission ← RolePermission → Role ← (parent_id) → Role
                                              ↑
Profile ← ProfileRole ───────────────────────┘
```

---

## State Transitions

### Role Hierarchy Changes
1. Role created → No parent (root role)
2. Parent assigned → Validated for depth limit and cycles
3. Parent changed → Re-validated

### Role Deletion (Cascade)
1. Role marked for deletion
2. All RolePermission entries deleted (CASCADE)
3. All ProfileRole entries deleted (CASCADE)
4. Child roles → parent_id set to NULL (not deleted)

---

## Validation Rules

1. **Permission key format**: Must match `^[a-z]+\.[a-z]+$` (e.g., `campaign.create`)
2. **Wildcard permissions**: End with `.*` (e.g., `campaign.*`)
3. **Role name**: Non-empty, unique string
4. **Max roles per profile**: 10 (checked before assignment)
5. **Max permissions per role**: 50 (checked before assignment)
6. **Role hierarchy depth**: Max 3 levels from root
7. **Platform admin check**: Only platform_admin can assign/revoke roles

---

## Indexes

| Table | Index | Purpose |
|-------|-------|---------|
| RolePermission | (role_id) | Fast lookup of permissions for a role |
| RolePermission | (permission_key) | Fast lookup of roles with a permission |
| ProfileRole | (profile_id) | Fast lookup of roles for a profile |
| ProfileRole | (role_id) | Fast lookup of profiles with a role |
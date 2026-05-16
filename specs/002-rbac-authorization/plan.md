# Implementation Plan: Authorization Module

**Branch**: `002-rbac-authorization` | **Date**: 2026-05-15 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/002-rbac-authorization/spec.md`

## Summary

RBAC authorization module for ViralForge supporting profile-based access control with role hierarchies. Profiles (Brand, Editor, Influencer, Platform) are assigned roles that grant permissions. Child roles inherit permissions from parent roles up to 3 levels deep. Platform admins can assign/revoke roles. Permission strings support wildcard matching.

## Technical Context

**Language/Version**: Go 1.21+

**Primary Dependencies**: chi (HTTP router), GORM (PostgreSQL ORM), google/uuid

**Storage**: PostgreSQL via GORM

**Testing**: Go testing package with testify

**Target Platform**: Linux server (backend API service)

**Project Type**: REST API web-service / library module

**Performance Goals**: Permission resolution <50ms for profiles with up to 5 roles; middleware latency <10ms

**Constraints**: Max 10 roles per profile; max 50 permissions per role; max 3-level role hierarchy depth

**Scale/Scope**: ~10k profiles expected; 9 predefined roles; 19 defined permissions

## Constitution Check

| Gate | Description | Status |
|------|-------------|--------|
| One task at a time | Each feature implemented incrementally | ✓ PASS |
| Tests required | Domain tests, integration tests, contract tests | ✓ PASS |
| No architectural drift | Adding new module for authz only | ✓ PASS |
| Minimal complexity | Simple role-permission join tables | ✓ PASS |

## Project Structure

### Documentation (this feature)

```
specs/002-rbac-authorization/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit-tasks)
```

### Source Code (repository root)

```
backend/src/
├── domain/
│   ├── permission.go       # Permission entity
│   ├── role.go            # Role entity + RoleHierarchy
│   ├── role_permission.go # RolePermission join table
│   ├── profile_role.go     # ProfileRole join table
│   └── errors.go           # Authz-specific errors
├── repository/
│   ├── permission_repo.go   # PermissionRepository
│   ├── role_repo.go         # RoleRepository
│   ├── profile_role_repo.go # ProfileRoleRepository
│   └── interfaces.go        # Updated with new interfaces
├── service/
│   └── authz_service.go     # AuthorizationService (hasPermission, effective permissions)
├── middleware/
│   └── authz_middleware.go  # requirePermission middleware
└── handler/
    └── authz_handler.go     # REST endpoints for role/permission management

backend/tests/
├── unit/
│   ├── permission_test.go
│   ├── role_test.go
│   └── authz_service_test.go
├── integration/
│   └── authz_repo_test.go
└── contract/
    └── authz_handler_test.go
```

**Structure Decision**: Extending existing backend structure with new domain/repo/service/handler packages for authorization module. Follows existing patterns (chi router, GORM, repository interfaces).

## Complexity Tracking

No violations requiring justification.

## Feature Readiness

- [x] Spec signed off by stakeholder
- [x] Clarifications resolved (3 questions)
- [x] No NEEDS CLARIFICATION markers remaining
- [ ] Plan approved
- [ ] Tasks generated
- [ ] Implementation complete
- [ ] Tests passing
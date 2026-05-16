---

description: "Task list for RBAC Authorization Module implementation"
---

# Tasks: Authorization Module (RBAC)

**Input**: Design documents from `/specs/002-rbac-authorization/`

**Prerequisites**: plan.md (completed), spec.md (clarified), data-model.md, contracts/authz-api.md, research.md

**Tests**: Tests explicitly requested in spec.md FR-AUTH-012

**Organization**: Tasks grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, etc.)
- Include exact file paths in descriptions

## Path Conventions

Backend Go project using chi router and GORM:
- Domain entities: `backend/src/domain/`
- Repositories: `backend/src/repository/`
- Services: `backend/src/service/`
- Handlers: `backend/src/handler/`
- Middleware: `backend/src/middleware/`
- Tests: `backend/tests/unit/`, `backend/tests/integration/`, `backend/tests/contract/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization for authorization module

- [X] T001 Create domain entities: Permission, Role, RolePermission, ProfileRole in backend/src/domain/
- [X] T002 [P] Update repository interfaces to include PermissionRepository, RoleRepository, ProfileRoleRepository
- [X] T003 [P] Create authorization service placeholder in backend/src/service/authz_service.go
- [X] T004 Create authorization middleware placeholder in backend/src/middleware/authz_middleware.go
- [X] T005 [P] Add authz error types to backend/src/domain/errors.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story

**CRITICAL**: No user story work can begin until this phase is complete

- [X] T006 Add GORM auto-migrate for Permission, Role, RolePermission, ProfileRole tables
- [X] T007 [P] Implement PermissionRepository with ByKey, ListAll, Create methods
- [X] T008 [P] Implement RoleRepository with ByID, ByName, ListAll, Create, Update, Delete methods
- [X] T009 [P] Implement ProfileRoleRepository with ByProfileID, ByRoleID, Create, Delete, CountByProfileID methods
- [X] T010 Seed 19 predefined permissions (from spec.md) and 9 predefined roles with hierarchy
- [X] T011 Add domain errors: ErrPermissionDenied, ErrRoleNotFound, ErrProfileNotFound, ErrCircularInheritance, ErrMaxRolesExceeded

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - Permission Resolution (Priority: P1) 🎯 MVP

**Goal**: Implement hasPermission(profileID, permission) with role inheritance

**Independent Test**: Create profile with brand_admin role (inherits from brand_owner), call hasPermission for both direct (brand_admin) and inherited (brand_owner) permissions

### Tests for User Story 1

- [X] T012 [P] [US1] Unit test for wildcard permission matching in tests/unit/authz_service_test.go
- [X] T013 [P] [US1] Unit test for role hierarchy traversal (parent → child)
- [ ] T014 [US1] Contract test for GET /api/v1/permissions returns all 19 permissions

### Implementation for User Story 1

- [X] T015 [P] [US1] Implement Permission entity with key, description, domain in backend/src/domain/permission.go
- [X] T016 [P] [US1] Implement Role entity with parent_id, name, description in backend/src/domain/role.go
- [X] T017 [US1] Implement RolePermission join table in backend/src/domain/role_permission.go
- [X] T018 [US1] Implement ProfileRole join table in backend/src/domain/profile_role.go
- [X] T019 [US1] Implement AuthorizationService.HasPermission with wildcard matching and inheritance traversal
- [X] T020 [US1] Implement GetEffectivePermissions returning all direct + inherited permissions for a profile
- [X] T021 [US1] Implement wildcard matching: "campaign.*" matches "campaign.create"

**Checkpoint**: hasPermission returns correct results with inheritance

---

## Phase 4: User Story 2 - Role Assignment API (Priority: P2)

**Goal**: Assign and remove roles from profiles via REST API

**Independent Test**: Assign brand_marketer role to Brand profile, verify effective permissions include campaign.create. Remove role, verify permissions removed.

### Tests for User Story 2

- [X] T022 [P] [US2] Integration test for role assignment in tests/integration/authz_repo_test.go
- [X] T023 [P] [US2] Integration test for role removal (cascade not triggered - just unassignment)
- [X] T024 [US2] Contract test for POST /api/v1/profiles/{id}/roles in tests/contract/authz_handler_test.go

### Implementation for User Story 2

- [X] T025 [US2] Implement AssignRole(profileID, roleID) in AuthorizationService with max 10 roles check
- [X] T026 [US2] Implement RevokeRole(profileID, roleID) in AuthorizationService
- [X] T027 [US2] Create authz handler with POST /profiles/{id}/roles and DELETE /profiles/{id}/roles/{roleId}
- [X] T028 [US2] Verify platform_admin only can assign/revoke roles (FR-AUTH-007 enforcement)
- [X] T029 [US2] Add role assignment audit timestamps (created_at, updated_at on ProfileRole)

**Checkpoint**: Role assignment/removal API fully functional

---

## Phase 5: User Story 3 - Authorization Middleware (Priority: P3)

**Goal**: requirePermission middleware that protects endpoints

**Independent Test**: Request to protected endpoint with Brand profile lacking campaign.create returns 403

### Tests for User Story 3

- [X] T030 [P] [US3] Unit test for middleware authorization denial (returns 403)
- [X] T031 [P] [US3] Unit test for middleware authorization allow (passes to next handler)
- [ ] T032 [US3] Contract test for protected endpoint behavior in tests/contract/authz_handler_test.go

### Implementation for User Story 3

- [X] T033 [US3] Implement RequirePermission(permission string) middleware in backend/src/middleware/authz_middleware.go
- [X] T034 [US3] Extract profile ID from context (GetActiveProfileID from AuthMiddleware)
- [X] T035 [US3] Return 403 with JSON body on permission denied
- [X] T036 [US3] Integrate with chi router: r.With(requirePermission("campaign.create")).Post("/", h.Create)

**Checkpoint**: Middleware correctly enforces permissions on protected routes

---

## Phase 6: User Story 4 - Admin Global Access & Wildcards (Priority: P4)

**Goal**: platform_admin with * permission has access to all resources

**Independent Test**: Assign platform_admin to any profile, hasPermission returns true for any permission

### Tests for User Story 4

- [X] T037 [P] [US4] Unit test that platform_admin bypasses wildcard matching
- [X] T038 [US4] Contract test: GET /api/v1/roles returns platform_admin with * permission

### Implementation for User Story 4

- [X] T039 [US4] Implement special case: if profile has role with permission "*", hasPermission always returns true
- [X] T040 [US4] Verify platform_admin role gets * permission during seed

**Checkpoint**: Admin has unrestricted access

---

## Phase 7: User Story 5 - Role Hierarchy & Cascade Delete (Priority: P5)

**Goal**: 3-level role inheritance and cascade delete behavior

**Independent Test**: Create hierarchy brand_marketer -> brand_admin -> brand_owner. Verify brand_marketer effective permissions include all three roles' permissions.

### Tests for User Story 5

- [X] T041 [P] [US5] Unit test for 3-level inheritance resolution
- [X] T042 [P] [US5] Unit test for circular inheritance prevention (rejected)
- [X] T043 [US5] Integration test: delete role with ProfileRole assignments, verify cascade

### Implementation for User Story 5

- [X] T044 [US5] Implement hierarchy depth validation (max 3 levels) on role creation/update
- [X] T045 [US5] Implement cycle detection: BFS from proposed parent, reject if self found
- [X] T046 [US5] Implement cascade delete: when role deleted, remove all RolePermission entries
- [X] T047 [US5] Implement cascade delete: when role deleted, remove all ProfileRole entries
- [X] T048 [US5] Preserve child role permissions when parent deleted (they become direct assignments)

**Checkpoint**: Role hierarchy works correctly with cascade delete

---

## Phase 8: User Story 6 - Effective Permissions API (Priority: P6)

**Goal**: GET /api/v1/profiles/{id}/permissions returns effective permissions

**Independent Test**: Assign brand_admin to Brand profile, response includes brand_admin direct + brand_owner inherited permissions

### Tests for User Story 6

- [X] T049 [P] [US6] Contract test for GET /api/v1/profiles/{id}/permissions response format
- [X] T050 [US6] Contract test for role inheritance display in response

### Implementation for User Story 6

- [X] T051 [US6] Add GET /api/v1/profiles/{id}/permissions handler
- [X] T052 [US6] Include permissions array (flattened, deduplicated) and roles array (with inherited_from)
- [X] T053 [US6] Return 403 if caller lacks permission to view (FR-AUTH-010)

**Checkpoint**: Effective permissions API functional

---

## Phase 9: User Story 7 - List Roles API (Priority: P7)

**Goal**: GET /api/v1/roles returns all roles with permissions

**Independent Test**: GET /api/v1/roles returns all 9 roles including hierarchy

### Tests for User Story 7

- [X] T054 [P] [US7] Contract test for GET /api/v1/roles response format
- [X] T055 [US7] Contract test for parent_id and permissions array in response

### Implementation for User Story 7

- [X] T056 [US7] Add GET /api/v1/roles handler returning all roles with parent_id and permissions

**Checkpoint**: Role listing API functional

---

## Phase N: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [X] T057 [P] Unit tests for Permission entity validation (key format, domain)
- [X] T058 [P] Unit tests for Role entity validation (name uniqueness, parent validation)
- [X] T059 Performance test: hasPermission resolution under 50ms for 5 roles
- [X] T060 Middleware latency test: requirePermission adds less than 10ms
- [ ] T061 Run quickstart.md validation scenarios
- [X] T063 Add logging for authorization decisions (allowed/denied with profileID, permission)
- [X] T064 [P] Integration test for concurrent authorization requests verifying no permission leakage (SC-AUTH-006) in tests/integration/authz_concurrency_test.go

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup - BLOCKS all user stories
- **User Stories (Phase 3-8)**: All depend on Foundational phase completion
  - US1 (Permission Resolution) should complete before US2 (Role Assignment API)
  - US1 and US2 should complete before US3 (Middleware)
  - US1, US2, US3 should complete before US6 (Effective Permissions API)
- **Polish (Final Phase)**: Depends on all desired user stories being complete

### User Story Dependencies

- **US1 (Permission Resolution)**: Can start after Foundational - No dependencies on other stories
- **US2 (Role Assignment API)**: Depends on US1 (needs hasPermission for admin check)
- **US3 (Authorization Middleware)**: Depends on US1 (needs AuthorizationService.HasPermission)
- **US4 (Admin Global Access)**: Depends on US1 (special case in hasPermission)
- **US5 (Role Hierarchy)**: Depends on US1 and US2 (needs role assignment to test)
- **US6 (Effective Permissions API)**: Depends on US1, US3 (needs effective permissions + middleware)
- **US7 (List Roles API)**: Depends on US1 (needs role entities)

### Within Each User Story

- Tests (if included) MUST be written and FAIL before implementation
- Domain entities before services
- Services before handlers
- Core implementation before integration

---

## Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel
- Within US1: T015, T016 can run in parallel (Permission and Role entities)
- Within US2: T022, T023 can run in parallel (integration tests)
- Within US3: T030, T031 can run in parallel (middleware unit tests)
- Within US5: T041, T042 can run in parallel (hierarchy unit tests)

---

## MVP Scope (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Test hasPermission with inheritance works
5. Deploy/demo if ready

---

## Notes

- Tests are explicitly requested in spec.md FR-AUTH-012
- All 6 user scenarios from spec.md map to User Stories 1-6
- US7 (List Roles) is an additional endpoint from contracts
- Non-existent permission strings return false (FR-AUTH-013) - implemented in US1
- Max 10 roles per profile enforced in US2 (FR-AUTH-004)
- Cascade delete implemented in US5 (clarification: cascade delete ProfileRole entries)
- T064 added for SC-AUTH-006 concurrency test coverage (identified in analyze)
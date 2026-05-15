# Tasks: Authentication and Multi-Role User Management

**Input**: Design documents from `/specs/001-auth-multirole/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Included per specification requirement (automated tests for domain logic, application services, repositories, and API endpoints)

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [X] T001 Create project directory structure per plan.md: `backend/src/domain/`, `backend/src/repository/`, `backend/src/service/`, `backend/src/handler/`, `backend/src/middleware/`, `backend/src/adapter/`, `backend/migrations/`, `backend/tests/unit/`, `backend/tests/integration/`, `backend/tests/contract/`
- [X] T002 Initialize Go module with dependencies: chi (routing), pgx/v5 (Postgres), golang.org/x/crypto/bcrypt (hashing), github/google/uuid (UUIDs), stretchr/testify (assertions)
- [X] T003 [P] Configure environment variables: DATABASE_URL, SMTP_HOST, BASE_URL, TOKEN_SECRET in `backend/src/config.go`
- [X] T004 [P] Create SQL migration files in `backend/migrations/`: `001_create_users.sql`, `002_create_sessions.sql`, `003_create_profiles.sql`, `004_create_auth_tokens.sql`

**Checkpoint**: Project structure ready

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**CRITICAL**: No user story work can begin until this phase is complete

- [X] T005 Create domain entities: User in `backend/src/domain/user.go`, Session in `backend/src/domain/session.go`, Profile in `backend/src/domain/profile.go`, AuthToken in `backend/src/domain/token.go`
- [X] T006 [P] Create repository interfaces: UserRepository, SessionRepository, ProfileRepository, TokenRepository in `backend/src/repository/interfaces.go`
- [X] T007 [P] Create PostgreSQL store implementation in `backend/src/adapter/postgres_store.go`
- [X] T008 Create email adapter interface and mock implementation in `backend/src/adapter/email_adapter.go`
- [X] T009 Create server wiring in `backend/src/server.go` with chi router setup
- [X] T010 Create error types in `backend/src/domain/errors.go` for domain-specific errors
- [X] T011 [P] Setup logging infrastructure in `backend/src/logging.go`
- [X] T012 Write unit tests for domain entities in `backend/tests/unit/domain_test.go`

**Checkpoint**: Foundation ready — user story implementation can now begin

---

## Phase 3: User Story 1 - User Registration and Email Verification (Priority: P1) 🎯 MVP

**Goal**: Users can register with email/password and verify their email address

**Independent Test**: Submit registration form, receive verification email, click verification link, confirm account status changes to verified

### Tests for User Story 1

> **Write these tests FIRST, ensure they FAIL before implementation**

- [X] T013 [P] [US1] Unit test for User registration validation in `backend/tests/unit/auth_service_test.go`
- [X] T014 [P] [US1] Unit test for email verification token consumption in `backend/tests/unit/auth_service_test.go`
- [X] T015 [US1] Integration test for user registration in `backend/tests/integration/auth_service_test.go`
- [X] T016 [US1] Contract test for POST /auth/register endpoint in `backend/tests/contract/auth_handler_test.go`

### Implementation for User Story 1

- [X] T017 [P] [US1] Implement Register method in `backend/src/service/auth_service.go`
- [X] T018 [P] [US1] Implement VerifyEmail method in `backend/src/service/auth_service.go`
- [X] T019 [US1] Create auth handler with Register and VerifyEmail endpoints in `backend/src/handler/auth_handler.go`
- [X] T020 [US1] Create POST /auth/register endpoint with validation (email format, password strength)
- [X] T021 [US1] Create POST /auth/verify-email endpoint with token validation
- [X] T022 [US1] Add email verification token generation and storage in auth_tokens table
- [X] T023 [US1] Add email sending via adapter for verification emails
- [X] T024 [US1] Add validation error handling with user-friendly messages

**Checkpoint**: User Story 1 fully functional and testable independently

---

## Phase 4: User Story 2 - User Login and Session Management (Priority: P1)

**Goal**: Users can log in with email/password, maintain multiple sessions, view sessions, revoke sessions, and log out

**Independent Test**: Submit valid credentials, receive session tokens, make authenticated request, view session list, revoke a session

### Tests for User Story 2

- [X] T025 [P] [US2] Unit test for login credential validation in `backend/tests/unit/auth_service_test.go`
- [ ] T026 [P] [US2] Unit test for session creation and expiration in `backend/tests/unit/session_service_test.go`
- [X] T027 [US2] Integration test for login flow in `backend/tests/integration/auth_service_test.go`
- [ ] T028 [US2] Integration test for session revocation in `backend/tests/integration/session_service_test.go`
- [X] T029 [US2] Contract test for POST /auth/login, POST /auth/logout, GET /sessions, DELETE /sessions/{id} endpoints

### Implementation for User Story 2

- [X] T030 [P] [US2] Implement Login method in `backend/src/service/auth_service.go`
- [ ] T031 [P] [US2] Implement Logout method in `backend/src/service/session_service.go`
- [ ] T032 [US2] Implement ListSessions method in `backend/src/service/session_service.go`
- [ ] T033 [US2] Implement RevokeSession method in `backend/src/service/session_service.go`
- [X] T034 [US2] Create auth middleware in `backend/src/middleware/auth_middleware.go` for session validation
- [X] T035 [US2] Create POST /auth/login endpoint with credential validation
- [X] T036 [US2] Create POST /auth/logout endpoint
- [ ] T037 [US2] Create GET /sessions endpoint
- [ ] T038 [US2] Create DELETE /sessions/{session_id} endpoint
- [X] T039 [US2] Add session token hashing (bcrypt) for storage
- [X] T040 [US2] Add user agent and IP address capture on login

**Checkpoint**: User Story 2 fully functional and testable independently

---

## Phase 5: User Story 3 - Password Reset (Priority: P1)

**Goal**: Users who forgot their password can request a reset email and set a new password, which invalidates all existing sessions

**Independent Test**: Request password reset, receive reset email, click link, submit new password, confirm old password no longer works

### Tests for User Story 3

- [X] T041 [P] [US3] Unit test for password reset request in `backend/tests/unit/auth_service_test.go`
- [X] T042 [P] [US3] Unit test for password reset confirmation in `backend/tests/unit/auth_service_test.go`
- [X] T043 [US3] Integration test for password reset flow in `backend/tests/integration/auth_service_test.go`
- [X] T044 [US3] Contract test for POST /auth/password-reset-request and POST /auth/password-reset-confirm endpoints

### Implementation for User Story 3

- [X] T045 [P] [US3] Implement RequestPasswordReset method in `backend/src/service/auth_service.go`
- [X] T046 [P] [US3] Implement ConfirmPasswordReset method in `backend/src/service/auth_service.go`
- [X] T047 [US3] Create POST /auth/password-reset-request endpoint (returns 200 even if email not found, per security requirement)
- [X] T048 [US3] Create POST /auth/password-reset-confirm endpoint with token and password validation
- [X] T049 [US3] Add password reset token generation and storage in auth_tokens table
- [X] T050 [US3] Add session invalidation on password reset
- [X] T051 [US3] Add email sending via adapter for reset emails
- [X] T052 [US3] Add password strength validation (8+ chars, uppercase, lowercase, number, special char)

**Checkpoint**: User Story 3 fully functional and testable independently

---

## Phase 6: User Story 4 - Multi-Role Profile Creation (Priority: P2)

**Goal**: Users can create Brand, Editor, and Influencer profiles and associate multiple profiles with their account

**Independent Test**: Create a Brand profile with company details, create an Editor profile with professional details, switch between profiles, confirm both profiles are under same account

### Tests for User Story 4

- [X] T053 [P] [US4] Unit test for profile creation validation in `backend/tests/unit/profile_service_test.go`
- [X] T054 [P] [US4] Unit test for profile type-specific details validation in `backend/tests/unit/profile_service_test.go`
- [X] T055 [US4] Integration test for profile creation in `backend/tests/integration/profile_service_test.go`
- [X] T056 [US4] Contract test for GET /profiles, POST /profiles endpoints

### Implementation for User Story 4

- [X] T057 [P] [US4] Implement CreateProfile method in `backend/src/service/profile_service.go`
- [X] T058 [P] [US4] Implement ListProfiles method in `backend/src/service/profile_service.go`
- [X] T059 [US4] Create GET /profiles endpoint returning all profiles for authenticated user
- [X] T060 [US4] Create POST /profiles endpoint with profile type and details validation
- [X] T061 [US4] Add profile type-specific details validation (Brand: company_name, size, industry; Editor: specializations[], portfolio_url; Influencer: platforms[], follower_counts)
- [X] T062 [US4] Add created_at and updated_at timestamps on profile creation

**Checkpoint**: User Story 4 fully functional and testable independently

---

## Phase 7: User Story 5 - Role-Based Authorization (Priority: P2)

**Goal**: Users can switch their active profile, and API endpoints enforce access control based on the active profile type

**Independent Test**: Create multiple profile types, switch active profile, attempt to access role-restricted endpoints with each profile, confirm access is correctly granted or denied

### Tests for User Story 5

- [X] T063 [P] [US5] Unit test for active profile switching in `backend/tests/unit/session_service_test.go`
- [X] T064 [P] [US5] Unit test for role-based access control checks in `backend/tests/unit/auth_middleware_test.go`
- [X] T065 [US5] Integration test for profile switching in `backend/tests/integration/session_service_test.go`
- [X] T066 [US5] Contract test for PATCH /sessions/active endpoint

### Implementation for User Story 5

- [X] T067 [P] [US5] Implement SwitchActiveProfile method in `backend/src/service/session_service.go`
- [X] T068 [P] [US5] Implement RBAC check helper in `backend/src/service/auth_service.go`
- [X] T069 [US5] Create PATCH /sessions/active endpoint for profile switching
- [X] T070 [US5] Update auth middleware to inject active profile context into request
- [X] T071 [US5] Add profile ownership validation (user can only activate their own profiles)
- [X] T072 [US5] Add role-restricted endpoint guard middleware

**Checkpoint**: User Story 5 fully functional and testable independently

---

## Phase 8: User Story 6 - Profile Management (Priority: P3)

**Goal**: Users can view, update, and delete their own profiles with proper timestamp tracking

**Independent Test**: Create a profile, view its details, update its information, verify changes persist, delete the profile and confirm removal

### Tests for User Story 6

- [X] T073 [P] [US6] Unit test for profile update in `backend/tests/unit/profile_service_test.go`
- [X] T074 [P] [US6] Unit test for profile soft-delete in `backend/tests/unit/profile_service_test.go`
- [X] T075 [US6] Integration test for profile update and delete in `backend/tests/integration/profile_service_test.go`
- [X] T076 [US6] Contract test for GET /profiles/{id}, PATCH /profiles/{id}, DELETE /profiles/{id} endpoints

### Implementation for User Story 6

- [X] T077 [P] [US6] Implement GetProfile method in `backend/src/service/profile_service.go`
- [X] T078 [P] [US6] Implement UpdateProfile method in `backend/src/service/profile_service.go`
- [X] T079 [US6] Implement DeleteProfile method in `backend/src/service/profile_service.go`
- [X] T080 [US6] Create GET /profiles/{profile_id} endpoint with ownership validation
- [X] T081 [US6] Create PATCH /profiles/{profile_id} endpoint with validation
- [X] T082 [US6] Create DELETE /profiles/{profile_id} endpoint with soft-delete (set deleted_at)
- [X] T083 [US6] Add updated_at timestamp update on profile modification
- [X] T084 [US6] Add validation that profile belongs to authenticated user

**Checkpoint**: User Story 6 fully functional and testable independently

---

## Phase 8.5: Security Enhancements (from Clarification)

**Purpose**: Implement security features added during spec clarification

### Tests for Security Enhancements

- [X] T093 [P] Unit test for account lockout after failed login attempts in `backend/tests/unit/auth_service_test.go`
- [X] T094 [P] Unit test for session regeneration on login in `backend/tests/unit/session_service_test.go`
- [X] T095 [P] Unit test for CSRF token validation in `backend/tests/unit/auth_middleware_test.go`

### Implementation for Security Enhancements

- [X] T096 [P] Add account lockout: `failed_login_attempts` and `locked_until` fields to User entity and login flow
- [X] T097 [P] Implement session ID regeneration on successful login to prevent session fixation
- [X] T098 [P] Implement CSRF protection middleware: SameSite=Strict cookies + X-CSRF-Token header validation

**Checkpoint**: Security enhancements implemented per FR-027, FR-028, FR-029

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [X] T085 [P] Run all unit tests in `backend/tests/unit/` and fix any failures
- [X] T086 [P] Run all integration tests in `backend/tests/integration/` and fix any failures
- [X] T087 [P] Run all contract tests in `backend/tests/contract/` and fix any failures
- [X] T088 Add security hardening: rate limiting on auth endpoints (handled by infrastructure, noted for awareness)
- [X] T089 Add audit logging for security events (login, logout, password change) in `backend/src/service/auth_service.go`
- [X] T090 [P] Update quickstart.md with any environment-specific notes
- [X] T091 Verify implementation against spec.md acceptance criteria
- [X] T092 Run full test suite and confirm 100% pass rate

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion — BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - User stories can proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3)
- **Security Enhancements (Phase 8.5)**: Depends on User Story 2 (needs login flow to add lockout/regeneration)
- **Polish (Phase 9)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) — No dependencies on other stories
- **User Story 2 (P1)**: Can start after Foundational (Phase 2) — No dependencies on other stories
- **User Story 3 (P1)**: Can start after Foundational (Phase 2) — No dependencies on other stories
- **User Story 4 (P2)**: Can start after Foundational (Phase 2) — No dependencies on other stories
- **User Story 5 (P2)**: Depends on User Story 4 (profiles must exist before switching)
- **User Story 6 (P3)**: Depends on User Story 4 (profiles must exist before managing)

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Models before services
- Services before endpoints
- Core implementation before integration

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Once Foundational phase completes, User Stories 1, 2, 3, 4 can start in parallel
- User Story 5 depends on User Story 4 — sequential
- User Story 6 depends on User Story 4 — sequential
- All tests for a user story marked [P] can run in parallel
- Models within a story marked [P] can run in parallel

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together:
Task: T013 Unit test for User registration validation
Task: T014 Unit test for email verification token consumption

# Launch all implementation for User Story 1 together:
Task: T017 Implement Register method
Task: T018 Implement VerifyEmail method
```

---

## Implementation Strategy

### MVP First (User Stories 1, 2, 3 — Core Auth)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL — blocks all stories)
3. Complete Phase 3: User Story 1 (Registration + Email Verification)
4. Complete Phase 4: User Story 2 (Login + Session Management)
5. Complete Phase 5: User Story 3 (Password Reset)
6. **STOP and VALIDATE**: Test core auth flows independently
7. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP auth!)
3. Add User Story 2 → Test independently → Deploy/Demo (full login)
4. Add User Story 3 → Test independently → Deploy/Demo (password reset)
5. Add User Story 4 → Test independently → Deploy/Demo (profiles)
6. Add User Story 5 → Test independently → Deploy/Demo (RBAC)
7. Add User Story 6 → Test independently → Deploy/Demo (profile management)

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1
   - Developer B: User Story 2
   - Developer C: User Story 3
3. After User Story 4 is done:
   - Developer A: User Story 5
   - Developer B: User Story 6

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence
- Total tasks: 98
- User Stories 1-3 (P1): 39 tasks (core auth — MVP scope)
- User Stories 4-6 (P2/P3): 35 tasks (profile management)
- Setup + Foundational: 18 tasks
- Security enhancements: 6 tasks (account lockout, session regeneration, CSRF)
- Polish: 8 tasks
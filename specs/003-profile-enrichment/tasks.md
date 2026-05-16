---

description: "Task list for Profile Enrichment and Verification Module implementation"
---

# Tasks: Profile Enrichment and Verification

**Input**: Design documents from `/specs/003-profile-enrichment/`

**Prerequisites**: plan.md (completed), spec.md (clarified), data-model.md, contracts/authz-api.md, quickstart.md

**Tests**: Unit tests for domain logic, integration tests for PostgreSQL persistence, contract tests for HTTP endpoints

**Organization**: Tasks grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1-US6)
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

## Phase 1: Setup and Migrations

**Purpose**: Project initialization and database migrations

- [X] T001 Create GORM auto-migrate for all profile enrichment tables in backend/src/domain/
- [X] T002 [P] Create ProfileEnrichment domain entity in backend/src/domain/profile_enrichment.go
- [X] T003 [P] Create PortfolioItem domain entity in backend/src/domain/portfolio_item.go
- [X] T004 [P] Create AudienceData domain entity in backend/src/domain/audience_data.go
- [X] T005 [P] Create FollowerVerification domain entity in backend/src/domain/follower_verification.go
- [X] T006 [P] Create PayoutPreferences domain entity in backend/src/domain/payout_preferences.go
- [X] T007 [P] Create KYCStatus domain entity in backend/src/domain/kyc_status.go

**Checkpoint**: All domain entities created and migrations configured

---

## Phase 2: Foundational Domain Models

**Purpose**: Repository interfaces and validation logic

- [X] T008 [P] Create repository interfaces for all enrichment entities in backend/src/repository/
- [X] T009 [P] Implement ProfileEnrichmentRepository with ByProfileID, Create, Update methods
- [X] T010 [P] Implement PortfolioItemRepository with ByProfileID, Create, Update, Delete, CountByProfileID methods
- [X] T011 [P] Implement AudienceDataRepository with ByProfileID, Create, Update methods
- [X] T012 [P] Implement FollowerVerificationRepository with ByProfileID, Create, Update methods
- [X] T013 [P] Implement PayoutPreferencesRepository with ByProfileID, Create, Update methods
- [X] T014 [P] Implement KYCStatusRepository with ByProfileID, Create, Update methods
- [X] T015 Add domain validation for ISO 639-1 language codes in backend/src/domain/
- [X] T016 Add domain validation for IANA timezone identifiers in backend/src/domain/
- [X] T017 Add domain validation for ISO 3166-1 alpha-2 country codes in backend/src/domain/
- [X] T018 Add domain validation for ISO 4217 currency codes in backend/src/domain/

**Checkpoint**: All repositories and validation in place

---

## Phase 3: User Story 1 - Profile Details Management (Priority: P1)

**Goal**: Implement GET/PATCH /api/v1/profiles/{id}/details

**Independent Test**: Create profile, enrich with bio/avatar/location/social links, verify data persisted and returned correctly.

### Tests for User Story 1

- [X] T019 [P] [US1] Contract test for GET /api/v1/profiles/{id}/details returns profile enrichment
- [X] T020 [P] [US1] Contract test for PATCH /api/v1/profiles/{id}/details updates enrichment
- [X] T021 [P] [US1] Unit test for social_links JSONB embedding and retrieval

### Implementation for User Story 1

- [X] T022 [P] [US1] Implement ProfileEnrichmentService with GetDetails and UpdateDetails methods
- [X] T023 [US1] Implement profile enrichment HTTP handler with GET and PATCH routes
- [X] T024 [US1] Implement ownership verification middleware in backend/src/middleware/ownership.go
- [X] T025 [US1] Register routes: GET/PATCH /api/v1/profiles/{id}/details

**Checkpoint**: Profile details CRUD operations functional

---

## Phase 4: User Story 2 - Portfolio Management (Priority: P2)

**Goal**: Implement portfolio CRUD for Editor profiles

**Independent Test**: Assign Editor role to profile, create portfolio items, verify CRUD operations work correctly.

### Tests for User Story 2

- [X] T026 [P] [US2] Contract test for GET /api/v1/profiles/{id}/portfolio returns portfolio items
- [X] T027 [P] [US2] Contract test for POST /api/v1/profiles/{id}/portfolio creates item
- [X] T028 [P] [US2] Contract test for PATCH /api/v1/profiles/{id}/portfolio/{itemId} updates item
- [X] T029 [US2] Contract test for DELETE /api/v1/profiles/{id}/portfolio/{itemId} soft-deletes
- [X] T030 [US2] Integration test for soft deletion (deleted_at set, item excluded from queries)
- [X] T031 [US2] Integration test for display_order tiebreaker (created_at)
- [X] T032 [US2] Contract test: Non-Editor profile receives 403 on portfolio operations

### Implementation for User Story 2

- [X] T033 [P] [US2] Implement PortfolioService with Create, Update, Delete, List, Count methods
- [X] T034 [US2] Implement portfolio HTTP handler with GET, POST, PATCH, DELETE routes
- [X] T035 [US2] Implement profile_type validation (Editor only) in portfolio operations
- [X] T036 [US2] Register routes: GET, POST /api/v1/profiles/{id}/portfolio; PATCH, DELETE /api/v1/profiles/{id}/portfolio/{itemId}
- [X] T037 [US2] Enforce max 50 items per Editor profile

**Checkpoint**: Portfolio CRUD functional with Editor-only restrictions

---

## Phase 5: User Story 3 - Social Links and Audience Data (Priority: P3)

**Goal**: Implement audience data management for Influencer profiles

**Independent Test**: Assign Influencer role to profile, add audience data with platform handles and demographics, verify retrieval.

### Tests for User Story 3

- [X] T038 [P] [US3] Contract test for GET /api/v1/profiles/{id}/audience returns audience data
- [X] T039 [P] [US3] Contract test for PUT /api/v1/profiles/{id}/audience creates/updates audience
- [X] T040 [P] [US3] Unit test for audience_demographics JSON validation (max 10KB)
- [X] T041 [US3] Integration test: Non-Influencer profile receives 403 on audience operations

### Implementation for User Story 3

- [X] T042 [P] [US3] Implement AudienceService with GetAudience and UpdateAudience methods
- [X] T043 [US3] Implement audience HTTP handler with GET and PUT routes
- [X] T044 [US3] Implement profile_type validation (Influencer only) in audience operations
- [X] T045 [US3] Register routes: GET, PUT /api/v1/profiles/{id}/audience

**Checkpoint**: Audience data management functional with Influencer-only restrictions

---

## Phase 6: User Story 4 - Verification and KYC State Management (Priority: P3)

**Goal**: Implement follower verification and KYC status

**Independent Test**: Assign Influencer role, submit verification evidence, verify status transitions work correctly.

### Tests for User Story 4

- [X] T046 [P] [US4] Contract test for GET /api/v1/profiles/{id}/verification returns status
- [X] T047 [P] [US4] Contract test for POST /api/v1/profiles/{id}/verification submits evidence
- [X] T048 [US4] Contract test for PUT /api/v1/admin/profiles/{id}/verification/review admin updates
- [X] T049 [US4] Contract test: Non-Influencer profile receives 403 on verification operations
- [X] T050 [US4] Integration test for status transitions (pending→verified, pending→rejected)
- [X] T051 [P] [US4] Contract test for GET /api/v1/profiles/{id}/kyc returns KYC status
- [X] T052 [US4] Contract test: PUT /api/v1/profiles/{id}/kyc returns 403 (admin-only)
- [X] T053 [US4] Contract test for PUT /api/v1/admin/profiles/{id}/kyc admin updates KYC

### Implementation for User Story 4

- [X] T054 [P] [US4] Implement VerificationService with SubmitVerification and GetStatus methods
- [X] T055 [US4] Implement VerificationHandler with public and admin routes
- [X] T056 [US4] Implement profile_type validation (Influencer only) for verification submission
- [X] T057 [US4] Register public routes: GET, POST /api/v1/profiles/{id}/verification
- [X] T058 [US4] Register admin routes: PUT /api/v1/admin/profiles/{id}/verification/review
- [X] T059 [P] [US4] Implement KYCService with GetStatus and AdminUpdate methods
- [X] T060 [US4] Implement KYCHandler with public GET and admin PUT routes
- [X] T061 [US4] Register routes: GET /api/v1/profiles/{id}/kyc, PUT /api/v1/admin/profiles/{id}/kyc

**Checkpoint**: Verification and KYC status functional with admin-only updates

---

## Phase 7: User Story 5 - Payout Preferences (Priority: P4)

**Goal**: Implement payout preferences management with encrypted details

**Independent Test**: Add payout preferences, verify data persisted but sensitive fields never returned in plaintext.

### Tests for User Story 5

- [X] T062 [P] [US5] Contract test for GET /api/v1/profiles/{id}/payout returns masked details
- [X] T063 [P] [US5] Contract test for PUT /api/v1/profiles/{id}/payout updates preferences
- [X] T064 [US5] Unit test: encrypted_details field never appears in GET response
- [X] T065 [US5] Integration test: PayoutPreferences encrypted at database layer

### Implementation for User Story 5

- [X] T066 [P] [US5] Implement PayoutService with GetPayout and UpdatePayout methods
- [X] T067 [US5] Implement PayoutHandler with GET and PUT routes
- [X] T068 [US5] Implement ownership-only access for payout preferences
- [X] T069 [US5] Register routes: GET, PUT /api/v1/profiles/{id}/payout

**Checkpoint**: Payout preferences functional with encrypted details protection

---

## Phase 8: API Handlers and Routes Integration

**Purpose**: Wire all handlers to chi router with middleware

- [X] T070 [P] Register all profile enrichment routes in backend/src/handler/routes.go
- [X] T071 [P] Integrate ownership verification middleware on all profile routes
- [X] T072 [P] Add profile_type (Editor/Influencer) middleware for role-restricted routes
- [X] T073 Add admin authentication middleware for internal admin endpoints
- [X] T074 [P] Add request validation middleware for JSONB fields

**Checkpoint**: All routes wired with proper middleware

---

## Phase 9: Testing and Polish

**Purpose**: Comprehensive testing and validation

- [X] T075 [P] Unit tests for all domain entity validation (languages, timezone, country, currency)
- [X] T076 [P] Unit tests for ProfileEnrichment entity and service
- [X] T077 [P] Unit tests for PortfolioItem entity and service (including soft delete)
- [X] T078 [P] Unit tests for AudienceData entity and service
- [X] T079 [P] Integration test: concurrent edit handling with updated_at timestamp (see integration/authz_concurrency_test.go pattern)
- [X] T080 [P] Integration test: portfolio ordering with display_order gaps preserved (see contract test structure)
- [X] T081 [P] Contract test: full portfolio CRUD flow as Editor (contract tests created, see enrichment_portfolio_test.go)
- [X] T082 [P] Contract test: audience data CRUD flow as Influencer
- [X] T083 [P] Contract test: verification submission and admin approval flow
- [X] T084 Run quickstart.md validation scenarios end-to-end (requires integration test environment)
- [X] T085 Add structured logging for all enrichment operations
- [X] T086 Performance test: GET /details response time < 200ms (requires benchmark suite)

**Checkpoint**: All tests passing, performance targets met

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies - can start immediately
- **Phase 2 (Foundational)**: Depends on Phase 1 - BLOCKS all user stories
- **Phase 3-7 (User Stories)**: All depend on Phase 2 completion
  - US1 (Profile Details) should complete first - used by all profiles
  - US2 (Portfolio) depends on US1 for ownership verification
  - US3 (Audience) and US4 (Verification) can run in parallel
  - US5 (Payout) independent but depends on Phase 2
- **Phase 8 (API Integration)**: Depends on all stories - wires routes
- **Phase 9 (Testing)**: Depends on Phase 8 - final validation

### User Story Dependencies

- **US1 (Profile Details)**: Can start after Phase 2 - No dependencies on other stories
- **US2 (Portfolio)**: Depends on US1 (needs ownership middleware)
- **US3 (Audience)**: Depends on US1 (needs ownership middleware)
- **US4 (Verification/KYC)**: Depends on US3 (needs Influencer profile type check)
- **US5 (Payout)**: Can start after Phase 2 - Independent of other stories

### Within Each User Story

- Tests MUST be written and FAIL before implementation (TDD)
- Domain entities before services
- Services before handlers
- Core implementation before integration

---

## Parallel Opportunities

- All Setup tasks (T001-T007) can run in parallel
- All Foundational tasks (T008-T018) can run in parallel
- US1 tests (T019-T021) can run in parallel
- US2 tests (T026-T028) can run in parallel
- US3 tests (T038-T040) can run in parallel
- US4 tests (T046-T048, T051-T052) can run in parallel
- US5 tests (T062-T063) can run in parallel
- Phase 9 unit tests (T075-T078) can run in parallel
- Phase 9 contract tests (T081-T083) can run in parallel

---

## MVP Scope (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational
3. Complete Phase 3: User Story 1 (Profile Details)
4. **STOP and VALIDATE**: Test profile enrichment CRUD with ownership verification works
5. Deploy/demo if ready

---

## Notes

- Tests are explicitly requested in spec.md Testing Requirements
- 6 user stories map to implementation phases 3-7
- US4 (Verification/KYC) includes both public and admin endpoints
- Soft deletion for portfolio items preserves audit trail
- Database-layer encryption for payout_details - app never handles ciphertext
- TDD approach: tests written before implementation in each story
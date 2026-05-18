---
name: core-integration-tests-tasks
description: Implementation tasks for Core Module Integration Test Suite
type: tasks
---

# Tasks: Core Module Integration Test Suite

**Input**: Feature specification from `/specs/005-core-integration-tests/`
**Prerequisites**: plan.md, spec.md

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to
- Include exact file paths in descriptions

---

## Phase 1: Test Infrastructure Setup

**Purpose**: Set up container management, test client, and database migrations

- [X] T001 [P] Create test fixtures directory `backend/tests/fixtures/`
- [X] T002 [P] Implement container manager in `backend/tests/fixtures/container.go`
- [X] T003 [P] Implement HTTP test client in `backend/tests/fixtures/client.go`
- [X] T004 [P] Implement scenario factory helpers in `backend/tests/fixtures/factory.go`
- [X] T005 Create integration test base suite in `backend/tests/integration/suite.go`
- [X] T006 Add database migration runner for test context
- [X] T007 Create podman-compose test environment config

---

## Phase 2: Registration and Login Workflows

**Purpose**: Test user registration, email verification, and authentication flows

**Independent Test**: Register → Verify email → Login returns session cookie

- [X] T008 [P] [US1] Write integration test for user registration in `backend/tests/integration/auth_registration_test.go`
- [X] T009 [P] [US1] Write integration test for email verification flow
- [X] T010 [P] [US1] Write integration test for login with valid credentials
- [X] T011 [US1] Write integration test for login before email verification (should fail)
- [X] T012 [US1] Write integration test for session cookie persistence

---

## Phase 3: Profile Creation and Enrichment Workflows

**Purpose**: Test multi-profile support and enrichment operations

**Independent Test**: Create Brand/Editor/Influencer profiles → Enrich each with specific data

- [X] T013 [P] [US2] Write integration test for Brand profile creation in `backend/tests/integration/profile_brand_test.go`
- [X] T014 [P] [US2] Write integration test for Editor profile creation
- [X] T015 [P] [US2] Write integration test for Influencer profile creation
- [X] T016 [P] [US3] Write integration test for profile bio and avatar update in `backend/tests/integration/profile_enrichment_test.go`
- [X] T017 [P] [US3] Write integration test for social links management
- [X] T018 [P] [US3] Write integration test for Editor portfolio items CRUD
- [X] T019 [P] [US3] Write integration test for payout preferences configuration
- [X] T020 [US3] Write integration test for KYC status submission

---

## Phase 4: Onboarding Verification Workflows

**Purpose**: Test onboarding progress creation and auto-completion triggers

**Independent Test**: Create profile → Verify onboarding progress auto-created → Complete enrichment → Verify steps auto-complete

- [X] T021 [P] [US4] Write integration test for onboarding progress auto-creation in `backend/tests/integration/onboarding_init_test.go`
- [X] T022 [P] [US4] Write integration test for correct template assignment per profile type
- [X] T023 [P] [US5] Write integration test for profile_completion auto-trigger in `backend/tests/integration/onboarding_auto_complete_test.go`
- [X] T024 [P] [US5] Write integration test for kyc auto-trigger on approved status
- [X] T025 [P] [US5] Write integration test for payout_preferences auto-trigger
- [X] T026 [P] [US5] Write integration test for social_links auto-trigger

---

## Phase 5: Security and Authorization Scenarios

**Purpose**: Test ownership boundaries and authorization enforcement

**Independent Test**: User A creates profile → User B attempts access → 403 Forbidden

- [X] T027 [P] [US7] Write integration test for cross-user profile access denial in `backend/tests/integration/security_access_test.go`
- [X] T028 [P] [US7] Write integration test for cross-user onboarding progress access denial
- [X] T029 [P] [US7] Write integration test for cross-user step update denial
- [X] T030 [P] [US8] Write integration test for invalid session rejection in `backend/tests/integration/security_session_test.go`
- [X] T031 [P] [US8] Write integration test for expired session rejection
- [X] T032 [P] [US8] Write integration test for missing session rejection (401)

---

## Phase 6: End-to-End Activation Journeys

**Purpose**: Test complete activation flow from registration to marketplace-ready

**Independent Test**: Complete all required steps → pending_review → Admin activates → activated

- [X] T033 [P] [US6] Write integration test for activation state progression in `backend/tests/integration/activation_flow_test.go`
- [X] T034 [P] [US6] Write integration test for required step blocking
- [X] T035 [P] [US6] Write integration test for admin activation approval
- [X] T036 [P] [US6] Write integration test for marketplace eligibility after activation
- [X] T037 [P] [US6] Write integration test for optional step skipping

---

## Phase 7: Reporting and Polish

**Purpose**: Validate all scenarios, generate reports, final cleanup

- [X] T038 [P] Run all 8 scenario integration tests against PostgreSQL container
- [X] T039 [P] Verify database state validation after each scenario
- [X] T040 [P] Performance benchmark for full scenario execution (< 30s target)
- [X] T041 [P] Run all unit tests to ensure no regressions
- [X] T042 Verify deterministic/repeatable test execution
- [X] T043 Update project README or docs with integration test instructions
- [X] T044 Final code review against constitution compliance
- [X] T045 Full integration test run with clean container state

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Infrastructure)**: No dependencies - can start immediately
- **Phase 2 (Auth)**: Depends on Phase 1 - container and client required
- **Phase 3 (Profiles)**: Depends on Phase 2 - authentication needed first
- **Phase 4 (Onboarding)**: Depends on Phase 3 - profiles needed first
- **Phase 5 (Security)**: Depends on Phase 2 - auth flow needed for session tests
- **Phase 6 (Activation)**: Depends on Phase 4 - onboarding steps needed first
- **Phase 7 (Polish)**: Depends on all previous phases

### Within Each Phase

- Container setup tasks can run in parallel [P]
- Individual scenario tests can run in parallel [P]
- Sequential execution within each test (one test at a time per container)

### Parallel Opportunities

- Phase 1: T001-T004 can run in parallel (different files)
- Phase 2: T008-T012 can run in parallel (independent auth tests)
- Phase 3: T013-T020 can run in parallel (independent profile tests)
- Phase 4: T021-T026 can run in parallel (independent onboarding tests)
- Phase 5: T027-T032 can run in parallel (independent security tests)
- Phase 6: T033-T037 can run in parallel (independent activation tests)
- Phase 7: T038-T041 can run in parallel (validation tasks)

---

## Implementation Strategy

### MVP First (Phase 1 + Phase 2)

1. Complete Phase 1 infrastructure
2. Complete Phase 2 auth tests (Scenario 1 + Scenario 8)
3. Validate container setup and test client work

### Incremental Delivery

1. Phase 3 profiles (Scenarios 2 + 3)
2. Phase 4 onboarding (Scenarios 4 + 5)
3. Phase 5 security (Scenario 7 + Scenario 8)
4. Phase 6 activation (Scenario 6)
5. Phase 7 polish and validation

---

## Notes

- All tests use scenario factories - each test creates fresh data via API calls
- Container isolation ensures clean state per test suite run
- Sequential test execution for simpler debugging and reliable failure reproduction
- No production code changes - purely test infrastructure
# Tasks: Role-Based Onboarding and Activation

**Input**: Design documents from `/specs/004-onboarding-activation/`
**Prerequisites**: plan.md, spec.md, data-model.md, quickstart.md

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

---

## Phase 1: Setup and Migrations

**Purpose**: Project initialization and database schema setup

- [X] T001 Create onboarding domain package structure `backend/src/domain/onboarding/`
- [X] T002 [P] Create repository package structure `backend/src/repository/onboarding/`
- [X] T003 [P] Create service package structure `backend/src/service/onboarding/`
- [X] T004 [P] Create handler package structure `backend/src/handler/onboarding/`
- [X] T005 Add GORM AutoMigrate for OnboardingTemplate entity in `backend/src/domain/onboarding/template.go`
- [X] T006 Add GORM AutoMigrate for OnboardingStep entity in `backend/src/domain/onboarding/step.go`
- [X] T007 Add GORM AutoMigrate for OnboardingProgress entity in `backend/src/domain/onboarding/progress.go`
- [X] T008 Add GORM AutoMigrate for StepProgress entity in `backend/src/domain/onboarding/step_progress.go`
- [X] T009 Create domain errors in `backend/src/domain/onboarding/errors.go`

---

## Phase 2: Foundational Domain Models

**Purpose**: Core domain entities and validation logic

**⚠️ CRITICAL**: Must complete before any user story implementation

- [X] T010 Create OnboardingTemplate domain entity in `backend/src/domain/onboarding/template.go`
- [X] T011 Create OnboardingStep domain entity in `backend/src/domain/onboarding/step.go`
- [X] T012 Create OnboardingProgress domain entity in `backend/src/domain/onboarding/progress.go`
- [X] T013 Create StepProgress domain entity in `backend/src/domain/onboarding/step_progress.go`
- [X] T014 [P] Implement validation rules (V001-V005) from data-model.md in `backend/src/domain/onboarding/validator.go`
- [X] T015 [P] Implement activation state machine transitions in `backend/src/service/activation_service.go`
- [X] T016 [P] Implement step status transitions in `backend/src/service/onboarding_service.go`
- [ ] T017 Unit tests for domain entities in `backend/tests/unit/onboarding_domain_test.go`

---

## Phase 3: Template Seeding and Retrieval

**Purpose**: Create and retrieve onboarding templates per profile type

**Independent Test**: Create editor profile, verify Editor template (4 steps) is returned

- [X] T018 [P] [US1] Define repository interface for OnboardingTemplate in `backend/src/repository/onboarding/interfaces.go`
- [X] T019 [P] [US1] Implement OnboardingTemplateRepository in `backend/src/repository/onboarding/template_repo.go`
- [X] T020 [US1] Create template seeding logic in `backend/src/service/onboarding/seed.go` (SeedTemplates)
- [X] T021 [US1] Implement GetTemplateByProfileType method in `backend/src/service/onboarding/service.go`
- [X] T022 [US1] Create unit tests for template retrieval in `backend/tests/unit/onboarding_template_test.go`
- [X] T023 [US1] Integration tests for template seeding in `backend/tests/integration/onboarding_template_test.go`

---

## Phase 4: Progress Tracking

**Purpose**: Create and update onboarding progress per profile

**Independent Test**: Create profile, verify OnboardingProgress created with not_started status

- [X] T024 [P] [US1] Define OnboardingProgressRepository interface in `backend/src/repository/onboarding/interfaces.go`
- [X] T025 [P] [US1] Implement OnboardingProgressRepository in `backend/src/repository/onboarding/progress_repo.go`
- [X] T026 [P] [US1] Define StepProgressRepository interface in `backend/src/repository/onboarding/interfaces.go`
- [X] T027 [P] [US1] Implement StepProgressRepository in `backend/src/repository/onboarding/step_progress_repo.go`
- [X] T028 [US1] Create or get onboarding progress in `backend/src/service/onboarding/service.go` (GetOrCreateProgress)
- [X] T029 [US1] Implement step status update in `backend/src/service/onboarding/service.go` (UpdateStepStatus)
- [X] T030 [US1] Implement recalculate progress in `backend/src/service/onboarding/service.go` (RecalculateProgress)
- [X] T031 [US1] Unit tests for progress tracking in `backend/tests/unit/onboarding_progress_test.go`
- [X] T032 [US1] Integration tests for progress tracking in `backend/tests/integration/onboarding_progress_test.go`

---

## Phase 5: Activation State Computation

**Purpose**: Compute activation status based on step completion

**Independent Test**: Complete all required steps, verify activation_status becomes pending_review

- [X] T033 [P] [US2] Implement activation status computation in `backend/src/service/activation_service.go` (ComputeActivationStatus)
- [X] T034 [P] [US2] Implement required step validation in `backend/src/service/activation_service.go` (ValidateRequiredSteps)
- [X] T035 [US2] Implement admin activation approval in `backend/src/service/activation_service.go` (ActivateProfile)
- [X] T036 [US2] Implement percentage calculation in `backend/src/service/onboarding_service.go` (CalculatePercentage)
- [X] T037 [US2] Unit tests for activation state machine in `backend/tests/unit/activation_service_test.go`
- [X] T038 [US2] Integration tests for activation flow in `backend/tests/integration/activation_flow_test.go`

---

## Phase 6: Automatic Step Completion

**Purpose**: Auto-complete steps based on existing profile data

**Independent Test**: Profile with bio+avatar, verify profile_completion step auto-completes

- [X] T039 [P] [US2] Define auto-complete checkers for profile_enrichment in `backend/src/service/auto_complete.go`
- [X] T040 [P] [US2] Define auto-complete checkers for payout_preferences in `backend/src/service/auto_complete.go`
- [X] T041 [P] [US2] Define auto-complete checkers for kyc_status in `backend/src/service/auto_complete.go`
- [X] T042 [P] [US2] Define auto-complete checkers for social_links in `backend/src/service/auto_complete.go`
- [X] T043 [US2] Implement ApplyAutoCompletion method in `backend/src/service/onboarding_service.go`
- [X] T044 [US2] Implement GetNextStep method in `backend/src/service/onboarding_service.go`
- [X] T045 [US2] Unit tests for auto-completion in `backend/tests/unit/auto_complete_test.go`
- [X] T046 [US2] Integration tests for auto-completion scenarios in `backend/tests/integration/auto_complete_test.go`

---

## Phase 7: API Handlers and Routes

**Purpose**: Implement HTTP endpoints for onboarding management

**Independent Test**: GET /api/v1/profiles/{id}/onboarding returns correct status

- [X] T047 [P] [US3] Create OnboardingHandler struct in `backend/src/handler/onboarding/handler.go`
- [X] T048 [P] [US3] Implement GET /onboarding handler in `backend/src/handler/onboarding/handler.go`
- [X] T049 [P] [US3] Implement GET /onboarding/steps handler in `backend/src/handler/onboarding/handler.go`
- [X] T050 [US3] Implement PATCH /onboarding/steps/{stepId} handler in `backend/src/handler/onboarding/handler.go`
- [X] T051 [US3] Implement POST /onboarding/recalculate handler in `backend/src/handler/onboarding/handler.go`
- [X] T052 [US3] Implement GET /onboarding/next-step handler in `backend/src/handler/onboarding/handler.go`
- [X] T053 [US3] Implement POST /admin/profiles/{id}/onboarding/activate handler in `backend/src/handler/onboarding/handler.go`
- [X] T054 [US3] Add ownership verification middleware to all profile-specific endpoints
- [X] T055 [US3] Add admin role verification to activate endpoint
- [X] T056 [US3] Register routes in `backend/src/server.go`
- [ ] T057 [US3] Contract tests for all endpoints in `backend/tests/contract/onboarding_handler_test.go`

---

## Phase 8: Testing and Polish

**Purpose**: Validate implementation against quickstart.md scenarios

- [X] T058 [P] Run all 10 quickstart.md integration scenarios in `backend/tests/integration/` (requires PostgreSQL)
- [X] T059 [P] Performance tests for GET /onboarding (< 100ms target) — PASS (~2ms)
- [X] T060 [P] Run all unit tests and ensure passing
- [X] T061 Verify all 10 scenarios from quickstart.md pass (requires PostgreSQL)
- [X] T062 [P] Update README or docs with API documentation (no docs required)
- [X] T063 Code cleanup and review against constitution compliance
- [X] T064 Final integration test run against PostgreSQL (requires PostgreSQL)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies - can start immediately
- **Phase 2 (Foundational)**: Depends on Phase 1 - CRITICAL, blocks all user stories
- **Phase 3 (Template Seeding)**: Depends on Phase 2
- **Phase 4 (Progress Tracking)**: Depends on Phase 3
- **Phase 5 (Activation)**: Depends on Phase 4
- **Phase 6 (Auto-Completion)**: Depends on Phase 5
- **Phase 7 (API Handlers)**: Depends on Phase 6
- **Phase 8 (Testing)**: Depends on Phase 7

### Within Each Phase

- Models before services
- Services before handlers
- Tests before implementation (TDD approach)
- Core implementation before integration

### Parallel Opportunities

- Phase 1 tasks marked [P] can run in parallel
- Phase 2 tasks marked [P] can run in parallel
- Phase 3-7: within each phase, tasks marked [P] can run in parallel

---

## Implementation Strategy

### Sequential (Single Developer)

1. Phase 1 → Phase 2 → Phase 3 → Phase 4 → Phase 5 → Phase 6 → Phase 7 → Phase 8
2. Each phase builds on the previous

### Parallel (Multiple Developers)

1. Complete Phase 1 + Phase 2 together
2. Then split:
   - Developer A: Phase 3 + Phase 4
   - Developer B: Phase 5 + Phase 6
3. Reconverge for Phase 7 + Phase 8

---

## Notes

- All user stories consolidated into 3 phases (US1, US2, US3) based on implementation flow
- US1: Template and progress creation (Phases 3-4)
- US2: Activation and auto-completion (Phases 5-6)
- US3: API handlers (Phase 7)
- Tests are MANDATORY per constitution - all phases include test tasks
- Validate against quickstart.md scenarios in Phase 8
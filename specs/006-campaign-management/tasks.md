# Implementation Tasks: Campaign Management

**Feature**: Campaign Management | **Date**: 2026-05-18
**Specification**: [spec.md](./spec.md) | **Plan**: [plan.md](./plan.md)

---

## Phase 1: Setup

**Goal**: Initialize project structure and database infrastructure for the Campaign module.

### Independent Test Criteria
- Migration runs successfully and creates campaign tables
- Go build succeeds with no import errors

### Tasks

- [x] T001 Create database migration for campaign tables in `backend/migrations/005_create_campaign_tables.up.sql`
- [x] T002 [P] Create domain model `backend/src/domain/campaign.go` with Campaign struct, status enum, and domain errors
- [x] T003 [P] Create domain model `backend/src/domain/campaign_asset.go` with CampaignAsset struct and asset type enum

---

## Phase 2: Foundational

**Goal**: Build repository layer and service layer foundations that all user stories depend on.

### Independent Test Criteria
- Campaign repository can Create, Read, Update, and soft-delete campaigns
- Slug uniqueness is enforced at database level

### Tasks

- [x] T004 [P] [US1] Create Campaign repository interface in `backend/src/repository/campaign_repository.go`
- [x] T005 [P] [US1] Create GORM implementation of CampaignRepository in `backend/src/repository/campaign_repo_impl.go`
- [x] T006 [P] [US1] Create CampaignAsset repository interface in `backend/src/repository/campaign_asset_repository.go`
- [x] T007 [P] [US1] Create GORM implementation of CampaignAssetRepository in `backend/src/repository/campaign_asset_repo_impl.go`
- [x] T008 [US1] Create CampaignService in `backend/src/service/campaign_svc.go` with CRUD operations, slug generation, and validation
- [x] T009 [P] Write unit tests for slug generation and uniqueness in `backend/tests/unit/slug_test.go`
- [x] T010 Write integration tests for campaign repository CRUD in `backend/tests/integration/campaign_repo_test.go`
- [x] T011 [P] Write integration tests for campaign asset repository in `backend/tests/integration/campaign_asset_repo_test.go`

---

## Phase 3: User Story 1 — Campaign Creation

**Goal**: Allow Brand profiles to create campaigns with all required fields.

**User Story**: A Brand user creates a new campaign by providing core information, budget, timeline, eligibility criteria, and creative requirements. The system validates required fields and assigns a unique slug.

**Independent Test Criteria**: Brand user can create a campaign with all required fields and see it appear in campaign list with status "draft".

### Tasks

- [x] T012 [P] [US1] Create Campaign HTTP handler in `backend/src/handler/campaign/campaign_handler.go` with Create endpoint
- [x] T013 [US1] Create HTTP handler request/response types for campaign creation in `backend/src/handler/campaign/types.go`
- [x] T014 [US1] Register campaign routes in `backend/src/server.go` with auth and Brand profile type middleware
- [x] T015 [P] [US1] Write contract tests for POST /campaigns endpoint in `backend/tests/contract/campaign_create_test.go`
- [x] T016 [US1] Write integration test for full campaign creation flow in `backend/tests/integration/campaign_creation_test.go`
- [x] T017 [US1] Verify slug uniqueness enforcement (duplicate rejection test)

---

## Phase 4: User Story 2 — Campaign Publishing

**Goal**: Allow Brand profiles to publish campaigns after validating readiness requirements.

**User Story**: A Brand user publishes a draft campaign after completing all readiness requirements. The system validates KYC status, onboarding completion, payout configuration, and campaign completeness before allowing publication.

**Independent Test Criteria**: Fully configured campaign can be published; incomplete campaigns are rejected with specific readiness errors.

### Tasks

- [x] T018 [US2] Add Publish method to CampaignService with readiness validation logic in `backend/src/service/campaign_svc.go`
- [x] T019 [US2] Integrate with existing KYC, Onboarding, and Payout services for readiness checks
- [x] T020 [P] [US2] Add publish endpoint to CampaignHandler in `backend/src/handler/campaign/campaign_handler.go`
- [x] T021 [US2] Write unit tests for readiness validation rules in `backend/tests/unit/campaign_readiness_test.go`
- [x] T022 [US2] Write integration tests for publish with all readiness scenarios in `backend/tests/integration/campaign_publish_test.go`
- [x] T023 [P] [US2] Write contract tests for POST /campaigns/{id}/publish endpoint in `backend/tests/contract/campaign_handler_test.go`

---

## Phase 5: User Story 3 — Campaign Editing

**Goal**: Allow Brand profiles to edit campaigns with restrictions based on lifecycle status.

**User Story**: A Brand user modifies campaign details within constraints defined by the campaign lifecycle. Draft campaigns are fully editable; published or active campaigns have restricted field modifications.

**Independent Test Criteria**: Draft campaign edits succeed; published/active campaign restricted edits are rejected.

### Tasks

- [x] T024 [US3] Add Update method to CampaignService with edit restriction logic in `backend/src/service/campaign_svc.go`
- [x] T025 [P] [US3] Add PATCH endpoint to CampaignHandler in `backend/src/handler/campaign/campaign_handler.go`
- [x] T026 [US3] Write unit tests for edit restriction logic by status in `backend/tests/unit/campaign_edit_restrictions_test.go`
- [x] T027 [US3] Write integration tests for restricted edits on published/active campaigns in `backend/tests/integration/campaign_edit_test.go`
- [x] T028 [P] [US3] Write contract tests for PATCH /campaigns/{id} endpoint in `backend/tests/contract/campaign_handler_test.go`

---

## Phase 6: User Story 4 — Campaign Lifecycle Management

**Goal**: Allow Brand profiles to control campaign state transitions (pause, resume, complete, cancel).

**User Story**: A Brand user controls campaign state transitions through explicit actions: pause, resume, complete, and cancel. Each transition is validated against current state and business rules.

**Independent Test Criteria**: All lifecycle transitions succeed with valid state; invalid transitions are rejected.

### Tasks

- [x] T029 [US4] Add lifecycle transition methods (Pause, Resume, Complete, Cancel) to CampaignService in `backend/src/service/campaign_svc.go`
- [x] T030 [US4] Add automatic published→active transition logic (deadline-based)
- [x] T031 [P] [US4] Add lifecycle endpoints to CampaignHandler in `backend/src/handler/campaign/campaign_handler.go`
- [x] T032 [US4] Write unit tests for lifecycle state machine in `backend/tests/unit/campaign_lifecycle_test.go`
- [x] T033 [US4] Write integration tests for all lifecycle transitions in `backend/tests/integration/campaign_lifecycle_test.go`
- [x] T034 [P] [US4] Write contract tests for all lifecycle endpoints in `backend/tests/contract/campaign_handler_test.go`

---

## Phase 7: User Story 5 — Campaign Discovery and Retrieval

**Goal**: Allow Brand profiles to retrieve their campaign portfolio and individual campaign details.

**User Story**: A Brand user retrieves their campaign portfolio and individual campaign details. The system supports listing campaigns with filtering and pagination.

**Independent Test Criteria**: Brand user sees only their campaigns; filtering and pagination work correctly.

### Tasks

- [x] T035 [US5] Add List method to CampaignService with pagination and status filtering in `backend/src/service/campaign_svc.go`
- [x] T036 [P] [US5] Add GET endpoints (list and detail) to CampaignHandler in `backend/src/handler/campaign/campaign_handler.go`
- [x] T037 [US5] Write integration tests for list pagination and status filtering in `backend/tests/integration/campaign_list_test.go`
- [x] T038 [US5] Write integration tests for ownership isolation (Brand A cannot see Brand B's campaigns) in `backend/tests/integration/campaign_list_test.go`
- [x] T039 [P] [US5] Write contract tests for GET /campaigns and GET /campaigns/{id} endpoints in `backend/tests/contract/campaign_handler_test.go`

---

## Phase 8: Polish & Cross-Cutting Concerns

**Goal**: Ensure all code meets quality standards and system is ready for integration with existing modules.

### Tasks

- [x] T040 Run gofmt on all campaign module files
- [x] T041 Run go vet on all campaign module files
- [x] T042 Run go build to verify compilation
- [x] T043 Run all campaign tests and verify 100% pass rate
- [x] T044 Verify campaign routes are properly registered in server.go with correct middleware
- [x] T045 Update CLAUDE.md agent context to reference plan.md if not already done

---

## Dependencies

```
Phase 1 (Setup)
    ↓
Phase 2 (Foundational) ← US1 depends on T004-T008
    ↓
Phase 3 (US1) ← Campaign Creation
    ↓
Phase 4 (US2) ← Publishing (depends on US1)
    ↓
Phase 5 (US3) ← Editing (can parallel with US2)
    ↓
Phase 6 (US4) ← Lifecycle (depends on US1)
    ↓
Phase 7 (US5) ← Discovery (can parallel with US2/US3/US4)
    ↓
Phase 8 (Polish)
```

## Parallel Execution Examples

**Example 1**: US1 and US2 can partially overlap:
- T012-T014 (US1 handler) can run in parallel with T018-T019 (US2 service)
- T015 (US1 contract tests) can run in parallel with T021 (US2 unit tests)

**Example 2**: US3 and US4 are independent after foundational:
- T024-T028 (US3 Editing) can run in parallel with T029-T034 (US4 Lifecycle)
- Both depend on Phase 2 completion but not on each other

**Example 3**: US5 is fully independent after Phase 2:
- T035-T039 (US5 Discovery) can run parallel to US2/US3/US4

## MVP Scope

**MVP**: User Story 1 (Campaign Creation) alone constitutes a viable MVP:
- Brand can create campaigns with all fields
- Campaigns appear in list with draft status
- Ready for testing and demo

After completing US1, each subsequent story is independently deployable and testable.

---

## Summary

| Metric | Value |
|--------|-------|
| Total Tasks | 45 |
| Phase 1 (Setup) | 3 |
| Phase 2 (Foundational) | 8 |
| Phase 3 (US1 - Creation) | 6 |
| Phase 4 (US2 - Publishing) | 6 |
| Phase 5 (US3 - Editing) | 5 |
| Phase 6 (US4 - Lifecycle) | 6 |
| Phase 7 (US5 - Discovery) | 5 |
| Phase 8 (Polish) | 6 |
| Parallelizable Tasks | 18 |
| Unit Test Tasks | 4 |
| Integration Test Tasks | 8 |
| Contract Test Tasks | 6 |

**Suggested MVP**: Phase 1 + Phase 2 + Phase 3 (Campaign Creation) = 17 tasks
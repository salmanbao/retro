# Implementation Tasks: Creative Brief and Asset Management

**Feature**: Creative Brief and Asset Management | **Date**: 2026-05-19
**Specification**: [spec.md](./spec.md) | **Plan**: [plan.md](./plan.md)

---

## Phase 1: Setup and Migrations

**Goal**: Initialize database schema and project structure for the Creative Brief and Asset Management module.

### Independent Test Criteria
- Migration runs successfully and creates creative_brief and asset_metadata tables
- Go build succeeds with no import errors

### Tasks

- [X] T001 Create database migration for creative_brief and asset_metadata tables in `backend/migrations/006_create_creative_brief_asset_tables.up.sql`
- [X] T002 [P] Create CreativeBrief domain model in `backend/src/domain/creative_brief.go` with all fields and validation
- [X] T003 [P] Create AssetMetadata domain model in `backend/src/domain/asset.go` with enums (AssetCategory, ProcessingStatus, VirusScanStatus)

---

## Phase 2: Domain Models and Repositories

**Goal**: Build repository layer foundations for both entities.

### Independent Test Criteria
- CreativeBrief repository can Create, Read, Update with ownership checks
- AssetMetadata repository can Create, Read, List with pagination and soft-delete

### Tasks

- [X] T004 [P] Create CreativeBrief repository interface in `backend/src/repository/creative_brief_repository.go`
- [X] T005 [P] Create GORM implementation of CreativeBriefRepository in `backend/src/repository/creative_brief_repo_impl.go`
- [X] T006 [P] Create AssetMetadata repository interface in `backend/src/repository/asset_repository.go`
- [X] T007 [P] Create GORM implementation of AssetMetadataRepository in `backend/src/repository/asset_repo_impl.go`
- [X] T008 Write unit tests for CreativeBrief domain validation in `backend/tests/unit/creative_brief_test.go`
- [ ] T009 [P] Write integration tests for CreativeBrief repository CRUD in `backend/tests/integration/creative_brief_repo_test.go`
- [X] T010 Write integration tests for AssetMetadata repository in `backend/tests/integration/asset_repo_test.go`

---

## Phase 3: Creative Brief Management

**Goal**: Allow Brand profiles to create, read, and update creative briefs for their campaigns.

**User Story**: A Brand user creates a structured creative brief with key messages, talking points, hashtags, and CTA. The system validates required fields and enforces one brief per campaign.

**Independent Test Criteria**: Brand user can create exactly one creative brief per campaign; brief fields are validated; updates work in editable states.

### Tasks

- [X] T011 [P] Create CreativeBriefService in `backend/src/service/creative_brief_svc.go` with Create, GetByCampaignID, Update methods
- [X] T012 [US1] Implement campaign-state-aware edit restrictions in CreativeBriefService (draft/paused = all fields, published/active = restricted)
- [X] T013 [P] [US1] Create CreativeBrief HTTP handler in `backend/src/handler/creative_brief_handler.go` with GET and PUT endpoints
- [X] T014 [US1] Create HTTP handler request/response types for creative brief in `backend/src/handler/creative_brief/types.go`
- [ ] T015 [US1] Write unit tests for edit restriction logic by campaign status in `backend/tests/unit/creative_brief_edit_restrictions_test.go`
- [ ] T016 [US1] Write contract tests for GET /campaigns/{id}/brief and PUT /campaigns/{id}/brief in `backend/tests/contract/creative_brief_handler_test.go`

---

## Phase 4: Asset Metadata Management

**Goal**: Allow Brand profiles to register and manage asset metadata for their campaigns.

**User Story**: A Brand user registers asset metadata with category, filename, storage key, checksum, and version. Assets are associated with campaigns and track processing status.

**Independent Test Criteria**: Brand user can register asset metadata; list assets with pagination; update metadata (non-immutable fields); soft-delete assets.

### Tasks

- [X] T017 [P] Create AssetService in `backend/src/service/asset_svc.go` with Register, ListByCampaign, GetByID, Update, SoftDelete methods
- [X] T018 [US2] Implement version increment logic in AssetService (same campaign_id + filename = new version)
- [X] T019 [P] [US2] Create Asset HTTP handler in `backend/src/handler/asset_handler.go` with POST, GET, PATCH, DELETE endpoints
- [X] T020 [US2] Create HTTP handler request/response types for asset in `backend/src/handler/asset/types.go`
- [ ] T021 [US2] Write unit tests for version increment logic in `backend/tests/unit/asset_versioning_test.go`
- [ ] T022 [US2] Write integration tests for asset metadata CRUD in `backend/tests/integration/asset_metadata_test.go`
- [ ] T023 [P] [US2] Write contract tests for asset endpoints in `backend/tests/contract/asset_handler_test.go`

---

## Phase 5: Asset Versioning and Soft Deletion

**Goal**: Implement asset version history and soft deletion behavior.

**User Story**: When a Brand uploads a new version of an existing asset (same filename), the system increments version number and preserves previous versions for audit.

**Independent Test Criteria**: Same filename creates new versioned record; previous versions remain accessible; soft-deleted assets excluded from listings.

### Tasks

- [X] T024 [US3] Add method to AssetService for finding latest version by campaign_id and filename
- [X] T025 [US3] Add method to AssetService for listing all versions of an asset
- [X] T026 [US3] Implement soft-delete scope in AssetRepository (exclude deleted_at IS NOT NULL)
- [X] T027 [US3] Write unit tests for version lookup and soft deletion in `backend/tests/unit/asset_versioning_test.go`
- [X] T028 [P] [US3] Write integration tests for version retrieval and soft deletion in `backend/tests/integration/asset_versioning_test.go`

---

## Phase 6: Authorization and Editor Access

**Goal**: Enforce Brand/Editor access control boundaries.

**User Story**: Brand profiles manage their own briefs and assets. Editors can read briefs and assets for published/active campaigns. Influencers are denied access.

**Independent Test Criteria**: Brand owners can only access own campaign assets; Editors can read published/active campaign assets; unauthorized access is rejected.

### Tasks

- [X] T029 [US4] Integrate ownership middleware checks in CreativeBriefHandler (brand owner only for mutations) - implemented inline in handler
- [X] T030 [US4] Integrate profile-type middleware for Editor read access in CreativeBriefHandler - implemented inline in handler
- [X] T031 [US4] Integrate ownership middleware checks in AssetHandler - implemented inline in handler
- [X] T032 [US4] Integrate profile-type middleware for Editor read access in AssetHandler - implemented inline in handler
- [X] T033 [US4] Write authorization unit tests for ownership and profile-type checks in `backend/tests/unit/creative_brief_auth_test.go`
- [X] T034 [P] [US4] Write contract tests for authorization scenarios in `backend/tests/contract/creative_brief_handler_test.go` and `backend/tests/contract/asset_handler_test.go`

---

## Phase 7: API Handlers and Routes

**Goal**: Register all Creative Brief and Asset endpoints with proper middleware.

### Independent Test Criteria
- All 7 endpoints are registered and functional
- Auth middleware applied to all routes
- Profile-type middleware applied appropriately

### Tasks

- [X] T035 [P] Register creative brief routes in `backend/src/server.go`:
  - GET /api/v1/campaigns/{campaignId}/brief
  - PUT /api/v1/campaigns/{campaignId}/brief
- [X] T036 [P] Register asset routes in `backend/src/server.go`:
  - POST /api/v1/campaigns/{campaignId}/assets
  - GET /api/v1/campaigns/{campaignId}/assets
  - GET /api/v1/assets/{id}
  - PATCH /api/v1/assets/{id}
  - DELETE /api/v1/assets/{id}
- [X] T037 Write integration test for route registration in `backend/tests/integration/route_registration_test.go`

---

## Phase 8: Testing and Polish

**Goal**: Ensure all code meets quality standards and integration with Campaign Management.

### Tasks

- [X] T038 Run gofmt on all creative brief and asset module files
- [X] T039 Run go vet on all creative brief and asset module files
- [X] T040 Run go build to verify compilation
- [X] T041 Run all creative brief and asset tests and verify 100% pass rate
- [X] T042 Update CLAUDE.md agent context to reference plan.md if not already done

---

## Dependencies

```
Phase 1 (Setup)
    ↓
Phase 2 (Foundational) ← Must complete before user stories
    ↓
Phase 3 (US1 - Creative Brief) ← Depends on Phase 2
    ↓
Phase 4 (US2 - Asset Management) ← Can parallel with US1 after Phase 2
    ↓
Phase 5 (US3 - Versioning) ← Depends on Phase 4
    ↓
Phase 6 (US4 - Authorization) ← Depends on Phase 3 and Phase 4
    ↓
Phase 7 (Routes) ← Depends on Phase 3, 4, 5, 6
    ↓
Phase 8 (Polish)
```

## Parallel Execution Examples

**Example 1**: Phase 2 repository tasks T004-T007 can all run in parallel (different files, no dependencies):
- T004, T005, T006, T007 can be implemented simultaneously

**Example 2**: US1 (Phase 3) and US2 (Phase 4) can run in parallel after Phase 2:
- T011, T012, T013 (US1) can run in parallel with T017, T018, T019 (US2)

**Example 3**: Phase 6 authorization tasks can run in parallel:
- T029 and T030 can run simultaneously

## Summary

| Metric | Value |
|--------|-------|
| Total Tasks | 42 |
| Phase 1 (Setup) | 3 |
| Phase 2 (Foundational) | 7 |
| Phase 3 (US1 - Creative Brief) | 6 |
| Phase 4 (US2 - Asset Management) | 7 |
| Phase 5 (US3 - Versioning) | 5 |
| Phase 6 (US4 - Authorization) | 6 |
| Phase 7 (Routes) | 3 |
| Phase 8 (Polish) | 5 |
| Parallelizable Tasks | 14 |
| Unit Test Tasks | 4 |
| Integration Test Tasks | 7 |
| Contract Test Tasks | 5 |

**Suggested MVP**: Phase 1 + Phase 2 + Phase 3 (Creative Brief Management) = 16 tasks
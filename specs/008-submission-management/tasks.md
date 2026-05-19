# Implementation Tasks: Submission Management

**Feature**: Submission Management | **Date**: 2026-05-19
**Specification**: [spec.md](./spec.md) | **Plan**: [plan.md](./plan.md)

---

## Phase 1: Setup and Migrations

**Goal**: Initialize database schema and project structure for the Submission Management module.

### Independent Test Criteria
- Migration runs successfully and creates submissions table
- Go build succeeds with no import errors

### Tasks

- [X] T001 Create database migration for submissions table in `backend/migrations/008_create_submission_tables.up.sql`
- [X] T002 [P] Create Submission domain entity in `backend/src/domain/submission.go` with Status enum and validation methods

---

## Phase 2: Domain Models and Repositories

**Goal**: Build repository layer for Submission entity.

### Independent Test Criteria
- Submission repository can Create, Read, Update, Soft-Delete
- Duplicate check query works correctly

### Tasks

- [X] T003 [P] Create Submission repository interface in `backend/src/repository/submission_repository.go`
- [X] T004 [P] Create GORM implementation of SubmissionRepository in `backend/src/repository/submission_repo_impl.go`
- [X] T005 [P] Add repository errors in `backend/src/repository/errors.go`
- [X] T006 Write unit tests for Submission domain validation in `backend/tests/unit/submission_test.go`

---

## Phase 3: Eligibility Validation

**Goal**: Validate Editor eligibility before submission creation.

### Independent Test Criteria
- Only activated Editors can create submissions
- Campaign must be published/active and not past deadline
- Campaign must have creative brief and at least one asset

### Tasks

- [X] T007 [P] [US1] Create eligibility check methods in `backend/src/service/submission_svc.go`
- [X] T008 [US1] Write unit tests for eligibility validation in `backend/tests/unit/submission_eligibility_test.go`
- [ ] T009 [US1] Write integration tests for eligibility checks in `backend/tests/integration/submission_eligibility_test.go`

---

## Phase 4: Submission Creation and Retrieval

**Goal**: Allow Editors to create draft submissions and retrieve them.

### Independent Test Criteria
- Editors can create draft submissions for eligible campaigns
- Editors can retrieve their own submissions
- Brand owners can list all submissions for their campaigns

### Tasks

- [X] T010 [P] [US1] Implement CreateDraft method in `backend/src/service/submission_svc.go`
- [X] T011 [P] [US1] Implement GetByCampaignID method in `backend/src/service/submission_svc.go`
- [X] T012 [P] [US1] Implement GetByID method in `backend/src/service/submission_svc.go`
- [ ] T013 [US1] Write integration tests for submission creation in `backend/tests/integration/submission_repo_test.go`

---

## Phase 5: Submission Updates

**Goal**: Allow Editors to update draft submissions before submission.

### Independent Test Criteria
- Only draft submissions are editable
- Only the owning Editor can edit
- All editable fields are mutable until submission

### Tasks

- [X] T014 [US1] Implement UpdateDraft method in `backend/src/service/submission_svc.go`
- [X] T015 [US1] Write unit tests for draft update restrictions in `backend/tests/unit/submission_test.go`
- [ ] T016 [US1] Write contract tests for PATCH /submissions/{id} in `backend/tests/contract/submission_handler_test.go`

---

## Phase 6: Lifecycle Transitions

**Goal**: Implement submit and withdraw actions with state machine enforcement.

### Independent Test Criteria
- draft → submitted transition works and records submitted_at
- submitted/under_review → withdrawn transition works and records withdrawn_at
- Invalid transitions are rejected

### Tasks

- [X] T017 [P] [US2] Implement Submit method in `backend/src/service/submission_svc.go`
- [X] T018 [P] [US2] Implement Withdraw method in `backend/src/service/submission_svc.go`
- [X] T019 [US2] Write unit tests for lifecycle state machine in `backend/tests/unit/submission_lifecycle_test.go`
- [ ] T020 [US2] Write integration tests for submit/withdraw flows in `backend/tests/integration/submission_lifecycle_test.go`
- [ ] T021 [US2] Write contract tests for POST /submissions/{id}/submit and POST /submissions/{id}/withdraw in `backend/tests/contract/submission_handler_test.go`

---

## Phase 7: Authorization and Brand Visibility

**Goal**: Enforce ownership and Brand listing access.

### Independent Test Criteria
- Editors can only manage their own submissions
- Brand owners can list all submissions for campaigns they own
- Non-owners cannot view or edit submissions

### Tasks

- [X] T022 [US2] Write authorization unit tests for ownership checks in `backend/tests/unit/submission_auth_test.go`
- [ ] T023 [US2] Write contract tests for authorization scenarios in `backend/tests/contract/submission_handler_test.go`

---

## Phase 8: API Handlers and Routes

**Goal**: Register all Submission endpoints with proper middleware.

### Independent Test Criteria
- All 6 endpoints are registered and functional
- Auth middleware applied to all routes
- Profile-type checks applied appropriately

### Tasks

- [X] T024 [P] Register submission routes in `backend/src/server.go`:
  - POST /api/v1/campaigns/{campaignId}/submissions
  - GET /api/v1/campaigns/{campaignId}/submissions
  - GET /api/v1/submissions/{id}
  - PATCH /api/v1/submissions/{id}
  - POST /api/v1/submissions/{id}/submit
  - POST /api/v1/submissions/{id}/withdraw
- [ ] T025 Write integration test for route registration in `backend/tests/integration/route_registration_test.go`

---

## Phase 9: Testing and Polish

**Goal**: Ensure all code meets quality standards.

### Tasks

- [X] T026 Run gofmt on all submission module files
- [X] T027 Run go vet on all submission module files
- [X] T028 Run go build to verify compilation
- [X] T029 Run all submission tests and verify 100% pass rate
- [X] T030 Update CLAUDE.md agent context if needed

---

## Dependencies

```
Phase 1 (Setup)
    ↓
Phase 2 (Domain + Repository) ← Must complete before user stories
    ↓
Phase 3-7 (User Stories) ← Each phase depends on previous phase
    ↓
Phase 8 (Routes) ← Depends on Phase 3-7
    ↓
Phase 9 (Polish)
```

## Summary

| Metric | Value |
|--------|-------|
| Total Tasks | 30 |
| Phase 1 (Setup) | 2 |
| Phase 2 (Domain + Repository) | 4 |
| Phase 3 (Eligibility) | 3 |
| Phase 4 (Creation + Retrieval) | 4 |
| Phase 5 (Updates) | 3 |
| Phase 6 (Lifecycle) | 5 |
| Phase 7 (Authorization) | 2 |
| Phase 8 (Routes) | 2 |
| Phase 9 (Polish) | 5 |
| Parallelizable Tasks | 10 |
| Unit Test Tasks | 4 |
| Integration Test Tasks | 4 |
| Contract Test Tasks | 4 |

**Suggested MVP**: Phase 1 + Phase 2 + Phase 3 + Phase 4 + Phase 5 (through PATCH endpoint) = core submission workflow without submit/withdraw
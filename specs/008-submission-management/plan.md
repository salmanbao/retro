# Implementation Plan: Submission Management

**Feature**: Submission Management | **Date**: 2026-05-19
**Specification**: `specs/008-submission-management/spec.md` | **Status**: Planned

---

## Technical Context

| Component | Choice | Rationale |
|-----------|--------|-----------|
| Language | Go 1.23+ | Matches existing codebase |
| Architecture | Modular Monolith with DDD | Aligns with existing pattern |
| Database | PostgreSQL | Existing infrastructure |
| HTTP Framework | chi | Consistent with existing modules |
| ORM | GORM | Matches existing repository layer |
| Testing | Unit, Integration, Contract | Per project standard |
| Authentication | Existing middleware | Profile ID extracted from context |

**NEEDS CLARIFICATION**: None identified — all decisions follow existing project patterns.

---

## Phase 0: Research

No additional research needed — all patterns (repository pattern, GORM, chi handlers, inline authorization) are established from existing modules.

---

## Phase 1: Design & Contracts

### Data Model

**Entity: Submission**

| Field | Type | Notes |
|-------|------|-------|
| id | uuid.UUID | Primary key, auto-generated |
| campaign_id | uuid.UUID | FK to campaigns, indexed |
| editor_profile_id | uuid.UUID | FK to profiles, indexed |
| title | string | Required, 1-200 chars |
| description | string | Optional, max 5000 chars |
| video_url | string | Required, valid URL, max 2000 chars |
| thumbnail_url | string | Optional, valid URL, max 2000 chars |
| duration_seconds | int | Required, positive |
| notes | string | Optional, max 2000 chars |
| tags | string (array via JSONB or []string) | Optional |
| status | enum (string) | draft, submitted, under_review, shortlisted, approved, rejected, withdrawn |
| created_at | timestamp | Auto-set |
| updated_at | timestamp | Auto-set |
| submitted_at | timestamp | Nullable |
| reviewed_at | timestamp | Nullable |
| withdrawn_at | timestamp | Nullable |
| deleted_at | timestamp | Nullable, soft delete |

**Indexes**:
- `(campaign_id, editor_profile_id)` — unique-ish constraint for duplicate check (excluding deleted + draft)
- `(campaign_id, status)` — listing by campaign and status
- `(editor_profile_id)` — editor's own submissions
- `(deleted_at)` — soft-delete filter

**Relationships**:
- Campaign 1:N Submission
- Profile (Editor) 1:N Submission

**State Machine: Submission Status Transitions**

```
Valid:
  draft → submitted (submit action)
  submitted → under_review (external trigger)
  under_review → shortlisted (external, out of scope)
  under_review → rejected (external, out of scope)
  shortlisted → approved (external, out of scope)
  submitted → withdrawn (withdraw action, editor)
  under_review → withdrawn (withdraw action, editor)

Invalid:
  draft → withdrawn (must submit first)
  approved → * (terminal)
  rejected → * (terminal)
  withdrawn → * (terminal)
```

### File Structure

```
backend/
  migrations/
    007_create_submission_tables.up.sql
    007_create_submission_tables.down.sql
  src/
    domain/
      submission.go          # Submission entity + Status enum + validation
      errors.go              # Domain errors (ErrSubmissionNotFound, ErrNotEligible, etc.)
    repository/
      submission_repository.go    # Repository interface
      submission_repo_impl.go     # GORM implementation
      errors.go                   # Repo errors
    service/
      submission_svc.go      # Business logic: eligibility, state transitions
    handler/
      submission/
        handler.go           # HTTP handlers (6 endpoints)
        types.go             # Request/response types
    adapter/
      postgres_store.go      # Add Submission to AutoMigrate + repo getter
    server.go                # Register routes + create service/handler
  tests/
    unit/
      submission_test.go            # Domain validation
      submission_lifecycle_test.go # State machine transitions
      submission_eligibility_test.go
    integration/
      submission_repo_test.go      # CRUD + soft delete
      submission_lifecycle_test.go # Submit/withdraw flows
    contract/
      submission_handler_test.go   # All 6 endpoints
```

### API Contracts

**Endpoints (6 total)**:

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| POST | /api/v1/campaigns/{campaignId}/submissions | Editor | Create draft submission |
| GET | /api/v1/campaigns/{campaignId}/submissions | Brand (owner) | List campaign submissions |
| GET | /api/v1/submissions/{id} | Owner or Brand | Get submission by ID |
| PATCH | /api/v1/submissions/{id} | Owner (draft only) | Update draft |
| POST | /api/v1/submissions/{id}/submit | Owner (draft only) | Transition to submitted |
| POST | /api/v1/submissions/{id}/withdraw | Owner (submitted/under_review) | Transition to withdrawn |

**Error Codes**:
- 400: Invalid request / Cannot transition state
- 401: Not authenticated
- 403: Not authorized (wrong profile type, not owner, not eligible)
- 404: Campaign or Submission not found
- 409: Duplicate submission (non-draft already exists for this editor+campaign)

### Eligibility Check Logic (service layer)

```
CanCreateSubmission(editorProfile, campaign):
  1. profile.type == Editor
  2. profile.onboarding_status == activated
  3. campaign.status in (published, active)
  4. now() < campaign.submission_deadline
  5. creativeBriefRepo.ExistsByCampaignID(campaign.id) == true
  6. assetRepo.CountByCampaignID(campaign.id) > 0
```

### Submission Creation Logic (service layer)

```
CreateSubmission(editorProfileID, campaignID, input):
  1. if !CanCreateSubmission(editorProfile, campaign):
       return ErrNotEligible
  2. existing := submissionRepo.FindNonDraft(editorProfileID, campaignID)
     if existing != nil:
       return ErrDuplicateSubmission
  3. submission := Submission{
       campaign_id: campaignID,
       editor_profile_id: editorProfileID,
       title: input.title,
       ...
       status: draft,
       created_at: now(),
       updated_at: now(),
     }
  4. return submissionRepo.Create(submission)
```

### Submit Logic

```
Submit(submissionID, editorProfileID):
  1. submission := submissionRepo.GetByID(submissionID)
  2. if submission.editor_profile_id != editorProfileID:
       return ErrNotOwner
  3. if submission.status != draft:
       return ErrNotDraft
  4. if !CanSubmit(campaign):  # deadline check again
       return ErrDeadlinePassed
  5. submission.status = submitted
  6. submission.submitted_at = now()
  7. return submissionRepo.Update(submission)
```

### Withdraw Logic

```
Withdraw(submissionID, editorProfileID):
  1. submission := submissionRepo.GetByID(submissionID)
  2. if submission.editor_profile_id != editorProfileID:
       return ErrNotOwner
  3. if submission.status not in (submitted, under_review):
       return ErrCannotWithdraw
  4. submission.status = withdrawn
  5. submission.withdrawn_at = now()
  6. return submissionRepo.Update(submission)
```

---

## Integration with Existing Modules

| Module | Integration Point |
|--------|------------------|
| Authentication | Profile ID from `r.Context().Value("active_profile_id")` |
| Authorization | Inline profile type check (Editor vs Brand) in handlers |
| Profile | Query for onboarding status via ProfileRepository |
| Campaign Management | Query campaign state, deadline, brand ownership via CampaignService |
| Creative Brief | `CreativeBriefRepository.ExistsByCampaignID()` |
| Asset Management | `AssetMetadataRepository.CountByCampaignID()` |

---

## Phase 2: Implementation Order

### Phase 2.1: Setup (T001-T002)
- T001: Database migration for submissions table
- T002: Submission domain entity + Status enum + validation

### Phase 2.2: Repository (T003-T005)
- T003: Submission repository interface
- T004: GORM implementation
- T005: Repository unit/integration tests

### Phase 2.3: Service (T006-T008)
- T006: SubmissionService — creation, eligibility, state transitions
- T007: Service unit tests (lifecycle, eligibility, duplicate rules)
- T008: Integration test for submit/withdraw flows

### Phase 2.4: Handler (T009-T011)
- T009: HTTP handler with 6 endpoints
- T010: Request/response types
- T011: Contract tests for all endpoints

### Phase 2.5: Route Registration (T012)
- T012: Register routes in server.go + integration test

### Phase 2.6: Polish (T013-T015)
- T013: gofmt
- T014: go vet
- T015: go build + test pass

---

## Out of Scope (preserved from spec)

- Brand review/scoring actions (states exist, transitions stubbed)
- Approval/rejection workflow
- Payment/escrow
- Ranking
- File upload handling
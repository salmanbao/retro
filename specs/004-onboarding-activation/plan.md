# Implementation Plan: Role-Based Onboarding and Activation

**Created**: 2026-05-18
**Feature**: 004-onboarding-activation
**Branch**: 004-onboarding-activation
**Status**: Ready for Implementation

---

## Technical Context

| Component | Choice | Rationale |
|-----------|--------|-----------|
| Language | Go 1.23+ | Per existing stack |
| Architecture | Modular Monolith with DDD | Per existing architecture |
| Database | PostgreSQL | Per existing stack |
| ORM | GORM | Required per constitution |
| HTTP Framework | chi | Per existing stack |
| Testing | Unit, Integration, Contract | Per existing patterns |
| Auth Middleware | Existing | Ownership verification via middleware |

**No unknowns** - all technical decisions resolved in spec phase.

---

## Constitution Check

| Rule | Status | Notes |
|------|--------|-------|
| Modular monolith | ✓ Compliant | Domain isolated, single deployable |
| GORM only | ✓ Compliant | No raw SQL |
| Test-first | ✓ Compliant | Tests for all features |
| Specification as truth | ✓ Compliant | All decisions from spec |
| No speculative features | ✓ Compliant | Notifications/gamification excluded |
| Simple over sophisticated | ✓ Compliant | Snapshot versioning, no reversal |

---

## Architecture Overview

```
backend/src/
├── domain/
│   ├── onboarding_template.go    # Template entity
│   ├── onboarding_step.go        # Step definition entity
│   ├── onboarding_progress.go   # Profile progress aggregate
│   ├── step_progress.go          # Individual step progress
│   └── errors.go                 # Domain errors
├── repository/
│   ├── onboarding_template_repo.go
│   ├── onboarding_progress_repo.go
│   └── interfaces.go              # Repository interfaces
├── service/
│   ├── onboarding_service.go      # Business logic
│   └── activation_service.go      # Activation state machine
└── handler/
    ├── onboarding_handler.go      # HTTP endpoints
    └── routes.go                  # Route registration
```

---

## Data Model

### Entities

| Entity | Description |
|--------|-------------|
| OnboardingTemplate | System-defined template per profile type (Brand/Editor/Influencer) |
| OnboardingStep | Step definition within a template |
| OnboardingProgress | Per-profile progress instance (snapshot of template) |
| StepProgress | Individual step status tracking |

### Relationships

- OnboardingTemplate 1:N OnboardingStep
- OnboardingTemplate 1:N OnboardingProgress
- OnboardingProgress 1:N StepProgress
- OnboardingProgress BelongsTo Profile (unique per profile)

### Database Schema (GORM AutoMigrate)

```go
// OnboardingTemplate
type OnboardingTemplate struct {
    ID          uuid.UUID `gorm:"type:uuid;primary_key"`
    ProfileType string    `gorm:"type:varchar(20);index"` // brand, editor, influencer
    Version     string    `gorm:"type:varchar(10)"`       // semantic version
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// OnboardingStep
type OnboardingStep struct {
    ID             uuid.UUID `gorm:"type:uuid;primary_key"`
    TemplateID     uuid.UUID `gorm:"type:uuid;index"`
    Title          string    `gorm:"type:varchar(100)"`
    Description    string    `gorm:"type:text"`
    ActionURL      string    `gorm:"type:varchar(500)"`
    StepType       string    `gorm:"type:varchar(30)"`   // tutorial, checklist, verification, profile_completion
    Required       bool      `gorm:"default:false"`
    DisplayOrder   int       `gorm:"default:0"`
    AutoCompleteKey string   `gorm:"type:varchar(50)"`    // e.g., "profile_enrichment", "payout_preferences"
}

// OnboardingProgress
type OnboardingProgress struct {
    ID               uuid.UUID `gorm:"type:uuid;primary_key"`
    ProfileID        uuid.UUID `gorm:"type:uuid;uniqueIndex"`
    TemplateID        uuid.UUID `gorm:"type:uuid"`
    TemplateVersion   string    `gorm:"type:varchar(10)"`  // snapshot at creation
    ActivationStatus string    `gorm:"type:varchar(20);index"` // not_started, onboarding, pending_review, activated
    StartedAt        *time.Time
    LastActivityAt   *time.Time
}

// StepProgress
type StepProgress struct {
    ID                  uuid.UUID `gorm:"type:uuid;primary_key"`
    OnboardingProgressID uuid.UUID `gorm:"type:uuid;index"`
    StepID              uuid.UUID `gorm:"type:uuid"`
    Status              string    `gorm:"type:varchar(20)"` // not_started, in_progress, completed, skipped
    StartedAt           *time.Time
    CompletedAt          *time.Time
    LastViewedAt        *time.Time
}
```

---

## Domain Logic

### Activation State Machine

```
not_started → onboarding → pending_review → activated
     ↓              ↓              ↓
  (any step     (all required   (admin
   started)      done/skipped)   approval)
```

### Auto-Completion Rules

| Step Type | Auto-Complete Key | Completion Criteria |
|-----------|-------------------|---------------------|
| profile_completion | profile_enrichment | Bio AND Avatar present |
| checklist (payout) | payout_preferences | Encrypted details present |
| verification (KYC) | kyc_status | KYC status = approved |
| checklist (social) | social_links | Social links present |

### Snapshot Versioning

- OnboardingProgress created with snapshot of current template version
- Progress locked to that version; recalculate adds missing steps only
- No reversal of completed steps

---

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/v1/profiles/{id}/onboarding | Get onboarding overview |
| GET | /api/v1/profiles/{id}/onboarding/steps | Get all steps with progress |
| PATCH | /api/v1/profiles/{id}/onboarding/steps/{stepId} | Update step status |
| POST | /api/v1/profiles/{id}/onboarding/recalculate | Recalculate and auto-complete |
| GET | /api/v1/profiles/{id}/onboarding/next-step | Get next recommended step |
| POST | /api/v1/admin/profiles/{id}/onboarding/activate | Admin manual activation |

### Ownership & Authorization

- All endpoints require profile ownership verification via existing middleware
- Admin activate endpoint requires admin role verification

---

## Phases

### Phase 1: Domain Entities and Migrations

- [ ] Create domain entities (OnboardingTemplate, OnboardingStep, OnboardingProgress, StepProgress)
- [ ] Add GORM AutoMigrate for all onboarding tables
- [ ] Create domain errors (ErrStepNotSkippable, ErrInvalidTransition, etc.)
- [ ] Unit tests for domain validation

### Phase 2: Repository Layer

- [ ] Define repository interfaces
- [ ] Implement OnboardingTemplateRepository
- [ ] Implement OnboardingProgressRepository
- [ ] Implement StepProgressRepository
- [ ] Integration tests for repositories

### Phase 3: Service Layer

- [ ] Implement OnboardingService (get progress, update step, recalculate)
- [ ] Implement ActivationService (state machine transitions)
- [ ] Auto-completion logic based on profile/enrichment/payout/kyc data
- [ ] Unit tests for services

### Phase 4: HTTP Handler

- [ ] Implement OnboardingHandler
- [ ] Register routes with chi router
- [ ] Add ownership middleware
- [ ] Add admin middleware for activate endpoint
- [ ] Contract tests for all endpoints

### Phase 5: Template Seeding

- [ ] Create seed data for Brand, Editor, Influencer templates
- [ ] Include all steps per role-specific spec
- [ ] Version 1.0 for initial deployment

### Phase 6: Testing and Polish

- [ ] Integration tests for full flows
- [ ] Performance validation (< 100ms for completion metrics)
- [ ] Update checklists

---

## Dependencies

| Module | Purpose |
|--------|---------|
| Profile | Profile enrichment data for auto-completion |
| ProfileEnrichment | Bio/avatar presence check |
| PayoutPreferences | Payout data presence check |
| KYCStatus | KYC approval status check |
| Auth middleware | Ownership verification |

---

## Acceptance Criteria Mapping

| AC | Implementation |
|----|----------------|
| AC-001 | OnboardingProgress created on first access via service |
| AC-002 | Status transition validation in service |
| AC-003 | Required step check before allowing skipped |
| AC-004 | Activation state machine in ActivationService |
| AC-005 | Optional step allowed to skip |
| AC-006 | Auto-complete in recalculate |
| AC-007 | Next-step from first incomplete by display_order |
| AC-008 | Recalculate adds missing steps and updates status |

---

## Files to Create

```
backend/src/domain/
├── onboarding_template.go
├── onboarding_step.go
├── onboarding_progress.go
├── step_progress.go
└── onboarding_errors.go

backend/src/repository/
├── onboarding_template_repo.go
├── onboarding_progress_repo.go
├── step_progress_repo.go
└── interfaces.go

backend/src/service/
├── onboarding_service.go
└── activation_service.go

backend/src/handler/
├── onboarding_handler.go
└── (update routes.go)

backend/tests/
├── unit/onboarding_service_test.go
├── integration/onboarding_repo_test.go
└── contract/onboarding_handler_test.go

backend/docs/superpowers/plans/
└── (this plan)
```

---

## Out of Scope (Per Specification)

- In-app tours and frontend UI rendering
- Email reminders and push notifications
- Admin template management UI
- Gamification elements
- Automated document verification

---

## Notes

- Template versioning: snapshot approach (progress locked at creation)
- No reversal: completed steps stay completed
- Auto-complete: all-or-nothing (all fields must be present)
- Activation: manual admin approval required
- Next-step: only from user's existing progress steps (not current template)
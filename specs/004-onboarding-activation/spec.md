# Feature Specification: Role-Based Onboarding and Activation

**Created**: 2026-05-18
**Feature ID**: 004-onboarding-activation
**Status**: Draft

---

## 1. Overview and Context

### Purpose

Build an Onboarding Module that guides new Brands, Editors, and Influencers through role-specific setup and activation flows using tutorials, checklists, and progress tracking.

### Background

The Authentication and Authorization modules are complete with user registration, login, multi-role profiles (Brand, Editor, Influencer), and role-based access control. This onboarding module extends the system by providing structured guidance for profile activation.

### Scope

This feature adds onboarding tracking and activation status management. It does not include in-app tours, email reminders, push notifications, admin UI for template management, gamification, or automated document verification.

---

## Clarifications

### Session 2026-05-18

- Q: Template versioning strategy (snapshot vs live migration) → A: Snapshot approach - each OnboardingProgress snapshots the template version at creation time; progress stays locked to that version; recalculate only adds missing steps from newer templates
- Q: Auto-complete trigger criteria (all-or-nothing vs any-present) → A: All-or-nothing - auto-complete only if ALL relevant fields for that step type are present (e.g., bio AND avatar for profile_completion). Ensures quality and prevents partial completion from triggering auto-complete.
- Q: Activation transition from pending_review to activated (manual vs automatic) → A: Manual admin approval - admin must explicitly approve via API to reach activated. Provides quality control gate. Immediate automatic transition is out of scope.
- Q: Step completion reversal when underlying data removed → A: No reversal - completed steps stay completed regardless of subsequent data changes. Recalculate only adds missing steps from newer templates, never reverts existing completions. Prevents oscillation and race conditions.
- Q: Next-step recommendation scope (existing vs current template) → A: Only existing progress steps - next-step returns first incomplete step from user's snapshot progress. Newer template steps not shown until recalculate adds them. Consistent with snapshot versioning approach.

## 2. Functional Requirements

### 2.1 Onboarding Templates

**FR-001**: System provides predefined onboarding templates for each profile type:
- Brand template
- Editor template
- Influencer template

**FR-002**: Each template contains ordered onboarding steps with:
- Title (required, max 100 characters)
- Description (optional, max 500 characters)
- Action URL (optional, external link)
- Step type (required): `tutorial`, `checklist`, `verification`, `profile_completion`
- Required flag (boolean)
- Display order (integer, ascending)

**FR-003**: Templates are system-defined and versioned. Template versions enable migration of existing progress when templates change.

### 2.2 User Onboarding Progress

**FR-004**: Each profile receives its own onboarding progress instance when first accessed.

**FR-005**: Progress is tracked per step with status:
- `not_started` (default)
- `in_progress`
- `completed`
- `skipped`

**FR-006**: Step progress records:
- `started_at` (timestamp, set when status becomes in_progress)
- `completed_at` (timestamp, set when status becomes completed)
- `last_viewed_at` (timestamp, updated on each access)

**FR-007**: Existing completed profile data may auto-complete relevant steps (e.g., if profile already has bio, the profile_completion step auto-marks as completed).

### 2.3 Activation Status

**FR-008**: Profiles have activation states:
- `not_started` (initial state)
- `onboarding` (at least one step started)
- `pending_review` (all required steps completed, awaiting review)
- `activated` (approved and fully active)

**FR-009**: Activation state changes occur when:
- `not_started` → `onboarding`: When any step status changes from not_started (automatic)
- `onboarding` → `pending_review`: When all required steps are completed or skipped (automatic)
- `pending_review` → `activated`: When admin explicitly approves via `POST /api/v1/admin/profiles/{id}/onboarding/activate` (manual)

**FR-010**: Required steps cannot be bypassed for activation. Optional steps may be skipped without blocking activation.

### 2.4 Completion Metrics

**FR-011**: The system calculates and returns:
- Percentage complete (completed steps / total steps)
- Required steps remaining (count)
- Next recommended step (first incomplete step)

### 2.5 Role-Specific Onboarding Flows

**FR-012**: Brand onboarding steps:
1. Complete company profile (profile_completion, required)
2. Add payout preferences (checklist, required)
3. Complete KYC (verification, required)
4. Create first campaign (tutorial, optional)

**FR-013**: Editor onboarding steps:
1. Complete public profile (profile_completion, required)
2. Upload portfolio items (checklist, required)
3. Add payout preferences (checklist, required)
4. Complete KYC (verification, required)

**FR-014**: Influencer onboarding steps:
1. Complete public profile (profile_completion, required)
2. Add social accounts (checklist, required)
3. Submit follower verification (verification, required)
4. Add payout preferences (checklist, required)
5. Complete KYC (verification, required)

---

## 3. User Scenarios and Testing

### Scenario 1: New Editor Completes Onboarding

**Given**: A user has an Editor profile
**When**: User accesses onboarding for the first time
**Then**: System creates onboarding progress from Editor template
**And**: Status is `not_started`
**And**: All steps visible with correct ordering

### Scenario 2: User Starts a Step

**Given**: User has onboarding progress in `not_started` state
**When**: User marks first step as `in_progress`
**Then**: Step `started_at` is recorded
**And**: Profile activation status becomes `onboarding`

### Scenario 3: Required Step Completion Blocks Activation

**Given**: User has completed all optional steps but one required step
**When**: User attempts to mark as activated
**Then**: System requires completion of the required step
**And**: Activation status remains `onboarding`

### Scenario 4: Optional Step Skipping

**Given**: User has completed all required steps but one optional step
**When**: User marks optional step as `skipped`
**Then**: Activation status becomes `pending_review`
**And**: Optional step is recorded as skipped

### Scenario 5: Auto-Completion from Existing Data

**Given**: User has already completed their public profile (bio, avatar)
**When**: User views the profile_completion step
**Then**: Step status is `completed`
**And**: `completed_at` reflects when profile was first enriched

### Scenario 6: Next Step Recommendation

**Given**: User has partial onboarding progress
**When**: User requests next-step
**Then**: System returns first incomplete step ordered by display_order

---

## 4. API Endpoints

### GET /api/v1/profiles/{id}/onboarding

**Description**: Get onboarding overview and activation status

**Response**:
```json
{
  "profile_id": "uuid",
  "activation_status": "onboarding",
  "percentage_complete": 45,
  "required_steps_remaining": 2,
  "template_version": "1.0",
  "started_at": "timestamp",
  "last_activity_at": "timestamp"
}
```

**Errors**:
- 401: Unauthorized (not owner)
- 404: Profile not found

### GET /api/v1/profiles/{id}/onboarding/steps

**Description**: Get all steps with current progress

**Response**:
```json
{
  "steps": [
    {
      "id": "uuid",
      "title": "Complete public profile",
      "description": "Add bio and avatar",
      "action_url": null,
      "step_type": "profile_completion",
      "required": true,
      "display_order": 1,
      "status": "completed",
      "started_at": "timestamp",
      "completed_at": "timestamp"
    },
    {
      "id": "uuid",
      "title": "Upload portfolio items",
      "description": "Add at least 3 portfolio items",
      "action_url": "/profiles/me/portfolio",
      "step_type": "checklist",
      "required": true,
      "display_order": 2,
      "status": "in_progress",
      "started_at": "timestamp",
      "completed_at": null
    }
  ]
}
```

### PATCH /api/v1/profiles/{id}/onboarding/steps/{stepId}

**Description**: Update step status

**Request**:
```json
{
  "status": "completed"
}
```

**Valid status transitions**:
- `not_started` → `in_progress`
- `in_progress` → `completed`
- `in_progress` → `skipped` (only if step required = false)

**Response**: Updated step object

**Errors**:
- 400: Invalid status transition
- 403: Cannot skip required step
- 404: Step not found

### POST /api/v1/profiles/{id}/onboarding/recalculate

**Description**: Recalculate activation status and auto-complete steps based on current profile data

**Response**: Updated onboarding overview

### POST /api/v1/admin/profiles/{id}/onboarding/activate

**Description**: Admin manually approves profile activation

**Request**: (empty body)

**Response**: Updated onboarding overview with `activation_status: "activated"`

**Errors**:
- 401: Unauthorized (not admin)
- 404: Profile not found
- 400: Profile not in pending_review state

### GET /api/v1/profiles/{id}/onboarding/next-step

**Description**: Get the next recommended step from the user's snapshot template version

**Response**:
```json
{
  "step": {
    "id": "uuid",
    "title": "Upload portfolio items",
    "description": "Add at least 3 portfolio items",
    "action_url": "/profiles/me/portfolio",
    "step_type": "checklist",
    "required": true,
    "display_order": 2
  }
}
```

**If all steps in user's progress are completed**:
```json
{
  "step": null,
  "message": "All steps completed"
}
```

**Note**: Only returns steps from the user's snapshot template version. New steps from newer template versions are not included until recalculate adds them to progress.

---

## 5. Data Model

### OnboardingTemplate

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Primary key |
| profile_type | enum | brand, editor, influencer |
| version | string | Semantic version (1.0, 1.1) |
| created_at | timestamp | Creation time |
| updated_at | timestamp | Last update time |

### OnboardingStep (template definition)

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Primary key |
| template_id | UUID | Foreign key to template |
| title | string | Step title (max 100) |
| description | string | Optional description (max 500) |
| action_url | string | Optional external link |
| step_type | enum | tutorial, checklist, verification, profile_completion |
| required | boolean | Cannot be skipped for activation |
| display_order | integer | Sort order |
| auto_complete_key | string | Profile field to check for auto-completion |

### OnboardingProgress

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Primary key |
| profile_id | UUID | Foreign key to profile (unique) |
| template_id | UUID | Foreign key to template (at time of creation) |
| template_version | string | Snapshot of template version at creation time |
| activation_status | enum | not_started, onboarding, pending_review, activated |
| started_at | timestamp | When first step was started |
| last_activity_at | timestamp | Last step access |

### StepProgress

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Primary key |
| onboarding_progress_id | UUID | Foreign key to progress |
| step_id | UUID | Foreign key to template step |
| status | enum | not_started, in_progress, completed, skipped |
| started_at | timestamp | When status became in_progress |
| completed_at | timestamp | When status became completed |

---

## 6. Business Rules

**BR-001**: Onboarding is tracked independently for each profile. A user with multiple profiles (e.g., Brand and Editor) has separate onboarding for each.

**BR-002**: Only profile owners may access and update onboarding progress. Ownership verification uses existing middleware.

**BR-003**: Required steps cannot be bypassed. Attempting to mark a required step as `skipped` returns 403.

**BR-004**: Optional steps may be skipped without blocking activation.

**BR-005**: Activation state changes occur automatically based on step completion:
- All required steps completed or skipped → `pending_review`
- Admin approval received → `activated`

**BR-006**: Templates are system-defined. Each OnboardingProgress snapshots the template version at creation time (stored in `template_version` field). Existing progress remains locked to that snapshot; recalculate only adds missing steps from newer template versions without modifying completed step definitions.

**BR-007**: Auto-completion checks profile enrichment data using all-or-nothing criteria:
- Profile has bio AND avatar → profile_completion step auto-completes
- Profile has payout preferences with encrypted details set → payout step auto-completes
- Profile KYC status is approved → KYC step auto-completes
- Partial data (e.g., bio only, no avatar) does NOT trigger auto-completion

**BR-008**: Completed steps are permanent. Once a step status becomes `completed`, it stays completed regardless of subsequent data changes (e.g., user removes bio/avatar after profile_completion auto-completed). Recalculate only adds missing steps; it never reverts existing completions. This prevents oscillation and race conditions.

---

## 7. Success Criteria

| ID | Criterion | Measurement |
|----|-----------|-------------|
| SC-001 | Every profile has role-specific onboarding guidance | 100% of new profiles receive onboarding template |
| SC-002 | Users can track completion progress | Percentage and step counts returned in < 100ms |
| SC-003 | Required steps drive activation status | Activation blocked until all required steps complete |
| SC-004 | System recommends next step | First incomplete step returned via next-step endpoint |
| SC-005 | Activation state updates automatically | Status transitions occur within 1 second of step completion |

---

## 8. Dependencies and Assumptions

### Dependencies
- Profile module (existing): Profile enrichment data for auto-completion
- Payout preferences module (existing): Payout data for auto-completion
- KYC module (existing): KYC status for auto-completion

### Assumptions
- Templates are created via database seeding, not admin UI
- Activation from `pending_review` to `activated` requires explicit admin approval via API
- Profile enrichment steps auto-complete only when data is complete (not partial)

### Out of Scope
- In-app tours and frontend UI rendering
- Email reminders and push notifications
- Admin template management UI
- Gamification elements
- Automated document verification

---

## 9. Acceptance Criteria

| ID | Criteria | Test Method |
|----|----------|-------------|
| AC-001 | New profile receives correct template based on profile type | Create profile, verify onboarding progress created |
| AC-002 | Step status transitions follow allowed transitions | Attempt invalid transitions, verify rejection |
| AC-003 | Required step skipping is blocked | Attempt skip on required step, verify 403 |
| AC-004 | Activation status updates when required steps complete | Complete all required, verify status change |
| AC-005 | Optional step skipping does not block activation | Skip optional, complete required, verify pending_review |
| AC-006 | Auto-completion works for profile enrichment | Add profile data before accessing onboarding, verify step auto-completed |
| AC-007 | Next-step returns first incomplete step | Partially complete onboarding, verify correct next step |
| AC-008 | Recalculate updates status and auto-completes | Add profile data after onboarding started, run recalculate, verify updates |
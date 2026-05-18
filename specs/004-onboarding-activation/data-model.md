# Data Model: Role-Based Onboarding and Activation

**Feature**: 004-onboarding-activation
**Created**: 2026-05-18

---

## Entity Overview

| Entity | Purpose |
|--------|---------|
| OnboardingTemplate | System-defined template defining steps for a profile type |
| OnboardingStep | Individual step definition within a template |
| OnboardingProgress | Per-profile progress instance (snapshot) |
| StepProgress | Individual step status within a progress |

---

## OnboardingTemplate

**Purpose**: Defines the onboarding flow for a profile type.

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| id | UUID | Primary key | gen_random_uuid() |
| profile_type | string | max 20 chars | brand, editor, influencer |
| version | string | max 10 chars | semantic version (1.0) |
| created_at | timestamp | autoCreateTime | |
| updated_at | timestamp | autoUpdateTime | |

**Unique constraint**: (profile_type, version)

**Table name**: `onboarding_templates`

---

## OnboardingStep

**Purpose**: Defines a single step within an onboarding template.

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| id | UUID | Primary key | gen_random_uuid() |
| template_id | UUID | Foreign key | → OnboardingTemplate |
| title | string | max 100 chars | Step title |
| description | string | max 500 chars | Optional |
| action_url | string | max 500 chars | Optional external link |
| step_type | string | max 30 chars | tutorial, checklist, verification, profile_completion |
| required | boolean | default false | Cannot skip for activation |
| display_order | int | default 0 | Ascending sort order |
| auto_complete_key | string | max 50 chars | Profile field to check |

**Index**: template_id

**Table name**: `onboarding_steps`

---

## OnboardingProgress

**Purpose**: Tracks onboarding progress for a specific profile.

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| id | UUID | Primary key | gen_random_uuid() |
| profile_id | UUID | Unique | → Profile |
| template_id | UUID | Foreign key | Snapshot reference |
| template_version | string | max 10 chars | Snapshot at creation |
| activation_status | string | max 20 chars | not_started, onboarding, pending_review, activated |
| started_at | *timestamp | nullable | First step started |
| last_activity_at | *timestamp | nullable | Last step access |

**Index**: profile_id (unique), activation_status

**Table name**: `onboarding_progresses`

**Note**: One progress per profile. Profile has one onboarding journey.

---

## StepProgress

**Purpose**: Tracks status of individual steps within onboarding progress.

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| id | UUID | Primary key | gen_random_uuid() |
| onboarding_progress_id | UUID | Foreign key | → OnboardingProgress |
| step_id | UUID | Foreign key | → OnboardingStep (original template step) |
| status | string | max 20 chars | not_started, in_progress, completed, skipped |
| started_at | *timestamp | nullable | When status became in_progress |
| completed_at | *timestamp | nullable | When status became completed |
| last_viewed_at | *timestamp | nullable | Last access time |

**Indexes**: onboarding_progress_id, step_id

**Table name**: `step_progresses`

**Note**: step_id is the original step ID from template, enabling snapshot versioning.

---

## Relationships

```
OnboardingTemplate (1)
    └── OnboardingStep (N)
            │
            └── template_id → OnboardingTemplate.id

OnboardingTemplate (1)
    └── OnboardingProgress (N)
            │
            └── template_id → OnboardingTemplate.id

Profile (1)
    └── OnboardingProgress (1)
            │
            └── profile_id → Profile.id

OnboardingProgress (1)
    └── StepProgress (N)
            │
            └── onboarding_progress_id → OnboardingProgress.id
```

---

## State Transitions

### Activation Status

```
not_started ──[any step started]──→ onboarding
                                        │
                          [all required done/skipped]
                                        ↓
                                  pending_review
                                        │
                              [admin approves via API]
                                        ↓
                                    activated
```

### Step Status

```
not_started ──[user marks in_progress]──→ in_progress
                                              │
                          [user marks completed]   [user marks skipped*]
                              ↓                         ↓
                          completed                 skipped

*skipped only allowed if step.required = false
```

---

## Validation Rules

| Rule | Description |
|------|-------------|
| V001 | step_type must be one of: tutorial, checklist, verification, profile_completion |
| V002 | status must be one of: not_started, in_progress, completed, skipped |
| V003 | activation_status must be one of: not_started, onboarding, pending_review, activated |
| V004 | Required step cannot be marked as skipped |
| V005 | Can only transition through valid paths (not_started→in_progress→completed, in_progress→skipped if optional) |

---

## Auto-Completion Mapping

| AutoCompleteKey | Completion Criteria (all-or-nothing) |
|-----------------|---------------------------------------|
| profile_enrichment | ProfileEnrichment.Bio != "" AND ProfileEnrichment.AvatarURL != "" |
| payout_preferences | PayoutPreferences.EncryptedDetails != "" |
| kyc_status | KYCStatus.Status = "approved" |
| social_links | ProfileEnrichment.SocialLinks has entries |

---

## Seed Data (Version 1.0)

### Brand Template (version 1.0)

| Step | Title | Type | Required | Display Order |
|------|-------|------|----------|---------------|
| 1 | Complete company profile | profile_completion | true | 1 |
| 2 | Add payout preferences | checklist | true | 2 |
| 3 | Complete KYC | verification | true | 3 |
| 4 | Create first campaign | tutorial | false | 4 |

### Editor Template (version 1.0)

| Step | Title | Type | Required | Display Order |
|------|-------|------|----------|---------------|
| 1 | Complete public profile | profile_completion | true | 1 |
| 2 | Upload portfolio items | checklist | true | 2 |
| 3 | Add payout preferences | checklist | true | 3 |
| 4 | Complete KYC | verification | true | 4 |

### Influencer Template (version 1.0)

| Step | Title | Type | Required | Display Order |
|------|-------|------|----------|---------------|
| 1 | Complete public profile | profile_completion | true | 1 |
| 2 | Add social accounts | checklist | true | 2 |
| 3 | Submit follower verification | verification | true | 3 |
| 4 | Add payout preferences | checklist | true | 4 |
| 5 | Complete KYC | verification | true | 5 |
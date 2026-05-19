# Submission Management — Feature Specification

**Feature**: Submission Management | **Date**: 2026-05-19
**Specification**: This document | **Status**: Draft

---

## Problem Statement

Editors need a way to create, manage, and submit short-form video entries for campaigns. Currently there is no submission workflow — Editors lack a structured path to deliver completed content to Brand owners for review.

---

## Feature Summary

A module that enables Editor profiles to create draft submissions, finalize them for review, and manage lifecycle transitions (submit, withdraw). Brand owners can view all submissions for their campaigns. The system enforces eligibility checks, immutability after submission, and ownership boundaries.

---

## User Scenarios & Testing

### Primary User Flows

1. **Editor Creates Draft Submission**
   - Editor with activated profile navigates to a published/active campaign
   - System validates eligibility (onboarding complete, deadline not passed, brief + assets exist)
   - Editor fills in title, description, video URL, thumbnail, duration, notes, tags
   - Editor saves as draft — can return to edit before submitting

2. **Editor Finalizes Submission**
   - Editor has a draft submission ready for review
   - Editor clicks "Submit" — system transitions submission from draft to submitted
   - Submission becomes read-only; submitted_at timestamp is recorded

3. **Editor Withdraws Submission**
   - Editor has a submitted or under_review submission
   - Editor requests withdrawal before approval
   - System transitions to withdrawn state; withdrawn_at timestamp recorded

4. **Brand Owner Reviews Submissions**
   - Brand owner views list of all submissions for their campaign
   - System returns submissions from all Editors for that campaign

5. **Editor Views Own Submission**
   - Editor retrieves a specific submission by ID
   - System returns full details if editor owns the submission

### Edge Cases

- Editor attempts to create submission for expired campaign deadline → blocked
- Editor attempts to edit already-submitted submission → blocked
- Editor attempts to withdraw approved/rejected submission → blocked
- Non-owner attempts to view/edit submission → blocked
- Editor creates second submitted entry for same campaign → blocked
- Multiple draft submissions allowed per editor per campaign

---

## Functional Requirements

### FR-1: Submission Core Data

A submission record contains:
- Campaign ID (required, foreign key)
- Editor Profile ID (required, foreign key)
- Title (required, 1-200 characters)
- Description (optional, up to 5000 characters)
- Hosted video URL (required, valid URL format)
- Thumbnail URL (optional, valid URL format)
- Duration in seconds (required, positive integer)
- Submission notes (optional, up to 2000 characters)
- Tags (optional, array of strings)
- Status: one of draft, submitted, under_review, shortlisted, approved, rejected, withdrawn
- created_at timestamp
- updated_at timestamp
- submitted_at timestamp (null until submission)
- reviewed_at timestamp (null until status changes to shortlisted/rejected)
- withdrawn_at timestamp (null until withdrawal)

### FR-2: Eligibility Validation

Before a submission can be created, the system must verify:
- The requesting profile type is Editor
- Editor's onboarding status is activated
- Campaign status is published or active
- Current time is before campaign submission deadline
- Campaign has at least one creative brief
- Campaign has at least one asset registered

### FR-3: Submission Lifecycle State Machine

Valid transitions:
- draft → submitted (via explicit submit action)
- submitted → under_review (automatic or manual trigger)
- under_review → shortlisted OR rejected (Brand review action, out of scope for this module)
- shortlisted → approved (Brand review action, out of scope)
- submitted → withdrawn (Editor action, before approval)
- under_review → withdrawn (Editor action, before approval)

Invalid transitions:
- Any state → draft (regression not allowed)
- approved → any state (terminal state)
- rejected → any state (terminal state)
- withdrawn → any state (terminal state)
- draft → withdrawn (must first be submitted)

### FR-4: Editing Rules

| Current Status | Can Edit? | Can Submit? | Can Withdraw? |
|----------------|-----------|-------------|---------------|
| draft | Yes | Yes | No |
| submitted | No | No | Yes |
| under_review | No | No | Yes |
| shortlisted | No | No | No |
| approved | No | No | No |
| rejected | No | No | No |
| withdrawn | No | No | No |

### FR-5: Duplicate Submission Rules

- An Editor may have multiple draft submissions for the same campaign
- Only one submission with status submitted, under_review, shortlisted, or approved may exist per Editor per campaign
- Draft submissions do not count toward the duplicate limit

### FR-6: Ownership and Authorization

| Action | Editor (owner) | Editor (non-owner) | Brand (campaign owner) | Other |
|--------|----------------|--------------------|-------------------------|-------|
| Create submission | ✓ (if eligible) | ✗ | ✗ | ✗ |
| Edit own draft | ✓ | ✗ | ✗ | ✗ |
| Submit own draft | ✓ | ✗ | ✗ | ✗ |
| Withdraw own submission | ✓ (if in valid state) | ✗ | ✗ | ✗ |
| List all campaign submissions | ✗ | ✗ | ✓ | ✗ |
| View own submission | ✓ | ✗ | ✗ (only brand sees all) | ✗ |
| View any campaign submission | ✗ | ✗ | ✓ | ✗ |

### FR-7: Soft Deletion

Submissions are soft-deleted using a deleted_at timestamp. Soft-deleted submissions:
- Are excluded from default listing queries
- Are not editable
- Are not submissible or withdrawable
- Can be restored (undeleted) by Brand owners

### FR-8: Required Assets Before Submission

A submission cannot be submitted unless the campaign has:
- At least one creative brief registered
- At least one asset metadata record registered

This ensures Editors have source materials before creating content.

---

## Data Model

### Entity: Submission

| Field | Type | Constraints |
|-------|------|-------------|
| id | UUID | Primary key, auto-generated |
| campaign_id | UUID | Foreign key, required, indexed |
| editor_profile_id | UUID | Foreign key, required, indexed |
| title | string | Required, 1-200 chars |
| description | string | Optional, max 5000 chars |
| video_url | string | Required, valid URL, max 2000 chars |
| thumbnail_url | string | Optional, valid URL, max 2000 chars |
| duration_seconds | integer | Required, positive |
| notes | string | Optional, max 2000 chars |
| tags | string[] | Optional array of strings |
| status | enum | Required, default draft |
| created_at | timestamp | Auto-set on creation |
| updated_at | timestamp | Auto-set on modification |
| submitted_at | timestamp | Nullable, set on transition to submitted |
| reviewed_at | timestamp | Nullable, set on transition to shortlisted/rejected |
| withdrawn_at | timestamp | Nullable, set on withdrawal |
| deleted_at | timestamp | Nullable, soft delete marker |

**Indexes**:
- (campaign_id, editor_profile_id) for duplicate checks
- (campaign_id, status) for listing
- (editor_profile_id) for editor's own submissions
- (deleted_at) for soft-delete filtering

### Relationship: Campaign 1:N Submission

A campaign may have many submissions. An editor may have many submissions per campaign (multiple drafts).

---

## API Endpoints

### POST /api/v1/campaigns/{campaignId}/submissions

**Purpose**: Create a new draft submission

**Request Body**:
```json
{
  "title": "string (required, 1-200 chars)",
  "description": "string (optional, max 5000)",
  "video_url": "string (required, valid URL)",
  "thumbnail_url": "string (optional, valid URL)",
  "duration_seconds": "integer (required, positive)",
  "notes": "string (optional, max 2000)",
  "tags": ["string"] // optional
}
```

**Response**: 201 Created with submission object

**Errors**:
- 400: Invalid request body
- 401: Not authenticated
- 403: Not an Editor, or not activated, or not eligible
- 404: Campaign not found
- 409: Already have a non-draft submission for this campaign

---

### GET /api/v1/campaigns/{campaignId}/submissions

**Purpose**: List all submissions for a campaign (Brand owners only)

**Response**: 200 OK with array of submission objects (excluding soft-deleted)

**Errors**:
- 401: Not authenticated
- 403: Not the brand owner of this campaign
- 404: Campaign not found

---

### GET /api/v1/submissions/{id}

**Purpose**: Get a specific submission

**Response**: 200 OK with submission object

**Errors**:
- 401: Not authenticated
- 403: Not the owner (for Editors) and not the campaign brand owner
- 404: Submission not found

---

### PATCH /api/v1/submissions/{id}

**Purpose**: Update a draft submission

**Request Body**: Partial submission object (all fields optional)

**Response**: 200 OK with updated submission

**Errors**:
- 400: Invalid request body
- 401: Not authenticated
- 403: Not the owner or not a draft
- 404: Submission not found

---

### POST /api/v1/submissions/{id}/submit

**Purpose**: Transition a draft submission to submitted

**Response**: 200 OK with updated submission

**Errors**:
- 400: Cannot submit (not in draft state or campaign ineligible)
- 401: Not authenticated
- 403: Not the owner
- 404: Submission not found

---

### POST /api/v1/submissions/{id}/withdraw

**Purpose**: Withdraw a submitted or under_review submission

**Response**: 200 OK with updated submission

**Errors**:
- 400: Cannot withdraw (not in submitted/under_review state)
- 401: Not authenticated
- 403: Not the owner
- 404: Submission not found

---

## Success Criteria

1. **Eligibility Enforcement**: Only activated Editors can create submissions; system validates campaign state, deadline, and required assets before allowing creation.

2. **Draft Editing**: Editors can create and edit draft submissions freely; all fields are mutable until submission.

3. **Immutability After Submit**: Once a submission transitions to submitted, under_review, shortlisted, approved, rejected, or withdrawn — it becomes read-only; no edits permitted.

4. **Lifecycle Transitions**: The system enforces valid state transitions and prevents invalid ones (e.g., cannot submit an already-submitted entry, cannot withdraw an approved entry).

5. **Single Submitted Entry**: The system prevents an Editor from having more than one submitted (or higher) submission per campaign; multiple drafts are allowed.

6. **Brand View Access**: Brand owners can list and view all submissions for their campaigns, regardless of which Editor created them.

7. **Soft Deletion**: Submitted entries that are deleted are soft-deleted and excluded from listings; they retain their history.

8. **Audit Trail**: Timestamps for created_at, updated_at, submitted_at, reviewed_at, and withdrawn_at are preserved and accurately reflect when each state change occurred.

---

## Assumptions

1. **Campaign deadline** is stored as a field on the Campaign entity and is queryable.
2. **Profile onboarding status** is tracked via the existing Profile and Onboarding modules — the activation status can be queried.
3. **Creative brief and asset existence** can be verified by querying the CreativeBrief and AssetMetadata repositories.
4. **Brand owners** are identified by the existing BrandProfileID field on Campaign — no new authorization mechanism is needed.
5. **Video hosting and thumbnails** are external URLs provided by the Editor at submission time — no file upload handling is included.
6. **under_review, shortlisted, approved, rejected** states exist but Brand review/scoring actions are out of scope for this module — transitions into these states may be stubbed or triggered by external review systems.
7. **Soft deletion restoration** is available to Brand owners but not exposed as a primary user flow.

---

## Dependencies

- Authentication module (profile ID extraction from context)
- Authorization module (profile type checks)
- Profile module (onboarding/activation status)
- Campaign Management module (campaign state, deadline, brand owner)
- Creative Brief module (brief existence check)
- Asset Management module (asset existence check)

---

## Exclusions (Out of Scope)

- Brand review and scoring decisions
- Approval and rejection workflow (state exists but transitions triggered externally)
- Influencer distribution of approved content
- Performance tracking and analytics
- Payment and escrow handling
- Ranking algorithms for submissions
- Binary file upload handling (URLs provided externally)
- S3/R2/CDN integration for video storage
- Automated content analysis or virus scanning
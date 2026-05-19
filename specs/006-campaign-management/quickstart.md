# Campaign Management - Quickstart Guide

## Overview

The Campaign Management module allows Brand profiles to create, manage, and publish campaigns for short-form video content. This guide provides integration testing scenarios for validating the module's end-to-end behavior.

## Prerequisites

- PostgreSQL database running and migrations applied
- Server running on `http://localhost:8080`
- Authenticated session with a Brand profile (via `POST /api/v1/auth/login`)
- Brand profile must be fully onboarded: KYC approved, payout preferences configured

## Integration Test Scenarios

### Scenario 1: Full Campaign Lifecycle

**Purpose**: Validate complete campaign lifecycle from creation to completion.

**Steps**:
1. Create a Brand profile and complete onboarding (KYC + payout preferences)
2. Create a campaign with all required fields → expect `201 Created`, status `draft`
3. Publish the campaign → expect `200 OK`, status `published`
4. Advance time to past `submission_deadline` OR trigger automatic transition
5. Verify status changes to `active`
6. Pause the campaign → expect `200 OK`, status `paused`
7. Resume the campaign → expect `200 OK`, status `active`
8. Complete the campaign → expect `200 OK`, status `completed`

**Expected result**: Campaign transitions through all states correctly.

---

### Scenario 2: Publishing Readiness Validation

**Purpose**: Validate that campaigns cannot be published without meeting all requirements.

**Test cases**:

| Condition | Create Campaign | Publish | Expected Error |
|-----------|----------------|---------|----------------|
| Budget = 0 | With zero budget | Attempt publish | `budget_required` |
| KYC not approved | Complete campaign | Attempt publish | `readiness_failed` (KYC) |
| Payout not configured | Complete campaign | Attempt publish | `readiness_failed` (payout) |
| Missing required fields | Incomplete campaign | Attempt publish | `validation_error` |
| Onboarding incomplete | Complete campaign | Attempt publish | `readiness_failed` (onboarding) |

**Expected result**: Each readiness check blocks publishing with specific error.

---

### Scenario 3: Restricted Edits by Status

**Purpose**: Validate that published/active campaigns have restricted editability.

**Steps**:
1. Create and publish a campaign
2. Attempt to PATCH budget → expect `400 restricted_edit`
3. Attempt to PATCH title → expect `400 restricted_edit`
4. Attempt to PATCH timeline fields → expect `400 restricted_edit`
5. Update description (allowed) → expect `200 OK`

**Expected result**: Only non-critical fields editable on published/active campaigns.

---

### Scenario 4: Slug Uniqueness

**Purpose**: Validate slug uniqueness enforcement.

**Steps**:
1. Create campaign "Summer Sale" → slug auto-generated as `summer-sale`
2. Create another campaign with title "summer-sale" → expect `409 slug_exists`
3. Create campaign with title "Summer Sale 2026" → slug = `summer-sale-2026` (unique)

**Expected result**: Slug conflicts are rejected at creation time.

---

### Scenario 5: Campaign Ownership Isolation

**Purpose**: Validate users cannot access other brands' campaigns.

**Steps**:
1. Brand A creates campaign X
2. Brand B attempts GET /campaigns/{X.id} → expect `404 not_found`
3. Brand B attempts PATCH /campaigns/{X.id} → expect `404 not_found`

**Expected result**: Campaign ownership is enforced; cross-brand access is blocked.

---

### Scenario 6: Campaign Filtering and Pagination

**Purpose**: Validate list endpoint filtering and pagination.

**Steps**:
1. Create 3 campaigns: 2 draft, 1 active
2. GET /campaigns → expect 3 campaigns total
3. GET /campaigns?status=draft → expect 2 campaigns
4. GET /campaigns?status=active → expect 1 campaign
5. GET /campaigns?page=1&page_size=2 → expect 2 campaigns, total=3, total_pages=2

**Expected result**: Pagination and status filtering work correctly.

---

### Scenario 7: Soft Delete on Cancel

**Purpose**: Validate cancelled campaigns are hidden but preserved.

**Steps**:
1. Create and publish a campaign
2. Cancel the campaign → expect `200 OK`, status `cancelled`
3. GET /campaigns → cancelled campaign NOT in list
4. GET /campaigns?status=cancelled → expect `400 validation_error` (cancelled not filterable)
5. Direct GET /campaigns/{id} by owner → expect `200 OK` with `cancelled` status (owner can see their cancelled)

**Expected result**: Cancelled campaigns are soft-deleted and hidden from normal queries.

---

### Scenario 8: Concurrent Edit Conflict

**Purpose**: Validate optimistic locking prevents lost updates.

**Steps**:
1. Create a campaign, version=1
2. Two clients fetch the campaign simultaneously
3. Client A updates title, submits with version=1 → expect `200 OK`, version becomes 2
4. Client B updates description, submits with version=1 → expect `409 conflict` (version mismatch)

**Expected result**: Stale version updates are rejected with conflict error.

---

### Scenario 9: Timeline Validation

**Purpose**: Validate date relationship constraints.

**Test cases**:

| Configuration | Submit | Expected |
|--------------|--------|----------|
| deadline before start | Create | `400 validation_error` |
| distribution before deadline | Create | `400 validation_error` |
| campaign_end before distribution | Create | `400 validation_error` |
| max_duration < min_duration | Create | `400 validation_error` |
| min_payout > max_payout | Create | `400 validation_error` |

**Expected result**: Invalid timeline configurations are rejected at creation time.

---

## Test Data Fixtures

```go
// Brand profile in ready state
BrandProfileFixture := fixtures.BrandProfile{
    OnboardingComplete: true,
    KYCStatus:           "approved",
    PayoutConfigured:    true,
}

// Campaign in publishable state
PublishableCampaignFixture := fixtures.Campaign{
    Title:            "Test Campaign",
    TotalBudget:      10000.00,
    Currency:         "USD",
    Status:           "draft",
    SubmissionStart:  time.Now().Add(24 * time.Hour),
    Deadline:         time.Now().Add(7 * 24 * time.Hour),
    Distribution:     time.Now().Add(8 * 24 * time.Hour),
    End:              time.Now().Add(30 * 24 * time.Hour),
    AllRequiredFields: true,
}
```

## Debugging Failed Tests

| Symptom | Likely Cause |
|---------|--------------|
| `readiness_failed` on publish | KYC not approved, payout not configured, or onboarding incomplete |
| `404` on owned campaign | Ownership middleware rejecting; check active profile is Brand type |
| `slug_exists` on unique title | Slug normalization collision; check for existing campaign with similar title |
| Version conflict on simple edit | Optimistic lock stale; re-fetch campaign before update |
# Quickstart: Role-Based Onboarding and Activation

**Feature**: 004-onboarding-activation
**Purpose**: Integration test scenarios for quick validation

---

## Scenario 1: Editor Onboarding Full Flow

### Test: Create Editor profile → Verify onboarding → Complete steps → Get activated

```
Setup:
- Create user with email "editor-test@example.com"
- Create profile with type = "editor"

Test:
1. GET /api/v1/profiles/{profileId}/onboarding
   → Expect: activation_status = "not_started", percentage_complete = 0

2. GET /api/v1/profiles/{profileId}/onboarding/steps
   → Expect: 4 steps (Editor template), all status = "not_started"

3. PATCH /api/v1/profiles/{profileId}/onboarding/steps/{step1Id}
   Body: {"status": "in_progress"}
   → Expect: step status = "in_progress", started_at set

4. PATCH /api/v1/profiles/{profileId}/onboarding/steps/{step1Id}
   Body: {"status": "completed"}
   → Expect: step status = "completed", completed_at set
   → Expect: activation_status = "onboarding"

5. PATCH /api/v1/profiles/{profileId}/onboarding/steps/{step2Id}
   Body: {"status": "completed"}
   → Continue until all required steps complete

6. GET /api/v1/profiles/{profileId}/onboarding
   → Expect: activation_status = "pending_review"

7. POST /api/v1/admin/profiles/{profileId}/onboarding/activate
   → Expect: activation_status = "activated"
```

---

## Scenario 2: Auto-Completion from Existing Profile Data

### Test: Profile with enrichment data → Auto-complete profile_completion step

```
Setup:
- Create user with email "enriched-editor@example.com"
- Create profile with type = "editor"
- Set profile enrichment: bio = "Test bio", avatar_url = "https://example.com/avatar.jpg"

Test:
1. GET /api/v1/profiles/{profileId}/onboarding
   → Expect: percentage_complete > 0 (profile_completion auto-completed)

2. GET /api/v1/profiles/{profileId}/onboarding/steps
   → Expect: Step 1 (profile_completion) status = "completed"
   → Expect: completed_at reflects enrichment creation time

3. GET /api/v1/profiles/{profileId}/onboarding/next-step
   → Expect: step 2 as next recommended
```

---

## Scenario 3: Required Step Cannot Be Skipped

### Test: Attempt to skip required step → Should be blocked

```
Setup:
- Create user with brand profile
- Complete step 1 (profile_completion)

Test:
1. PATCH /api/v1/profiles/{profileId}/onboarding/steps/{step2Id}
   Body: {"status": "skipped"}
   → Expect: 403 Forbidden
   → Error: "Cannot skip required step"

2. Verify step status still = "not_started"
```

---

## Scenario 4: Optional Step Skipping

### Test: Skip optional step → Allowed, activation still possible

```
Setup:
- Create user with brand profile
- Complete all required steps (1, 2, 3)

Test:
1. PATCH /api/v1/profiles/{brandProfileId}/onboarding/steps/{step4Id}
   Body: {"status": "skipped"}
   → Expect: 200 OK, status = "skipped"

2. GET /api/v1/profiles/{brandProfileId}/onboarding
   → Expect: activation_status = "pending_review"
   → Verify optional step 4 is skipped but not blocking activation
```

---

## Scenario 5: Next-Step Returns Correct Step

### Test: Partially complete onboarding → Next-step is correct

```
Setup:
- Create user with influencer profile
- Start and complete steps 1, 2

Test:
1. GET /api/v1/profiles/{profileId}/onboarding/next-step
   → Expect: Step 3 (follower verification) returned

2. Mark step 3 as in_progress, then completed

3. GET /api/v1/profiles/{profileId}/onboarding/next-step
   → Expect: Step 4 (payout preferences) returned

4. Mark step 4 as completed

5. GET /api/v1/profiles/{profileId}/onboarding/next-step
   → Expect: Step 5 (KYC) returned

6. Complete step 5

7. GET /api/v1/profiles/{profileId}/onboarding/next-step
   → Expect: {"step": null, "message": "All steps completed"}
```

---

## Scenario 6: Recalculate Adds Missing Steps

### Test: Template updated → Recalculate adds new steps without reverting

```
Setup:
- Create editor profile with v1.0 template
- Complete steps 1, 2

Simulate: New template version v1.1 with additional step

Test:
1. POST /api/v1/profiles/{profileId}/onboarding/recalculate
   → Expect: New step added to progress (step 5 if new)
   → Expect: Existing completed steps remain completed
   → Expect: activation_status unchanged

2. GET /api/v1/profiles/{profileId}/onboarding/steps
   → Expect: Original 4 steps + 1 new step
   → Expect: Steps 1, 2 still completed
   → Expect: New step status = "not_started"
```

---

## Scenario 7: Non-Owner Access Denied

### Test: User attempts to access another user's onboarding → 401

```
Setup:
- Create two users: userA, userB
- userA has editor profile with onboarding started

Test:
1. As userB, GET /api/v1/profiles/{userAProfileId}/onboarding
   → Expect: 401 Unauthorized

2. As userB, PATCH /api/v1/profiles/{userAProfileId}/onboarding/steps/...
   → Expect: 401 Unauthorized
```

---

## Scenario 8: Brand Onboarding Activation

### Test: Brand profile completes all required steps → Admin activates

```
Setup:
- Create brand profile

Test:
1. Complete all 4 steps (1, 2, 3 required; 4 optional)
2. GET /api/v1/profiles/{brandProfileId}/onboarding
   → Expect: activation_status = "pending_review"

3. As admin, POST /api/v1/admin/profiles/{brandProfileId}/onboarding/activate
   → Expect: 200 OK

4. GET /api/v1/profiles/{brandProfileId}/onboarding
   → Expect: activation_status = "activated"
   → Expect: activated timestamp set
```

---

## Scenario 9: Influencer Onboarding with Social Links

### Test: Influencer profile → Auto-complete social accounts check

```
Setup:
- Create influencer profile
- Set profile enrichment with social_links set

Test:
1. GET /api/v1/profiles/{profileId}/onboarding/steps
   → Expect: Step 2 (Add social accounts) status = "completed"
   → Auto-completed because social_links has entries

2. GET /api/v1/profiles/{profileId}/onboarding
   → Expect: percentage_complete > 0
```

---

## Scenario 10: Step Completion Is Permanent

### Test: Complete step → Remove source data → Step stays completed

```
Setup:
- Create editor profile
- Set profile enrichment (bio + avatar)
- Verify profile_completion auto-completes

Simulate: User removes bio from profile enrichment

Test:
1. GET /api/v1/profiles/{profileId}/onboarding/steps
   → Expect: Step 1 (profile_completion) status = "completed"
   → completed_at still set (not reverted)

2. POST /api/v1/profiles/{profileId}/onboarding/recalculate
   → Expect: Step 1 still "completed"
   → Verify no reversal occurred
```

---

## Running Quickstart Scenarios

```bash
# Run integration tests
go test ./tests/integration/... -v -run "TestOnboarding"

# Run specific scenario
go test ./tests/integration/... -v -run "TestOnboardingScenario1"

# Run with PostgreSQL
DATABASE_URL=postgres://... go test ./tests/integration/... -v
```

---

## Expected Results Summary

| Scenario | Key Assertions |
|----------|----------------|
| 1 | Full flow: not_started → onboarding → pending_review → activated |
| 2 | Auto-complete works from existing enrichment |
| 3 | Required step skip blocked with 403 |
| 4 | Optional skip allowed |
| 5 | Next-step returns correct step by display_order |
| 6 | Recalculate adds new steps, preserves existing |
| 7 | Non-owner access 401 |
| 8 | Admin activation works |
| 9 | Social links auto-complete |
| 10 | Completed steps stay completed (no reversal) |
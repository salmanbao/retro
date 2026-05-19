# Review Checklist: Submission Management Requirements

**Purpose**: Validate quality of Submission Management requirements
**Created**: 2026-05-19
**Feature**: [spec.md](../spec.md)

---

## Requirement Completeness

- [ ] CHK001 - Are all required submission fields (FR-1) specified with exact constraints (character limits, nullability)? [Completeness, Spec §FR-1]
- [ ] CHK002 - Are all 6 eligibility preconditions (FR-2) individually enumerated? [Completeness, Spec §FR-2]
- [ ] CHK003 - Are all 7 submission lifecycle states defined? [Completeness, Spec §FR-3]
- [ ] CHK004 - Are deadline validation requirements specified with a concrete field name and type? [Completeness, Spec §FR-2, Assumption #1]
- [ ] CHK005 - Are the required creative brief and asset existence checks specified with repository method names? [Completeness, Spec §FR-2, Assumption #3]
- [ ] CHK006 - Are soft deletion restoration requirements defined for Brand owners? [Completeness, Spec §FR-7, Assumption #7]

## Requirement Clarity

- [ ] CHK007 - Is "activated" onboarding status clearly defined and queryable? [Clarity, Spec §FR-2, Assumption #2]
- [ ] CHK008 - Are the terms "submitted" and "under_review" distinguished with explicit behavioral differences? [Clarity, Spec §FR-3]
- [ ] CHK009 - Is "shortlisted" state purpose explicitly defined relative to approved/rejected? [Clarity, Spec §FR-3]
- [ ] CHK010 - Is the "submission deadline" field name specified on the Campaign entity? [Clarity, Spec §FR-2, Assumption #1]
- [ ] CHK011 - Are tag limits (max count, max length per tag) specified? [Clarity, Spec §FR-1]
- [ ] CHK012 - Is "duration in seconds" bounded with a maximum value? [Clarity, Spec §FR-1]

## Requirement Consistency

- [ ] CHK013 - Does FR-4 Editing Rules table align with FR-3 State Machine invalid transitions? [Consistency, Spec §FR-3 vs §FR-4]
- [ ] CHK014 - Do FR-6 Ownership table actions align with all 6 API endpoints defined? [Consistency, Spec §FR-6 vs API Endpoints]
- [ ] CHK015 - Does FR-5 Duplicate Rules statement match the index definition in Data Model? [Consistency, Spec §FR-5 vs Data Model]
- [ ] CHK016 - Are "shortlisted" and "under_review" state transition triggers consistent (both mention external Brand review)? [Consistency, Spec §FR-3]

## Acceptance Criteria Quality

- [ ] CHK017 - Is criterion "Eligibility Enforcement" measurable with specific validation checkpoints? [Measurability, Spec §Success Criteria #1]
- [ ] CHK018 - Is "Immutability After Submit" criterion testable without implementation details? [Measurability, Spec §Success Criteria #3]
- [ ] CHK019 - Does "Single Submitted Entry" criterion explicitly exclude drafts from the count? [Measurability, Spec §Success Criteria #5]
- [ ] CHK020 - Are all 8 Success Criteria technology-agnostic (no frameworks, databases, or APIs mentioned)? [Measurability, Spec §Success Criteria]

## Scenario Coverage

- [ ] CHK021 - Are primary flows (Editor creates draft, submits, withdraws) each mapped to a user scenario? [Coverage, Spec §User Scenarios]
- [ ] CHK022 - Is the Brand owner viewing submissions flow covered as a separate user scenario? [Coverage, Spec §User Scenarios #4]
- [ ] CHK023 - Are all 6 edge cases in the Edge Cases section addressed by specific FRs or success criteria? [Coverage, Spec §Edge Cases]
- [ ] CHK024 - Is the case of Editor with multiple drafts for same campaign explicitly covered in FR-5? [Coverage, Spec §FR-5]
- [ ] CHK025 - Is there a scenario for when under_review submission transitions to shortlisted/rejected by external Brand action? [Coverage, Spec §FR-3]

## Edge Case Coverage

- [ ] CHK026 - Is behavior defined when an Editor attempts to create a submission after the deadline? [Edge Case, Spec §FR-2]
- [ ] CHK027 - Is behavior defined when an Editor tries to submit a campaign with no creative brief or assets? [Edge Case, Spec §FR-2, FR-8]
- [ ] CHK028 - Is behavior defined when a Brand owner attempts to create a submission? [Edge Case, Spec §FR-6]
- [ ] CHK029 - Is behavior defined when a non-editor profile type attempts any submission action? [Edge Case, Spec §FR-2, FR-6]
- [ ] CHK030 - Are concurrent submission attempts by the same Editor handled consistently? [Edge Case, Spec §FR-5, Conflict]

## Authorization Boundaries

- [ ] CHK031 - Does the spec define what happens when an Editor tries to edit another Editor's draft? [Authorization, Spec §FR-6]
- [ ] CHK032 - Does the spec define what a Brand owner can do with submissions (beyond listing)? [Authorization, Spec §FR-6]
- [ ] CHK033 - Is the Influencer profile type explicitly denied access in all scenarios? [Authorization, Spec §FR-6, Exclusion]
- [ ] CHK034 - Does Brand ownership check use the existing BrandProfileID field on Campaign? [Authorization, Spec §FR-6, Assumption #4]

## API Contract Quality

- [ ] CHK035 - Are all 6 API endpoints defined with method, path, auth requirement, and success response? [Completeness, Spec §API Endpoints]
- [ ] CHK036 - Is the 409 error (duplicate submission) defined for the POST endpoint? [Completeness, Spec §API Endpoints]
- [ ] CHK037 - Are PATCH /submissions/{id} field update constraints explicitly limited to draft status? [Clarity, Spec §API Endpoints]
- [ ] CHK038 - Is the error response format specified consistently across all endpoints? [Consistency, Spec §API Endpoints]
- [ ] CHK039 - Are the POST /submit and POST /withdraw actions idempotent or documented as such? [Clarity, Spec §API Endpoints]

## Data Model Integrity

- [ ] CHK040 - Are all 9 timestamps (created, updated, submitted, reviewed, withdrawn, deleted) accounted for with nullability rules? [Completeness, Spec §Data Model]
- [ ] CHK041 - Is the (campaign_id, editor_profile_id) unique constraint condition documented (excluding deleted + draft)? [Clarity, Spec §Data Model, FR-5]
- [ ] CHK042 - Are all 4 indexes defined with their specific query use cases? [Completeness, Spec §Data Model]
- [ ] CHK043 - Is the relationship cardinality between Campaign and Submission explicitly stated (1:N)? [Completeness, Spec §Data Model]

## Test Coverage Expectations

- [ ] CHK044 - Are unit test requirements defined for lifecycle state machine transitions? [Coverage, Spec §Testing Requirements]
- [ ] CHK045 - Are integration test requirements defined for PostgreSQL persistence? [Coverage, Spec §Testing Requirements]
- [ ] CHK046 - Are contract test requirements defined for all HTTP endpoints? [Coverage, Spec §Testing Requirements]
- [ ] CHK047 - Are authorization tests required for both Editor and Brand access patterns? [Coverage, Spec §Testing Requirements]

## Scope Exclusions

- [ ] CHK048 - Is "Brand review and scoring" correctly excluded from submission lifecycle transitions (only external triggers)? [Exclusion, Spec §FR-3, Exclusions]
- [ ] CHK049 - Are file upload handling and S3/CDN integration correctly excluded? [Exclusion, Spec §Exclusions, Assumption #5]
- [ ] CHK050 - Is payment/escrow correctly excluded with no references to financial state? [Exclusion, Spec §Exclusions]
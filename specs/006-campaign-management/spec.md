# Feature Specification: Campaign Management

**Feature Branch**: `006-campaign-management`

**Created**: 2026-05-18

**Status**: Draft

**Input**: User description: "Create a specification for 'Campaign Management' for ViralForge."

## User Scenarios & Testing

### User Story 1 - Campaign Creation (Priority: P1)

A Brand user creates a new campaign by providing core information, budget, timeline, eligibility criteria, and creative requirements. The system validates required fields and assigns a unique slug.

**Why this priority**: Campaign creation is the foundational use case; without it, no other campaign operations are possible.

**Independent Test**: Can be fully tested by creating a campaign with all required fields and verifying it appears in the campaign list with status "draft".

**Acceptance Scenarios**:

1. **Given** a Brand profile is activated and onboarded, **When** the user submits a campaign with all required fields (title, slug, budget, timeline), **Then** the system creates the campaign with status "draft" and returns the campaign ID.
2. **Given** a Brand profile is activated and onboarded, **When** the user submits a campaign with a duplicate slug, **Then** the system rejects the request with an error indicating slug uniqueness violation.
3. **Given** a Brand profile is activated and onboarded, **When** the user submits a campaign missing required fields, **Then** the system rejects the request with validation errors for each missing field.
4. **Given** a non-Brand profile (Editor or Influencer), **When** the user attempts to create a campaign, **Then** the system rejects the request with a forbidden error.

---

### User Story 2 - Campaign Publishing (Priority: P2)

A Brand user publishes a draft campaign after completing all readiness requirements. The system validates KYC status, onboarding completion, payout configuration, and campaign completeness before allowing publication.

**Why this priority**: Publishing transitions a campaign from draft to a live state where influencers can discover and engage with it. This is the core value-delivery moment.

**Independent Test**: Can be fully tested by creating a fully-configured campaign and publishing it, verifying status transitions from "draft" to "published" to "active".

**Acceptance Scenarios**:

1. **Given** a draft campaign with complete fields and zero budget, **When** the user attempts to publish, **Then** the system rejects with a validation error indicating budget must be greater than zero.
2. **Given** a draft campaign with complete fields and positive budget, but KYC not approved, **When** the user attempts to publish, **Then** the system rejects with a readiness error indicating KYC requirement not met.
3. **Given** a draft campaign that passes all readiness checks, **When** the user publishes the campaign, **Then** the system transitions the status to "published" and subsequently to "active" once the submission deadline is reached.
4. **Given** a Brand profile that is not fully onboarded, **When** the user attempts to publish any campaign, **Then** the system rejects with a readiness error indicating onboarding must be completed.

---

### User Story 3 - Campaign Editing (Priority: P3)

A Brand user modifies campaign details within constraints defined by the campaign lifecycle. Draft campaigns are fully editable; published or active campaigns have restricted field modifications.

**Why this priority**: Campaign optimization after initial creation is a common workflow; allowing edits while protecting critical lifecycle constraints balances flexibility with operational integrity.

**Independent Test**: Can be fully tested by creating a draft campaign, editing it, and verifying changes persist; then creating an active campaign and attempting restricted edits.

**Acceptance Scenarios**:

1. **Given** a draft campaign, **When** the user updates the title, description, budget, and eligibility criteria, **Then** the system saves all changes and reflects them in subsequent API responses.
2. **Given** a published campaign (status "published"), **When** the user attempts to change the total budget, **Then** the system rejects the request with a restricted-edit error.
3. **Given** an active campaign (status "active"), **When** the user attempts to modify the submission deadline, **Then** the system rejects the request with a restricted-edit error.
4. **Given** a paused campaign (status "paused"), **When** the user attempts to modify core eligibility criteria, **Then** the system accepts the changes and transitions the campaign back to "active" status.

---

### User Story 4 - Campaign Lifecycle Management (Priority: P2)

A Brand user controls campaign state transitions through explicit actions: pause, resume, complete, and cancel. Each transition is validated against current state and business rules.

**Why this priority**: Campaign lifecycle management enables Brands to respond to real-world conditions (budget exhaustion, performance issues, strategic changes) while maintaining operational control.

**Independent Test**: Can be fully tested by creating and publishing a campaign, then executing each lifecycle transition and verifying status changes and system behavior.

**Acceptance Scenarios**:

1. **Given** an active campaign, **When** the user sends a pause request, **Then** the system transitions the campaign to "paused" status and stops accepting new submissions.
2. **Given** a paused campaign, **When** the user sends a resume request, **Then** the system transitions the campaign to "active" status and resumes accepting submissions if within deadline.
3. **Given** an active campaign that has passed its end date, **When** the user sends a complete request, **Then** the system transitions the campaign to "completed" status and finalizes all metrics.
4. **Given** any non-cancelled campaign, **When** the user sends a cancel request, **Then** the system soft-deletes the campaign and transitions it to "cancelled" status.
5. **Given** a cancelled campaign, **When** any lifecycle action is attempted (pause, resume, complete), **Then** the system rejects the request with an invalid-transition error.

---

### User Story 5 - Campaign Discovery and Retrieval (Priority: P1)

A Brand user retrieves their campaign portfolio and individual campaign details. The system supports listing campaigns with filtering and pagination.

**Why this priority**: Campaign management requires visibility into all campaigns and the ability to inspect individual campaign state and details.

**Independent Test**: Can be fully tested by creating multiple campaigns and retrieving them via list and detail endpoints, verifying correct data and access control.

**Acceptance Scenarios**:

1. **Given** a Brand user with multiple campaigns, **When** the user requests a list of all their campaigns, **Then** the system returns only campaigns owned by the user's Brand profiles, ordered by creation date descending.
2. **Given** a Brand user with campaigns in multiple states (draft, published, completed), **When** the user filters by status "active", **Then** the system returns only active campaigns owned by the user.
3. **Given** a Brand user owns a specific campaign, **When** the user requests campaign details by ID, **Then** the system returns the full campaign object including all fields.
4. **Given** a Brand user attempts to access another Brand's campaign, **When** the user requests campaign details by ID, **Then** the system rejects with a not-found or forbidden error.

---

### Edge Cases

- What happens when campaign slug is auto-generated from title but title contains special characters? Slug generation should normalize to lowercase alphanumeric with hyphens.
- How does system handle concurrent campaign edits from the same user? Last-write-wins with optimistic concurrency control using version/timestamp.
- What occurs when a campaign's submission deadline passes while the campaign is still "published" (not yet "active")? System auto-transitions to "active" status on deadline.
- How does system behave when distribution start date is before submission deadline? System validates timeline consistency and rejects invalid configurations.
- What happens when budget is edited to zero for an "active" campaign? Restricted edit prevents this scenario.
- How does system handle minimum payout limit exceeding maximum payout limit? Validation rejects the configuration at creation/edit time.
- What occurs when allowed countries list is empty? System defaults to all countries allowed.

## Requirements

### Functional Requirements

- **FR-001**: System MUST allow Brand profiles to create campaigns with fields: title, slug, summary, description, objective, product/service name, landing page URL.
- **FR-002**: System MUST auto-generate a unique slug from the campaign title if not explicitly provided, normalized to lowercase alphanumeric with hyphens.
- **FR-003**: System MUST enforce slug uniqueness across all campaigns; duplicate slugs result in rejection.
- **FR-004**: System MUST allow Brand profiles to set budget fields: total budget, currency, target approved clips, target influencer posts, CPV, min/max payout limits.
- **FR-005**: System MUST allow Brand profiles to set timeline fields: submission start, submission deadline, distribution start, campaign end date.
- **FR-006**: System MUST validate that submission deadline is after submission start date, and campaign end date is after distribution start date.
- **FR-007**: System MUST allow Brand profiles to set eligibility criteria: allowed countries, allowed languages, minimum follower thresholds, supported platforms, creator categories.
- **FR-008**: System MUST allow Brand profiles to set creative requirements: video duration range, aspect ratio, mandatory talking points, prohibited claims, required hashtags, CTA instructions.
- **FR-009**: System MUST allow Brand profiles to reference campaign assets via URLs to uploaded files and supporting documents.
- **FR-010**: System MUST only allow Brand profile types to create campaigns; Editor and Influencer profiles receive a forbidden response.
- **FR-011**: System MUST enforce ownership validation: users can only access campaigns owned by their Brand profiles.
- **FR-012**: System MUST implement campaign lifecycle states: draft, published, active, paused, completed, cancelled.
- **FR-013**: System MUST allow lifecycle transitions only when valid from current state (e.g., draft→published, published→active, active→paused, etc.).
- **FR-014**: System MUST validate publishing readiness: brand must be fully onboarded, KYC approved, payout configured, required fields complete, budget > 0.
- **FR-015**: System MUST restrict edits to published or active campaigns: title, description, budget, eligibility criteria changes are rejected; only certain fields allowed.
- **FR-016**: System MUST soft-delete campaigns when cancelled, preserving data for audit but hiding from normal queries.
- **FR-017**: System MUST list campaigns with pagination (page, page_size) and filtering by status.
- **FR-018**: System MUST return campaign details including current status, progress metrics, and all configuration fields.
- **FR-019**: System MUST validate minimum payout limit is less than or equal to maximum payout limit.
- **FR-020**: System MUST support currency codes per ISO 4217 standard.

### Key Entities

- **Campaign**: Represents a brand's campaign containing core info (title, slug, description, objective, product, landing URL), budget configuration (total budget, currency, targets, CPV, payout limits), timeline (submission/distribution windows, end date), eligibility (countries, languages, followers, platforms, categories), creative requirements (duration, aspect, talking points, prohibited, hashtags, CTA), and asset references. Linked to Brand profile via `brand_profile_id`. Has lifecycle status and soft-delete timestamp.
- **CampaignAsset**: Represents a reference to an uploaded asset or document linked to a campaign. Contains URL, asset type, and optional description. Linked to Campaign via `campaign_id`.

## Success Criteria

### Measurable Outcomes

- **SC-001**: Brand users can create a campaign with all required fields in under 60 seconds from start to confirmation.
- **SC-002**: System prevents publishing of campaigns missing required fields, budget of zero, or unapproved KYC with 100% reliability.
- **SC-003**: Campaign lifecycle transitions (pause, resume, complete, cancel) execute within 5 seconds with consistent state updates.
- **SC-004**: Brand users can retrieve a paginated list of their campaigns with 100% accuracy in filtering by status.
- **SC-005**: Slug uniqueness is enforced with zero collisions across all campaigns in the system.
- **SC-006**: Soft-deleted campaigns are excluded from list queries but preserved in database for audit purposes.
- **SC-007**: Restricted edit validation prevents budget or eligibility modifications to published/active campaigns.
- **SC-008**: Ownership validation ensures users cannot access campaigns from other Brand profiles.
- **SC-009**: Timeline validation rejects invalid date configurations (deadline before start, end before distribution).
- **SC-010**: System handles concurrent campaign edits without data loss or corruption through optimistic concurrency control.

## Assumptions

- Users have completed email verification before interacting with campaign features.
- Campaign asset upload is handled by a separate file management service; this spec only references asset URLs.
- Influencer discovery of campaigns is handled by a separate module and is out of scope.
- Escrow and payment processing is handled by a separate financial module and is out of scope.
- Performance analytics for campaign effectiveness is handled by a separate analytics module and is out of scope.
- Default currency is USD if not explicitly specified by the Brand.
- Minimum follower threshold defaults to 0 (no minimum) if not specified.
- Video duration range defaults to 15-60 seconds if not specified.
- Campaign slugs are generated server-side using a normalized version of the title; Brands cannot manually set slugs that violate normalization rules.
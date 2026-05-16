# Feature Specification: Profile Enrichment and Verification

**Feature Branch**: `003-profile-enrichment`

**Created**: 2026-05-16

**Status**: Draft

**Input**: User description: "Profile Enrichment and Verification module for ViralForge marketplace"

## User Scenarios & Testing

### User Story 1 - User Enriches Public Profile (Priority: P1)

A Brand, Editor, or Influencer user updates their profile with public-facing information to establish their identity on the platform.

**Why this priority**: Core profile functionality needed by all profile types before marketplace participation.

**Independent Test**: Create a profile, enrich with bio/avatar/location/social links, verify data persisted and returned correctly.

**Acceptance Scenarios**:

1. **Given** an authenticated user with an existing profile, **When** they submit updated public profile information (bio, avatar, cover, website, location, languages, timezone), **Then** the information is stored and returned on subsequent retrieval.

2. **Given** a user viewing another profile, **When** they fetch profile details, **Then** they see the public information but not sensitive payout or KYC data.

3. **Given** an authenticated user, **When** they attempt to modify another user's profile details, **Then** the request is denied with 403 Forbidden.

---

### User Story 2 - Editor Manages Portfolio (Priority: P2)

An Editor user creates, updates, and removes portfolio items to showcase their work to potential clients.

**Why this priority**: Portfolio is the primary discovery mechanism for Editors in the marketplace.

**Independent Test**: Assign Editor role to profile, create portfolio items, verify CRUD operations work correctly.

**Acceptance Scenarios**:

1. **Given** an Editor profile, **When** they create a portfolio item with title, description, thumbnail, video URL, external link, and display order, **Then** the item is stored and associated with their profile.

2. **Given** an Editor with existing portfolio items, **When** they update one item's description or reorder position, **Then** changes are persisted.

3. **Given** an Editor with existing portfolio items, **When** they delete an item, **Then** the item is soft-deleted (preserved for audit, hidden from display).

4. **Given** a Brand or Influencer profile, **When** they attempt to create a portfolio item, **Then** the request is denied with 403 Forbidden.

5. **Given** portfolio items exist for an Editor, **When** a Brand views the Editor's profile, **Then** portfolio items are visible in the response.

---

### User Story 3 - Influencer Manages Audience Data (Priority: P3)

An Influencer user stores audience demographics and platform metrics to help Brands make informed partnership decisions.

**Why this priority**: Audience data is key differentiator for Influencer profiles in the marketplace.

**Independent Test**: Assign Influencer role to profile, add audience data with platform handles and demographics, verify retrieval.

**Acceptance Scenarios**:

1. **Given** an Influencer profile, **When** they submit audience data including platform handles, claimed follower counts, engagement rate, and demographics JSON, **Then** the data is stored and associated with their profile.

2. **Given** an Influencer with existing audience data, **When** they update follower counts or demographics, **Then** changes are persisted.

3. **Given** a Brand or Editor profile, **When** they attempt to add audience data, **Then** the request is denied with 403 Forbidden.

---

### User Story 4 - Influencer Submits Follower Verification (Priority: P3)

An Influencer submits evidence of their follower count for platform verification to build trust with Brands.

**Why this priority**: Verification status is a trust signal that affects conversion rates.

**Independent Test**: Assign Influencer role, submit verification evidence, verify status transitions work correctly.

**Acceptance Scenarios**:

1. **Given** an Influencer profile, **When** they submit verification evidence URLs and notes, **Then** the submission is stored with status "pending" and timestamp.

2. **Given** an Influencer with pending verification, **When** an internal admin reviews and approves/rejects, **Then** status updates to "verified" or "rejected" with timestamp and notes.

3. **Given** a Brand or Editor profile, **When** they attempt to submit verification, **Then** the request is denied with 403 Forbidden.

---

### User Story 5 - User Manages Payout Preferences (Priority: P4)

A user configures their payout destination and method to receive payments from the platform.

**Why this priority**: Required before first payout can be processed.

**Independent Test**: Add payout preferences, verify data persisted but sensitive fields never returned in plaintext.

**Acceptance Scenarios**:

1. **Given** any profile type, **When** they submit payout preferences with method, beneficiary name, country, currency, and encrypted details, **Then** the preferences are stored securely.

2. **Given** payout preferences exist, **When** a user retrieves their own payout preferences, **Then** the response does not contain decrypted payout details (returns masked or redacted).

3. **Given** an authenticated user, **When** they view another user's profile, **Then** payout preferences are not included in the response.

---

### User Story 6 - User Views KYC Status (Priority: P4)

A user checks their Know Your Customer (KYC) verification status to understand platform compliance standing.

**Why this priority**: KYC status affects出金 and certain platform features.

**Independent Test**: Set KYC status via admin service, verify user can read their status but cannot modify it directly.

**Acceptance Scenarios**:

1. **Given** any profile type, **When** they check their KYC status, **Then** they see current status (not_started, pending, approved, rejected, suspended) and review notes if any.

2. **Given** a user with KYC status of "rejected" or "suspended", **When** they view their status, **Then** review notes explaining the decision are visible.

3. **Given** an authenticated user, **When** they attempt to directly modify their own KYC status, **Then** the request is denied with 403 Forbidden (admin-only operation).

---

### Edge Cases

- What happens when a user without a profile type tries to access profile details? → Return 404 with clear message to create profile first.
- How does the system handle malformed JSON in audience demographics? → Return 400 with validation error.
- What happens when a portfolio item is deleted but its ID is requested directly? → Return 404 (soft-deleted items are not accessible by ID).
- How does the system handle concurrent edits to the same profile? → Last-write-wins with optimistic concurrency (updated_at timestamp).
- What happens when payout details are submitted with an invalid country/currency combination? → Return 400 with validation error.
- How does the system handle empty social media link fields? → Allow null/empty, skip validation for empty fields.
- What happens when verification evidence URLs are invalid? → Accept as-is (external validation not performed). URL format validation is performed but not host accessibility validation.
- How does the system handle role type changes (e.g., Editor becomes Influencer)? → Portfolio items from previous role type become inaccessible but not deleted. Role type change is detected when profile_type field is updated; system validates access based on current profile_type on each request.
- How does the system handle status transitions for verification and KYC? → Any state transition is allowed; status_history table records all transitions for audit purposes. Rejected/verified status can be revisited via new submission.
- How does the system handle portfolio items with identical display_order values? → Use created_at timestamp as secondary sort key (earlier items first).
- How does the system validate JSONB field inputs (social_links, demographics, evidence_urls)? → String content validated for SQL injection patterns before storage. Max field size enforced: social_links 5KB, demographics 10KB, evidence_urls 50KB per entry.

---

## Clarifications

### Session 2026-05-16

- Q: Status transition rules for verification and KYC → A: Any state transition allowed; status_history table records all transitions. Rejected status can return to pending via new submission.
- Q: Portfolio ordering tiebreaker when display_order values are identical → A: Use created_at timestamp as secondary sort (earlier created_at first).
- Q: Encryption responsibility for sensitive payout data → A: Database-layer encryption. DB handles transparent encryption/decryption (PostgreSQL TDE or pgcrypto); app receives plaintext; app never stores or handles raw ciphertext.
- Q: Admin API boundary for verification/KYC status changes → A: Admin endpoints use internal service-to-service authentication (JWT or mTLS). Admin service bypasses normal profile ownership checks. Admin endpoints not exposed via public REST API.
- Q: Profile Details storage structure → A: JSON embedded in ProfileDetails. SocialLinks stored as JSON column within ProfileDetails. Single table query for full profile details (avoids joins). Atomic updates. Schema migration simplicity preferred over independent social link queries.
- Q: Portfolio item display order reset behavior → A: No auto-renumber after delete. Gaps in display_order are preserved. Delete is a single operation (no cascading updates to other items). Users can explicitly reorder via PATCH if desired.

---

## Requirements

### Functional Requirements

- **FR-001**: System MUST allow authenticated users to retrieve their own profile details including bio, avatar URL, cover image URL, website URL, location, languages, and timezone.
- **FR-002**: System MUST allow authenticated users to update their own profile details via PATCH request with partial data.
- **FR-003**: System MUST enforce ownership verification: users may only access and modify profiles they own (same profile ID as authenticated session).
- **FR-004**: System MUST allow Editor profiles to create portfolio items with title, description, thumbnail URL, video URL, external link, and display order.
- **FR-005**: System MUST allow Editor profiles to update their own portfolio items via PATCH request.
- **FR-006**: System MUST allow Editor profiles to delete portfolio items via soft-delete (preserving record with deleted_at timestamp).
- **FR-007**: System MUST list portfolio items ordered by display_order ascending for Editor profiles.
- **FR-008**: System MUST reject portfolio operations for non-Editor profiles with 403 Forbidden.
- **FR-009**: System MUST allow Influencer profiles to store and update audience data including platform handles, follower counts, engagement rate, and demographics JSON.
- **FR-010**: System MUST reject audience data operations for non-Influencer profiles with 403 Forbidden.
- **FR-011**: System MUST allow Influencer profiles to submit follower verification evidence with status "pending" and timestamp.
- **FR-012**: System MUST allow internal admin services to update follower verification status to verified/rejected with notes and timestamp.
- **FR-013**: System MUST reject follower verification submission from non-Influencer profiles with 403 Forbidden.
- **FR-014**: System MUST allow any profile type to store and update payout preferences including method, beneficiary name, country, currency, encrypted details, and readiness flag.
- **FR-015**: System MUST never return encrypted payout details in plaintext in any API response (masked or omitted entirely).
- **FR-016**: System MUST allow any profile type to view their own KYC status and review notes.
- **FR-017**: System MUST reject direct user modification of KYC status (admin-only operation) with 403 Forbidden.
- **FR-018**: System MUST soft-delete portfolio items (preserve record, set deleted_at, exclude from normal queries).
- **FR-019**: System MUST validate that languages field contains valid ISO language codes.
- **FR-020**: System MUST validate that timezone field contains valid IANA timezone identifiers.

### Key Entities

- **ProfileDetails**: Public profile information attached to a profile. Fields: bio (text), avatar_url (url), cover_image_url (url), website_url (url), location (string), languages (string[]), timezone (string), social_links (json - embedded tiktok, instagram, youtube, x_twitter, linkedin, website), updated_at (timestamp).
- **PortfolioItem**: Work samples for Editor profiles. Fields: id (uuid), profile_id (fk), title (string), description (text), thumbnail_url (url), video_url (url), external_link (url), display_order (int), created_at, updated_at, deleted_at (nullable).
- **AudienceData**: Audience metrics for Influencer profiles. Fields: profile_id (fk, unique), platform_handles (json), claimed_followers (json), engagement_rate (decimal), audience_demographics (json), updated_at (timestamp).
- **FollowerVerification**: Follower count verification evidence. Fields: profile_id (fk, unique), status (enum: unverified/pending/verified/rejected), evidence_urls (text[]), verification_notes (text, nullable), reviewed_at (timestamp, nullable), reviewed_by (string, nullable).
- **PayoutPreferences**: Payment destination settings. Fields: profile_id (fk, unique), preferred_method (enum: bank_transfer/paypal/crypto), beneficiary_name (string), country (string), currency (string), encrypted_details (text), payout_ready (boolean), updated_at (timestamp).
- **KYCStatus**: Know Your Customer compliance status. Fields: profile_id (fk, unique), status (enum: not_started/pending/approved/rejected/suspended), review_notes (text, nullable), reviewed_at (timestamp, nullable), reviewed_by (string, nullable).

### Non-Functional Requirements

- **NFR-001**: Payout details encryption must use AES-256-GCM or equivalent at-rest encryption.
- **NFR-002**: API responses for profile details must complete within 200ms under normal load.
- **NFR-003**: System must support portfolio items up to 50 items per Editor profile.
- **NFR-004**: Rate limiting: Profile enrichment endpoints limited to 100 requests per minute per user; portfolio endpoints limited to 60 requests per minute per user.

---

## Success Criteria

### Measurable Outcomes

- **SC-001**: Users can successfully enrich their profile with all public information fields within 5 minutes of starting the flow.
- **SC-002**: Editors can create, view, update, and delete portfolio items with each operation completing within 2 seconds.
- **SC-003**: Influencers can store audience data with demographics JSON up to 10KB without errors.
- **SC-004**: Verification submissions are acknowledged within 1 second with pending status confirmation.
- **SC-005**: Payout preferences are saved and retrieved without exposing sensitive details (100% masking of encrypted fields).
- **SC-006**: All profile type restrictions are enforced correctly (zero unauthorized access to role-specific features).
- **SC-007**: Soft-deleted items are excluded from normal queries but preserved in database for audit purposes.

---

## Assumptions

- Existing profile entities are already stored in the database and share the same ID system.
- Authentication middleware already extracts active profile ID from session/token.
- Admin services are internal services not exposed via public REST API for verification/KYC status changes.
- Encryption key management is handled by existing infrastructure (Vault or similar).
- Timezone validation uses standard IANA timezone database.
- Languages field accepts ISO 639-1 two-letter codes.
- Currency codes follow ISO 4217 standard.
- Country codes follow ISO 3166-1 alpha-2 standard.
- Existing user registration flow creates a default profile if one does not exist.
- Database handles transparent encryption for payout_details (application receives plaintext only; never stores or handles raw ciphertext).
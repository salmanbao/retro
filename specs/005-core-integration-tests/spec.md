# Feature Specification: Core Module Integration Test Suite

**Feature**: 005-core-integration-tests
**Created**: 2026-05-18
**Objective**: Build a comprehensive integration and end-to-end test suite that verifies Authentication, Authorization, Profile Enrichment, and Role-Based Onboarding modules work correctly together.

---

## Clarifications

### Session 2026-05-18

- Q: Database isolation strategy for parallel test execution → A: Test containers (docker/podman spawned per test suite run)
- Q: Test data management approach for scenarios → A: Scenario factories (each test creates fresh data via API calls it needs)
- Q: Parallel test execution strategy → A: Sequential execution (one test at a time, simpler debugging)

### Completed Modules

| Module | Status |
|--------|--------|
| Authentication | Complete |
| Authorization | Complete |
| Profile Enrichment and Verification | Complete |
| Role-Based Onboarding and Activation | Complete |

### Integration Goal

Verify that all modules function together without contract mismatches, authorization boundaries are enforced, and onboarding reacts correctly to profile enrichment actions.

---

## User Scenarios & Testing

### Scenario 1: User Registration and Verification Flow

**Test**: Register new user → Verify email → Log in successfully

```
Given a new email address
When I submit registration with valid credentials
Then I receive a verification email
And my account is in "pending verification" state

When I click the verification link
Then my account becomes "verified"
And I can log in with valid credentials
And I receive an authentication session
```

### Scenario 2: Role Profile Creation

**Test**: Create Brand, Editor, and Influencer profiles with proper authorization

```
Given an authenticated user
When I create a Brand profile
Then the profile is created with type "Brand"
And I am the owner of the profile

When I create an Editor profile
Then the profile is created with type "Editor"
And I am the owner of the profile

When I create an Influencer profile
Then the profile is created with type "Influencer"
And I am the owner of the profile
```

### Scenario 3: Profile Enrichment Operations

**Test**: Enrich profiles with details, links, portfolio items, payout preferences, and KYC status

```
Given an authenticated user with an Editor profile
When I update profile details (bio, avatar)
Then the details are persisted

When I add social links (Twitter, Instagram)
Then the links are stored

When I create portfolio items
Then the items are associated with my profile

When I configure payout preferences
Then the preferences are stored securely

When I submit KYC status
Then the status is recorded
```

### Scenario 4: Onboarding Initialization

**Test**: Onboarding progress is automatically created for new profiles

```
Given an authenticated user
When I create a new profile
Then onboarding progress is automatically created
And the profile type matches the correct template

For Brand profile:
- Template includes: company_profile, payout_preferences, kyc, first_campaign

For Editor profile:
- Template includes: public_profile, portfolio, payout_preferences, kyc

For Influencer profile:
- Template includes: public_profile, social_accounts, follower_verification, payout_preferences, kyc
```

### Scenario 5: Automatic Step Completion

**Test**: Completing profile data satisfies related onboarding steps

```
Given an authenticated user with an Editor profile
And incomplete onboarding progress
When I complete profile enrichment (bio + avatar)
Then the profile_completion step is auto-completed
And the completion percentage is updated

When I submit KYC with "approved" status
Then the kyc step is auto-completed
And the completion percentage is updated
```

### Scenario 6: Activation Progression

**Test**: Required steps drive activation state transitions

```
Given an authenticated user with an Editor profile
When I complete all required steps
Then the activation_status becomes "pending_review"

When an admin activates the profile
Then the activation_status becomes "activated"
And the profile is marketplace-ready
```

### Scenario 7: Security and Access Control

**Test**: Users cannot access or modify other users' resources

```
Given User A and User B with separate profiles
When User A attempts to GET User B's profile details
Then the request is rejected with 403 Forbidden

When User A attempts to PATCH User B's onboarding progress
Then the request is rejected with 403 Forbidden

When User A attempts to access User B's onboarding steps
Then the request is rejected with 403 Forbidden
```

### Scenario 8: Session-Based Access

**Test**: Only authenticated requests succeed; invalid sessions are rejected

```
Given an invalid or expired session token
When I make an authenticated request
Then the request is rejected with 401 Unauthorized

Given a valid session token
When I make an authenticated request
Then the request succeeds
```

---

## Functional Requirements

### R001: Registration and Authentication Flow

- New users can register with email and password
- Registration triggers email verification workflow
- Users cannot log in until email is verified
- Successful login establishes an authenticated session

### R002: Multi-Profile Support

- Users can create multiple profiles of different types (Brand, Editor, Influencer)
- Each profile is independently owned by the creating user
- Profile type determines onboarding template assignment

### R003: Profile Enrichment CRUD

- Profile owners can update their profile details
- Profile owners can add/remove social links
- Editors can create, update, delete portfolio items
- Profile owners can configure payout preferences
- KYC status can be submitted and updated

### R004: Onboarding Progress Lifecycle

- Onboarding progress is created automatically upon profile creation
- Progress reflects the correct template for the profile type
- Step statuses can be updated (not_started → in_progress → completed)
- Required steps cannot be skipped

### R005: Auto-Completion Triggers

- Profile enrichment completion (bio + avatar) auto-completes profile_completion step
- KYC "approved" status auto-completes kyc step
- Payout preferences with encrypted_details auto-completes payout_preferences step
- Social links presence auto-completes social_links step

### R006: Activation State Transitions

- Activation transitions: not_started → onboarding → pending_review → activated
- State changes occur when required steps are completed
- Admin approval transitions pending_review to activated
- Activated profiles are marketplace-eligible

### R007: Authorization Boundaries

- Profile owners can only access/modify their own profiles
- Onboarding progress is accessible only to profile owners
- Admin endpoints require admin role verification
- All protected endpoints reject requests without valid sessions

### R008: Session Management

- Authenticated sessions are validated on each request
- Invalid/expired sessions return 401 Unauthorized
- Ownership middleware enforces resource-level access control

---

## Success Criteria

### SC1: Module Integration

- All four modules (Auth, AuthZ, Enrichment, Onboarding) function together
- No contract mismatches between module interfaces
- Data flows correctly across module boundaries

### SC2: Core User Journeys

- User can register, verify email, and log in
- User can create profiles of all three types
- User can enrich profiles with all supported data types
- Onboarding progress is created and updates correctly

### SC3: Authorization Enforcement

- Users cannot access other users' profiles or data
- Ownership rules are enforced at all endpoints
- Admin-only endpoints reject non-admin users

### SC4: Onboarding Reactivity

- Profile enrichment triggers auto-completion of related steps
- Completion percentage accurately reflects progress
- Activation state transitions are deterministic

### SC5: Test Coverage

- Integration tests cover all cross-module interactions
- Contract tests verify HTTP endpoint behavior
- End-to-end tests cover complete user journeys

---

## Key Entities

### User
- id (UUID)
- email
- password_hash
- email_verified
- created_at, updated_at

### Session
- id (UUID)
- user_id (FK)
- token_hash
- expires_at

### Profile
- id (UUID)
- user_id (FK)
- type (Brand|Editor|Influencer)
- created_at, updated_at

### ProfileEnrichment
- profile_id (FK)
- bio, avatar_url
- social_links (JSONB)
- updated_at

### PortfolioItem
- id (UUID)
- profile_id (FK)
- title, description, url
- display_order

### PayoutPreferences
- profile_id (FK)
- encrypted_details
- created_at, updated_at

### KYCStatus
- profile_id (FK)
- status (pending|approved|rejected)
- submitted_at

### OnboardingProgress
- id (UUID)
- profile_id (FK)
- profile_type
- template_id
- activation_status
- started_at, last_activity_at

### StepProgress
- id (UUID)
- onboarding_progress_id (FK)
- step_id (FK)
- status (not_started|in_progress|completed|skipped)
- started_at, completed_at

---

## Assumptions

- PostgreSQL is used as the database for all integration tests
- HTTP endpoints are tested using contract test patterns
- Email verification uses token-based links (not SMS)
- Admin role is assigned directly to user accounts
- Session tokens are stored in HTTP-only cookies or headers
- Onboarding templates are seeded via database migration
- **Database isolation**: Test containers (podman) spawned per test suite run to ensure clean state and parallel execution safety
- **Test data management**: Scenario factories — each test creates fresh data via API calls it needs, ensuring true test isolation
- **Test execution**: Sequential (one test at a time) for simpler debugging and reliable reproduction of failures

---

## Explicitly Excluded

- Payment processing and escrow functionality
- External KYC provider integrations
- Social media API integrations
- Push notifications or email reminders
- Gamification features

---

## Dependencies

This feature depends on all four modules being complete and deployed. It cannot be implemented until:
- Authentication module is stable
- Authorization module is stable  
- Profile Enrichment module is stable
- Onboarding module is stable
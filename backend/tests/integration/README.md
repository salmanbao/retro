# Integration Test Suite

**Feature**: Core Module Integration Test Suite (005)
**Location**: `backend/tests/integration/`

## Overview

This test suite verifies the integration between completed ViralForge modules:
- Authentication
- Authorization
- Profile Enrichment and Verification
- Role-Based Onboarding and Activation

## Running Tests

### Prerequisites

1. **Podman installed** and running:
   ```bash
   podman --version
   ```

2. **Start PostgreSQL container**:
   ```bash
   cd backend
   podman-compose -f podman-compose.test.yml up -d
   ```

### Run All Integration Tests

```bash
cd backend
go test ./tests/integration/... -v
```

### Run Specific Test Files

```bash
# Auth tests
go test ./tests/integration/auth_*.go -v

# Profile tests
go test ./tests/integration/profile_*.go -v

# Onboarding tests
go test ./tests/integration/onboarding_*.go -v

# Security tests
go test ./tests/integration/security_*.go -v

# Activation tests
go test ./tests/integration/activation_*.go -v
```

### Run Unit Tests

```bash
go test ./tests/unit/... -v
```

## Test Structure

### Phase 1: Infrastructure
- `tests/fixtures/container.go` - Podman container management
- `tests/fixtures/client.go` - HTTP test client with session handling
- `tests/fixtures/factory.go` - Scenario factory helpers
- `tests/integration/suite.go` - TestSuite base class

### Phase 2: Auth Workflows (US1)
- `tests/integration/auth_registration_test.go` - User registration
- `tests/integration/auth_verification_test.go` - Email verification
- `tests/integration/auth_login_test.go` - Login with credentials
- `tests/integration/auth_login_before_verification_test.go` - Session management

### Phase 3: Profile Creation & Enrichment (US2, US3)
- `tests/integration/profile_brand_test.go` - Brand profile creation
- `tests/integration/profile_editor_test.go` - Editor profile creation
- `tests/integration/profile_influencer_test.go` - Influencer profile creation
- `tests/integration/profile_enrichment_test.go` - Bio/avatar updates
- `tests/integration/profile_social_links_test.go` - Social links management
- `tests/integration/profile_portfolio_test.go` - Editor portfolio CRUD
- `tests/integration/profile_payout_test.go` - Payout preferences
- `tests/integration/profile_kyc_test.go` - KYC status submission

### Phase 4: Onboarding Verification (US4, US5)
- `tests/integration/onboarding_init_test.go` - Auto-creation of onboarding
- `tests/integration/onboarding_template_test.go` - Template assignment
- `tests/integration/onboarding_auto_complete_test.go` - Profile completion triggers
- `tests/integration/onboarding_kyc_auto_trigger_test.go` - KYC auto-trigger
- `tests/integration/onboarding_payout_auto_trigger_test.go` - Payout auto-trigger
- `tests/integration/onboarding_social_links_auto_trigger_test.go` - Social links auto-trigger

### Phase 5: Security & Authorization (US7, US8)
- `tests/integration/security_access_test.go` - Cross-user access denial
- `tests/integration/security_session_test.go` - Session validation

### Phase 6: End-to-End Activation (US6)
- `tests/integration/activation_flow_test.go` - State progression
- `tests/integration/activation_required_step_blocking_test.go` - Required steps block activation
- `tests/integration/activation_admin_test.go` - Admin approval/rejection
- `tests/integration/activation_marketplace_test.go` - Marketplace eligibility
- `tests/integration/activation_optional_step_test.go` - Optional step skipping

## Test Coverage

| Scenario | Description | Test File(s) |
|----------|-------------|--------------|
| 1 | User Registration and Verification | `auth_registration_test.go`, `auth_verification_test.go`, `auth_login_test.go` |
| 2 | Role Profile Creation | `profile_brand_test.go`, `profile_editor_test.go`, `profile_influencer_test.go` |
| 3 | Profile Enrichment | `profile_enrichment_test.go`, `profile_social_links_test.go`, `profile_portfolio_test.go`, `profile_payout_test.go`, `profile_kyc_test.go` |
| 4 | Onboarding Initialization | `onboarding_init_test.go`, `onboarding_template_test.go` |
| 5 | Automatic Step Completion | `onboarding_auto_complete_test.go`, `onboarding_*_auto_trigger_test.go` |
| 6 | Activation Progression | `activation_flow_test.go`, `activation_*_test.go` |
| 7 | Security and Access Control | `security_access_test.go` |
| 8 | Session-Based Access | `security_session_test.go` |

## Notes

- Integration tests require a running server (tests skip if server unavailable)
- Some endpoints may not be implemented yet - tests handle this gracefully with `t.Skip()`
- Test client uses cookie-jar for session management
- Each test creates fresh data via API calls (no shared state)

## Troubleshooting

**Tests skip with "Login failed"**
- Server may require email verification
- Check if `AUTH_BYPASS_VERIFICATION` is set in test mode

**Container not starting**
```bash
podman-compose -f podman-compose.test.yml down
podman-compose -f podman-compose.test.yml up -d
```

**Port conflicts**
- Check if port 5432 is already in use
- Modify `podman-compose.test.yml` to use different port
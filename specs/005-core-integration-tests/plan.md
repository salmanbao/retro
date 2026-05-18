# Implementation Plan: Core Module Integration Test Suite

**Branch**: `005-core-integration-tests` | **Date**: 2026-05-18 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/005-core-integration-tests/spec.md`

## Summary

Build a comprehensive integration and end-to-end test suite that verifies Authentication, Authorization, Profile Enrichment, and Role-Based Onboarding modules work correctly together. Tests execute via HTTP-based workflows against PostgreSQL using container-based isolation, with scenario factories for test data management.

## Technical Context

**Language/Version**: Go 1.23+

**Primary Dependencies**: chi router, GORM PostgreSQL driver, testcontainers-go (container management)

**Storage**: PostgreSQL (via test containers)

**Testing**: Go testing package with integration/contract test patterns

**Target Platform**: Linux server (CI/CD)

**Project Type**: Integration test suite (internal testing infrastructure)

**Performance Goals**: Full scenario execution < 30s per scenario

**Constraints**: No production code introduced solely for testing; sequential test execution

**Scale/Scope**: 8 user scenarios across 4 modules

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Status | Notes |
|------|--------|-------|
| Test-First Development | PASS | Integration tests are the primary deliverable |
| No production code for testing | PASS | Test infrastructure only, no feature code changes |
| Simplicity Over Sophistication | PASS | Sequential execution, container isolation, scenario factories |
| Specification as Single Source of Truth | PASS | All scenarios defined in spec.md |

## Project Structure

### Documentation (this feature)

```text
specs/005-core-integration-tests/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output (test entities only)
├── quickstart.md        # Phase 1 output (test scenarios)
├── contracts/           # Phase 1 output (HTTP endpoint contracts)
└── tasks.md             # Phase 2 output (/speckit-tasks command)
```

### Source Code (repository root)

```text
backend/
└── tests/
    ├── integration/         # Full HTTP-based integration tests
    │   └── *_integration_test.go
    ├── contract/            # HTTP endpoint contract tests
    │   └── *_contract_test.go
    └── fixtures/             # Test helpers and factories
        ├── container.go      # Podman container management
        ├── factory.go        # Scenario factory helpers
        └── client.go         # HTTP test client
```

**Structure Decision**: Test infrastructure lives under `backend/tests/` alongside existing unit tests. Integration tests use HTTP client to exercise full stack. Contract tests verify endpoint behavior.

## Research

### Container-Based Database Isolation

**Decision**: Use `testcontainers` Go library with podman

**Rationale**:
- Provides clean database per test run
- Automatically handles container lifecycle (start/stop/cleanup)
- Supports PostgreSQL natively
- Go-native solution

**Alternatives considered**:
- Raw podman CLI: More control but manual lifecycle management
- Docker SDK: Not podman-native, may have compatibility issues
- Transaction rollback: Doesn't work with true HTTP tests that span multiple requests

### Test Data Management

**Decision**: Scenario factories — each test creates fresh data via API calls

**Rationale**:
- Ensures true test isolation
- Exercises the full API stack including auth flows
- Each test is independent and can run in any order
- Easier to debug (no shared state)

**Alternatives considered**:
- Centralized seed data: Creates coupling between tests, harder to debug
- Golden files: Brittle, requires manual updates when API changes

## Phase 1: Design

### Data Model (Test Entities)

No new production entities. Integration tests validate existing domain entities:
- User (registration, verification)
- Session (authentication)
- Profile (multi-profile support)
- ProfileEnrichment (bio, avatar, social links)
- PortfolioItem (editor-specific)
- PayoutPreferences (encrypted details)
- KYCStatus (verification status)
- OnboardingProgress / StepProgress (onboarding flow)

### Contracts

**HTTP Test Client Interface**:
```go
type TestClient struct {
    BaseURL    string
    HttpClient *http.Client
    SessionCookie *http.Cookie
}

func (c *TestClient) Register(email, password string) (*RegisterResponse, error)
func (c *TestClient) VerifyEmail(token string) error
func (c *TestClient) Login(email, password string) error
func (c *TestClient) CreateProfile(profileType string) (*ProfileResponse, error)
func (c *TestClient) EnrichProfile(profileID string, data EnrichmentData) error
func (c *TestClient) GetOnboarding(profileID string) (*OnboardingResponse, error)
func (c *TestClient) UpdateStep(profileID, stepID, status string) error
```

**Container Lifecycle Interface**:
```go
type ContainerManager struct {
    ContainerID string
    DSN         string
}

func (m *ContainerManager) Start(ctx context.Context) error
func (m *ContainerManager) Stop(ctx context.Context) error
func (m *ContainerManager) RunMigrations(db *gorm.DB) error
```

### Quickstart (Test Scenarios)

1. **User Registration and Verification Flow**
2. **Role Profile Creation**
3. **Profile Enrichment Operations**
4. **Onboarding Initialization**
5. **Automatic Step Completion**
6. **Activation Progression**
7. **Security and Access Control**
8. **Session-Based Access**

## Complexity Tracking

No violations requiring justification.

## Implementation Phases

| Phase | Tasks |
|-------|-------|
| Phase 1 | Container management, test client, migrations |
| Phase 2 | Scenario factories for each module |
| Phase 3 | Integration tests for each scenario |
| Phase 4 | Contract tests for HTTP endpoints |
| Phase 5 | Final validation and cleanup |

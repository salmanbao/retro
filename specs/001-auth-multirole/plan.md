# Implementation Plan: Authentication and Multi-Role User Management

**Branch**: `main` | **Date**: 2026-05-15 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification updated with security enhancements from clarification session

## Summary

Build authentication and multi-role user management for ViralForge — a three-sided marketplace. Users register with email/password, verify via email, maintain multiple sessions, reset passwords, and manage Brand/Editor/Influencer profiles with role-based authorization.

**Technical approach**: Go backend using chi router, PostgreSQL for persistence, bcrypt for password hashing (cost 12), server-side session tokens (UUID), adapter pattern for email.

**Security additions (from clarification)**:
- Account lockout: 5 failed attempts → 15-minute lock
- Session regeneration on login (fixation prevention)
- CSRF protection: SameSite=Strict + X-CSRF-Token header
- Token expiry stored in `expires_at` column (not business logic)

## Technical Context

**Language/Version**: Go 1.21+

**Primary Dependencies**:
- `github.com/go-chi/chi/v5` — HTTP routing
- `github.com/jackc/pgx/v5` — PostgreSQL driver
- `golang.org/x/crypto/bcrypt` — Password hashing
- `github.com/google/uuid` — UUID generation

**Storage**: PostgreSQL (sessions, users, profiles, auth_tokens)

**Testing**: Go stdlib `testing` + `github.com/stretchr/testify`

**Target Platform**: Linux server (REST API)

**Project Type**: Web service (REST API)

**Performance Goals**: 10k concurrent sessions, 99% auth API success rate

**Constraints**: Session timeout 24h, password reset tokens expire in 1h, verification tokens expire in 24h

**Scale/Scope**: 10k concurrent users, 6 user stories

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Rule | Status | Notes |
|------|--------|-------|
| Specification is law | ✓ PASS | All features traceable to spec |
| Simplicity mandatory | ✓ PASS | Single DB, no unnecessary abstractions |
| One task at a time | ✓ PASS | Task list defines execution order |
| Tests required | ✓ PASS | Unit, integration, contract tests exist |
| No architectural drift | ✓ PASS | No new modules beyond spec |
| Financial correctness | N/A | No financial operations |

## Project Structure

### Documentation (this feature)

```text
specs/001-auth-multirole/
├── plan.md              # This file
├── research.md          # Technical decisions
├── data-model.md        # Entity definitions
├── quickstart.md        # Integration scenarios
├── contracts/           # API contracts
├── checklists/          # Validation checklists
└── tasks.md             # Task breakdown
```

### Source Code (backend/)

```text
backend/
├── src/
│   ├── domain/          # User, Session, Profile, errors
│   ├── repository/      # Repository interfaces
│   ├── adapter/         # PostgresStore, EmailAdapter
│   ├── service/        # AuthService, SessionService, ProfileService
│   ├── handler/        # AuthHandler, SessionHandler, ProfileHandler
│   ├── middleware/     # AuthMiddleware, CSRFMiddleware
│   ├── config.go       # Environment config
│   ├── logging.go      # Logging setup
│   └── server.go       # Router wiring
├── migrations/          # SQL migrations
└── tests/
    ├── unit/           # Domain & service tests
    ├── integration/    # Repository tests
    └── contract/       # API endpoint tests
```

**Structure Decision**: Standard layered Go architecture with clear separation: domain → repository → service → handler → middleware.

## Complexity Tracking

> No violations requiring justification.

## New Security Implementation Details

### Account Lockout (FR-027)

| Field | Detail |
|-------|--------|
| Threshold | 5 failed login attempts |
| Lockout duration | 15 minutes |
| Reset | Automatic after 15 minutes |
| Storage | `failed_login_attempts`, `locked_until` columns on users table |

### Session Regeneration (FR-028)

| Event | Action |
|-------|--------|
| Login success | Generate new session ID, invalidate old |
| Session ID format | UUID v4 (same as new sessions) |

### CSRF Protection (FR-029)

| Layer | Implementation |
|-------|---------------|
| Cookie | SameSite=Strict; HttpOnly; Secure |
| Header | `X-CSRF-Token` required on POST/PATCH/DELETE |
| Generation | New CSRF token returned on login and profile switch |

### Token Expiry (Clarification)

| Token Type | Storage | Validation |
|------------|---------|------------|
| Email verification | `expires_at` column | Checked on lookup: `expires_at > NOW()` |
| Password reset | `expires_at` column | Checked on lookup: `expires_at > NOW()` |

## Phase 1 Updates Required

- [ ] Update research.md with new security decisions
- [ ] Update data-model.md: add `failed_login_attempts`, `locked_until` to User; update auth_tokens schema
- [ ] Update tasks.md with new security tasks
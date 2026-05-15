# Implementation Plan: Authentication and Multi-Role User Management

**Branch**: `[001-auth-multirole]` | **Date**: 2026-05-14 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/001-auth-multirole/spec.md`

## Summary

Build the Authentication and Multi-Role User Management module for ViralForge — a three-sided marketplace connecting Brands, Editors, and Influencers. Users register with email/password, verify email, log in/out, reset passwords, manage sessions, and maintain multiple role profiles (Brand, Editor, Influencer) under a single account. REST APIs drive all operations. Go/PostgreSQL backend with clean architecture, full test coverage.

## Technical Context

**Language/Version**: Go 1.21+

**Primary Dependencies**:
- stdlib (HTTP routing via chi, bcrypt for hashing, pgx for Postgres)
- No external auth frameworks — sessions stored in Postgres
- [NEEDS CLARIFICATION: email sending — external service adapter or SMTP?]

**Storage**: PostgreSQL (sessions, users, profiles, tokens)

**Testing**: Go stdlib `testing` package + `testify/assert`

**Target Platform**: Linux server (Docker container)

**Project Type**: REST/JSON web-service (modular monolith per constitution)

**Performance Goals**: 10,000 concurrent sessions, 99% success rate for auth APIs, sub-second RBAC checks

**Constraints**: <5s login, <2s profile switch, <3s session revocation

**Scale/Scope**: All three user roles (Brand, Editor, Influencer), multi-session per user, soft-delete for profiles

## Constitution Check

| Gate | Status | Notes |
|------|--------|-------|
| Modular monolith only | PASS | Single deployable backend |
| Go + PostgreSQL only | PASS | Per constitution tech stack |
| No new modules beyond spec | PASS | Only auth, session, profile entities |
| Domain isolation from frameworks | PASS | Domain packages have no framework deps |
| Tests required for all public interfaces | PASS | Unit/integration/contract tests specified |
| Secrets via env vars only | PASS | No hardcoded secrets |
| Input validation at all boundaries | PASS | API validation per spec |

No violations. Proceeding to Phase 0.

## Project Structure

### Documentation (this feature)

```text
specs/001-auth-multirole/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
└── contracts/           # Phase 1 output
    └── auth-api.yaml    # REST API contract
```

### Source Code (repository root)

```text
backend/
├── src/
│   ├── domain/              # Domain entities (User, Session, Profile, Token)
│   │   ├── user.go
│   │   ├── session.go
│   │   ├── profile.go
│   │   └── token.go
│   ├── repository/          # Repository interfaces
│   │   ├── user_repository.go
│   │   ├── session_repository.go
│   │   └── profile_repository.go
│   ├── service/             # Application services
│   │   ├── auth_service.go
│   │   ├── session_service.go
│   │   └── profile_service.go
│   ├── handler/             # HTTP handlers
│   │   ├── auth_handler.go
│   │   ├── session_handler.go
│   │   └── profile_handler.go
│   ├── middleware/          # Auth middleware
│   │   └── auth_middleware.go
│   ├── adapter/             # External adapters (email, pg store)
│   │   ├── email_adapter.go
│   │   └── postgres_store.go
│   └── server.go            # HTTP server wiring
├── migrations/             # SQL migration files
│   ├── 001_create_users.sql
│   ├── 002_create_sessions.sql
│   ├── 003_create_profiles.sql
│   └── 004_create_tokens.sql
└── tests/
    ├── unit/
    ├── integration/
    └── contract/
```

**Structure Decision**: Option 2 (Web application) — backend REST service with Go, Postgres persistence, clean architecture layers.

## Phase 0: Research

### Research Questions

1. **Session token strategy**: opaque bearer tokens (UUID) vs JWT. Sessions stored server-side in Postgres — simpler, revocable, fits RBAC context.
2. **Email verification/password reset token format**: single-use UUID tokens stored in Postgres with expiration.
3. **Password hashing**: bcrypt with cost factor 12 (industry standard per SC-011).
4. **Active profile in session**: stored in session record, switchable via API without re-auth.
5. **Email adapter**: interface-only (external service injected) — spec assumes external provider.

### Dispatched Agents

- golang-database: Postgres schema design for auth/session/profile tables
- golang-security: Password hashing best practices, session management
- golang-dependency-injection: Clean architecture wiring pattern

## Phase 1: Design Artifacts

*(Generated below — research.md, data-model.md, contracts/, quickstart.md)*

---

## research.md

### Session Token Strategy

**Decision**: UUID-based opaque bearer tokens stored server-side in PostgreSQL sessions table.

**Rationale**:
- Server-side sessions allow instant revocation (critical for security)
- JWTs are stateless and harder to revoke — inappropriate for multi-session, password-reset-invalidate-all requirement
- UUIDs are computationally unpredictable without cryptographic randomness requirement
- Trade-off: slightly more DB queries, but Postgres handles 10k sessions trivially

**Alternatives considered**:
- JWT stateless sessions: rejected — cannot revoke without distributed blacklist
- opaque random strings: same as UUID approach but less standard

### Token Storage Schema

**Decision**: Single `auth_tokens` table with type discriminator (verification, password_reset).

```sql
auth_tokens(uuid, user_id, token_type, expires_at, used_at, created_at)
```

**Rationale**: Unified table for all one-time tokens. Index on (user_id, token_type) for lookup. `used_at` set on consumption to prevent reuse.

### Active Profile in Session

**Decision**: `active_profile_id` column on `sessions` table, nullable. Switch via PATCH /sessions/active with profile_id.

**Rationale**: Profile switch without re-auth is a spec requirement (SC-006). Session record is the right home for active context.

### Password Hashing

**Decision**: bcrypt with cost factor 12.

**Rationale**: Industry standard, widely deployed, cost factor 12 meets SC-011 (minimum 12 rounds). Go's `golang.org/x/crypto/bcrypt` is the standard library.

### Data Model Summary

See `data-model.md` for full entity definitions.

### API Contract

See `contracts/auth-api.yaml` for OpenAPI 3.0 contract covering all endpoints.

---

## data-model.md

### Entity: User

| Field | Type | Constraints |
|-------|------|-------------|
| id | UUID | PK |
| email | VARCHAR(255) | UNIQUE, NOT NULL |
| password_hash | VARCHAR(255) | NOT NULL |
| verified | BOOLEAN | DEFAULT FALSE |
| created_at | TIMESTAMPTZ | NOT NULL |
| updated_at | TIMESTAMPTZ | NOT NULL |

### Entity: Session

| Field | Type | Constraints |
|-------|------|-------------|
| id | UUID | PK |
| user_id | UUID | FK → users.id, NOT NULL |
| active_profile_id | UUID | FK → profiles.id, NULLABLE |
| token_hash | VARCHAR(255) | UNIQUE, NOT NULL |
| user_agent | VARCHAR(512) | NULLABLE |
| ip_address | VARCHAR(45) | NULLABLE |
| expires_at | TIMESTAMPTZ | NOT NULL |
| created_at | TIMESTAMPTZ | NOT NULL |

### Entity: Profile

| Field | Type | Constraints |
|-------|------|-------------|
| id | UUID | PK |
| user_id | UUID | FK → users.id, NOT NULL |
| profile_type | ENUM(brand, editor, influencer) | NOT NULL |
| name | VARCHAR(255) | NOT NULL |
| details | JSONB | NOT NULL |
| created_at | TIMESTAMPTZ | NOT NULL |
| updated_at | TIMESTAMPTZ | NOT NULL |
| deleted_at | TIMESTAMPTZ | NULLABLE (soft delete) |

### Entity: AuthToken

| Field | Type | Constraints |
|-------|------|-------------|
| id | UUID | PK |
| user_id | UUID | FK → users.id, NOT NULL |
| token_type | ENUM(verification, password_reset) | NOT NULL |
| token_hash | VARCHAR(255) | UNIQUE, NOT NULL |
| expires_at | TIMESTAMPTZ | NOT NULL |
| used_at | TIMESTAMPTZ | NULLABLE |
| created_at | TIMESTAMPTZ | NOT NULL |

### Relationships

- User 1:N Session (user has many sessions)
- User 1:N Profile (user has many profiles)
- Profile 1:1 Session (session has optional active profile — foreign key in sessions table)
- User 1:N AuthToken (one-time tokens for email verification and password reset)

### Validation Rules

- Email: RFC 5322 compliant format, max 255 chars
- Password: 8+ chars, uppercase, lowercase, number, special char
- Profile name: 1-255 chars, non-empty
- Profile details JSONB: type-specific schema validation (Brand: company_name, size, industry; Editor: specializations[], portfolio_url; Influencer: platforms[], follower_counts)

---

## contracts/auth-api.yaml

```yaml
openapi: 3.0.3
info:
  title: ViralForge Auth API
  version: 1.0.0
paths:
  /auth/register:
    post:
      summary: Register new account
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [email, password]
              properties:
                email: { type: string, format: email }
                password: { type: string, minLength: 8 }
      responses:
        201:
          description: Account created, verification email sent
        400:
          description: Validation error
        409:
          description: Email already registered
  /auth/verify-email:
    post:
      summary: Verify email address
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [token]
              properties:
                token: { type: string }
      responses:
        200:
          description: Email verified
        400:
          description: Invalid or expired token
  /auth/login:
    post:
      summary: Log in
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [email, password]
              properties:
                email: { type: string }
                password: { type: string }
      responses:
        200:
          description: Login successful
          headers:
            Set-CAuthorization: { schema: { type: string } }
        401:
          description: Invalid credentials
        403:
          description: Email not verified
  /auth/logout:
    post:
      summary: Log out current session
      security:
        - BearerAuth: []
      responses:
        204:
          description: Logged out
        401:
          description: Not authenticated
  /auth/password-reset-request:
    post:
      summary: Request password reset
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [email]
              properties:
                email: { type: string }
      responses:
        200:
          description: Reset email sent if account exists
  /auth/password-reset-confirm:
    post:
      summary: Confirm password reset
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [token, new_password]
              properties:
                token: { type: string }
                new_password: { type: string }
      responses:
        200:
          description: Password reset complete
        400:
          description: Invalid or expired token
  /sessions:
    get:
      summary: List active sessions
      security:
        - BearerAuth: []
      responses:
        200:
          description: List of sessions
        401:
          description: Not authenticated
  /sessions/{session_id}:
    delete:
      summary: Revoke a session
      security:
        - BearerAuth: []
      responses:
        204:
          description: Session revoked
        401:
          description: Not authenticated
        404:
          description: Session not found
  /sessions/active:
    patch:
      summary: Switch active profile
      security:
        - BearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [profile_id]
              properties:
                profile_id: { type: string, format: uuid }
      responses:
        200:
          description: Active profile switched
        400:
          description: Invalid profile_id
        403:
          description: Profile does not belong to user
  /profiles:
    get:
      summary: List user's profiles
      security:
        - BearerAuth: []
      responses:
        200:
          description: List of profiles
    post:
      summary: Create a new profile
      security:
        - BearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [profile_type, name, details]
              properties:
                profile_type: { type: string, enum: [brand, editor, influencer] }
                name: { type: string }
                details: { type: object }
      responses:
        201:
          description: Profile created
  /profiles/{profile_id}:
    get:
      summary: Get profile details
      security:
        - BearerAuth: []
      responses:
        200:
          description: Profile details
        403:
          description: Profile does not belong to user
        404:
          description: Profile not found
    patch:
      summary: Update profile
      security:
        - BearerAuth: []
      responses:
        200:
          description: Profile updated
    delete:
      summary: Delete profile
      security:
        - BearerAuth: []
      responses:
        204:
          description: Profile deleted
components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
```

---

## quickstart.md

### Running Auth Module Locally

```bash
# 1. Start PostgreSQL
podman run -d --name vf-postgres -p 5432:5432 postgres:16

# 2. Run migrations
psql $DATABASE_URL -f migrations/001_create_users.sql
psql $DATABASE_URL -f migrations/002_create_sessions.sql
psql $DATABASE_URL -f migrations/003_create_profiles.sql
psql $DATABASE_URL -f migrations/004_create_tokens.sql

# 3. Start server
go run ./backend/src/server.go

# 4. Run tests
go test ./backend/src/... -v
```

### Key Environment Variables

| Variable | Description |
|----------|-------------|
| DATABASE_URL | PostgreSQL connection string |
| SMTP_HOST | Email service SMTP host |
| BASE_URL | Public URL for email links |
| TOKEN_SECRET | HMAC secret for token signing |
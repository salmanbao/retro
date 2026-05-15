# Research: Authentication and Multi-Role User Management

## Session Token Strategy

**Decision**: UUID-based opaque bearer tokens stored server-side in PostgreSQL sessions table.

**Rationale**:
- Server-side sessions allow instant revocation (critical for security)
- JWTs are stateless and harder to revoke — inappropriate for multi-session, password-reset-invalidate-all requirement
- UUIDs are computationally unpredictable without cryptographic randomness requirement
- Trade-off: slightly more DB queries, but Postgres handles 10k sessions trivially

**Alternatives considered**:
- JWT stateless sessions: rejected — cannot revoke without distributed blacklist
- opaque random strings: same as UUID approach but less standard

## Token Storage Schema

**Decision**: Single `auth_tokens` table with type discriminator (verification, password_reset).

```sql
auth_tokens(uuid, user_id, token_type, expires_at, used_at, created_at)
```

**Rationale**: Unified table for all one-time tokens. Index on (user_id, token_type) for lookup. `expires_at` stored in DB and validated on lookup (per clarification). `used_at` set on consumption to prevent reuse.

**Update (2026-05-15)**: `expires_at` column now validated in SQL query, not business logic alone — ensures consistency even if app crashes before business logic check.

## Active Profile in Session

**Decision**: `active_profile_id` column on `sessions` table, nullable. Switch via PATCH /sessions/active with profile_id.

**Rationale**: Profile switch without re-auth is a spec requirement (SC-006). Session record is the right home for active context.

## Password Hashing

**Decision**: bcrypt with cost factor 12.

**Rationale**: Industry standard, widely deployed, cost factor 12 meets OWASP recommendation and SC-011. Go's `golang.org/x/crypto/bcrypt` is the standard library.

## Account Lockout

**Decision**: 5 failed login attempts → 15-minute lockout, stored in `users` table.

```sql
users(failed_login_attempts INT, locked_until TIMESTAMP)
```

**Rationale**: Balances brute-force protection against denial-of-service. 15-minute window auto-resets without admin intervention. Prevents rapid-fire attacks while allowing legitimate users to retry after cooldown.

**Alternatives considered**:
- Infinite lockout: rejected — too easy to DOS any account
- No lockout: rejected — vulnerable to brute force
- Exponential backoff: more complex, 15-min is simpler

## Session Regeneration

**Decision**: Generate new UUID session ID on successful login, invalidate old.

**Rationale**: Prevents session fixation attacks (OWASP recommendation). Attacker cannot pre-set session ID since new ID is generated server-side after authentication.

## CSRF Protection

**Decision**: SameSite=Strict cookies + X-CSRF-Token header validation on state-changing requests.

**Implementation**:
- Cookie: `SameSite=Strict; HttpOnly; Secure`
- Header: `X-CSRF-Token` required on POST/PATCH/DELETE
- Token generated on login and profile switch

**Rationale**: Chi/Go has no built-in CSRF. SameSite=Strict alone blocks most CSRF, but header validation provides defense-in-depth for browser-compatible clients.

**Alternatives considered**:
- Double-submit cookie: more complex, SameSite+header is sufficient
- CSRF middleware library: adds dependency; custom is simpler

## Email Adapter Pattern

**Decision**: Interface-only (`EmailService` interface), concrete adapter injected via dependency injection.

**Rationale**: Keeps domain isolated from email provider specifics. Adapter pattern enables testing with mock email adapter without network calls.

## API Framework Choice

**Decision**: Go stdlib `net/http` with `github.com/go-chi/chi` for routing.

**Rationale**: Chi is lightweight, idiomatic Go. No ORM coupling. Stdlib testing with `net/http/httptest`.

## Email Verification Links

**Decision**: UUID tokens stored with `expires_at` column. Single-use: `used_at` set on consumption.

**Rationale**: Verification links expire based on DB `expires_at` value, not in-memory timestamp. 24-hour window per spec assumption.

**Update (2026-05-15)**: Expiry now enforced at DB level via `expires_at > NOW()` check — consistent even if app crashes.

## Password Reset Flow

**Decision**: UUID reset tokens stored with `expires_at` column. All existing sessions invalidated on successful reset.

**Rationale**: Single-use tokens with 1-hour expiry (per spec assumption). Session invalidation prevents attacker maintaining access after legitimate password change.

## Session Timeout

**Decision**: 24-hour inactivity timeout per spec assumption.

**Rationale**: Reasonable balance between security and UX. Sessions expire automatically if user is inactive.
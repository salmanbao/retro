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

**Rationale**: Unified table for all one-time tokens. Index on (user_id, token_type) for lookup. `used_at` set on consumption to prevent reuse.

## Active Profile in Session

**Decision**: `active_profile_id` column on `sessions` table, nullable. Switch via PATCH /sessions/active with profile_id.

**Rationale**: Profile switch without re-auth is a spec requirement (SC-006). Session record is the right home for active context.

## Password Hashing

**Decision**: bcrypt with cost factor 12.

**Rationale**: Industry standard, widely deployed, cost factor 12 meets SC-011 (minimum 12 rounds). Go's `golang.org/x/crypto/bcrypt` is the standard library.

## Email Adapter Pattern

**Decision**: Interface-only (`EmailService` interface), concrete adapter injected via dependency injection.

**Rationale**: Keeps domain isolated from email provider specifics. Adapter pattern enables testing with mock email adapter without network calls.

## API Framework Choice

**Decision**: Go stdlib `net/http` with `github.com/go-chi/chi` for routing.

**Rationale**: Chi is lightweight, idiomatic Go. No ORM coupling. Stdlib testing with `net/http/httptest`.

## Email Verification Links

**Decision**: UUID tokens, never expire in the database until used — separate expiration tracked in business logic (24h per spec assumption).

**Rationale**: Simpler than JWT for verification links. Tokens are single-use: `used_at` set on consumption.
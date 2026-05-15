# Data Model: Authentication and Multi-Role User Management

## Entity: User

| Field | Type | Constraints |
|-------|------|-------------|
| id | UUID | PK |
| email | VARCHAR(255) | UNIQUE, NOT NULL |
| password_hash | VARCHAR(255) | NOT NULL |
| verified | BOOLEAN | DEFAULT FALSE |
| failed_login_attempts | INT | DEFAULT 0 |
| locked_until | TIMESTAMPTZ | NULLABLE |
| created_at | TIMESTAMPTZ | NOT NULL |
| updated_at | TIMESTAMPTZ | NOT NULL |

**Lockout rules**:
- After 5 consecutive failed login attempts, `locked_until` is set to NOW() + 15 minutes
- If `locked_until > NOW()`, login attempts are rejected with "account locked" error
- On successful login, `failed_login_attempts` is reset to 0

## Entity: Session

| Field | Type | Constraints |
|-------|------|-------------|
| id | UUID | PK |
| user_id | UUID | FK → users.id, NOT NULL |
| active_profile_id | UUID | FK → profiles.id, NULLABLE |
| token_hash | VARCHAR(255) | UNIQUE, NOT NULL |
| csrf_token | VARCHAR(255) | NULLABLE |
| user_agent | VARCHAR(512) | NULLABLE |
| ip_address | VARCHAR(45) | NULLABLE |
| expires_at | TIMESTAMPTZ | NOT NULL |
| created_at | TIMESTAMPTZ | NOT NULL |

**Note**: `csrf_token` stored for validation against `X-CSRF-Token` header on state-changing requests.

## Entity: Profile

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

## Entity: AuthToken

| Field | Type | Constraints |
|-------|------|-------------|
| id | UUID | PK |
| user_id | UUID | FK → users.id, NOT NULL |
| token_type | ENUM(verification, password_reset) | NOT NULL |
| token_hash | VARCHAR(255) | UNIQUE, NOT NULL |
| expires_at | TIMESTAMPTZ | NOT NULL |
| used_at | TIMESTAMPTZ | NULLABLE |
| created_at | TIMESTAMPTZ | NOT NULL |

**Expiry enforcement**: Token is valid only if `expires_at > NOW() AND used_at IS NULL`. Expired or already-used tokens are rejected.

## Entity: LoginHistory

| Field | Type | Constraints |
|-------|------|-------------|
| id | UUID | PK |
| user_id | UUID | FK → users.id, NOT NULL |
| ip_address | VARCHAR(45) | NOT NULL |
| user_agent | VARCHAR(512) | NOT NULL |
| device_fingerprint | VARCHAR(255) | NOT NULL |
| city | VARCHAR(100) | NULLABLE |
| country | VARCHAR(100) | NULLABLE |
| latitude | DECIMAL(9,6) | NULLABLE |
| longitude | DECIMAL(9,6) | NULLABLE |
| logged_in_at | TIMESTAMPTZ | NOT NULL |

**Note**: `device_fingerprint` is a hash of browser characteristics (user agent, accept language, screen, timezone) — human-readable identifier, not cryptographic. Geolocation derived from IP address using MaxMind GeoLite2 or similar.

## Entity: TwoFactorSettings

| Field | Type | Constraints |
|-------|------|-------------|
| id | UUID | PK |
| user_id | UUID | FK → users.id, UNIQUE, NOT NULL |
| totp_secret_encrypted | VARCHAR(512) | NOT NULL |
| enabled | BOOLEAN | DEFAULT FALSE |
| backup_codes_hash | JSONB | NOT NULL |
| created_at | TIMESTAMPTZ | NOT NULL |
| updated_at | TIMESTAMPTZ | NOT NULL |

**Note**: `totp_secret_encrypted` is AES-256-GCM encrypted. `backup_codes_hash` is an array of bcrypt hashes, one per backup code. Each backup code can only be used once — marked used on consumption.

## Relationships

- User 1:N Session (user has many sessions)
- User 1:N Profile (user has many profiles)
- Profile 1:1 Session (session has optional active profile — foreign key in sessions table)
- User 1:N AuthToken (one-time tokens for email verification and password reset)
- User 1:N LoginHistory (user has many login history entries)
- User 1:1 TwoFactorSettings (user has optional 2FA settings)

## Validation Rules

- **Email**: RFC 5322 compliant format, max 255 chars
- **Password**: 8+ chars, uppercase, lowercase, number, special char
- **Profile name**: 1-255 chars, non-empty
- **Profile details JSONB**: type-specific schema validation:
  - Brand: company_name, size, industry
  - Editor: specializations[], portfolio_url
  - Influencer: platforms[], follower_counts

## Indexes

- `users(email)` — unique index for login lookup
- `sessions(user_id)` — list sessions per user
- `sessions(token_hash)` — session authentication
- `profiles(user_id)` — list profiles per user
- `auth_tokens(user_id, token_type)` — token lookup
- `auth_tokens(token_hash)` — token consumption
- `auth_tokens(expires_at)` — cleanup expired tokens
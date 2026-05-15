# Data Model: Authentication and Multi-Role User Management

## Entity: User

| Field | Type | Constraints |
|-------|------|-------------|
| id | UUID | PK |
| email | VARCHAR(255) | UNIQUE, NOT NULL |
| password_hash | VARCHAR(255) | NOT NULL |
| verified | BOOLEAN | DEFAULT FALSE |
| created_at | TIMESTAMPTZ | NOT NULL |
| updated_at | TIMESTAMPTZ | NOT NULL |

## Entity: Session

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

## Relationships

- User 1:N Session (user has many sessions)
- User 1:N Profile (user has many profiles)
- Profile 1:1 Session (session has optional active profile — foreign key in sessions table)
- User 1:N AuthToken (one-time tokens for email verification and password reset)

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
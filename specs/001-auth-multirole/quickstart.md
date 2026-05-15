# Quickstart: Authentication and Multi-Role User Management

## Running Locally

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

## Key Environment Variables

| Variable | Description |
|----------|-------------|
| DATABASE_URL | PostgreSQL connection string |
| SMTP_HOST | Email service SMTP host |
| BASE_URL | Public URL for email links |
| TOKEN_SECRET | HMAC secret for token signing |

## Key Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | /auth/register | Register new account |
| POST | /auth/verify-email | Verify email address |
| POST | /auth/login | Log in |
| POST | /auth/logout | Log out current session |
| POST | /auth/password-reset-request | Request password reset |
| POST | /auth/password-reset-confirm | Confirm password reset |
| GET | /sessions | List active sessions |
| DELETE | /sessions/{id} | Revoke a session |
| PATCH | /sessions/active | Switch active profile |
| GET | /profiles | List user's profiles |
| POST | /profiles | Create a new profile |
| GET | /profiles/{id} | Get profile details |
| PATCH | /profiles/{id} | Update profile |
| DELETE | /profiles/{id} | Delete profile |
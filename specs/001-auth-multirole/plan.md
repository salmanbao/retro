# Implementation Plan: Authentication and Multi-Role User Management

**Branch**: `main` | **Date**: 2026-05-15 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `spec.md` with User Stories 1-8

## Summary

Go backend with GORM ORM providing email/password authentication, multi-role profiles (Brand/Editor/Influencer), session management, password reset, login history tracking, and TOTP-based 2FA. PostgreSQL database with server-side sessions. Chi router for HTTP routing. bcrypt for password hashing, AES-256-GCM for TOTP secrets.

## Technical Context

| Field | Value |
|-------|-------|
| Language/Version | Go 1.23 |
| Primary Dependencies | chi/v5, gorm, pgx/v5, bcrypt, otp library, qrcode library |
| Storage | PostgreSQL with GORM |
| Testing | stretchr/testify |
| Target Platform | Linux server |
| Project Type | Web service (REST API) |
| Performance Goals | 10k concurrent sessions, 1000 req/s auth endpoints |
| Constraints | <200ms p95 latency, 24hr session timeout |
| Scale/Scope | 10k users, 6 entities |

## Constitution Check

| Requirement | Status |
|-------------|--------|
| GORM as sole DB access mechanism | PASS — using gorm.io/gorm |
| No raw SQL in application code | PASS — repository layer uses GORM methods |
| Domain isolation (no framework deps in domain) | PASS — domain package is pure |
| ORM queries in adapter layer | PASS |
| AutoMigrate for schema management | PASS |

## Project Structure

```text
backend/
├── src/
│   ├── domain/          # Pure domain entities (User, Session, Profile, AuthToken, LoginHistory, TwoFactorSettings)
│   ├── repository/      # Repository interfaces
│   ├── adapter/         # GORM implementations (PostgresStore)
│   ├── service/         # AuthService, ProfileService, SessionService, LoginHistoryService, TwoFactorService
│   ├── handler/         # HTTP handlers (AuthHandler, ProfileHandler, SessionHandler, TwoFactorHandler)
│   ├── middleware/      # AuthMiddleware, RequireCSRF
│   └── config.go         # Environment-driven configuration
├── migrations/          # SQL migration files (reference only, GORM AutoMigrate primary)
├── tests/
│   ├── unit/
│   ├── integration/
│   └── contract/
└── go.mod

## Complexity Tracking

> No violations. Constitution compliance achieved with GORM-only approach.

## Technology Decisions

| Decision | Choice | Rationale |
|----------|--------|----------|
| ORM | GORM (gorm.io/gorm) | Constitution mandate; good Go support |
| Session tokens | UUID (server-side) | Revocable, instant logout support |
| Password hashing | bcrypt cost 12 | OWASP recommended |
| 2FA algorithm | TOTP (RFC 6238) | Widest support, works offline |
| 2FA libraries | github.com/pquerna/otp | Standard Go TOTP library |
| QR code generation | github.com/skip2/go-qrcode | Simple, no external service |
| Geolocation | MaxMind GeoLite2 (optional) | Industry standard, free tier |
| Encryption | AES-256-GCM | For TOTP secrets at rest |

## New Entities Summary

| Entity | Purpose |
|--------|---------|
| LoginHistory | Audit trail of login events with geolocation and device fingerprint |
| TwoFactorSettings | TOTP secret (encrypted), enabled flag, backup codes (hashed) |

## Phased Implementation Order

1. **Phase 1**: Existing auth (US1-US3) + profiles (US4-US6) — already implemented
2. **Phase 2**: Login History (US7) — new entity, service, handler, endpoints
3. **Phase 3**: 2FA via Authenticator App (US8) — TOTP setup, verification, backup codes
# Implementation Plan: Profile Enrichment and Verification

**Branch**: `003-profile-enrichment` | **Date**: 2026-05-16 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/003-profile-enrichment/spec.md`

## Summary

Extend the existing Profile aggregate with marketplace, compliance, and payout metadata. ProfileEnrichment module adds public profile info (bio, social links), Editor portfolio items, Influencer audience data and follower verification, payout preferences, and KYC status. Sensitive payout details protected via encryption interface abstraction. PostgreSQL JSONB for flexible structures. Strict ownership and authorization enforced via existing auth middleware.

## Technical Context

**Language/Version**: Go 1.23+

**Primary Dependencies**:
- chi (HTTP routing)
- GORM (PostgreSQL ORM)
- pgcrypto (transparent encryption for payout details)

**Storage**: PostgreSQL with JSONB for flexible fields (social_links, demographics, evidence_urls)

**Testing**: go test with unit/integration/contract subpackages

**Target Platform**: Linux server (Modular Monolith)

**Project Type**: Web service (REST API)

**Performance Goals**: API responses < 200ms for profile details

**Constraints**: No external integrations, no over-engineering, strict ownership enforcement

**Scale/Scope**: Up to 50 portfolio items per Editor, 10KB max audience demographics JSON

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Rule | Status | Notes |
|------|--------|-------|
| One task at a time | ✓ PASS | Phases 3+ organized by user story |
| Tests required | ✓ PASS | Unit, integration, contract tests specified |
| No architectural drift | ✓ PASS | Extends existing Profile aggregate |
| Simplicity mandatory | ✓ PASS | JSONB for flexible fields, single DB encryption |
| No new modules | ✓ PASS | Adds to existing Profile domain |

## Project Structure

### Documentation (this feature)

```
specs/003-profile-enrichment/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (authz-api.md)
└── tasks.md             # Phase 2 output (/speckit-tasks)
```

### Source Code (repository root)

```
backend/
├── src/
│   ├── domain/          # Entities: ProfileEnrichment, PortfolioItem, AudienceData, FollowerVerification, PayoutPreferences, KYCStatus
│   ├── repository/      # Repository interfaces and implementations
│   ├── service/         # ProfileEnrichmentService, PortfolioService, AudienceService, VerificationService, PayoutService, KYCService
│   ├── handler/         # HTTP handlers for /profiles/{id}/details, /profiles/{id}/portfolio, /profiles/{id}/payout, /profiles/{id}/kyc
│   └── middleware/      # Ownership verification middleware
└── tests/
    ├── unit/            # Domain and service unit tests
    ├── integration/     # Repository integration tests
    └── contract/        # HTTP endpoint contract tests
```

**Structure Decision**: Extends existing backend/src structure per conventions. ProfileEnrichment, PortfolioItem, AudienceData, FollowerVerification, PayoutPreferences, KYCStatus added as domain entities. Separate services per concern (portfolio, audience, verification, payout, kyc). Handlers use chi router per existing patterns.

## Complexity Tracking

> None required - no constitution violations

---

# Phase 0: Research

## Unknowns to Resolve

| Unknown | Resolution |
|---------|------------|
| Encryption interface abstraction design | Use go制剂interface with nil implementation (DB handles encryption transparently per clarification) |
| JSONB validation approach in Go/GORM | Use json.RawMessage with custom validation in domain entity setters |
| Profile aggregate extension pattern | Embed or reference existing Profile entity; add ProfileEnrichment as separate aggregate |

## Research Findings

### Encryption Abstraction

**Decision**: No application-level encryption interface needed. Database handles transparent encryption for payout_details (per spec clarification: database-layer encryption). Application receives plaintext, database stores encrypted blob.

**Rationale**: Simplest architecture. App doesn't handle raw ciphertext. Leverage PostgreSQL TDE/pgcrypto.

**Alternatives considered**:
- App-layer encryption with Vault: Rejected - adds external dependency, latency
- Hybrid: Rejected - over-engineering for marketplace MVP

### JSONB Validation

**Decision**: Custom validation in domain entity setters using validator library.

**Rationale**: GORM handles JSONB column mapping; validation happens in domain layer before repository persist.

### Profile Aggregate Extension

**Decision**: ProfileEnrichment is a separate aggregate with ProfileID foreign key, not embedded in Profile.

**Rationale**: Clear ownership boundaries; profile-type-specific data (portfolio, audience) can be queried independently. Existing Profile aggregate remains unchanged.

---

# Phase 1: Design & Contracts

## Data Model

### Entities

**ProfileEnrichment** (root aggregate for public profile info)
- id: uuid (pk)
- profile_id: uuid (fk, unique)
- bio: text
- avatar_url: url
- cover_image_url: url
- website_url: url
- location: string
- languages: string[] (ISO 639-1)
- timezone: string (IANA)
- social_links: jsonb (embedded {tiktok, instagram, youtube, x_twitter, linkedin, website})
- created_at, updated_at: timestamps

**PortfolioItem** (Editor-only, separate aggregate)
- id: uuid (pk)
- profile_id: uuid (fk)
- title: string (max 200 chars)
- description: text
- thumbnail_url: url
- video_url: url (nullable)
- external_link: url (nullable)
- display_order: int
- created_at, updated_at, deleted_at: timestamps

**AudienceData** (Influencer-only, separate aggregate)
- profile_id: uuid (fk, unique)
- platform_handles: jsonb ({tiktok, instagram, youtube, x_twitter, linkedin})
- claimed_followers: jsonb ({platform: count})
- engagement_rate: decimal
- audience_demographics: jsonb (max 10KB)
- updated_at: timestamp

**FollowerVerification** (Influencer-only)
- profile_id: uuid (fk, unique)
- status: enum (unverified, pending, verified, rejected)
- evidence_urls: jsonb (text array)
- verification_notes: text (nullable)
- reviewed_at: timestamp (nullable)
- reviewed_by: string (nullable)
- created_at, updated_at: timestamps

**PayoutPreferences**
- profile_id: uuid (fk, unique)
- preferred_method: enum (bank_transfer, paypal, crypto)
- beneficiary_name: string
- country: string (ISO 3166-1 alpha-2)
- currency: string (ISO 4217)
- encrypted_details: text (DB-layer encryption, never returned in plaintext)
- payout_ready: boolean
- updated_at: timestamp

**KYCStatus**
- profile_id: uuid (fk, unique)
- status: enum (not_started, pending, approved, rejected, suspended)
- review_notes: text (nullable)
- reviewed_at: timestamp (nullable)
- reviewed_by: string (nullable)
- updated_at: timestamp

### Relationships

- Profile (existing) → ProfileEnrichment (1:1)
- Profile → PortfolioItem (1:many, soft-delete)
- Profile → AudienceData (1:1)
- Profile → FollowerVerification (1:1)
- Profile → PayoutPreferences (1:1)
- Profile → KYCStatus (1:1)

### Validation Rules

| Field | Validation |
|-------|------------|
| languages | ISO 639-1 two-letter codes |
| timezone | IANA timezone identifier |
| country | ISO 3166-1 alpha-2 |
| currency | ISO 4217 |
| audience_demographics | Max 10KB, valid JSON object |
| encrypted_details | Never returned in plaintext |
| display_order | Positive integer, no auto-renumber on delete |

## Contracts

### REST API Endpoints

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | /api/v1/profiles/{id}/details | Get profile enrichment | Owner only |
| PATCH | /api/v1/profiles/{id}/details | Update profile enrichment | Owner only |
| GET | /api/v1/profiles/{id}/portfolio | List portfolio items | Editor only |
| POST | /api/v1/profiles/{id}/portfolio | Create portfolio item | Editor only |
| PATCH | /api/v1/profiles/{id}/portfolio/{itemId} | Update portfolio item | Editor only |
| DELETE | /api/v1/profiles/{id}/portfolio/{itemId} | Soft-delete portfolio item | Editor only |
| GET | /api/v1/profiles/{id}/audience | Get audience data | Owner only |
| PUT | /api/v1/profiles/{id}/audience | Create/update audience data | Influencer only |
| GET | /api/v1/profiles/{id}/verification | Get verification status | Owner only |
| POST | /api/v1/profiles/{id}/verification | Submit verification evidence | Influencer only |
| GET | /api/v1/profiles/{id}/payout | Get payout preferences | Owner only |
| PUT | /api/v1/profiles/{id}/payout | Create/update payout preferences | Owner only |
| GET | /api/v1/profiles/{id}/kyc | Get KYC status | Owner only |
| PUT | /api/v1/admin/profiles/{id}/kyc | Admin update KYC status | Internal admin only |

### Request/Response Shapes

**GET /api/v1/profiles/{id}/details**
```json
{
  "profile_id": "uuid",
  "bio": "string",
  "avatar_url": "url",
  "cover_image_url": "url",
  "website_url": "url",
  "location": "string",
  "languages": ["en", "es"],
  "timezone": "America/New_York",
  "social_links": {
    "tiktok": "handle",
    "instagram": "handle",
    "youtube": "handle",
    "x_twitter": "handle",
    "linkedin": "url",
    "website": "url"
  },
  "updated_at": "timestamp"
}
```

**POST /api/v1/profiles/{id}/portfolio**
```json
{
  "title": "string",
  "description": "string",
  "thumbnail_url": "url",
  "video_url": "url",
  "external_link": "url",
  "display_order": 1
}
```

**Response (portfolio list)**
```json
{
  "items": [
    {
      "id": "uuid",
      "title": "string",
      "description": "string",
      "thumbnail_url": "url",
      "video_url": "url",
      "external_link": "url",
      "display_order": 1,
      "created_at": "timestamp"
    }
  ],
  "total": 5
}
```

### Internal Admin API (not public REST)

**PUT /api/v1/admin/profiles/{id}/kyc**
```json
{
  "status": "approved|rejected|suspended",
  "review_notes": "string"
}
```
Internal service-to-service auth (JWT or mTLS). Not exposed to public API.

## Quickstart Scenarios

1. **Profile enrichment flow**: Create profile → PATCH details → GET details
2. **Portfolio CRUD**: Assign Editor role → Create item → List items → Update → Delete
3. **Audience data flow**: Assign Influencer role → PUT audience → GET audience
4. **Verification submission**: Influencer submits evidence → Status becomes pending → Admin approves → Status becomes verified
5. **Payout preferences**: Set preferences → GET returns masked details only
6. **KYC status**: View status (not_started) → Admin updates to approved

---

## Phase 1 Output Summary

**Generated Artifacts**:
- `data-model.md` - Entity definitions and relationships
- `contracts/authz-api.md` - API endpoint contracts
- `quickstart.md` - Test scenarios for validation

**Agent Context Update**: `CLAUDE.md` plan reference updated to point to `specs/003-profile-enrichment/plan.md`
# Data Model: Profile Enrichment and Verification

**Feature**: `specs/003-profile-enrichment/`
**Generated**: 2026-05-16

## Entity Overview

| Entity | Aggregate Root | Profile-Type | Notes |
|--------|---------------|--------------|-------|
| ProfileEnrichment | Yes | All | Public profile info |
| PortfolioItem | Yes | Editor only | Work samples |
| AudienceData | Yes | Influencer only | Audience metrics |
| FollowerVerification | Yes | Influencer only | Verification evidence |
| PayoutPreferences | Yes | All | Payment settings |
| KYCStatus | Yes | All | Compliance status |

## ProfileEnrichment

Public profile information attached to a Profile.

```
Table: profile_enrichments
Columns:
  id              UUID PRIMARY KEY
  profile_id      UUID NOT NULL UNIQUE FK → profiles(id)
  bio             TEXT
  avatar_url      VARCHAR(500)
  cover_image_url VARCHAR(500)
  website_url     VARCHAR(500)
  location        VARCHAR(200)
  languages       JSONB (array of ISO 639-1 codes)
  timezone        VARCHAR(100) (IANA identifier)
  social_links    JSONB {tiktok, instagram, youtube, x_twitter, linkedin, website}
  created_at      TIMESTAMP
  updated_at      TIMESTAMP

Indexes:
  UNIQUE profile_id
```

**Validation**:
- languages: Valid ISO 639-1 two-letter codes
- timezone: Valid IANA timezone identifier
- social_links: All URL fields must be valid URLs or null

## PortfolioItem

Work samples for Editor profiles.

```
Table: portfolio_items
Columns:
  id              UUID PRIMARY KEY
  profile_id      UUID NOT NULL FK → profiles(id)
  title           VARCHAR(200) NOT NULL
  description     TEXT
  thumbnail_url   VARCHAR(500)
  video_url       VARCHAR(500)
  external_link   VARCHAR(500)
  display_order   INTEGER NOT NULL DEFAULT 0
  created_at      TIMESTAMP
  updated_at      TIMESTAMP
  deleted_at      TIMESTAMP NULL (soft delete)

Indexes:
  profile_id + display_order (for ordering)
  deleted_at (for filtering soft-deleted)

Constraint:
  MAX 50 items per profile_id (enforced in service layer)
```

**Validation**:
- display_order: No auto-renumber on delete (gaps preserved)
- display_order: Use created_at as tiebreaker for identical values

## AudienceData

Audience metrics for Influencer profiles.

```
Table: audience_data
Columns:
  profile_id          UUID PRIMARY KEY FK → profiles(id)
  platform_handles    JSONB {tiktok, instagram, youtube, x_twitter, linkedin}
  claimed_followers   JSONB {platform: integer}
  engagement_rate     DECIMAL(5,2)
  audience_demographics JSONB (max 10KB)
  updated_at          TIMESTAMP

Validation:
  engagement_rate: 0.00 to 100.00
  audience_demographics: Valid JSON object, max 10KB
```

## FollowerVerification

Follower count verification evidence.

```
Table: follower_verifications
Columns:
  profile_id          UUID PRIMARY KEY FK → profiles(id)
  status              ENUM (unverified, pending, verified, rejected) NOT NULL DEFAULT 'unverified'
  evidence_urls       JSONB (array of URLs)
  verification_notes  TEXT
  reviewed_at         TIMESTAMP
  reviewed_by         VARCHAR(200)
  created_at          TIMESTAMP
  updated_at          TIMESTAMP

Status Transitions:
  Any transition allowed (flexible with history)
  status_history table records all transitions

Relationship:
  One Profile → One FollowerVerification
  Rejected can return to pending via new submission
```

## PayoutPreferences

Payment destination settings.

```
Table: payout_preferences
Columns:
  profile_id          UUID PRIMARY KEY FK → profiles(id)
  preferred_method    ENUM (bank_transfer, paypal, crypto) NOT NULL
  beneficiary_name     VARCHAR(200) NOT NULL
  country             VARCHAR(2) NOT NULL (ISO 3166-1 alpha-2)
  currency            VARCHAR(3) NOT NULL (ISO 4217)
  encrypted_details   TEXT NOT NULL (DB-layer encryption)
  payout_ready        BOOLEAN NOT NULL DEFAULT false
  updated_at          TIMESTAMP

Security:
  encrypted_details: Never returned in plaintext
  DB handles transparent encryption (PostgreSQL TDE/pgcrypto)
  App receives plaintext only

Validation:
  country: ISO 3166-1 alpha-2
  currency: ISO 4217
```

## KYCStatus

Know Your Customer compliance status.

```
Table: kyc_statuses
Columns:
  profile_id      UUID PRIMARY KEY FK → profiles(id)
  status          ENUM (not_started, pending, approved, rejected, suspended) NOT NULL DEFAULT 'not_started'
  review_notes    TEXT
  reviewed_at     TIMESTAMP
  reviewed_by     VARCHAR(200)
  updated_at      TIMESTAMP

Update Restriction:
  Direct user modification NOT allowed
  Admin-only via internal service API
```

## Relationships Diagram

```
Profile (existing)
├── 1:1 ProfileEnrichment
├── 1:many PortfolioItem (soft-delete, Editor only)
├── 1:1 AudienceData (Influencer only)
├── 1:1 FollowerVerification (Influencer only)
├── 1:1 PayoutPreferences (All types)
└── 1:1 KYCStatus (All types)
```

## Soft Deletion Rules

- PortfolioItem: deleted_at set on delete, excluded from normal queries
- Soft-deleted items still returned by ID lookup (404 for direct access)
- No cascade delete (profile deletion handled by profile aggregate rules)

## Indexes Summary

| Table | Index | Purpose |
|-------|-------|---------|
| profile_enrichments | unique (profile_id) | Lookup by profile |
| portfolio_items | (profile_id, display_order) | Ordered listing |
| portfolio_items | (deleted_at) | Soft-delete filter |
| audience_data | unique (profile_id) | Lookup by profile |
| follower_verifications | unique (profile_id) | Lookup by profile |
| payout_preferences | unique (profile_id) | Lookup by profile |
| kyc_statuses | unique (profile_id) | Lookup by profile |
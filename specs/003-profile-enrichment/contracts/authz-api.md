# API Contracts: Profile Enrichment and Verification

**Feature**: `specs/003-profile-enrichment/`
**Generated**: 2026-05-16

## Overview

Public REST API endpoints for profile enrichment, portfolio management, audience data, follower verification, payout preferences, and KYC status. Admin endpoints for KYC/verification updates use internal service-to-service API.

---

## Public Endpoints

### Profile Details

#### GET /api/v1/profiles/{id}/details

Get profile enrichment data.

**Authorization**: Profile owner only (same profile ID as authenticated session)

**Response 200**:
```json
{
  "profile_id": "550e8400-e29b-41d4-a716-446655440000",
  "bio": "Award-winning video creator specializing in tech reviews",
  "avatar_url": "https://cdn.example.com/avatars/550e8400.jpg",
  "cover_image_url": "https://cdn.example.com/covers/550e8400.jpg",
  "website_url": "https://creator.example.com",
  "location": "San Francisco, CA",
  "languages": ["en", "es"],
  "timezone": "America/Los_Angeles",
  "social_links": {
    "tiktok": "creator_handle",
    "instagram": "@creator_handle",
    "youtube": "UC...",
    "x_twitter": "@creator_handle",
    "linkedin": "https://linkedin.com/in/creator",
    "website": "https://creator.example.com"
  },
  "updated_at": "2026-05-16T10:30:00Z"
}
```

**Response 403**: Not profile owner
```json
{"error": "Forbidden", "message": "You do not own this profile"}
```

**Response 404**: Profile not found
```json
{"error": "Not Found", "message": "Profile not found"}
```

---

#### PATCH /api/v1/profiles/{id}/details

Update profile enrichment data (partial update).

**Authorization**: Profile owner only

**Request Body** (all fields optional):
```json
{
  "bio": "Updated bio text",
  "avatar_url": "https://cdn.example.com/avatars/new.jpg",
  "location": "New York, NY",
  "languages": ["en"],
  "timezone": "America/New_York",
  "social_links": {
    "instagram": "@new_handle"
  }
}
```

**Response 200**: Updated profile details (same shape as GET)

**Response 400**: Validation error
```json
{"error": "Bad Request", "message": "Invalid timezone: not a valid IANA identifier"}
```

---

### Portfolio

#### GET /api/v1/profiles/{id}/portfolio

List portfolio items for Editor profile.

**Authorization**: Editor profile type only

**Query Parameters**:
- `limit`: int (default 20, max 50)
- `offset`: int (default 0)

**Response 200**:
```json
{
  "items": [
    {
      "id": "660e8400-e29b-41d4-a716-446655440001",
      "title": "Tech Review: Latest Smartphone",
      "description": "In-depth review of the latest flagship device",
      "thumbnail_url": "https://cdn.example.com/thumbs/660e8400.jpg",
      "video_url": "https://youtube.com/watch?v=...",
      "external_link": "https://creator.example.com/review",
      "display_order": 1,
      "created_at": "2026-05-10T08:00:00Z"
    }
  ],
  "total": 5,
  "limit": 20,
  "offset": 0
}
```

**Response 403**: Not an Editor profile
```json
{"error": "Forbidden", "message": "Portfolio items are only available for Editor profiles"}
```

---

#### POST /api/v1/profiles/{id}/portfolio

Create a new portfolio item.

**Authorization**: Editor profile type only

**Request Body**:
```json
{
  "title": "Campaign: Brand Collaboration",
  "description": "Work sample from recent brand partnership",
  "thumbnail_url": "https://cdn.example.com/thumbs/new.jpg",
  "video_url": "https://youtube.com/watch?v=...",
  "external_link": "https://creator.example.com/work",
  "display_order": 6
}
```

**Response 201**: Created item
```json
{
  "id": "770e8400-e29b-41d4-a716-446655440002",
  "title": "Campaign: Brand Collaboration",
  "description": "Work sample from recent brand partnership",
  "thumbnail_url": "https://cdn.example.com/thumbs/new.jpg",
  "video_url": "https://youtube.com/watch?v=...",
  "external_link": "https://creator.example.com/work",
  "display_order": 6,
  "created_at": "2026-05-16T11:00:00Z"
}
```

**Response 400**: Validation error or max items reached
```json
{"error": "Bad Request", "message": "Maximum portfolio items (50) reached"}
```

---

#### PATCH /api/v1/profiles/{id}/portfolio/{itemId}

Update a portfolio item.

**Authorization**: Editor profile owner only

**Request Body** (all fields optional):
```json
{
  "title": "Updated Title",
  "description": "Updated description",
  "display_order": 3
}
```

**Response 200**: Updated item

**Response 404**: Item not found (or soft-deleted)
```json
{"error": "Not Found", "message": "Portfolio item not found"}
```

---

#### DELETE /api/v1/profiles/{id}/portfolio/{itemId}

Soft-delete a portfolio item.

**Authorization**: Editor profile owner only

**Response 204**: No content (soft-delete successful)

**Behavior**:
- Sets deleted_at timestamp
- Item excluded from normal queries
- Returns 404 for direct ID lookup

---

### Audience Data

#### GET /api/v1/profiles/{id}/audience

Get audience data for Influencer profile.

**Authorization**: Profile owner only

**Response 200**:
```json
{
  "profile_id": "550e8400-e29b-41d4-a716-446655440000",
  "platform_handles": {
    "tiktok": "influencer_handle",
    "instagram": "@influencer",
    "youtube": "UC...",
    "x_twitter": "@influencer"
  },
  "claimed_followers": {
    "tiktok": 500000,
    "instagram": 250000,
    "youtube": 100000,
    "x_twitter": 50000
  },
  "engagement_rate": 4.5,
  "audience_demographics": {
    "age": {"18-24": 0.35, "25-34": 0.45, "35-44": 0.15, "45+": 0.05},
    "gender": {"male": 0.6, "female": 0.35, "other": 0.05},
    "regions": {"US": 0.5, "UK": 0.2, "CA": 0.1, "OTHER": 0.2}
  },
  "updated_at": "2026-05-16T09:00:00Z"
}
```

**Response 403**: Not an Influencer profile
```json
{"error": "Forbidden", "message": "Audience data is only available for Influencer profiles"}
```

---

#### PUT /api/v1/profiles/{id}/audience

Create or update audience data.

**Authorization**: Influencer profile type only

**Request Body**:
```json
{
  "platform_handles": {...},
  "claimed_followers": {...},
  "engagement_rate": 5.2,
  "audience_demographics": {...}
}
```

**Response 200**: Updated audience data

**Response 400**: Validation error (e.g., demographics > 10KB)
```json
{"error": "Bad Request", "message": "audience_demographics exceeds maximum size of 10KB"}
```

---

### Follower Verification

#### GET /api/v1/profiles/{id}/verification

Get verification status.

**Authorization**: Profile owner only

**Response 200**:
```json
{
  "profile_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "verified",
  "evidence_urls": ["https://evidence.example.com/screenshot1.png"],
  "verification_notes": "Follower counts verified via social platform APIs",
  "reviewed_at": "2026-05-15T14:00:00Z",
  "reviewed_by": "admin@viralforge.com"
}
```

---

#### POST /api/v1/profiles/{id}/verification

Submit verification evidence.

**Authorization**: Influencer profile type only

**Request Body**:
```json
{
  "evidence_urls": [
    "https://evidence.example.com/screenshot1.png",
    "https://evidence.example.com/screenshot2.png"
  ]
}
```

**Response 201**: Submission acknowledged
```json
{
  "profile_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "pending",
  "evidence_urls": [...],
  "submitted_at": "2026-05-16T11:30:00Z"
}
```

**Behavior**:
- Sets status to "pending"
- Clears previous reviewed_at and reviewed_by
- Any previous status (verified, rejected) returns to pending

---

### Payout Preferences

#### GET /api/v1/profiles/{id}/payout

Get payout preferences (masked sensitive data).

**Authorization**: Profile owner only

**Response 200**:
```json
{
  "profile_id": "550e8400-e29b-41d4-a716-446655440000",
  "preferred_method": "bank_transfer",
  "beneficiary_name": "John Doe",
  "country": "US",
  "currency": "USD",
  "payout_ready": true,
  "updated_at": "2026-05-16T10:00:00Z"
}
```

**Note**: `encrypted_details` NEVER returned in plaintext (masked/redacted)

---

#### PUT /api/v1/profiles/{id}/payout

Create or update payout preferences.

**Authorization**: Profile owner only

**Request Body**:
```json
{
  "preferred_method": "bank_transfer",
  "beneficiary_name": "John Doe",
  "country": "US",
  "currency": "USD",
  "encrypted_details": "...",
  "payout_ready": true
}
```

**Response 200**: Updated preferences

---

### KYC Status

#### GET /api/v1/profiles/{id}/kyc

Get KYC status.

**Authorization**: Profile owner only

**Response 200**:
```json
{
  "profile_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "approved",
  "review_notes": "Identity verified via manual review",
  "reviewed_at": "2026-05-14T16:00:00Z",
  "reviewed_by": "compliance@viralforge.com",
  "updated_at": "2026-05-14T16:00:00Z"
}
```

**Note**: Direct modification NOT allowed via public API

---

## Internal Admin Endpoints

### PUT /api/v1/admin/profiles/{id}/kyc

Admin update KYC status.

**Authorization**: Internal service-to-service (JWT or mTLS)

**Request Body**:
```json
{
  "status": "approved",
  "review_notes": "All documents verified successfully"
}
```

**Response 200**: Updated KYC status

---

### PUT /api/v1/admin/profiles/{id}/verification/review

Admin update verification status (approve/reject).

**Authorization**: Internal service-to-service (JWT or mTLS)

**Request Body**:
```json
{
  "status": "verified",
  "verification_notes": "Follower counts confirmed accurate"
}
```

**Response 200**: Updated verification status

---

## Error Response Format

All error responses follow this structure:

```json
{
  "error": "Error Type",
  "message": "Human-readable description",
  "details": {} // optional additional context
}
```

## Status Codes

| Code | Usage |
|------|-------|
| 200 | Successful GET/PATCH/PUT |
| 201 | Successful POST (created) |
| 204 | Successful DELETE (no content) |
| 400 | Validation error, bad request |
| 403 | Not authorized, forbidden |
| 404 | Resource not found |
| 500 | Internal server error |
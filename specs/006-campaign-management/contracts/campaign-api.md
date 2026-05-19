# Campaign Management - API Contracts

Base URL: `/api/v1`

## Authentication

All endpoints require `Authorization: Bearer <token>` header. Token is a session token obtained via `/api/v1/auth/login`.

## Common Types

### Campaign Status
```
draft | published | active | paused | completed | cancelled
```

### Error Response
```json
{
  "error": "error_code",
  "message": "Human readable message"
}
```

## Endpoints

---

### POST /campaigns

Create a new campaign.

**Request Body**:
```json
{
  "title": "Summer Product Launch",
  "summary": "Short description",
  "description": "Full campaign description...",
  "objective": "Brand awareness",
  "product_name": "New Skin Serum",
  "landing_url": "https://example.com/product",
  "total_budget": 50000.00,
  "currency": "USD",
  "target_clips": 100,
  "target_posts": 50,
  "cpv": 0.50,
  "min_payout": 100.00,
  "max_payout": 500.00,
  "submission_start": "2026-06-01T00:00:00Z",
  "submission_deadline": "2026-06-30T23:59:59Z",
  "distribution_start": "2026-07-01T00:00:00Z",
  "campaign_end": "2026-07-31T23:59:59Z",
  "allowed_countries": ["US", "GB", "CA"],
  "allowed_languages": ["en", "es"],
  "min_followers": 10000,
  "platforms": ["instagram", "tiktok"],
  "creator_categories": ["beauty", "lifestyle"],
  "min_duration_secs": 15,
  "max_duration_secs": 60,
  "aspect_ratio": "9:16",
  "talking_points": ["Mention product name", "Show demo"],
  "prohibited_claims": ["cure", "treat", "diagnose"],
  "hashtags": ["#SummerLaunch", "#Sponsored"],
  "cta_instructions": "Visit our landing page for more info"
}
```

**Response** `201 Created`:
```json
{
  "id": "uuid",
  "title": "Summer Product Launch",
  "slug": "summer-product-launch",
  "status": "draft",
  "created_at": "2026-05-18T12:00:00Z"
}
```

**Errors**:
- `400 validation_error` - Missing or invalid fields
- `403 forbidden` - Non-Brand profile type
- `409 slug_exists` - Slug already in use

---

### GET /campaigns

List campaigns for the authenticated user's Brand profiles.

**Query Parameters**:
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | 1 | Page number |
| `page_size` | int | 20 | Items per page (max 100) |
| `status` | string | - | Filter by status |

**Response** `200 OK`:
```json
{
  "campaigns": [
    {
      "id": "uuid",
      "title": "Summer Product Launch",
      "slug": "summer-product-launch",
      "status": "draft",
      "total_budget": 50000.00,
      "currency": "USD",
      "submission_deadline": "2026-06-30T23:59:59Z",
      "created_at": "2026-05-18T12:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 20,
  "total_pages": 1
}
```

---

### GET /campaigns/{id}

Get full campaign details.

**Response** `200 OK`:
```json
{
  "id": "uuid",
  "brand_profile_id": "uuid",
  "title": "Summer Product Launch",
  "slug": "summer-product-launch",
  "summary": "Short description",
  "description": "Full campaign description...",
  "objective": "Brand awareness",
  "product_name": "New Skin Serum",
  "landing_url": "https://example.com/product",
  "total_budget": 50000.00,
  "currency": "USD",
  "target_clips": 100,
  "target_posts": 50,
  "cpv": 0.50,
  "min_payout": 100.00,
  "max_payout": 500.00,
  "submission_start": "2026-06-01T00:00:00Z",
  "submission_deadline": "2026-06-30T23:59:59Z",
  "distribution_start": "2026-07-01T00:00:00Z",
  "campaign_end": "2026-07-31T23:59:59Z",
  "allowed_countries": ["US", "GB", "CA"],
  "allowed_languages": ["en", "es"],
  "min_followers": 10000,
  "platforms": ["instagram", "tiktok"],
  "creator_categories": ["beauty", "lifestyle"],
  "min_duration_secs": 15,
  "max_duration_secs": 60,
  "aspect_ratio": "9:16",
  "talking_points": ["Mention product name", "Show demo"],
  "prohibited_claims": ["cure", "treat", "diagnose"],
  "hashtags": ["#SummerLaunch", "#Sponsored"],
  "cta_instructions": "Visit our landing page for more info",
  "assets": [
    {
      "id": "uuid",
      "url": "https://cdn.example.com/assets/ref-1.jpg",
      "asset_type": "reference",
      "description": "Product reference image"
    }
  ],
  "status": "draft",
  "version": 1,
  "created_at": "2026-05-18T12:00:00Z",
  "updated_at": "2026-05-18T12:00:00Z"
}
```

**Errors**:
- `404 not_found` - Campaign not found or not owned by user

---

### PATCH /campaigns/{id}

Update campaign fields. Restrictions apply based on status.

**Request Body** (all fields optional):
```json
{
  "title": "Updated Title",
  "description": "Updated description",
  "total_budget": 60000.00,
  "min_payout": 150.00,
  "max_payout": 600.00,
  "submission_start": "2026-06-15T00:00:00Z"
}
```

**Response** `200 OK`:
```json
{
  "id": "uuid",
  "title": "Updated Title",
  "status": "draft",
  "version": 2,
  "updated_at": "2026-05-18T14:00:00Z"
}
```

**Errors**:
- `400 validation_error` - Invalid field values
- `400 restricted_edit` - Cannot edit this field in current status
- `404 not_found` - Campaign not found

**Edit Restrictions by Status**:

| Status | Allowed Edits | Rejected Edits |
|--------|--------------|----------------|
| draft | all fields | - |
| published | summary, description, talking_points, hashtags, cta_instructions | title, budget, timeline, eligibility, status |
| active | summary, description, talking_points, hashtags, cta_instructions | title, budget, timeline, eligibility, status |
| paused | all except title and timeline | title, submission_start, submission_deadline |
| completed | none | any field |
| cancelled | none | any field |

---

### POST /campaigns/{id}/publish

Publish a draft campaign. Validates readiness requirements.

**Response** `200 OK`:
```json
{
  "id": "uuid",
  "status": "published",
  "updated_at": "2026-05-18T15:00:00Z"
}
```

**Errors**:
- `400 budget_required` - total_budget must be > 0
- `400 readiness_failed` - KYC not approved / onboarding incomplete / payout not configured
- `400 validation_error` - Missing required campaign fields
- `404 not_found` - Campaign not found
- `409 invalid_transition` - Campaign not in draft status

---

### POST /campaigns/{id}/pause

Pause an active campaign.

**Response** `200 OK`:
```json
{
  "id": "uuid",
  "status": "paused",
  "updated_at": "2026-05-18T16:00:00Z"
}
```

**Errors**:
- `404 not_found` - Campaign not found
- `409 invalid_transition` - Campaign not in active status

---

### POST /campaigns/{id}/resume

Resume a paused campaign.

**Response** `200 OK`:
```json
{
  "id": "uuid",
  "status": "active",
  "updated_at": "2026-05-18T17:00:00Z"
}
```

**Errors**:
- `404 not_found` - Campaign not found
- `409 invalid_transition` - Campaign not in paused status

---

### POST /campaigns/{id}/complete

Complete an active campaign.

**Response** `200 OK`:
```json
{
  "id": "uuid",
  "status": "completed",
  "updated_at": "2026-05-18T18:00:00Z"
}
```

**Errors**:
- `404 not_found` - Campaign not found
- `409 invalid_transition` - Campaign not in active status, or campaign_end not reached

---

### POST /campaigns/{id}/cancel

Cancel a campaign (soft delete).

**Response** `200 OK`:
```json
{
  "id": "uuid",
  "status": "cancelled",
  "deleted_at": "2026-05-18T19:00:00Z"
}
```

**Errors**:
- `404 not_found` - Campaign not found
- `409 invalid_transition` - Campaign already cancelled

---

## Readiness Validation (for publish)

A campaign can be published only if ALL of the following are true:

1. Brand profile is fully onboarded and activated
2. KYC is approved for the Brand profile
3. Payout preferences are configured for the Brand profile
4. Required campaign fields are complete (title, description, objective, budget, timeline, eligibility, creative)
5. Total budget > 0

If any check fails, publish is rejected with `400 readiness_failed` and details of which requirement failed.
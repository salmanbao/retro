# Quickstart: Profile Enrichment and Verification

**Feature**: `specs/003-profile-enrichment/`
**Purpose**: Validation scenarios for implementation testing

## Prerequisites

- Running PostgreSQL database
- Existing Profile aggregate with profile_id
- Authentication middleware providing active_profile_id

## Validation Scenarios

### Scenario 1: Profile Enrichment Flow

**Purpose**: Verify profile details CRUD operations

```
1. Create a profile (or use existing test profile)
2. PATCH /api/v1/profiles/{id}/details with:
   {
     "bio": "Test creator",
     "avatar_url": "https://example.com/avatar.jpg",
     "location": "New York",
     "languages": ["en", "fr"],
     "timezone": "America/New_York"
   }
   Expected: 200 OK, data persisted
3. GET /api/v1/profiles/{id}/details
   Expected: 200 OK, all fields returned
4. Attempt to GET /api/v1/profiles/{different_id}/details as non-owner
   Expected: 403 Forbidden
```

### Scenario 2: Portfolio CRUD (Editor Profile)

**Purpose**: Verify portfolio operations for Editor profile type

```
1. Create or use an Editor profile
2. POST /api/v1/profiles/{editor_id}/portfolio with:
   {
     "title": "Campaign Video",
     "description": "Brand collaboration video",
     "thumbnail_url": "https://example.com/thumb.jpg",
     "display_order": 1
   }
   Expected: 201 Created
3. GET /api/v1/profiles/{editor_id}/portfolio
   Expected: 200 OK with items array containing created item
4. PATCH /api/v1/profiles/{editor_id}/portfolio/{item_id} with:
   {"display_order": 2}
   Expected: 200 OK, order updated
5. DELETE /api/v1/profiles/{editor_id}/portfolio/{item_id}
   Expected: 204 No Content
6. GET /api/v1/profiles/{editor_id}/portfolio/{deleted_item_id}
   Expected: 404 Not Found
```

### Scenario 3: Portfolio Rejection for Non-Editor

**Purpose**: Verify profile type enforcement

```
1. Create or use a Brand profile
2. POST /api/v1/profiles/{brand_id}/portfolio
   Expected: 403 Forbidden with message about Editor profiles only
3. GET /api/v1/profiles/{brand_id}/portfolio
   Expected: 403 Forbidden
```

### Scenario 4: Audience Data (Influencer Profile)

**Purpose**: Verify audience data operations

```
1. Create or use an Influencer profile
2. PUT /api/v1/profiles/{influencer_id}/audience with:
   {
     "platform_handles": {"tiktok": "handle", "instagram": "@handle"},
     "claimed_followers": {"tiktok": 100000, "instagram": 50000},
     "engagement_rate": 4.5,
     "audience_demographics": {"age": {"18-24": 0.5}}
   }
   Expected: 200 OK
3. GET /api/v1/profiles/{influencer_id}/audience
   Expected: 200 OK with all submitted data
4. Attempt from Brand profile
   Expected: 403 Forbidden
```

### Scenario 5: Verification Submission and Admin Review

**Purpose**: Verify verification workflow

```
1. Submit verification as Influencer:
   POST /api/v1/profiles/{influencer_id}/verification with:
   {"evidence_urls": ["https://example.com/proof.png"]}
   Expected: 201 Created, status is "pending"
2. Admin approves verification:
   PUT /api/v1/admin/profiles/{influencer_id}/verification/review with:
   {"status": "verified", "verification_notes": "Counts confirmed"}
   Expected: 200 OK, status is "verified"
3. Re-submit after rejection:
   POST /api/v1/profiles/{influencer_id}/verification (after admin rejects)
   Expected: 201 Created, status returns to "pending"
```

### Scenario 6: Payout Preferences (Masking)

**Purpose**: Verify sensitive data is never returned in plaintext

```
1. Set payout preferences:
   PUT /api/v1/profiles/{profile_id}/payout with encrypted details
   Expected: 200 OK
2. GET /api/v1/profiles/{profile_id}/payout
   Expected: 200 OK, encrypted_details field ABSENT or masked
   Verify: Response does not contain decrypted payout info
3. View as different user
   Expected: 403 Forbidden (payout preferences are owner-only)
```

### Scenario 7: KYC Status View and Admin Update

**Purpose**: Verify KYC workflow

```
1. Check initial KYC status:
   GET /api/v1/profiles/{profile_id}/kyc
   Expected: status is "not_started"
2. User attempts direct modification:
   PUT /api/v1/profiles/{profile_id}/kyc with {"status": "approved"}
   Expected: 403 Forbidden (admin-only)
3. Admin updates status:
   PUT /api/v1/admin/profiles/{profile_id}/kyc with:
   {"status": "approved", "review_notes": "Docs verified"}
   Expected: 200 OK
4. User views updated status:
   GET /api/v1/profiles/{profile_id}/kyc
   Expected: status is "approved", review_notes visible
```

### Scenario 8: Social Links Embedding

**Purpose**: Verify social_links stored as JSONB in profile details

```
1. Update profile with social links:
   PATCH /api/v1/profiles/{id}/details with:
   {
     "social_links": {
       "tiktok": "handle",
       "instagram": "@handle",
       "youtube": "channel",
       "x_twitter": "@handle"
     }
   }
   Expected: 200 OK
2. GET /api/v1/profiles/{id}/details
   Expected: social_links returned as embedded object
3. Partial update to social_links:
   PATCH /api/v1/profiles/{id}/details with:
   {"social_links": {"instagram": "@new_handle"}}
   Expected: Other social links preserved, only instagram updated
```

### Scenario 9: Portfolio Ordering with Gaps

**Purpose**: Verify display_order gaps preserved after delete

```
1. Create 3 portfolio items with display_order 1, 2, 3
2. Delete item with display_order = 2
   Expected: 204 No Content
3. List portfolio items
   Expected: Items with display_order 1 and 3 returned
   Verify: Gap at position 2 is preserved (no auto-renumber)
```

### Scenario 10: Max Portfolio Items Limit

**Purpose**: Verify 50 item limit enforced

```
1. Create 50 portfolio items (use bulk creation or loop)
2. Attempt to create 51st item:
   POST /api/v1/profiles/{editor_id}/portfolio
   Expected: 400 Bad Request with message about max limit
```

---

## Test Data Setup

```sql
-- Create test profile (if needed)
INSERT INTO profiles (id, user_id, profile_type, created_at, updated_at)
VALUES ('550e8400-e29b-41d4-a716-446655440000', 'user-123', 'editor', NOW(), NOW());

-- Create profile enrichment (if needed)
INSERT INTO profile_enrichments (id, profile_id, created_at, updated_at)
VALUES (gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440000', NOW(), NOW());
```

---

## Cleanup

```sql
-- Remove test data
DELETE FROM portfolio_items WHERE profile_id = '550e8400-e29b-41d4-a716-446655440000';
DELETE FROM audience_data WHERE profile_id = '550e8400-e29b-41d4-a716-446655440000';
DELETE FROM follower_verifications WHERE profile_id = '550e8400-e29b-41d4-a716-446655440000';
DELETE FROM payout_preferences WHERE profile_id = '550e8400-e29b-41d4-a716-446655440000';
DELETE FROM kyc_statuses WHERE profile_id = '550e8400-e29b-41d4-a716-446655440000';
DELETE FROM profile_enrichments WHERE profile_id = '550e8400-e29b-41d4-a716-446655440000';
```
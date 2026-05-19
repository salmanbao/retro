# Quickstart: Creative Brief and Asset Management

## Overview

The Creative Brief and Asset Management module allows Brand profiles to attach structured creative briefs and asset metadata to campaigns. Editors can access this information for published and active campaigns.

## Prerequisites

- Campaign Management module installed
- Authentication middleware configured
- Profile type middleware (Brand, Editor, Influencer) configured

## Key Concepts

### Creative Brief
- One brief per campaign (enforced at database level)
- JSONB fields for flexible arrays (key_messages, hashtags, talking_points)
- Edit restrictions based on campaign lifecycle state

### Asset Metadata
- Versioned by (campaign_id, original_filename) tuple
- Soft deletion preserves audit trail
- Editors have read-only access to published/active campaign assets

## Usage Examples

### 1. Create a Creative Brief (Brand)

```bash
# Create brief for a draft campaign
curl -X PUT "http://localhost:8080/api/v1/campaigns/{campaignId}/brief" \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{
    "key_messages": ["Our product solves X problem", "Made for Gen-Z"],
    "product_benefits": ["Fast", "Easy to use", "Affordable"],
    "mandatory_talking_points": ["Brand logo visible", "Mention discount code"],
    "required_hashtags": ["#brand", "#ad", "#campaign"],
    "call_to_action_text": "Click the link in bio!",
    "tone_and_style_guidelines": "Fun, energetic, Gen-Z vibes"
  }'
```

### 2. Get Creative Brief (Brand or Editor)

```bash
# Get brief - works for brand owner or editor on published/active campaign
curl "http://localhost:8080/api/v1/campaigns/{campaignId}/brief" \
  -H "Authorization: Bearer {token}"
```

### 3. Register Asset Metadata (Brand)

```bash
# Register new asset
curl -X POST "http://localhost:8080/api/v1/campaigns/{campaignId}/assets" \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{
    "category": "raw_footage",
    "original_filename": "product_demo_001.mp4",
    "display_name": "Product Demo Footage v1",
    "mime_type": "video/mp4",
    "file_size_bytes": 52428800,
    "storage_key": "campaigns/{campaignId}/raw/product_demo_001.mp4",
    "checksum": "sha256:abc123..."
  }'
```

### 4. Upload New Version (Same Filename)

```bash
# Registering same filename auto-increments version
curl -X POST "http://localhost:8080/api/v1/campaigns/{campaignId}/assets" \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{
    "category": "raw_footage",
    "original_filename": "product_demo_001.mp4",
    "display_name": "Product Demo Footage v2",
    "mime_type": "video/mp4",
    "file_size_bytes": 55120960,
    "storage_key": "campaigns/{campaignId}/raw/product_demo_001_v2.mp4",
    "checksum": "sha256:def456..."
  }'
# Response version will be 2
```

### 5. List Assets (Brand or Editor)

```bash
# Paginated list
curl "http://localhost:8080/api/v1/campaigns/{campaignId}/assets?page=1&page_size=20" \
  -H "Authorization: Bearer {token}"
```

### 6. Update Asset Status (Brand)

```bash
# Mark asset as ready
curl -X PATCH "http://localhost:8080/api/v1/assets/{assetId}" \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{
    "processing_status": "ready"
  }'
```

### 7. Soft-Delete Asset (Brand)

```bash
# Soft delete
curl -X DELETE "http://localhost:8080/api/v1/assets/{assetId}" \
  -H "Authorization: Bearer {token}"
# Returns 204 No Content
```

## Access Control Matrix

| Action | Brand Owner | Editor | Influencer |
|--------|-------------|--------|------------|
| Create brief | ✓ (own campaign) | ✗ | ✗ |
| Update brief | ✓ (draft/paused only) | ✗ | ✗ |
| View brief | ✓ (own campaign) | ✓ (pub/active) | ✗ |
| Register asset | ✓ (own campaign) | ✗ | ✗ |
| Update asset | ✓ (own campaign) | ✗ | ✗ |
| Delete asset | ✓ (own campaign) | ✗ | ✗ |
| View assets | ✓ (own campaign) | ✓ (pub/active) | ✗ |

## Campaign State Edit Restrictions

| Campaign State | Brief Fields Editable |
|----------------|----------------------|
| draft | All fields |
| paused | All fields |
| published | tone_and_style_guidelines, target_audience_description, example_video_links only |
| active | tone_and_style_guidelines, target_audience_description, example_video_links only |
| completed | None |
| cancelled | None |

## Asset Categories

- `raw_footage` - Unedited video footage
- `product_images` - Product photography
- `logos` - Brand and campaign logos
- `brand_guidelines` - Style guides
- `scripts` - Video scripts
- `example_videos` - Reference videos
- `music_references` - Music tracks
- `legal_documents` - Contracts, releases
- `other` - Miscellaneous

## Processing Statuses

- `pending` - Awaiting upload
- `uploaded` - Upload complete
- `processing` - Being processed
- `ready` - Ready for use
- `failed` - Processing failed
- `archived` - Archived (hidden by default)
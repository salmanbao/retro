# API Contracts: Creative Brief and Asset Management

## Creative Brief Endpoints

### GET /api/v1/campaigns/{campaignId}/brief

**Description**: Get the creative brief for a campaign

**Access**: Brand owner (own campaign), Editor (published/active campaigns)

**Request**:
- Path: `campaignId` (UUID, required)
- Headers: `Authorization: Bearer <token>`

**Response 200**:
```json
{
  "id": "uuid",
  "campaign_id": "uuid",
  "key_messages": ["message1", "message2"],
  "product_benefits": ["benefit1", "benefit2"],
  "mandatory_talking_points": ["point1", "point2"],
  "prohibited_claims": ["claim1"],
  "required_hashtags": ["#brand", "#campaign"],
  "call_to_action_text": "Click the link in bio!",
  "tone_and_style_guidelines": "Fun, energetic, Gen-Z",
  "target_audience_description": "18-24 year old Gen-Z",
  "competitor_references": ["@competitor"],
  "example_video_links": ["https://youtube.com/..."],
  "created_at": "2026-05-19T10:00:00Z",
  "updated_at": "2026-05-19T10:00:00Z"
}
```

**Response 404**:
```json
{
  "error": "creative_brief_not_found",
  "message": "No creative brief found for this campaign"
}
```

**Response 403**:
```json
{
  "error": "forbidden",
  "message": "You do not have access to this campaign's brief"
}
```

---

### PUT /api/v1/campaigns/{campaignId}/brief

**Description**: Create or replace a creative brief for a campaign

**Access**: Brand owner only

**Request**:
- Path: `campaignId` (UUID, required)
- Headers: `Authorization: Bearer <token>`
- Body:
```json
{
  "key_messages": ["message1", "message2"],
  "product_benefits": ["benefit1", "benefit2"],
  "mandatory_talking_points": ["point1", "point2"],
  "prohibited_claims": ["claim1"],
  "required_hashtags": ["#brand", "#campaign"],
  "call_to_action_text": "Click the link in bio!",
  "tone_and_style_guidelines": "Fun, energetic, Gen-Z",
  "target_audience_description": "18-24 year old Gen-Z",
  "competitor_references": ["@competitor"],
  "example_video_links": ["https://youtube.com/..."]
}
```

**Response 200** (update):
```json
{
  "id": "uuid",
  "campaign_id": "uuid",
  "key_messages": ["message1", "message2"],
  "...": "..."
  "updated_at": "2026-05-19T12:00:00Z"
}
```

**Response 201** (create):
```json
{
  "id": "uuid",
  "campaign_id": "uuid",
  "key_messages": ["message1", "message2"],
  "...": "..."
  "created_at": "2026-05-19T12:00:00Z"
  "updated_at": "2026-05-19T12:00:00Z"
}
```

**Response 400** (validation error):
```json
{
  "error": "validation_failed",
  "message": "Required field missing",
  "fields": ["key_messages"]
}
```

**Response 403**:
```json
{
  "error": "forbidden",
  "message": "Only the campaign owner can modify the brief"
}
```

**Response 409** (campaign not editable):
```json
{
  "error": "campaign_not_editable",
  "message": "Cannot modify brief for published or active campaign"
}
```

---

## Asset Metadata Endpoints

### POST /api/v1/campaigns/{campaignId}/assets

**Description**: Register a new asset metadata entry

**Access**: Brand owner only

**Request**:
- Path: `campaignId` (UUID, required)
- Headers: `Authorization: Bearer <token>`
- Body:
```json
{
  "category": "raw_footage",
  "original_filename": "footage_001.mp4",
  "display_name": "Main Product Footage",
  "mime_type": "video/mp4",
  "file_size_bytes": 104857600,
  "storage_key": "campaigns/abc123/footage_001.mp4",
  "checksum": "a1b2c3d4e5f6..."
}
```

**Response 201**:
```json
{
  "id": "uuid",
  "campaign_id": "uuid",
  "category": "raw_footage",
  "original_filename": "footage_001.mp4",
  "display_name": "Main Product Footage",
  "mime_type": "video/mp4",
  "file_size_bytes": 104857600,
  "storage_key": "campaigns/abc123/footage_001.mp4",
  "checksum": "a1b2c3d4e5f6...",
  "version": 1,
  "processing_status": "pending",
  "virus_scan_status": "not_scanned",
  "uploaded_by_profile_id": "uuid",
  "created_at": "2026-05-19T12:00:00Z",
  "updated_at": "2026-05-19T12:00:00Z"
}
```

**Response 400**:
```json
{
  "error": "validation_failed",
  "message": "Invalid category",
  "fields": ["category"]
}
```

---

### GET /api/v1/campaigns/{campaignId}/assets

**Description**: List all assets for a campaign (paginated)

**Access**: Brand owner (own campaign), Editor (published/active campaigns)

**Request**:
- Path: `campaignId` (UUID, required)
- Query: `page` (int, default 1), `page_size` (int, default 20, max 100)
- Headers: `Authorization: Bearer <token>`

**Response 200**:
```json
{
  "data": [
    {
      "id": "uuid",
      "campaign_id": "uuid",
      "category": "raw_footage",
      "original_filename": "footage_001.mp4",
      "display_name": "Main Product Footage",
      "version": 2,
      "processing_status": "ready",
      "created_at": "2026-05-19T12:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 5,
    "total_pages": 1
  }
}
```

**Response 403**:
```json
{
  "error": "forbidden",
  "message": "You do not have access to this campaign's assets"
}
```

---

### GET /api/v1/assets/{id}

**Description**: Get a single asset's metadata

**Access**: Brand owner (own campaign), Editor (published/active campaigns)

**Request**:
- Path: `assetId` (UUID, required)
- Headers: `Authorization: Bearer <token>`

**Response 200**:
```json
{
  "id": "uuid",
  "campaign_id": "uuid",
  "category": "raw_footage",
  "original_filename": "footage_001.mp4",
  "display_name": "Main Product Footage",
  "mime_type": "video/mp4",
  "file_size_bytes": 104857600,
  "storage_key": "campaigns/abc123/footage_001.mp4",
  "checksum": "a1b2c3d4e5f6...",
  "version": 2,
  "processing_status": "ready",
  "virus_scan_status": "clean",
  "uploaded_by_profile_id": "uuid",
  "created_at": "2026-05-19T12:00:00Z",
  "updated_at": "2026-05-19T12:00:00Z"
}
```

**Response 404**:
```json
{
  "error": "asset_not_found",
  "message": "Asset not found or has been deleted"
}
```

---

### PATCH /api/v1/assets/{id}

**Description**: Update asset metadata (not campaign_id, version, checksum)

**Access**: Brand owner only

**Request**:
- Path: `assetId` (UUID, required)
- Headers: `Authorization: Bearer <token>`
- Body (partial update):
```json
{
  "display_name": "Updated Display Name",
  "processing_status": "ready"
}
```

**Response 200**:
```json
{
  "id": "uuid",
  "display_name": "Updated Display Name",
  "processing_status": "ready",
  "...": "..."
  "updated_at": "2026-05-19T14:00:00Z"
}
```

**Response 400**:
```json
{
  "error": "validation_failed",
  "message": "Cannot modify immutable field",
  "fields": ["version"]
}
```

**Response 403**:
```json
{
  "error": "forbidden",
  "message": "Only the campaign owner can modify assets"
}
```

---

### DELETE /api/v1/assets/{id}

**Description**: Soft-delete an asset

**Access**: Brand owner only

**Request**:
- Path: `assetId` (UUID, required)
- Headers: `Authorization: Bearer <token>`

**Response 204**: No content

**Response 403**:
```json
{
  "error": "forbidden",
  "message": "Only the campaign owner can delete assets"
}
```

---

## Common Error Responses

| HTTP Code | Error Code | Description |
|-----------|------------|-------------|
| 400 | validation_failed | Invalid input data |
| 401 | unauthorized | Missing or invalid token |
| 403 | forbidden | Insufficient permissions |
| 404 | not_found | Resource does not exist |
| 409 | conflict | State conflict (e.g., campaign not editable) |
| 500 | internal_error | Server error |
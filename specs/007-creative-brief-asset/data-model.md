# Data Model: Creative Brief and Asset Management

## Entities

### CreativeBrief

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PK, default gen_random_uuid() | Unique identifier |
| campaign_id | UUID | FK, unique | Reference to campaign |
| key_messages | JSONB | not null | Array of strings |
| product_benefits | JSONB | not null | Array of strings |
| mandatory_talking_points | JSONB | not null | Array of strings |
| prohibited_claims | JSONB | default '[]' | Array of strings |
| required_hashtags | JSONB | not null | Array of strings |
| call_to_action_text | VARCHAR(200) | not null | CTA text |
| tone_and_style_guidelines | VARCHAR(1000) | | Style guidance |
| target_audience_description | VARCHAR(500) | | Audience description |
| competitor_references | JSONB | default '[]' | Array of strings |
| example_video_links | JSONB | default '[]' | Array of URLs |
| created_at | TIMESTAMP | not null, default now() | Creation timestamp |
| updated_at | TIMESTAMP | not null, default now() | Last update timestamp |

**Constraints**:
- UNIQUE(campaign_id) - one brief per campaign

### AssetMetadata

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | UUID | PK, default gen_random_uuid() | Unique identifier |
| campaign_id | UUID | FK, index | Reference to campaign |
| category | VARCHAR(50) | not null, index | Asset category enum |
| original_filename | VARCHAR(255) | not null | Original file name |
| display_name | VARCHAR(255) | not null | Display name |
| mime_type | VARCHAR(100) | not null | MIME type |
| file_size_bytes | BIGINT | not null, >= 0 | File size |
| storage_key | VARCHAR(500) | not null | Storage location key |
| checksum | VARCHAR(64) | not null | SHA-256 hex hash |
| version | INT | not null, default 1 | Version number |
| processing_status | VARCHAR(20) | not null | Processing status enum |
| virus_scan_status | VARCHAR(20) | not null | Virus scan status enum |
| uploaded_by_profile_id | UUID | FK | Uploader profile reference |
| created_at | TIMESTAMP | not null, default now() | Creation timestamp |
| updated_at | TIMESTAMP | not null, default now() | Last update timestamp |
| deleted_at | TIMESTAMP | nullable, index | Soft deletion timestamp |

**Constraints**:
- INDEX(campaign_id, original_filename) - for version lookup
- INDEX(category) - for category filtering
- INDEX(processing_status) - for status filtering
- INDEX(deleted_at) - for soft-delete queries

## Enums

### AssetCategory

| Value | Description |
|-------|-------------|
| raw_footage | Unedited video footage |
| product_images | Product photography and renders |
| logos | Brand and campaign logos |
| brand_guidelines | Brand style guides and fonts |
| scripts | Video scripts and storyboards |
| example_videos | Reference videos for style/quality |
| music_references | Music tracks and sound references |
| legal_documents | Contracts, releases, terms |
| other | Miscellaneous assets |

### ProcessingStatus

| Value | Description |
|-------|-------------|
| pending | Awaiting upload initiation |
| uploaded | Upload complete, awaiting processing |
| processing | Being processed or transcoded |
| ready | Ready for use |
| failed | Processing failed |
| archived | Archived and hidden by default |

### VirusScanStatus

| Value | Description |
|-------|-------------|
| not_scanned | Not yet scanned |
| scanning | Currently scanning |
| clean | No threats detected |
| infected | Threat detected |

## Relationships

```
Campaign (existing)
    │
    ├── 1:1 CreativeBrief
    │       └── campaign_id = FK (unique)
    │
    └── 1:N AssetMetadata
            └── campaign_id = FK (indexed)

AssetMetadata
    └── uploaded_by_profile_id = FK (Profile)
```

## Versioning Strategy

Asset versions are stored as separate rows with incrementing version numbers:

```sql
-- Query to find latest version of an asset
SELECT * FROM asset_metadata
WHERE campaign_id = ? AND original_filename = ?
ORDER BY version DESC
LIMIT 1;
```

When registering a new version:
1. Find max version for (campaign_id, original_filename)
2. Insert new row with version = max + 1
3. Previous versions remain accessible

## Soft Deletion Pattern

```go
// Default scope excludes deleted records
func NotDeleted(db *gorm.DB) *gorm.DB {
    return db.Where("deleted_at IS NULL")
}

// Query: Get all non-deleted assets for campaign
db.Where("campaign_id = ? AND deleted_at IS NULL").Find(&assets)
```

## Indexes

| Index | Columns | Type | Purpose |
|-------|---------|------|---------|
| idx_creative_brief_campaign | campaign_id | B-tree | Unique lookup |
| idx_asset_campaign | campaign_id | B-tree | List assets by campaign |
| idx_asset_campaign_filename | campaign_id, original_filename | B-tree | Version lookup |
| idx_asset_category | category | B-tree | Category filtering |
| idx_asset_status | processing_status | B-tree | Status filtering |
| idx_asset_deleted | deleted_at | B-tree | Soft-delete filtering |
# Feature Specification: Creative Brief and Asset Management

**Created**: 2026-05-19
**Feature ID**: 007-creative-brief-asset
**Status**: Draft

---

## 1. Overview and Context

### Purpose

Build a module that allows Brand profiles to attach structured creative briefs and asset metadata to campaigns so Editors can access all instructions and source materials required to produce short-form videos.

### Background

The Campaign Management module is complete and integration-tested. Brand profiles can create, publish, and manage campaigns through their lifecycle. This module extends campaign capabilities by allowing Brand profiles to attach structured creative briefs and track asset metadata that Editors need to produce compliant short-form video content.

### Scope

This feature adds:
- Creative brief management (create, read, update) per campaign
- Asset metadata registration and versioning
- Editor access to briefs and asset metadata for published/active campaigns

This feature explicitly excludes:
- Binary file upload handling
- S3/R2 storage integration
- Signed URLs
- CDN delivery
- Video transcoding
- Malware scanning
- Automatic content analysis

---

## 2. Functional Requirements

### 2.1 Creative Brief

**FR-001**: Each campaign may have at most one active creative brief.

**FR-002**: A creative brief contains:
- Key messages (array of strings, required)
- Product benefits (array of strings, required)
- Mandatory talking points (array of strings, required)
- Prohibited claims (array of strings, optional)
- Required hashtags (array of strings, required)
- Call-to-action text (string, required, max 200 characters)
- Tone and style guidelines (string, optional, max 1000 characters)
- Target audience description (string, optional, max 500 characters)
- Competitor references (array of strings, optional)
- Example video links (array of URLs, optional)

**FR-003**: Brief updates are allowed while the campaign is in editable states (draft, paused).

**FR-004**: Published and active campaigns restrict specific brief fields from modification:
- Active campaigns: only `tone_and_style_guidelines`, `target_audience_description`, `example_video_links` are editable
- Published campaigns: same restrictions as active campaigns

### 2.2 Asset Metadata

**FR-005**: Asset metadata records contain:
- Campaign ID (required, foreign key)
- Asset category (required): `raw_footage`, `product_images`, `logos`, `brand_guidelines`, `scripts`, `example_videos`, `music_references`, `legal_documents`, `other`
- Original filename (required, max 255 characters)
- Display name (required, max 255 characters)
- MIME type (required, validated against allowed types)
- File size in bytes (required, positive integer)
- Storage key (required, max 500 characters)
- Checksum (required, SHA-256 hex string)
- Version number (required, positive integer, starts at 1)
- Processing status (required): `pending`, `uploaded`, `processing`, `ready`, `failed`, `archived`
- Virus scan status (required): `not_scanned`, `scanning`, `clean`, `infected`
- Uploaded by profile ID (required, foreign key)

**FR-006**: New uploads of the same logical asset (same campaign_id + original filename) increment the version number. Previous versions remain accessible for audit purposes.

**FR-007**: Assets are soft-deleted (deleted_at timestamp set) and excluded from default listings.

### 2.3 Asset Categories

**FR-008**: Valid asset categories:
| Category | Description |
|----------|-------------|
| raw_footage | Unedited video footage |
| product_images | Product photography and renders |
| logos | Brand and campaign logos |
| brand_guidelines | Brand style guides and fonts |
| scripts | Video scripts and storyboards |
| example_videos | Reference videos for style/quality |
| music_references | Music tracks and sound references |
| legal_documents | Contracts, releases, terms |
| other | Miscellaneous assets |

### 2.4 Processing Status

**FR-009**: Asset processing status values:
| Status | Description |
|--------|-------------|
| pending | Awaiting upload initiation |
| uploaded | Upload complete, awaiting processing |
| processing | Being processed or transcoded |
| ready | Ready for use |
| failed | Processing failed |
| archived | Archived and hidden by default |

### 2.5 Access Control

**FR-010**: Brand profiles that own the campaign may:
- Create a creative brief
- Update a creative brief (subject to campaign state restrictions)
- Register new asset metadata
- Update asset metadata (only if not in `archived` state)
- Soft-delete asset metadata

**FR-011**: Authenticated Editors may:
- List briefs for campaigns where campaign status is `published` or `active`
- View brief details for eligible campaigns
- List assets for campaigns where campaign status is `published` or `active`
- View asset metadata (not raw file content) for eligible campaigns

**FR-012**: Influencers do not access briefs or asset metadata at this stage.

**FR-013**: Users may only access briefs and assets for campaigns owned by their Brand profile, except Editors who may access metadata for any published or active campaign.

---

## 3. User Scenarios and Testing

### 3.1 Brand Creates Creative Brief

**Scenario**: Brand owner creates a creative brief for a draft campaign
1. Brand authenticates and selects their campaign
2. System displays empty brief form
3. Brand fills all required fields
4. System validates required fields and saves brief
5. Brand sees confirmation that brief is saved

**Acceptance**: Brand can create one brief per campaign with all required fields validated.

### 3.2 Brand Updates Brief in Editable State

**Scenario**: Brand updates a brief while campaign is in draft
1. Brand authenticates and selects campaign with existing brief
2. System displays current brief with edit form
3. Brand modifies key_messages and mandatory_talking_points
4. System validates and saves updated brief
5. Brand sees confirmation of successful update

**Acceptance**: Updates succeed when campaign is in draft or paused state.

### 3.3 Brand Restricted from Updating Active Campaign Brief

**Scenario**: Brand attempts to modify key_messages on an active campaign
1. Brand authenticates and selects active campaign
2. System displays brief with restricted fields shown as read-only
3. Brand attempts to modify key_messages
4. System rejects the request with clear error message

**Acceptance**: Fields restricted by campaign state return appropriate error.

### 3.4 Editor Lists Assets for Published Campaign

**Scenario**: Editor views asset metadata for a published campaign
1. Editor authenticates with Editor profile type
2. Editor requests asset list for published campaign
3. System returns paginated list of asset metadata (no raw files)
4. Editor sees asset category, filename, version, and processing status

**Acceptance**: Editors can list assets for published/active campaigns.

### 3.5 Brand Registers New Asset Version

**Scenario**: Brand uploads new version of an existing asset
1. Brand authenticates and navigates to campaign assets
2. Brand initiates new asset registration with existing filename
3. System increments version number
4. Brand provides storage_key and checksum for new version
5. System creates new asset record with version = previous + 1

**Acceptance**: Same filename creates new versioned record; previous version remains accessible.

### 3.6 Brand Soft-Deletes Asset

**Scenario**: Brand removes an asset from campaign
1. Brand authenticates and navigates to asset list
2. Brand selects asset and initiates delete
3. System soft-deletes (sets deleted_at)
4. Asset no longer appears in default listing
5. Asset remains in database for audit purposes

**Acceptance**: Soft-deleted assets excluded from listings but preserved in database.

---

## 4. API Endpoints

### 4.1 Creative Brief

| Method | Endpoint | Description | Access |
|--------|----------|-------------|--------|
| GET | `/api/v1/campaigns/{campaignId}/brief` | Get campaign creative brief | Brand owner, Editor |
| PUT | `/api/v1/campaigns/{campaignId}/brief` | Create or replace brief | Brand owner only |

### 4.2 Asset Metadata

| Method | Endpoint | Description | Access |
|--------|----------|-------------|--------|
| POST | `/api/v1/campaigns/{campaignId}/assets` | Register new asset | Brand owner only |
| GET | `/api/v1/campaigns/{campaignId}/assets` | List campaign assets | Brand owner, Editor |
| GET | `/api/v1/assets/{id}` | Get asset details | Brand owner, Editor |
| PATCH | `/api/v1/assets/{id}` | Update asset metadata | Brand owner only |
| DELETE | `/api/v1/assets/{id}` | Soft-delete asset | Brand owner only |

---

## 5. Data Model

### 5.1 Creative Brief Entity

```
CreativeBrief
├── campaign_id: UUID (FK, unique)
├── key_messages: string[] (required)
├── product_benefits: string[] (required)
├── mandatory_talking_points: string[] (required)
├── prohibited_claims: string[] (optional)
├── required_hashtags: string[] (required)
├── call_to_action_text: string (required, max 200)
├── tone_and_style_guidelines: string (optional, max 1000)
├── target_audience_description: string (optional, max 500)
├── competitor_references: string[] (optional)
├── example_video_links: string[] (optional)
├── created_at: timestamp
├── updated_at: timestamp
```

### 5.2 Asset Metadata Entity

```
Asset
├── id: UUID (PK)
├── campaign_id: UUID (FK)
├── category: enum (asset_category)
├── original_filename: string (required, max 255)
├── display_name: string (required, max 255)
├── mime_type: string (required)
├── file_size_bytes: bigint (required, positive)
├── storage_key: string (required, max 500)
├── checksum: string (required, SHA-256 hex)
├── version: int (required, starts at 1)
├── processing_status: enum (processing_status)
├── virus_scan_status: enum (virus_scan_status)
├── uploaded_by_profile_id: UUID (FK)
├── created_at: timestamp
├── updated_at: timestamp
├── deleted_at: timestamp (nullable, for soft delete)
```

### 5.3 Relationships

- Campaign 1:1 CreativeBrief (one brief per campaign)
- Campaign 1:N Asset (many assets per campaign)
- Asset references Campaign through campaign_id
- Asset references Profile through uploaded_by_profile_id

---

## 6. Acceptance Criteria

### 6.1 Creative Brief Management

- [ ] Brand can create exactly one creative brief per campaign
- [ ] Brand can update brief fields when campaign is in draft or paused state
- [ ] Brand is restricted from modifying key_messages, product_benefits, mandatory_talking_points, required_hashtags, call_to_action_text when campaign is published or active
- [ ] Editors can view briefs for published and active campaigns
- [ ] Only one brief exists per campaign (enforced at database level)

### 6.2 Asset Metadata Management

- [ ] Brand can register asset metadata with all required fields
- [ ] Brand can update asset metadata fields (except campaign_id, version, checksum)
- [ ] Brand can soft-delete assets
- [ ] New uploads with same filename create new versioned records
- [ ] Previous asset versions remain accessible

### 6.3 Access Control

- [ ] Brand owners can only access briefs and assets for their own campaigns
- [ ] Editors can access briefs and assets for any published or active campaign
- [ ] Influencers are denied access to briefs and assets
- [ ] Unauthorized access attempts return appropriate HTTP errors

### 6.4 Editor Experience

- [ ] Editors can list assets with pagination
- [ ] Editors can view asset metadata (not raw file content)
- [ ] Editors cannot create, update, or delete briefs or assets

---

## 7. Success Criteria

1. **Brief Management**: Brand owners can create, read, and update one structured creative brief per campaign with field-level edit restrictions based on campaign lifecycle state.

2. **Asset Versioning**: Brand owners can register asset metadata with automatic version increment when the same filename is reused. Previous versions remain accessible for audit.

3. **Editor Access**: Authenticated Editors can list and view asset metadata for any published or active campaign, enabling content production workflows.

4. **Authorization Enforcement**: All endpoints enforce ownership and profile-type authorization. Brand owners manage their own campaign assets; Editors access metadata for eligible campaigns; Influencers are denied access.

5. **Test Coverage**: All APIs have contract tests; repository layer has integration tests; domain logic has unit tests. Authorization scenarios are covered.

---

## 8. Dependencies

- Campaign Management: Creative briefs and assets attach to existing campaigns; campaign ownership and lifecycle state determine access and editability
- Authentication: All endpoints require authenticated users
- Authorization: Brand/Editor/Influencer profile types determine access rights
- Profile: Asset records reference profile IDs for uploaded_by tracking

---

## 9. Assumptions

- Campaign ownership is verified via existing ownership middleware pattern
- Profile type verification follows existing profile type middleware pattern
- Soft deletion follows the same pattern as campaign soft deletion (deleted_at timestamp)
- Asset metadata registration happens after file upload to storage (storage key provided, not file content)
- Virus scan status is a placeholder field; actual scanning is out of scope
- File validation uses MIME type whitelist; no actual file content analysis

---

## 10. Out of Scope

The following are explicitly excluded from this feature:
- Binary file upload handling
- S3/R2 or any cloud storage integration
- Signed URLs or CDN delivery
- Video transcoding or processing
- Malware scanning (placeholder only)
- Automatic content analysis or approval workflows
- Influencer access to briefs or assets
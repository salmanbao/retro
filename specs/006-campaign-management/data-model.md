# Campaign Management - Data Model

## Entities

### Campaign

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| `id` | UUID | PK, not null | Auto-generated |
| `brand_profile_id` | UUID | FK to profiles, not null | Owner reference |
| `title` | string | not null, max 255 | Campaign title |
| `slug` | string | not null, unique, max 255 | URL-safe, normalized from title |
| `summary` | string | max 500 | Short description |
| `description` | text | not null | Detailed description |
| `objective` | string | max 255 | Campaign goal |
| `product_name` | string | max 255 | Product/service being promoted |
| `landing_url` | string | valid URL, max 2048 | Landing page |
| `total_budget` | decimal(12,2) | not null, > 0 | Total budget amount |
| `currency` | string(3) | not null, ISO 4217 | e.g., USD, EUR |
| `target_clips` | int | not null, >= 0 | Target approved clips |
| `target_posts` | int | not null, >= 0 | Target influencer posts |
| `cpv` | decimal(10,4) | not null | Cost per 1,000 views |
| `min_payout` | decimal(12,2) | nullable | Minimum payout per creator |
| `max_payout` | decimal(12,2) | nullable | Maximum payout per creator |
| `submission_start` | datetime | not null | When submissions open |
| `submission_deadline` | datetime | not null, > submission_start | When submissions close |
| `distribution_start` | datetime | not null, > submission_deadline | When distribution begins |
| `campaign_end` | datetime | not null, > distribution_start | When campaign ends |
| `allowed_countries` | string[] | not null, default all | ISO 3166-1 alpha-2 |
| `allowed_languages` | string[] | not null | ISO 639-1 |
| `min_followers` | int | not null, >= 0, default 0 | Minimum follower threshold |
| `platforms` | string[] | not null | instagram/tiktok/youtube/etc. |
| `creator_categories` | string[] | nullable | Creator type filters |
| `min_duration_secs` | int | not null, default 15 | Min video duration |
| `max_duration_secs` | int | not null, default 60 | Max video duration |
| `aspect_ratio` | string | not null | e.g., "9:16", "1:1" |
| `talking_points` | string[] | nullable | Required talking points |
| `prohibited_claims` | string[] | nullable | Forbidden claims |
| `hashtags` | string[] | nullable | Required hashtags |
| `cta_instructions` | string | nullable | Call-to-action text |
| `status` | enum | not null, default 'draft' | draft/published/active/paused/completed/cancelled |
| `version` | int | not null, default 1 | Optimistic locking |
| `created_at` | datetime | not null | Auto-generated |
| `updated_at` | datetime | not null | Auto-updated |
| `deleted_at` | datetime | nullable | Soft delete marker |

**Validation rules**:
- `min_payout <= max_payout` if both set
- `submission_deadline > submission_start`
- `distribution_start > submission_deadline`
- `campaign_end > distribution_start`
- `max_duration_secs >= min_duration_secs`

### CampaignAsset

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| `id` | UUID | PK, not null | Auto-generated |
| `campaign_id` | UUID | FK to campaigns, not null | Parent campaign |
| `url` | string | not null, valid URL, max 2048 | Asset location |
| `asset_type` | enum | not null | reference/raw_media/document |
| `description` | string | nullable, max 255 | Asset description |
| `created_at` | datetime | not null | Auto-generated |

## Relationships

- Campaign 1:N CampaignAsset (cascade delete)
- Profile 1:N Campaign (campaigns belong to Brand profiles)
- Campaign status transition rules enforced in service layer

## Indexes

- `idx_campaign_slug` ON campaigns(slug) WHERE deleted_at IS NULL
- `idx_campaign_brand_profile_id` ON campaigns(brand_profile_id) WHERE deleted_at IS NULL
- `idx_campaign_status` ON campaigns(status) WHERE deleted_at IS NULL
- `idx_campaign_asset_campaign_id` ON campaign_assets(campaign_id)

## Migrations

Migration file: `backend/migrations/XXXXXX_create_campaign_tables.up.sql`

Implements:
- campaigns table with all fields
- campaign_assets table
- All indexes
- Soft delete via deleted_at GORM callback
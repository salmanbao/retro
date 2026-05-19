-- Create campaign tables
CREATE TABLE campaigns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    brand_profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    summary VARCHAR(500),
    description TEXT NOT NULL,
    objective VARCHAR(255),
    product_name VARCHAR(255),
    landing_url VARCHAR(2048),
    total_budget DECIMAL(12, 2) NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    target_clips INTEGER NOT NULL DEFAULT 0,
    target_posts INTEGER NOT NULL DEFAULT 0,
    cpv DECIMAL(10, 4) NOT NULL DEFAULT 0,
    min_payout DECIMAL(12, 2),
    max_payout DECIMAL(12, 2),
    submission_start TIMESTAMPTZ NOT NULL,
    submission_deadline TIMESTAMPTZ NOT NULL,
    distribution_start TIMESTAMPTZ NOT NULL,
    campaign_end TIMESTAMPTZ NOT NULL,
    allowed_countries TEXT[] NOT NULL DEFAULT '{}',
    allowed_languages TEXT[] NOT NULL DEFAULT '{}',
    min_followers INTEGER NOT NULL DEFAULT 0,
    platforms TEXT[] NOT NULL DEFAULT '{}',
    creator_categories TEXT[],
    min_duration_secs INTEGER NOT NULL DEFAULT 15,
    max_duration_secs INTEGER NOT NULL DEFAULT 60,
    aspect_ratio VARCHAR(10) NOT NULL DEFAULT '9:16',
    talking_points TEXT[],
    prohibited_claims TEXT[],
    hashtags TEXT[],
    cta_instructions TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Indexes for campaigns
CREATE UNIQUE INDEX idx_campaign_slug ON campaigns(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_campaign_brand_profile_id ON campaigns(brand_profile_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_campaign_status ON campaigns(status) WHERE deleted_at IS NULL;

-- Campaign assets table
CREATE TABLE campaign_assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    url VARCHAR(2048) NOT NULL,
    asset_type VARCHAR(20) NOT NULL,
    description VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_campaign_asset_campaign_id ON campaign_assets(campaign_id);
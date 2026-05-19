-- Create creative_brief table
CREATE TABLE creative_briefs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL UNIQUE REFERENCES campaigns(id) ON DELETE CASCADE,
    key_messages JSONB NOT NULL DEFAULT '[]',
    product_benefits JSONB NOT NULL DEFAULT '[]',
    mandatory_talking_points JSONB NOT NULL DEFAULT '[]',
    prohibited_claims JSONB DEFAULT '[]',
    required_hashtags JSONB NOT NULL DEFAULT '[]',
    call_to_action_text TEXT,
    tone_and_style_guidelines TEXT,
    target_audience_description TEXT,
    competitor_references JSONB DEFAULT '[]',
    example_video_links JSONB DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create asset_metadata table
CREATE TABLE asset_metadata (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    category TEXT NOT NULL CHECK (category IN (
        'raw_footage', 'product_images', 'logos', 'brand_guidelines',
        'scripts', 'example_videos', 'music_references', 'legal_documents', 'other'
    )),
    original_filename TEXT NOT NULL,
    display_name TEXT NOT NULL,
    mime_type TEXT NOT NULL,
    file_size_bytes BIGINT NOT NULL,
    storage_key TEXT NOT NULL,
    checksum TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    processing_status TEXT NOT NULL DEFAULT 'pending' CHECK (processing_status IN (
        'pending', 'uploaded', 'processing', 'ready', 'failed', 'archived'
    )),
    virus_scan_status TEXT NOT NULL DEFAULT 'not_scanned' CHECK (virus_scan_status IN (
        'not_scanned', 'scanning', 'clean', 'infected'
    )),
    uploaded_by_profile_id UUID NOT NULL REFERENCES profiles(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Indexes for creative_briefs
CREATE INDEX idx_creative_briefs_campaign_id ON creative_briefs(campaign_id);

-- Indexes for asset_metadata
CREATE INDEX idx_asset_metadata_campaign_id ON asset_metadata(campaign_id);
CREATE INDEX idx_asset_metadata_category ON asset_metadata(category);
CREATE INDEX idx_asset_metadata_processing_status ON asset_metadata(processing_status);
CREATE INDEX idx_asset_metadata_deleted_at ON asset_metadata(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_asset_metadata_version_lookup ON asset_metadata(campaign_id, original_filename, version DESC);
CREATE INDEX idx_asset_metadata_uploaded_by ON asset_metadata(uploaded_by_profile_id);
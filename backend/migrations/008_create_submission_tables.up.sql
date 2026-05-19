-- Create submissions table for campaign submissions
CREATE TABLE IF NOT EXISTS submissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    editor_profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    video_url VARCHAR(2000) NOT NULL,
    thumbnail_url VARCHAR(2000),
    duration_seconds INTEGER NOT NULL CHECK (duration_seconds > 0),
    notes TEXT,
    tags JSONB DEFAULT '[]'::jsonb,
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN (
        'draft', 'submitted', 'under_review', 'shortlisted', 'approved', 'rejected', 'withdrawn'
    )),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    submitted_at TIMESTAMP WITH TIME ZONE,
    reviewed_at TIMESTAMP WITH TIME ZONE,
    withdrawn_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Indexes
CREATE INDEX idx_submissions_campaign_id ON submissions(campaign_id);
CREATE INDEX idx_submissions_editor_profile_id ON submissions(editor_profile_id);
CREATE INDEX idx_submissions_campaign_status ON submissions(campaign_id, status);
CREATE INDEX idx_submissions_deleted_at ON submissions(deleted_at);

-- Composite unique constraint: one non-draft submission per editor per campaign
-- Drafts are excluded (editor can have multiple drafts)
CREATE UNIQUE INDEX idx_unique_submission_per_editor_campaign
    ON submissions(campaign_id, editor_profile_id)
    WHERE status NOT IN ('draft') AND deleted_at IS NULL;
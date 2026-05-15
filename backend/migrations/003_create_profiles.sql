-- Migration: 003_create_profiles
-- Create profiles table for multi-role profiles

CREATE TYPE profile_type AS ENUM ('brand', 'editor', 'influencer');

CREATE TABLE profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    profile_type profile_type NOT NULL,
    name VARCHAR(255) NOT NULL,
    details JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_profiles_user_id ON profiles(user_id);

-- Add foreign key for active_profile_id in sessions after profiles table exists
ALTER TABLE sessions ADD CONSTRAINT fk_sessions_active_profile
    FOREIGN KEY (active_profile_id) REFERENCES profiles(id) ON DELETE SET NULL;
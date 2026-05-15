-- Migration: 004_create_auth_tokens
-- Create auth_tokens table for email verification and password reset tokens

CREATE TYPE auth_token_type AS ENUM ('verification', 'password_reset');

CREATE TABLE auth_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_type auth_token_type NOT NULL,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_auth_tokens_user_id ON auth_tokens(user_id, token_type);
CREATE INDEX idx_auth_tokens_token_hash ON auth_tokens(token_hash);
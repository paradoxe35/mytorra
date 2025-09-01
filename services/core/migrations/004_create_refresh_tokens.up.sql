-- Create refresh_tokens table for secure token storage
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL, -- Bcrypt hash of the actual token
    device_id VARCHAR(100), -- Optional device identifier
    device_name VARCHAR(255), -- Optional human-readable device name
    ip_address INET, -- IP address when token was created
    user_agent TEXT, -- Browser/app user agent
    expires_at TIMESTAMP NOT NULL,
    last_used_at TIMESTAMP,
    revoked_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Ensure token_hash is unique (each token should be unique)
    CONSTRAINT unique_token_hash UNIQUE (token_hash)
);

-- Indexes for efficient querying
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id) WHERE revoked_at IS NULL;
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at) WHERE revoked_at IS NULL;
CREATE INDEX idx_refresh_tokens_device_id ON refresh_tokens(device_id) WHERE revoked_at IS NULL;
CREATE INDEX idx_refresh_tokens_created_at ON refresh_tokens(created_at);

-- Add a comment explaining the security model
COMMENT ON TABLE refresh_tokens IS 'Stores hashed refresh tokens for secure authentication. Actual tokens are never stored, only their bcrypt hashes.';
COMMENT ON COLUMN refresh_tokens.token_hash IS 'Bcrypt hash of the actual refresh token. The actual token is never stored.';

-- Create OAuth profiles table for storing OAuth provider information
CREATE TABLE IF NOT EXISTS oauth_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    provider_id VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    name VARCHAR(255),
    picture TEXT,
    access_token TEXT,  -- Will be encrypted at application level
    refresh_token TEXT, -- Will be encrypted at application level
    token_expiry TIMESTAMP WITH TIME ZONE,
    scopes TEXT[],
    raw_profile JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT unique_provider_user UNIQUE(provider, provider_id)
);

-- Add indexes for performance
CREATE INDEX idx_oauth_profiles_user_id ON oauth_profiles(user_id);
CREATE INDEX idx_oauth_profiles_provider_email ON oauth_profiles(provider, email);

-- Create OAuth states table for CSRF protection and PKCE
CREATE TABLE IF NOT EXISTS oauth_states (
    state VARCHAR(255) PRIMARY KEY,
    code_verifier VARCHAR(255),  -- For PKCE flow
    redirect_uri TEXT,
    provider VARCHAR(50) NOT NULL,
    ip_address INET,
    user_agent TEXT,
    user_id UUID, -- Optional, for linking existing accounts
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);

-- Index for cleanup of expired states
CREATE INDEX idx_oauth_states_expires ON oauth_states(expires_at);

-- Create update trigger for updated_at column
CREATE OR REPLACE FUNCTION update_oauth_profiles_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_oauth_profiles_updated_at_trigger
    BEFORE UPDATE ON oauth_profiles
    FOR EACH ROW
    EXECUTE FUNCTION update_oauth_profiles_updated_at();

-- Add comments for documentation
COMMENT ON TABLE oauth_profiles IS 'Stores OAuth provider profiles linked to users';
COMMENT ON TABLE oauth_states IS 'Temporary storage for OAuth state parameters for CSRF protection';
COMMENT ON COLUMN oauth_profiles.provider IS 'OAuth provider name (google, facebook, etc.)';
COMMENT ON COLUMN oauth_profiles.provider_id IS 'User ID from the OAuth provider';
COMMENT ON COLUMN oauth_states.code_verifier IS 'PKCE code verifier for enhanced security';

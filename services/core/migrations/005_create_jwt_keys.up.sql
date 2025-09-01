-- Create jwt_keys table for storing RSA key pairs
CREATE TABLE IF NOT EXISTS jwt_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_id VARCHAR(100) UNIQUE NOT NULL, -- Unique identifier for the key (kid in JWK)
    private_key TEXT NOT NULL, -- PEM-encoded private key (encrypted in production)
    public_key TEXT NOT NULL, -- PEM-encoded public key
    algorithm VARCHAR(20) NOT NULL DEFAULT 'RS256',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    rotated_at TIMESTAMP, -- When this key was rotated (replaced by a new one)

    CONSTRAINT valid_algorithm CHECK (algorithm IN ('RS256', 'RS384', 'RS512'))
);

-- Index for active keys
CREATE INDEX idx_jwt_keys_active ON jwt_keys(is_active, expires_at) WHERE is_active = true;
CREATE INDEX idx_jwt_keys_key_id ON jwt_keys(key_id);
CREATE INDEX idx_jwt_keys_expires_at ON jwt_keys(expires_at);

-- Add a comment explaining the security model
COMMENT ON TABLE jwt_keys IS 'Stores RSA key pairs for JWT signing. Private keys should be encrypted at rest in production.';
COMMENT ON COLUMN jwt_keys.private_key IS 'PEM-encoded RSA private key. Must be encrypted at rest in production environments.';
COMMENT ON COLUMN jwt_keys.public_key IS 'PEM-encoded RSA public key. Used for JWT verification by other services.';
COMMENT ON COLUMN jwt_keys.key_id IS 'Unique key identifier used in JWT kid header and JWKS endpoint.';

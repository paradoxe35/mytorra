CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    phone VARCHAR(20) UNIQUE,
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    profile_image_url TEXT,

    -- Verification
    email_verified BOOLEAN DEFAULT FALSE,
    phone_verified BOOLEAN DEFAULT FALSE,
    id_verified BOOLEAN DEFAULT FALSE,
    verification_level INT DEFAULT 0,

    -- OAuth
    google_id VARCHAR(255) UNIQUE,
    facebook_id VARCHAR(255) UNIQUE,

    -- MFA
    mfa_enabled BOOLEAN DEFAULT FALSE,
    mfa_secret VARCHAR(255),

    -- Location
    area VARCHAR(50) NOT NULL,
    last_known_lat DECIMAL(10, 8),
    last_known_lng DECIMAL(11, 8),

    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    last_login_at TIMESTAMP,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,

    CONSTRAINT valid_area CHECK (area IN ('bukavu', 'goma')),
    CONSTRAINT valid_status CHECK (status IN ('active', 'suspended', 'banned', 'deleted'))
);

CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_phone ON users(phone) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_area ON users(area) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_status ON users(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_created_at ON users(created_at);

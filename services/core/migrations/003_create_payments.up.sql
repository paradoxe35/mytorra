CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    amount DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    type VARCHAR(20) NOT NULL,
    method VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',

    -- Provider details
    provider_id VARCHAR(255),
    provider_name VARCHAR(50),

    -- Transaction details
    description TEXT,
    metadata JSONB,

    -- Refund information
    refunded_amount DECIMAL(10, 2),
    refund_reason TEXT,

    -- Commission
    commission DECIMAL(10, 2),
    commission_rate DECIMAL(5, 4),

    -- Timestamps
    processed_at TIMESTAMP,
    failed_at TIMESTAMP,
    refunded_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_type CHECK (type IN ('purchase', 'subscription', 'wallet_topup', 'commission')),
    CONSTRAINT valid_method CHECK (method IN ('card', 'wallet', 'mobile_money', 'bank_transfer')),
    CONSTRAINT valid_status CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'refunded'))
);

CREATE INDEX idx_payments_user_id ON payments(user_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_created_at ON payments(created_at);
CREATE INDEX idx_payments_provider_id ON payments(provider_id);

-- Create subscription_payment_transactions table
CREATE TABLE IF NOT EXISTS subscription_payment_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    subscription_id UUID NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
    billing_cycle_id UUID REFERENCES billing_cycles(id) ON DELETE SET NULL,
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    type VARCHAR(50) NOT NULL, -- 'subscription', 'upgrade', 'downgrade', 'refund', 'proration'
    status VARCHAR(50) NOT NULL, -- 'pending', 'processing', 'succeeded', 'failed', 'canceled', 'refunded'
    gateway_transaction_id VARCHAR(255),
    gateway_response JSONB,
    processed_at TIMESTAMP,
    failure_reason TEXT,
    retry_count INTEGER DEFAULT 0,
    idempotency_key VARCHAR(255) UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_amount CHECK (amount >= 0),
    CONSTRAINT valid_retry_count CHECK (retry_count >= 0)
);

-- Create indexes
CREATE INDEX idx_spt_subscription ON subscription_payment_transactions(subscription_id);
CREATE INDEX idx_spt_billing_cycle ON subscription_payment_transactions(billing_cycle_id);
CREATE INDEX idx_spt_status ON subscription_payment_transactions(status);
CREATE INDEX idx_spt_type ON subscription_payment_transactions(type);
CREATE INDEX idx_spt_gateway_id ON subscription_payment_transactions(gateway_transaction_id);
CREATE INDEX idx_spt_processed_at ON subscription_payment_transactions(processed_at) WHERE processed_at IS NOT NULL;
CREATE INDEX idx_spt_created_at ON subscription_payment_transactions(created_at);

-- Add reference to payments table if it exists
ALTER TABLE payments ADD COLUMN IF NOT EXISTS subscription_id UUID REFERENCES subscriptions(id);
ALTER TABLE payments ADD COLUMN IF NOT EXISTS billing_cycle_id UUID REFERENCES billing_cycles(id);

-- Create payment_methods table for storing customer payment methods
CREATE TABLE IF NOT EXISTS payment_methods (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    gateway_method_id VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL, -- 'card', 'bank_account', 'paypal', etc.
    last4 VARCHAR(4),
    brand VARCHAR(50),
    exp_month INTEGER,
    exp_year INTEGER,
    is_default BOOLEAN DEFAULT false,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payment_methods_user ON payment_methods(user_id);
CREATE INDEX idx_payment_methods_gateway ON payment_methods(gateway_method_id);
CREATE INDEX idx_payment_methods_default ON payment_methods(user_id, is_default) WHERE is_default = true;

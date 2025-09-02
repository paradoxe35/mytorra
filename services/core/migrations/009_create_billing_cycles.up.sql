-- Create billing_cycles table
CREATE TABLE IF NOT EXISTS billing_cycles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    subscription_id UUID NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    invoice_id VARCHAR(255),
    invoice_number VARCHAR(100),
    paid_at TIMESTAMP,
    payment_transaction_id UUID,
    attempts INTEGER DEFAULT 0,
    last_attempt_at TIMESTAMP,
    next_retry_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_dates CHECK (end_date > start_date),
    CONSTRAINT valid_amount CHECK (amount >= 0)
);

-- Create indexes
CREATE INDEX idx_billing_cycles_subscription ON billing_cycles(subscription_id);
CREATE INDEX idx_billing_cycles_status ON billing_cycles(status);
CREATE INDEX idx_billing_cycles_dates ON billing_cycles(start_date, end_date);
CREATE INDEX idx_billing_cycles_paid_at ON billing_cycles(paid_at) WHERE paid_at IS NOT NULL;
CREATE INDEX idx_billing_cycles_next_retry ON billing_cycles(next_retry_at) WHERE next_retry_at IS NOT NULL AND status != 'paid';

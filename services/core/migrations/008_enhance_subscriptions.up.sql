-- Add missing columns to existing subscriptions table
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS tier_id UUID REFERENCES subscription_tiers(id);
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS stripe_subscription_id VARCHAR(255);
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS stripe_customer_id VARCHAR(255);
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS metadata JSONB DEFAULT '{}';
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS gateway_subscription_id VARCHAR(255);
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS payment_method_id VARCHAR(255);
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS proration_behavior VARCHAR(50) DEFAULT 'create_prorations';

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_entity_id ON subscriptions(entity_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(status);
CREATE INDEX IF NOT EXISTS idx_subscriptions_tier_id ON subscriptions(tier_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_trial_ends ON subscriptions(trial_ends_at)
    WHERE trial_ends_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_subscriptions_next_billing ON subscriptions(next_billing_date)
    WHERE next_billing_date IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_subscriptions_stripe_subscription ON subscriptions(stripe_subscription_id)
    WHERE stripe_subscription_id IS NOT NULL;

-- Add constraint for active subscriptions
ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS unique_active_subscription;
ALTER TABLE subscriptions ADD CONSTRAINT unique_active_subscription
    UNIQUE (user_id, entity_id, entity_type)
    WHERE status IN ('active', 'trialing');

-- Update existing subscriptions to link with tiers
UPDATE subscriptions s
SET tier_id = t.id
FROM subscription_tiers t
WHERE LOWER(s.tier) = LOWER(t.code)
AND s.tier_id IS NULL;

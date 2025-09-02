-- Remove added columns from subscriptions table
ALTER TABLE subscriptions DROP COLUMN IF EXISTS tier_id;
ALTER TABLE subscriptions DROP COLUMN IF EXISTS stripe_subscription_id;
ALTER TABLE subscriptions DROP COLUMN IF EXISTS stripe_customer_id;
ALTER TABLE subscriptions DROP COLUMN IF EXISTS metadata;
ALTER TABLE subscriptions DROP COLUMN IF EXISTS gateway_subscription_id;
ALTER TABLE subscriptions DROP COLUMN IF EXISTS payment_method_id;
ALTER TABLE subscriptions DROP COLUMN IF EXISTS proration_behavior;

-- Drop added indexes
DROP INDEX IF EXISTS idx_subscriptions_tier_id;
DROP INDEX IF EXISTS idx_subscriptions_stripe_subscription;

-- Drop constraint
ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS unique_active_subscription;

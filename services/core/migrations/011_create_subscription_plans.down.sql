-- Drop new subscription columns
DROP INDEX IF EXISTS idx_subscriptions_area_id;
DROP INDEX IF EXISTS idx_subscriptions_plan_id;

ALTER TABLE subscriptions DROP COLUMN IF EXISTS area_id;
ALTER TABLE subscriptions DROP COLUMN IF EXISTS plan_id;

-- Drop plan_area_pricing table
DROP INDEX IF EXISTS idx_plan_area_pricing_active;
DROP INDEX IF EXISTS idx_plan_area_pricing_area_id;
DROP INDEX IF EXISTS idx_plan_area_pricing_plan_id;

DROP TABLE IF EXISTS plan_area_pricing;

-- Drop subscription_plans table
DROP INDEX IF EXISTS idx_subscription_plans_sort_order;
DROP INDEX IF EXISTS idx_subscription_plans_active;
DROP INDEX IF EXISTS idx_subscription_plans_code;

DROP TABLE IF EXISTS subscription_plans;

-- Drop payment_methods table
DROP TABLE IF EXISTS payment_methods CASCADE;

-- Remove columns from payments table
ALTER TABLE payments DROP COLUMN IF EXISTS subscription_id;
ALTER TABLE payments DROP COLUMN IF EXISTS billing_cycle_id;

-- Drop subscription_payment_transactions table
DROP TABLE IF EXISTS subscription_payment_transactions CASCADE;

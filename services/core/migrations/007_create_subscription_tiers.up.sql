-- Create subscription_tiers table
CREATE TABLE IF NOT EXISTS subscription_tiers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) NOT NULL UNIQUE,
    price DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    features JSONB NOT NULL DEFAULT '{}',
    limits JSONB NOT NULL DEFAULT '{}',
    trial_days INTEGER DEFAULT 0,
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_subscription_tiers_code ON subscription_tiers(code);
CREATE INDEX idx_subscription_tiers_active ON subscription_tiers(is_active);
CREATE INDEX idx_subscription_tiers_sort_order ON subscription_tiers(sort_order);

-- Insert default tiers
INSERT INTO subscription_tiers (name, code, price, currency, features, limits, trial_days, sort_order) VALUES
('Basic', 'basic', 9.99, 'USD',
 '{"support": "email", "api_access": true, "max_projects": 3}',
 '{"api_calls": 1000, "storage_gb": 10}',
 7, 1),
('Professional', 'professional', 29.99, 'USD',
 '{"support": "priority", "api_access": true, "max_projects": 10, "analytics": true}',
 '{"api_calls": 10000, "storage_gb": 100}',
 14, 2),
('Premium', 'premium', 99.99, 'USD',
 '{"support": "24/7", "api_access": true, "max_projects": -1, "analytics": true, "custom_domain": true}',
 '{"api_calls": -1, "storage_gb": 1000}',
 30, 3),
('Enterprise', 'enterprise', 299.99, 'USD',
 '{"support": "dedicated", "api_access": true, "max_projects": -1, "analytics": true, "custom_domain": true, "sla": true}',
 '{"api_calls": -1, "storage_gb": -1}',
 30, 4);

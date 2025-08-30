CREATE TABLE IF NOT EXISTS areas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id UUID REFERENCES areas(id),

    -- Area hierarchy
    type VARCHAR(20) NOT NULL, -- 'country', 'province', 'town', 'subarea'
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) UNIQUE,

    -- Geographic data
    boundary_polygon JSONB, -- GeoJSON polygon for area boundaries
    center_lat DECIMAL(10, 8),
    center_lng DECIMAL(11, 8),
    operational_radius INT, -- in kilometers

    -- Configuration
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    service_availability JSONB, -- {"mytorra": true, "torra_shop": true, ...}
    settings JSONB, -- Area-specific settings

    -- Area Behavior Settings
    switching_mode VARCHAR(20) NOT NULL DEFAULT 'manual',
    allow_cross_trade BOOLEAN DEFAULT FALSE,
    nearby_area_ids UUID[], -- Array of related area IDs
    transition_zones JSONB, -- Overlapping boundary configurations

    -- Operational data
    timezone VARCHAR(50),
    languages TEXT[], -- Array of supported languages
    currency VARCHAR(3) DEFAULT 'USD',
    tax_rate DECIMAL(5, 4),

    -- Service limits
    min_order_value DECIMAL(10, 2),
    delivery_fee_structure JSONB,
    cross_area_fees JSONB, -- {"shipping": 5.00, "express": 10.00}
    operational_hours JSONB,

    -- Metadata
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    activated_at TIMESTAMP,
    deactivated_at TIMESTAMP,

    CONSTRAINT valid_area_type CHECK (type IN ('country', 'province', 'town', 'subarea')),
    CONSTRAINT valid_status CHECK (status IN ('active', 'inactive', 'testing', 'suspended')),
    CONSTRAINT valid_switching_mode CHECK (switching_mode IN ('automatic', 'manual', 'hybrid', 'locked'))
);

-- Additional tables for area relationships
CREATE TABLE IF NOT EXISTS area_relationships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    area_id UUID NOT NULL REFERENCES areas(id) ON DELETE CASCADE,
    related_area_id UUID NOT NULL REFERENCES areas(id) ON DELETE CASCADE,
    relationship_type VARCHAR(50) NOT NULL, -- 'sister_city', 'metro_zone', 'trade_partner'
    bidirectional BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_area_relationship UNIQUE (area_id, related_area_id, relationship_type)
);

CREATE INDEX idx_areas_parent_id ON areas(parent_id);
CREATE INDEX idx_areas_type ON areas(type);
CREATE INDEX idx_areas_status ON areas(status);
CREATE INDEX idx_areas_name ON areas(name);
CREATE INDEX idx_areas_code ON areas(code);
CREATE INDEX idx_areas_center_location ON areas(center_lat, center_lng);

-- Create hierarchical path view for easier querying
CREATE OR REPLACE VIEW area_hierarchy AS
WITH RECURSIVE area_path AS (
    SELECT
        id,
        parent_id,
        name,
        type,
        ARRAY[name] as path,
        1 as level
    FROM areas
    WHERE parent_id IS NULL

    UNION ALL

    SELECT
        a.id,
        a.parent_id,
        a.name,
        a.type,
        ap.path || a.name,
        ap.level + 1
    FROM areas a
    INNER JOIN area_path ap ON a.parent_id = ap.id
)
SELECT * FROM area_path;

-- Insert initial areas for DRC (Democratic Republic of Congo)
INSERT INTO areas (type, name, code, center_lat, center_lng, operational_radius, currency, languages, status) VALUES
    ('country', 'Democratic Republic of Congo', 'CD', -2.5, 23.5, 2000, 'CDF', ARRAY['fr', 'sw', 'ln', 'kg'], 'active');

-- Get the country ID for inserting provinces
WITH country AS (
    SELECT id FROM areas WHERE code = 'CD' LIMIT 1
)
INSERT INTO areas (parent_id, type, name, code, center_lat, center_lng, operational_radius, currency, languages, status)
SELECT
    country.id,
    'province',
    province.name,
    province.code,
    province.lat,
    province.lng,
    province.radius,
    'CDF',
    ARRAY['fr', 'sw'],
    'active'
FROM country, (VALUES
    ('South Kivu', 'CD-SK', -2.5083, 28.8608, 150),
    ('North Kivu', 'CD-NK', -0.7008, 29.2708, 150)
) AS province(name, code, lat, lng, radius);

-- Insert towns for South Kivu
WITH south_kivu AS (
    SELECT id FROM areas WHERE code = 'CD-SK' LIMIT 1
)
INSERT INTO areas (parent_id, type, name, code, center_lat, center_lng, operational_radius, currency, languages, status)
SELECT
    south_kivu.id,
    'town',
    town.name,
    town.code,
    town.lat,
    town.lng,
    town.radius,
    'CDF',
    ARRAY['fr', 'sw'],
    'active'
FROM south_kivu, (VALUES
    ('Bukavu', 'CD-SK-BKV', -2.5083, 28.8608, 30),
    ('Uvira', 'CD-SK-UVR', -3.3953, 29.1378, 20),
    ('Baraka', 'CD-SK-BRK', -4.1039, 29.0947, 15)
) AS town(name, code, lat, lng, radius);

-- Insert towns for North Kivu
WITH north_kivu AS (
    SELECT id FROM areas WHERE code = 'CD-NK' LIMIT 1
)
INSERT INTO areas (parent_id, type, name, code, center_lat, center_lng, operational_radius, currency, languages, status)
SELECT
    north_kivu.id,
    'town',
    town.name,
    town.code,
    town.lat,
    town.lng,
    town.radius,
    'CDF',
    ARRAY['fr', 'sw'],
    'active'
FROM north_kivu, (VALUES
    ('Goma', 'CD-NK-GOM', -1.6744, 29.2336, 25),
    ('Beni', 'CD-NK-BNI', 0.4917, 29.4731, 20),
    ('Butembo', 'CD-NK-BTB', 0.1419, 29.2914, 20)
) AS town(name, code, lat, lng, radius);

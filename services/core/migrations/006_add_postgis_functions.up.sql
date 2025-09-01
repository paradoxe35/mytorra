-- Create a function to find areas by coordinate using PostGIS
-- This replaces the manual Haversine formula calculation with PostGIS optimized spatial queries

CREATE OR REPLACE FUNCTION find_areas_by_coordinate(
    input_lat DOUBLE PRECISION,
    input_lng DOUBLE PRECISION,
    max_distance_km INTEGER DEFAULT 50
) RETURNS TABLE (
    id UUID,
    name VARCHAR(255),
    code VARCHAR(50),
    type VARCHAR(20),
    parent_id UUID,
    center_lat DECIMAL(10, 8),
    center_lng DECIMAL(11, 8),
    operational_radius INTEGER,
    distance_km DOUBLE PRECISION,
    status VARCHAR(20)
) AS $$
BEGIN
    RETURN QUERY
    WITH input_point AS (
        SELECT ST_SetSRID(ST_MakePoint(input_lng, input_lat), 4326)::geography AS geog
    )
    SELECT
        a.id,
        a.name,
        a.code,
        a.type,
        a.parent_id,
        a.center_lat,
        a.center_lng,
        a.operational_radius,
        ST_Distance(a.center_point::geography, ip.geog) / 1000 AS distance_km,
        a.status
    FROM areas a, input_point ip
    WHERE a.status = 'active'
        AND (
            -- Check if point is within operational radius
            ST_DWithin(a.center_point::geography, ip.geog, a.operational_radius * 1000)
            OR
            -- Or if point is within the boundary polygon (if defined)
            (a.boundary IS NOT NULL AND ST_Contains(a.boundary, ST_SetSRID(ST_MakePoint(input_lng, input_lat), 4326)))
        )
    ORDER BY
        CASE a.type
            WHEN 'subarea' THEN 1
            WHEN 'town' THEN 2
            WHEN 'province' THEN 3
            WHEN 'country' THEN 4
        END,
        distance_km ASC;

    -- If no areas found within operational radius, find the nearest one
    IF NOT FOUND THEN
        RETURN QUERY
        WITH input_point AS (
            SELECT ST_SetSRID(ST_MakePoint(input_lng, input_lat), 4326)::geography AS geog
        )
        SELECT
            a.id,
            a.name,
            a.code,
            a.type,
            a.parent_id,
            a.center_lat,
            a.center_lng,
            a.operational_radius,
            ST_Distance(a.center_point::geography, ip.geog) / 1000 AS distance_km,
            a.status
        FROM areas a, input_point ip
        WHERE a.status = 'active'
            AND ST_DWithin(a.center_point::geography, ip.geog, max_distance_km * 1000)
        ORDER BY distance_km ASC
        LIMIT 1;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Create a function to check if a point is within an area
CREATE OR REPLACE FUNCTION is_point_in_area(
    area_id UUID,
    lat DOUBLE PRECISION,
    lng DOUBLE PRECISION
) RETURNS BOOLEAN AS $$
DECLARE
    result BOOLEAN;
BEGIN
    SELECT
        CASE
            WHEN a.boundary IS NOT NULL THEN
                ST_Contains(a.boundary, ST_SetSRID(ST_MakePoint(lng, lat), 4326))
            ELSE
                ST_DWithin(
                    a.center_point::geography,
                    ST_SetSRID(ST_MakePoint(lng, lat), 4326)::geography,
                    a.operational_radius * 1000
                )
        END INTO result
    FROM areas a
    WHERE a.id = area_id;

    RETURN COALESCE(result, FALSE);
END;
$$ LANGUAGE plpgsql;

-- Create a function to find neighboring areas
CREATE OR REPLACE FUNCTION find_neighboring_areas(
    area_id UUID,
    max_distance_km INTEGER DEFAULT 100
) RETURNS TABLE (
    id UUID,
    name VARCHAR(255),
    code VARCHAR(50),
    type VARCHAR(20),
    distance_km DOUBLE PRECISION
) AS $$
BEGIN
    RETURN QUERY
    WITH source_area AS (
        SELECT center_point, boundary, operational_radius
        FROM areas
        WHERE id = area_id
    )
    SELECT
        a.id,
        a.name,
        a.code,
        a.type,
        ST_Distance(a.center_point::geography, sa.center_point::geography) / 1000 AS distance_km
    FROM areas a, source_area sa
    WHERE a.id != area_id
        AND a.status = 'active'
        AND (
            -- Check distance between centers
            ST_DWithin(a.center_point::geography, sa.center_point::geography, max_distance_km * 1000)
            OR
            -- Check if boundaries touch or overlap (if defined)
            (a.boundary IS NOT NULL AND sa.boundary IS NOT NULL
             AND (ST_Touches(a.boundary, sa.boundary) OR ST_Overlaps(a.boundary, sa.boundary)))
        )
    ORDER BY distance_km ASC;
END;
$$ LANGUAGE plpgsql;

-- Add comment for documentation
COMMENT ON FUNCTION find_areas_by_coordinate IS 'Find areas that contain or are near a given coordinate point using PostGIS spatial functions';
COMMENT ON FUNCTION is_point_in_area IS 'Check if a coordinate point is within a specific area boundary or operational radius';
COMMENT ON FUNCTION find_neighboring_areas IS 'Find all neighboring areas within a specified distance or with touching boundaries';

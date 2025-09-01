-- Enable required PostgreSQL extensions
-- These must be created before any other migrations that depend on them

-- PostGIS for spatial/geographic data and operations
CREATE EXTENSION IF NOT EXISTS postgis;

-- UUID generation functions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Trigram matching for text similarity searches
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Ensure SRID 4326 (WGS 84) exists in spatial_ref_sys
-- This is the standard GPS coordinate system (latitude/longitude)
-- Sometimes PostGIS doesn't populate this automatically in test/Docker environments
INSERT INTO spatial_ref_sys (srid, auth_name, auth_srid, proj4text, srtext)
SELECT 4326, 'EPSG', 4326,
    '+proj=longlat +datum=WGS84 +no_defs',
    'GEOGCS["WGS 84",DATUM["WGS_1984",SPHEROID["WGS 84",6378137,298.257223563,AUTHORITY["EPSG","7030"]],AUTHORITY["EPSG","6326"]],PRIMEM["Greenwich",0,AUTHORITY["EPSG","8901"]],UNIT["degree",0.0174532925199433,AUTHORITY["EPSG","9122"]],AUTHORITY["EPSG","4326"]]'
WHERE NOT EXISTS (SELECT 1 FROM spatial_ref_sys WHERE srid = 4326);

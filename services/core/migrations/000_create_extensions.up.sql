-- Enable required PostgreSQL extensions
-- These must be created before any other migrations that depend on them

-- PostGIS for spatial/geographic data and operations
CREATE EXTENSION IF NOT EXISTS postgis;

-- UUID generation functions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Trigram matching for text similarity searches
CREATE EXTENSION IF NOT EXISTS pg_trgm;

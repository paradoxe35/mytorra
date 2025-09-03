-- Core extensions that are always enabled

-- Enable UUID generation support (always enabled)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enable pgvector for vector similarity search (always enabled as we install it)
CREATE EXTENSION IF NOT EXISTS vector;

-- Enable fuzzystrmatch for fuzzy string matching (commonly used with PostGIS)
CREATE EXTENSION IF NOT EXISTS fuzzystrmatch;

-- Enable tablefunc for crosstab and other table functions
CREATE EXTENSION IF NOT EXISTS tablefunc;

-- Print enabled core extensions
DO $$
BEGIN
    RAISE NOTICE 'Core extensions enabled successfully';
    RAISE NOTICE 'Enabled: uuid-ossp, vector, fuzzystrmatch, tablefunc';
    RAISE NOTICE 'PostGIS extensions are pre-enabled by the base image';
END $$;

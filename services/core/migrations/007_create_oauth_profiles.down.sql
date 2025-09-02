-- Drop OAuth related tables and functions
DROP TRIGGER IF EXISTS update_oauth_profiles_updated_at_trigger ON oauth_profiles;
DROP FUNCTION IF EXISTS update_oauth_profiles_updated_at();
DROP TABLE IF EXISTS oauth_states;
DROP TABLE IF EXISTS oauth_profiles;

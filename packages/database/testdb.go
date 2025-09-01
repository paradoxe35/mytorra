// Package database provides test database utilities for all services.
//
// Environment Variables:
//   - SERVICE_NAME: Name of the service (e.g., "core", "payment", "delivery")
//     If not set, will try to auto-detect from working directory
//   - ENABLE_DB_TESTS: Set to "true" to enable database tests
//   - TEST_DB_HOST: PostgreSQL host (default: localhost)
//   - TEST_DB_PORT: PostgreSQL port (default: 5432)
//   - TEST_DB_USER: Database user (default: postgres)
//   - TEST_DB_PASSWORD: Database password (default: postgres)
//   - TEST_DB_NAME: Shared test database name (default: test_torra)
//   - TEST_DB_ISOLATE: Create isolated database per test (default: true)
//   - TEST_DB_DROP_AFTER: Drop database after test (default: false)
//
// Usage:
//
//	// In your test file
//	func TestRepository(t *testing.T) {
//	    // Set service name if running tests from non-standard location
//	    os.Setenv("SERVICE_NAME", "myservice")
//
//	    testDB, err := database.SetupTestDB(nil)
//	    if err != nil {
//	        t.Fatal(err)
//	    }
//	    defer testDB.Close()
//
//	    // Your test code here
//	}
package database

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/paradoxe35/torra/packages/config"
)

// TestDB represents a test database connection with cleanup capabilities
type TestDB struct {
	*DB
	dbName   string
	cleanup  func() error
	migrated bool
}

// TestConfig holds configuration for test database setup
type TestConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	SSLMode  string
	// If true, creates a unique database for each test
	IsolateTests bool
	// If true, drops databases and tables after tests (default: false - just clear data)
	DropAfterTest bool
	// Optional: Explicit path to migrations directory
	MigrationsPath string
	// Optional: Service name (overrides SERVICE_NAME env var)
	ServiceName string
}

// DefaultTestConfig returns a default test configuration
func DefaultTestConfig() *TestConfig {
	return &TestConfig{
		Host:          config.GetEnvOrDefault("TEST_DB_HOST", "localhost"),
		Port:          config.GetEnvOrDefaultInt("TEST_DB_PORT", 5432),
		User:          config.GetEnvOrDefault("TEST_DB_USER", "postgres"),
		Password:      config.GetEnvOrDefault("TEST_DB_PASSWORD", "postgres"),
		SSLMode:       config.GetEnvOrDefault("TEST_DB_SSL_MODE", "disable"),
		IsolateTests:  config.GetEnvOrDefaultBool("TEST_DB_ISOLATE", true),
		DropAfterTest: config.GetEnvOrDefaultBool("TEST_DB_DROP_AFTER", false),
	}
}

// IsTestingEnabled checks if database testing is enabled via environment variables
func IsTestingEnabled() bool {
	return config.GetEnvOrDefaultBool("ENABLE_DB_TESTS", false)
}

// SetupTestDB creates a test database connection with optional isolation and migrations
func SetupTestDB(cfg *TestConfig) (*TestDB, error) {
	if cfg == nil {
		cfg = DefaultTestConfig()
	}

	var dbName string
	var cleanup func() error

	if cfg.IsolateTests {
		// Create a unique database for this test
		uniqueDBName := fmt.Sprintf("test_torra_%d", time.Now().UnixNano())
		if err := createTestDatabase(cfg, uniqueDBName); err != nil {
			return nil, fmt.Errorf("failed to create test database: %w", err)
		}
		dbName = uniqueDBName
		cleanup = func() error {
			if cfg.DropAfterTest {
				return dropTestDatabase(cfg, uniqueDBName)
			}
			return cleanTestDatabase(cfg, uniqueDBName)
		}
	} else {
		// Use a shared test database
		dbName = config.GetEnvOrDefault("TEST_DB_NAME", "test_torra")
		if err := createTestDatabase(cfg, dbName); err != nil {
			// Database might already exist, which is fine for shared mode
		}
		cleanup = func() error {
			// For shared database, always just clean tables (never drop)
			return cleanTestDatabase(cfg, dbName)
		}
	}

	// Connect to the test database
	dbConfig := &config.DatabaseConfig{
		Host:            cfg.Host,
		Port:            cfg.Port,
		User:            cfg.User,
		Password:        cfg.Password,
		Database:        dbName,
		SSLMode:         cfg.SSLMode,
		MaxOpenConns:    5, // Lower for tests
		MaxIdleConns:    2, // Lower for tests
		ConnMaxLifetime: 5 * time.Minute,
	}

	db, err := New(dbConfig)
	if err != nil {
		cleanup() // Clean up on failure
		return nil, fmt.Errorf("failed to connect to test database: %w", err)
	}

	testDB := &TestDB{
		DB:      db,
		dbName:  dbName,
		cleanup: cleanup,
	}

	// Initialize PostGIS extension if not already present
	if err := initializePostGIS(db); err != nil {
		db.Close()
		cleanup()
		return nil, fmt.Errorf("failed to initialize PostGIS: %w", err)
	}

	// Run migrations for the current service
	if err := runServiceMigrations(db, cfg); err != nil {
		db.Close()
		cleanup()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	testDB.migrated = true
	return testDB, nil
}

// Close closes the test database connection and runs cleanup
func (tdb *TestDB) Close() error {
	if tdb.DB != nil {
		if err := tdb.DB.Close(); err != nil {
			return fmt.Errorf("failed to close database connection: %w", err)
		}
	}

	if tdb.cleanup != nil {
		if err := tdb.cleanup(); err != nil {
			return fmt.Errorf("failed to cleanup test database: %w", err)
		}
	}

	return nil
}

// GetDBName returns the test database name
func (tdb *TestDB) GetDBName() string {
	return tdb.dbName
}

// IsMigrated returns true if migrations were run on this test database
func (tdb *TestDB) IsMigrated() bool {
	return tdb.migrated
}

// TruncateAllTables truncates all tables in the test database (useful for cleanup between tests)
func (tdb *TestDB) TruncateAllTables(ctx context.Context) error {
	// Get all table names
	var tables []string
	query := `
		SELECT tablename
		FROM pg_tables
		WHERE schemaname = 'public'
		AND tablename NOT LIKE 'schema_migrations%'
	`

	err := tdb.SelectContext(ctx, &tables, query)
	if err != nil {
		return fmt.Errorf("failed to get table names: %w", err)
	}

	if len(tables) == 0 {
		return nil // No tables to truncate
	}

	// Truncate all tables
	truncateQuery := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE",
		strings.Join(tables, ", "))

	_, err = tdb.ExecContext(ctx, truncateQuery)
	if err != nil {
		return fmt.Errorf("failed to truncate tables: %w", err)
	}

	return nil
}

// createTestDatabase creates a new test database
func createTestDatabase(cfg *TestConfig, dbName string) error {
	// Connect to postgres database to create the test database
	adminConfig := &config.DatabaseConfig{
		Host:     cfg.Host,
		Port:     cfg.Port,
		User:     cfg.User,
		Password: cfg.Password,
		Database: "postgres", // Connect to default postgres database
		SSLMode:  cfg.SSLMode,
	}

	adminDB, err := New(adminConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to admin database: %w", err)
	}
	defer adminDB.Close()

	// Create the test database
	_, err = adminDB.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
	if err != nil {
		// Check if database already exists
		if !isDBExistsError(err) {
			return fmt.Errorf("failed to create test database %s: %w", dbName, err)
		}
	}

	return nil
}

// dropTestDatabase drops a test database
func dropTestDatabase(cfg *TestConfig, dbName string) error {
	// Connect to postgres database to drop the test database
	adminConfig := &config.DatabaseConfig{
		Host:     cfg.Host,
		Port:     cfg.Port,
		User:     cfg.User,
		Password: cfg.Password,
		Database: "postgres", // Connect to default postgres database
		SSLMode:  cfg.SSLMode,
	}

	adminDB, err := New(adminConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to admin database: %w", err)
	}
	defer adminDB.Close()

	// Terminate connections to the database before dropping
	_, err = adminDB.Exec(fmt.Sprintf(`
		SELECT pg_terminate_backend(pid)
		FROM pg_stat_activity
		WHERE datname = '%s' AND pid <> pg_backend_pid()
	`, dbName))
	if err != nil {
		// Log but don't fail - this is best effort
	}

	// Drop the test database
	_, err = adminDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
	if err != nil {
		return fmt.Errorf("failed to drop test database %s: %w", dbName, err)
	}

	return nil
}

// cleanTestDatabase cleans all tables in a shared test database
func cleanTestDatabase(cfg *TestConfig, dbName string) error {
	// Connect to the test database
	dbConfig := &config.DatabaseConfig{
		Host:     cfg.Host,
		Port:     cfg.Port,
		User:     cfg.User,
		Password: cfg.Password,
		Database: dbName,
		SSLMode:  cfg.SSLMode,
	}

	db, err := New(dbConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to test database for cleanup: %w", err)
	}
	defer db.Close()

	testDB := &TestDB{DB: db, dbName: dbName}
	return testDB.TruncateAllTables(context.Background())
}

// isDBExistsError checks if the error is due to database already existing
func isDBExistsError(err error) bool {
	return strings.Contains(err.Error(), "already exists") ||
		strings.Contains(err.Error(), "duplicate key")
}

// initializePostGIS ensures PostGIS extension is available and initialized
func initializePostGIS(db *DB) error {
	// Create PostGIS extension if not exists
	db.Exec("CREATE EXTENSION IF NOT EXISTS postgis")
	db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
	db.Exec("CREATE EXTENSION IF NOT EXISTS pg_trgm")

	// Ensure SRID 4326 (WGS84) exists in spatial_ref_sys
	// This is standard in PostGIS but might be missing in some test environments
	db.Exec(`
		INSERT INTO spatial_ref_sys (srid, auth_name, auth_srid, proj4text, srtext)
		SELECT 4326, 'EPSG', 4326,
			'+proj=longlat +datum=WGS84 +no_defs',
			'GEOGCS["WGS 84",DATUM["WGS_1984",SPHEROID["WGS 84",6378137,298.257223563,AUTHORITY["EPSG","7030"]],AUTHORITY["EPSG","6326"]],PRIMEM["Greenwich",0,AUTHORITY["EPSG","8901"]],UNIT["degree",0.0174532925199433,AUTHORITY["EPSG","9122"]],AUTHORITY["EPSG","4326"]]'
		WHERE NOT EXISTS (SELECT 1 FROM spatial_ref_sys WHERE srid = 4326)
	`)

	return nil
}

// runServiceMigrations runs the SQL migration files for the current service
func runServiceMigrations(db *DB, cfg *TestConfig) error {
	// Check if explicit migrations path is provided
	if cfg.MigrationsPath != "" {
		return runMigrationsFromPath(db, cfg.MigrationsPath)
	}

	// Get service name from config, then environment, then try to detect
	serviceName := cfg.ServiceName
	if serviceName == "" {
		serviceName = config.GetEnvOrDefault("SERVICE_NAME", "")
	}

	// If SERVICE_NAME not set, try to detect from current working directory
	if serviceName == "" {
		cwd, err := filepath.Abs(".")
		if err == nil {
			// Check if we're in a service directory (e.g., services/core)
			if strings.Contains(cwd, "services") {
				parts := strings.Split(cwd, string(filepath.Separator))
				for i, part := range parts {
					if part == "services" && i+1 < len(parts) {
						serviceName = parts[i+1]
						break
					}
				}
			}
		}
	}

	// If still no service name, check for migrations in current directory
	if serviceName == "" {
		// Try to find migrations in the current directory structure
		possiblePaths := []string{
			"migrations",
			filepath.Join(".", "migrations"),
			filepath.Join("..", "migrations"),
			filepath.Join("..", "..", "migrations"),
		}

		for _, path := range possiblePaths {
			if files, err := filepath.Glob(filepath.Join(path, "*.up.sql")); err == nil && len(files) > 0 {
				return runMigrationsFromPath(db, path)
			}
		}

		// No migrations found, which is OK for some tests
		return nil
	}

	// Build migration paths based on service name
	migrationPaths := []string{
		filepath.Join("services", serviceName, "migrations"),
		filepath.Join("..", "..", "services", serviceName, "migrations"),
		filepath.Join(".", "services", serviceName, "migrations"),
		filepath.Join("..", serviceName, "migrations"),
		filepath.Join(".", "migrations"), // Current directory migrations
	}

	// Try each path
	for _, path := range migrationPaths {
		files, err := filepath.Glob(filepath.Join(path, "*.up.sql"))
		if err == nil && len(files) > 0 {
			return runMigrationsFromPath(db, path)
		}
	}

	// If migrations are not found, just continue - the repository tests might not need all tables
	return nil
}

// runMigrationsFromPath executes all migration files from a given path
func runMigrationsFromPath(db *DB, migrationsPath string) error {
	files, err := filepath.Glob(filepath.Join(migrationsPath, "*.up.sql"))
	if err != nil {
		return fmt.Errorf("failed to list migration files in %s: %w", migrationsPath, err)
	}

	if len(files) == 0 {
		return nil // No migrations to run
	}

	// Sort files to ensure they run in order
	sort.Strings(files)

	// Run each migration file
	for _, file := range files {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		// Execute the migration
		_, err = db.Exec(string(content))
		if err != nil {
			// Ignore certain errors that might occur in test environments
			errStr := err.Error()
			if strings.Contains(errStr, "already exists") ||
				strings.Contains(errStr, "duplicate key") ||
				strings.Contains(errStr, "permission denied") {
				continue
			}
			return fmt.Errorf("failed to execute migration %s: %w", file, err)
		}
	}

	return nil
}

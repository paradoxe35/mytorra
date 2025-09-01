package database

import (
	"context"
	"fmt"
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
	return config.GetEnvOrDefaultBool("ENABLE_DB_TESTS", false) ||
		config.GetEnvOrDefaultBool("CI", false) ||
		config.GetEnvOrDefaultBool("GITHUB_ACTIONS", false)
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

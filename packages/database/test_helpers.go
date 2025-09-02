package database

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/paradoxe35/mytorra/packages/config"
)

// TestDB provides a test database instance
type TestDB struct {
	*GormDB
	testSchema string
	cleanup    func()
}

// NewTestDB creates a new test database instance using PostgreSQL
// It creates an isolated schema for each test to ensure test isolation
func NewTestDB(t *testing.T) *TestDB {
	t.Helper()

	// Get test database configuration from environment or use defaults
	dsn := getEnvOrDefault("TEST_DATABASE_URL",
		fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			getEnvOrDefault("TEST_DB_HOST", "localhost"),
			getEnvAsIntOrDefault("TEST_DB_PORT", 5432),
			getEnvOrDefault("TEST_DB_USER", "postgres"),
			getEnvOrDefault("TEST_DB_PASSWORD", "postgres"),
			getEnvOrDefault("TEST_DB_NAME", "torra_test"),
		))

	// Connect to PostgreSQL
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Create a unique test schema for isolation
	testSchema := fmt.Sprintf("test_%s_%s", t.Name(), uuid.New().String()[:8])
	// Replace any invalid characters in schema name
	testSchema = sanitizeSchemaName(testSchema)

	// Create the test schema
	if err := db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", testSchema)).Error; err != nil {
		t.Fatalf("failed to create test schema %s: %v", testSchema, err)
	}

	// Set search path to use test schema
	if err := db.Exec(fmt.Sprintf("SET search_path TO %s, public", testSchema)).Error; err != nil {
		t.Fatalf("failed to set search path: %v", err)
	}

	// Enable PostGIS extension (install in public schema if not exists, then use from test schema)
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS postgis SCHEMA public").Error; err != nil {
		// Try without specifying schema
		if err := db.Exec("CREATE EXTENSION IF NOT EXISTS postgis").Error; err != nil {
			t.Logf("warning: could not enable PostGIS extension: %v", err)
		}
	}

	// Enable UUID extension
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\" SCHEMA public").Error; err != nil {
		if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
			t.Logf("warning: could not enable uuid-ossp extension: %v", err)
		}
	}

	testDB := &TestDB{
		GormDB:     &GormDB{DB: db},
		testSchema: testSchema,
		cleanup: func() {
			// Drop the test schema and all its contents
			db.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", testSchema))

			// Close the connection
			if sqlDB, err := db.DB(); err == nil {
				sqlDB.Close()
			}
		},
	}

	// Register cleanup
	t.Cleanup(func() {
		testDB.Close()
	})

	return testDB
}

// Migrate runs auto-migration for the given models
func (tdb *TestDB) Migrate(models ...interface{}) error {
	return tdb.DB.AutoMigrate(models...)
}

// TruncateAll truncates all tables (useful for cleaning between tests)
func (tdb *TestDB) TruncateAll(models ...interface{}) error {
	// Disable foreign key checks temporarily
	tdb.DB.Exec("SET session_replication_role = 'replica'")
	defer tdb.DB.Exec("SET session_replication_role = 'origin'")

	for _, model := range models {
		tableName := tdb.DB.Model(model).Statement.Table
		if err := tdb.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", tableName)).Error; err != nil {
			// Fallback to DELETE if TRUNCATE fails
			if err := tdb.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(model).Error; err != nil {
				return fmt.Errorf("failed to truncate table %s: %w", tableName, err)
			}
		}
	}
	return nil
}

// Close closes the test database and performs cleanup
func (tdb *TestDB) Close() {
	if tdb.cleanup != nil {
		tdb.cleanup()
	}
}

// SeedData is a helper to seed test data
func (tdb *TestDB) SeedData(data interface{}) error {
	return tdb.DB.Create(data).Error
}

// AssertCount asserts the count of records in a table
func (tdb *TestDB) AssertCount(t *testing.T, model interface{}, expected int64, message string) {
	t.Helper()

	var count int64
	if err := tdb.DB.Model(model).Count(&count).Error; err != nil {
		t.Fatalf("failed to count records: %v", err)
	}

	if count != expected {
		t.Errorf("%s: expected %d records, got %d", message, expected, count)
	}
}

// AssertExists asserts that a record exists
func (tdb *TestDB) AssertExists(t *testing.T, model interface{}, conditions ...interface{}) {
	t.Helper()

	err := tdb.DB.First(model, conditions...).Error
	if err != nil {
		t.Errorf("record does not exist: %v", err)
	}
}

// AssertNotExists asserts that a record does not exist
func (tdb *TestDB) AssertNotExists(t *testing.T, model interface{}, conditions ...interface{}) {
	t.Helper()

	err := tdb.DB.First(model, conditions...).Error
	if err == nil {
		t.Error("record exists but should not")
	}
}

// AssertSoftDeleted asserts that a record is soft deleted
func (tdb *TestDB) AssertSoftDeleted(t *testing.T, model interface{}, id uuid.UUID) {
	t.Helper()

	// Check with unscoped to see if record exists
	err := tdb.DB.Unscoped().First(model, "id = ?", id).Error
	if err != nil {
		t.Errorf("record does not exist at all: %v", err)
		return
	}

	// Check that it's not found without unscoped (soft deleted)
	err = tdb.DB.First(model, "id = ?", id).Error
	if err == nil {
		t.Error("record is not soft deleted")
	}
}

// ExecSQL executes raw SQL (useful for complex test setup)
func (tdb *TestDB) ExecSQL(sql string, values ...interface{}) error {
	return tdb.DB.Exec(sql, values...).Error
}

// InTransaction runs a function in a transaction and rolls it back after
// Useful for testing transaction behavior
func (tdb *TestDB) InTransaction(t *testing.T, fn func(*gorm.DB)) {
	t.Helper()

	tx := tdb.DB.Begin()
	defer tx.Rollback()

	fn(tx)
}

// NewTestGormDB creates a test GORM database from config (for integration tests)
func NewTestGormDB(t *testing.T) *GormDB {
	t.Helper()

	cfg := &config.DatabaseConfig{
		Host:            getEnvOrDefault("TEST_DB_HOST", "localhost"),
		Port:            getEnvAsIntOrDefault("TEST_DB_PORT", 5432),
		Database:        getEnvOrDefault("TEST_DB_NAME", "torra_test"),
		User:            getEnvOrDefault("TEST_DB_USER", "postgres"),
		Password:        getEnvOrDefault("TEST_DB_PASSWORD", "postgres"),
		SSLMode:         "disable",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 30 * 60, // 30 minutes
	}

	db, err := NewGormDB(cfg)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// Helper functions
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		fmt.Sscanf(value, "%d", &intValue)
		return intValue
	}
	return defaultValue
}

// sanitizeSchemaName removes or replaces invalid characters in schema names
func sanitizeSchemaName(name string) string {
	// Replace any non-alphanumeric characters (except underscore) with underscore
	result := ""
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			result += string(r)
		} else {
			result += "_"
		}
	}

	// Ensure it doesn't start with a number
	if len(result) > 0 && result[0] >= '0' && result[0] <= '9' {
		result = "t_" + result
	}

	// Truncate if too long (PostgreSQL identifier limit is 63 chars)
	if len(result) > 63 {
		result = result[:63]
	}

	return result
}

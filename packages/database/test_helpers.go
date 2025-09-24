package database

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"github.com/paradoxe35/mytorra/packages/config"
)

// TestDB provides a test database instance
type TestDB struct {
	*DB
	testSchema string
	cleanup    func()
}

func DBTesEnabled() bool {
	return os.Getenv("ENABLE_DB_TESTS") == "true"
}

// NewTestDB creates a new test database instance using PostgreSQL
// It creates an isolated schema for each test to ensure test isolation
func NewTestDB(t *testing.T) *TestDB {
	t.Helper()
	ctx := context.Background()

	if !DBTesEnabled() {
		t.Skip("test database is not enabled")
	}

	// Get test database configuration from environment or use defaults
	cfg := &config.DatabaseConfig{
		Host:            getEnvOrDefault("TEST_DB_HOST", "localhost"),
		Port:            getEnvAsIntOrDefault("TEST_DB_PORT", 5432),
		Database:        getEnvOrDefault("TEST_DB_NAME", "mytorra_test"),
		User:            getEnvOrDefault("TEST_DB_USER", "postgres"),
		Password:        getEnvOrDefault("TEST_DB_PASSWORD", "postgres"),
		SSLMode:         "disable",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 30 * time.Minute,
	}

	// Connect to PostgreSQL using Bun
	db, err := NewDB(cfg)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Create a unique test schema for isolation
	testSchema := fmt.Sprintf("test_%s_%s", t.Name(), uuid.New().String()[:8])
	// Replace any invalid characters in schema name
	testSchema = sanitizeSchemaName(testSchema)

	// Create the test schema
	if _, err := db.DB.ExecContext(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", testSchema)); err != nil {
		t.Fatalf("failed to create test schema %s: %v", testSchema, err)
	}

	// Set search path to use test schema
	if _, err := db.DB.ExecContext(ctx, fmt.Sprintf("SET search_path TO %s, public", testSchema)); err != nil {
		t.Fatalf("failed to set search path: %v", err)
	}

	// Enable PostGIS extension (install in public schema if not exists, then use from test schema)
	if _, err := db.DB.ExecContext(ctx, "CREATE EXTENSION IF NOT EXISTS postgis SCHEMA public"); err != nil {
		// Try without specifying schema
		if _, err := db.DB.ExecContext(ctx, "CREATE EXTENSION IF NOT EXISTS postgis"); err != nil {
			t.Logf("warning: could not enable PostGIS extension: %v", err)
		}
	}

	// Enable UUID extension
	if _, err := db.DB.ExecContext(ctx, "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\" SCHEMA public"); err != nil {
		if _, err := db.DB.ExecContext(ctx, "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\""); err != nil {
			t.Logf("warning: could not enable uuid-ossp extension: %v", err)
		}
	}

	testDB := &TestDB{
		DB:         db,
		testSchema: testSchema,
		cleanup: func() {
			// Drop the test schema and all its contents
			db.DB.ExecContext(context.Background(), fmt.Sprintf("DROP SCHEMA IF NOT EXISTS %s CASCADE", testSchema))
			// Close the connection
			db.Close()
		},
	}

	// Register cleanup
	t.Cleanup(func() {
		testDB.Close()
	})

	return testDB
}

// CreateTables creates tables for the given models
func (tdb *TestDB) CreateTables(ctx context.Context, models ...interface{}) error {
	for _, model := range models {
		if _, err := tdb.DB.DB.NewCreateTable().
			Model(model).
			IfNotExists().
			Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}

// TruncateAll truncates all tables (useful for cleaning between tests)
func (tdb *TestDB) TruncateAll(ctx context.Context, models ...interface{}) error {
	// Disable foreign key checks temporarily
	if _, err := tdb.DB.DB.ExecContext(ctx, "SET session_replication_role = 'replica'"); err != nil {
		return err
	}
	defer tdb.DB.DB.ExecContext(ctx, "SET session_replication_role = 'origin'")

	for _, model := range models {
		tableName := tdb.DB.DB.NewSelect().Model(model).GetTableName()
		if _, err := tdb.DB.DB.ExecContext(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", tableName)); err != nil {
			// Fallback to DELETE if TRUNCATE fails
			if _, err := tdb.DB.DB.NewDelete().Model(model).Where("1=1").Exec(ctx); err != nil {
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
func (tdb *TestDB) SeedData(ctx context.Context, data interface{}) error {
	_, err := tdb.DB.DB.NewInsert().Model(data).Exec(ctx)
	return err
}

// AssertCount asserts the count of records in a table
func (tdb *TestDB) AssertCount(t *testing.T, ctx context.Context, model interface{}, expected int, message string) {
	t.Helper()

	count, err := tdb.DB.DB.NewSelect().Model(model).Count(ctx)
	if err != nil {
		t.Fatalf("failed to count records: %v", err)
	}

	if count != expected {
		t.Errorf("%s: expected %d records, got %d", message, expected, count)
	}
}

// AssertExists asserts that a record exists
func (tdb *TestDB) AssertExists(t *testing.T, ctx context.Context, model interface{}, conditions ...interface{}) {
	t.Helper()

	query := tdb.DB.DB.NewSelect().Model(model)
	for i := 0; i < len(conditions); i += 2 {
		if i+1 < len(conditions) {
			query = query.Where(fmt.Sprintf("%s = ?", conditions[i]), conditions[i+1])
		}
	}

	err := query.Scan(ctx)
	if err != nil {
		t.Errorf("record does not exist: %v", err)
	}
}

// AssertNotExists asserts that a record does not exist
func (tdb *TestDB) AssertNotExists(t *testing.T, ctx context.Context, model interface{}, conditions ...interface{}) {
	t.Helper()

	query := tdb.DB.DB.NewSelect().Model(model)
	for i := 0; i < len(conditions); i += 2 {
		if i+1 < len(conditions) {
			query = query.Where(fmt.Sprintf("%s = ?", conditions[i]), conditions[i+1])
		}
	}

	err := query.Scan(ctx)
	if err == nil {
		t.Error("record exists but should not")
	}
}

// AssertSoftDeleted asserts that a record is soft deleted
func (tdb *TestDB) AssertSoftDeleted(t *testing.T, ctx context.Context, model interface{}, id uuid.UUID) {
	t.Helper()

	// Check with deleted records to see if record exists
	err := tdb.DB.DB.NewSelect().
		Model(model).
		Where("id = ?", id).
		WhereDeleted().
		Scan(ctx)
	if err != nil {
		t.Errorf("record does not exist at all: %v", err)
		return
	}

	// Check that it's not found without WhereDeleted (soft deleted)
	err = tdb.DB.DB.NewSelect().
		Model(model).
		Where("id = ?", id).
		Scan(ctx)

	if err == nil {
		t.Error("record is not soft deleted")
	}
}

// ExecSQL executes raw SQL (useful for complex test setup)
func (tdb *TestDB) ExecSQL(ctx context.Context, sql string, values ...interface{}) error {
	_, err := tdb.DB.DB.ExecContext(ctx, sql, values...)
	return err
}

// InTransaction runs a function in a transaction and rolls it back after
// Useful for testing transaction behavior
func (tdb *TestDB) InTransaction(t *testing.T, ctx context.Context, fn func(ctx context.Context, tx bun.Tx)) {
	t.Helper()

	tx, err := tdb.DB.DB.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	fn(ctx, tx)
}

// NewTestBunDB creates a test Bun database from config (for integration tests)
func NewTestBunDB(t *testing.T) *DB {
	t.Helper()

	cfg := &config.DatabaseConfig{
		Host:            getEnvOrDefault("TEST_DB_HOST", "localhost"),
		Port:            getEnvAsIntOrDefault("TEST_DB_PORT", 5432),
		Database:        getEnvOrDefault("TEST_DB_NAME", "mytorra_test"),
		User:            getEnvOrDefault("TEST_DB_USER", "postgres"),
		Password:        getEnvOrDefault("TEST_DB_PASSWORD", "postgres"),
		SSLMode:         "disable",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 30 * time.Minute,
	}

	db, err := NewDB(cfg)
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

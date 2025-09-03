package database

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"

	"github.com/paradoxe35/mytorra/packages/config"
	pkgLogger "github.com/paradoxe35/mytorra/packages/logger"
)

// DB wraps the Bun database connection
type DB struct {
	*bun.DB
	pool *pgxpool.Pool
}

// New creates a new Bun database connection with pgx
func New(cfg *config.DatabaseConfig) (*DB, error) {
	return NewDB(cfg)
}

// NewDB creates a new Bun database connection with pgx pool
func NewDB(cfg *config.DatabaseConfig) (*DB, error) {
	ctx := context.Background()

	// Parse PostgreSQL configuration
	pgxConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Configure pgx for better performance
	pgxConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	// Configure connection pool
	maxOpenConns := cfg.MaxOpenConns
	if maxOpenConns == 0 {
		maxOpenConns = 4 * runtime.GOMAXPROCS(0)
	}

	pgxConfig.MaxConns = int32(maxOpenConns)
	pgxConfig.MinConns = int32(cfg.MaxIdleConns)
	if cfg.ConnMaxLifetime > 0 {
		pgxConfig.MaxConnLifetime = cfg.ConnMaxLifetime
	}
	pgxConfig.MaxConnIdleTime = 30 * time.Minute

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Create sql.DB from pgx pool
	sqldb := stdlib.OpenDBFromPool(pool)

	// Create Bun DB
	bunDB := bun.NewDB(sqldb, pgdialect.New())

	// Add debug logging in development
	if cfg.Database == "development" || cfg.Database == "test" {
		bunDB.AddQueryHook(bundebug.NewQueryHook(
			bundebug.WithVerbose(true),
			bundebug.FromEnv("BUNDEBUG"),
		))
	}

	db := &DB{
		DB:   bunDB,
		pool: pool,
	}

	// Enable required extensions
	extensions := []string{"uuid-ossp", "pg_trgm", "postgis"}
	log := pkgLogger.New("database", "production")
	for _, ext := range extensions {
		if _, err := bunDB.ExecContext(ctx, fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS \"%s\"", ext)); err != nil {
			log.WithError(err).Warn("Failed to enable %s extension", ext)
		}
	}

	return db, nil
}

// Transaction executes a function within a database transaction
func (db *DB) Transaction(ctx context.Context, fn func(ctx context.Context, tx bun.Tx) error) error {
	return db.DB.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		return fn(ctx, tx)
	})
}

// WithContext returns a new DB instance with the provided context
func (db *DB) WithContext(ctx context.Context) *bun.DB {
	return db.DB
}

// Close closes the database connection
func (db *DB) Close() error {
	if err := db.DB.Close(); err != nil {
		return err
	}

	if db.pool != nil {
		db.pool.Close()
	}

	return nil
}

// Paginate creates pagination for Bun queries
type Paginate struct {
	Page    int
	PerPage int
}

func (p *Paginate) Apply(q *bun.SelectQuery) *bun.SelectQuery {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PerPage <= 0 {
		p.PerPage = 20
	}
	if p.PerPage > 100 {
		p.PerPage = 100
	}

	offset := (p.Page - 1) * p.PerPage
	return q.Offset(offset).Limit(p.PerPage)
}

func NewPaginate(page, perPage int) *Paginate {
	return &Paginate{
		Page:    page,
		PerPage: perPage,
	}
}

// Pagination holds pagination parameters
type Pagination struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

// PaginationResult holds pagination metadata
type PaginationResult struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// NewPaginationResult creates a new pagination result
func NewPaginationResult(page, perPage int, total int64) *PaginationResult {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 20
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return &PaginationResult{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}
}

// IsUniqueViolation checks if the error is a unique constraint violation
func IsUniqueViolation(err error) bool {
	if err == nil {
		return false
	}

	if pgErr, ok := err.(pgdriver.Error); ok {
		return pgErr.IntegrityViolation()
	}

	errStr := err.Error()
	return contains(errStr, pgerrcode.UniqueViolation) ||
		contains(errStr, "duplicate key") ||
		contains(errStr, "unique constraint")
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

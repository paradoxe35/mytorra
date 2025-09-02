package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/paradoxe35/torra/packages/config"
)

// GormDB wraps the GORM database connection
type GormDB struct {
	*gorm.DB
}

// NewGormDB creates a new GORM database connection
func NewGormDB(cfg *config.DatabaseConfig) (*GormDB, error) {
	// Configure GORM
	gormConfig := &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		QueryFields: true,
		Logger:      logger.Default.LogMode(logger.Info),
	}

	// Open database connection
	db, err := gorm.Open(postgres.Open(cfg.DSN()), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL database to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Enable PostGIS extension if not already enabled
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS postgis").Error; err != nil {
		return nil, fmt.Errorf("failed to enable PostGIS extension: %w", err)
	}

	return &GormDB{db}, nil
}

// Transaction executes a function within a database transaction
func (db *GormDB) Transaction(fn func(*gorm.DB) error, opts ...*sql.TxOptions) error {
	return db.DB.Transaction(fn, opts...)
}

// WithContext returns a new DB instance with the provided context
func (db *GormDB) WithContext(ctx context.Context) *gorm.DB {
	return db.DB.WithContext(ctx)
}

// Close closes the database connection
func (db *GormDB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Paginate creates a paginated scope for GORM queries
func Paginate(page, perPage int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}
		if perPage <= 0 {
			perPage = 20
		}
		if perPage > 100 {
			perPage = 100
		}

		offset := (page - 1) * perPage
		return db.Offset(offset).Limit(perPage)
	}
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

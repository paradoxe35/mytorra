package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/paradoxe35/torra/packages/config"
)

type DB struct {
	*sqlx.DB
}

func New(cfg *config.DatabaseConfig) (*DB, error) {
	db, err := sqlx.Connect("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{db}, nil
}

func (db *DB) InTransaction(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(tx)
	return err
}

func (db *DB) ExistsContext(ctx context.Context, query string, args ...interface{}) (bool, error) {
	var exists bool
	query = fmt.Sprintf("SELECT EXISTS(%s)", query)
	err := db.GetContext(ctx, &exists, query, args...)
	return exists, err
}

type Pagination struct {
	Page    int
	PerPage int
}

func (p *Pagination) Offset() int {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PerPage <= 0 {
		p.PerPage = 20
	}
	return (p.Page - 1) * p.PerPage
}

func (p *Pagination) Limit() int {
	if p.PerPage <= 0 {
		return 20
	}
	if p.PerPage > 100 {
		return 100
	}
	return p.PerPage
}

type QueryBuilder struct {
	query string
	args  []interface{}
}

func NewQueryBuilder(base string) *QueryBuilder {
	return &QueryBuilder{
		query: base,
		args:  make([]interface{}, 0),
	}
}

func (qb *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder {
	qb.query += " WHERE " + condition
	qb.args = append(qb.args, args...)
	return qb
}

func (qb *QueryBuilder) And(condition string, args ...interface{}) *QueryBuilder {
	qb.query += " AND " + condition
	qb.args = append(qb.args, args...)
	return qb
}

func (qb *QueryBuilder) OrderBy(order string) *QueryBuilder {
	qb.query += " ORDER BY " + order
	return qb
}

func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.query += fmt.Sprintf(" LIMIT %d", limit)
	return qb
}

func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.query += fmt.Sprintf(" OFFSET %d", offset)
	return qb
}

func (qb *QueryBuilder) Build() (string, []interface{}) {
	return qb.query, qb.args
}

func IsNoRowsError(err error) bool {
	return err == sql.ErrNoRows
}

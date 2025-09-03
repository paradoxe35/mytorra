package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// BaseModel is a base model for all entities with UUID primary key and soft delete support
type BaseModel struct {
	ID        uuid.UUID    `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time    `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time    `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
	DeletedAt sql.NullTime `bun:"deleted_at,soft_delete,nullzero" json:"-"`
}

var _ bun.BeforeAppendModelHook = (*BaseModel)(nil)

// BeforeAppendModel hook to set UUID and timestamps
func (m *BaseModel) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		if m.ID == uuid.Nil {
			m.ID = uuid.New()
		}
		m.CreatedAt = time.Now().UTC()
		m.UpdatedAt = time.Now().UTC()
	case *bun.UpdateQuery:
		m.UpdatedAt = time.Now().UTC()
	}
	return nil
}

// UUIDModel is an alias for BaseModel for backward compatibility
type UUIDModel = BaseModel

// TimestampModel provides only timestamps without ID (for models that have composite keys)
type TimestampModel struct {
	CreatedAt time.Time    `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time    `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
	DeletedAt sql.NullTime `bun:"deleted_at,soft_delete,nullzero" json:"-"`
}

var _ bun.BeforeAppendModelHook = (*TimestampModel)(nil)

// BeforeAppendModel hook to set timestamps
func (m *TimestampModel) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		m.CreatedAt = time.Now().UTC()
		m.UpdatedAt = time.Now().UTC()
	case *bun.UpdateQuery:
		m.UpdatedAt = time.Now().UTC()
	}
	return nil
}

// SoftDelete interface for models that support soft delete
type SoftDelete interface {
	IsSoftDeleted() bool
}

// IsSoftDeleted checks if the model has been soft deleted
func (m *BaseModel) IsSoftDeleted() bool {
	return m.DeletedAt.Valid
}

// IsSoftDeleted checks if the model has been soft deleted
func (m *TimestampModel) IsSoftDeleted() bool {
	return m.DeletedAt.Valid
}

package database

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel is a base model for all entities with UUID primary key and soft delete support
// This replaces the default gorm.Model which uses uint ID
type BaseModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time      `gorm:"not null;default:now()"`
	UpdatedAt time.Time      `gorm:"not null;default:now()"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// BeforeCreate hook to set UUID if not provided
func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// UUIDModel is an alias for BaseModel for backward compatibility
type UUIDModel = BaseModel

// TimestampModel provides only timestamps without ID (for models that have composite keys)
type TimestampModel struct {
	CreatedAt time.Time      `gorm:"not null;default:now()"`
	UpdatedAt time.Time      `gorm:"not null;default:now()"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// SoftDelete interface for models that support soft delete
type SoftDelete interface {
	IsSoftDeleted() bool
}

// IsSoftDeleted checks if the model has been soft deleted
func (b *BaseModel) IsSoftDeleted() bool {
	return b.DeletedAt.Valid
}

// IsSoftDeleted checks if the model has been soft deleted
func (t *TimestampModel) IsSoftDeleted() bool {
	return t.DeletedAt.Valid
}

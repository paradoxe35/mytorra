package models

import (
	"context"
	"time"

	"github.com/uptrace/bun"

	"github.com/paradoxe35/mytorra/packages/database"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/entities"
)

// JWTKeyModel represents the Bun model for the jwt_keys table
type JWTKeyModel struct {
	database.BaseModel `bun:"table:jwt_keys,alias:jk"`

	KeyID      string     `bun:"key_id,notnull,unique"`
	PublicKey  string     `bun:"public_key,notnull"`
	PrivateKey string     `bun:"private_key,notnull"`
	Algorithm  string     `bun:"algorithm,notnull,default:'RS256'"`
	IsActive   bool       `bun:"is_active,default:true"`
	RotatedAt  *time.Time `bun:"rotated_at"`
	ExpiresAt  *time.Time `bun:"expires_at"`
}

// BeforeAppendModel is called before inserting or updating the model
func (j *JWTKeyModel) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	// Add any custom logic here if needed
	return nil
}

// ToEntity converts the GORM model to domain entity
func (j *JWTKeyModel) ToEntity() *entities.JWTKey {
	entity := &entities.JWTKey{
		ID:         j.ID,
		KeyID:      j.KeyID,
		PublicKey:  j.PublicKey,
		PrivateKey: j.PrivateKey,
		Algorithm:  j.Algorithm,
		IsActive:   j.IsActive,
		CreatedAt:  j.CreatedAt,
		RotatedAt:  j.RotatedAt,
	}

	if j.ExpiresAt != nil {
		entity.ExpiresAt = *j.ExpiresAt
	} else {
		// Default expiry if not set
		entity.ExpiresAt = j.CreatedAt.AddDate(1, 0, 0) // 1 year
	}

	return entity
}

// JWTKeyFromEntity creates a GORM model from domain entity
func JWTKeyFromEntity(entity *entities.JWTKey) *JWTKeyModel {
	return &JWTKeyModel{
		BaseModel: database.BaseModel{
			ID:        entity.ID,
			CreatedAt: entity.CreatedAt,
			UpdatedAt: time.Now(),
		},
		KeyID:      entity.KeyID,
		PublicKey:  entity.PublicKey,
		PrivateKey: entity.PrivateKey,
		Algorithm:  entity.Algorithm,
		IsActive:   entity.IsActive,
		RotatedAt:  entity.RotatedAt,
		ExpiresAt:  &entity.ExpiresAt,
	}
}

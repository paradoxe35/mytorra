package models

import (
	"time"

	"github.com/paradoxe35/torra/packages/database"
	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
)

// JWTKeyModel represents the GORM model for the jwt_keys table
type JWTKeyModel struct {
	database.BaseModel
	KeyID      string `gorm:"type:varchar(255);uniqueIndex;not null"`
	PublicKey  string `gorm:"type:text;not null"`
	PrivateKey string `gorm:"type:text;not null"`
	Algorithm  string `gorm:"type:varchar(10);not null;default:'RS256'"`
	IsActive   bool   `gorm:"type:boolean;default:true;index"`
	RotatedAt  *time.Time
	ExpiresAt  *time.Time
}

// TableName specifies the table name for the JWTKeyModel
func (JWTKeyModel) TableName() string {
	return "jwt_keys"
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

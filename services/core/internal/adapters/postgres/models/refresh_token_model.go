package models

import (
	"time"

	"github.com/google/uuid"

	"github.com/paradoxe35/mytorra/packages/database"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/entities"
)

// RefreshTokenModel represents the GORM model for the refresh_tokens table
type RefreshTokenModel struct {
	database.BaseModel
	UserID     uuid.UUID  `gorm:"type:uuid;not null;index"`
	User       *UserModel `gorm:"foreignKey:UserID"`
	TokenHash  string     `gorm:"type:varchar(512);uniqueIndex;not null"`
	DeviceID   string     `gorm:"type:varchar(255)"`
	DeviceName string     `gorm:"type:varchar(255)"`
	IPAddress  string     `gorm:"type:varchar(45)"`
	UserAgent  string     `gorm:"type:text"`
	ExpiresAt  time.Time  `gorm:"not null"`
	LastUsedAt *time.Time
	RevokedAt  *time.Time
}

// TableName specifies the table name for the RefreshTokenModel
func (RefreshTokenModel) TableName() string {
	return "refresh_tokens"
}

// ToEntity converts the GORM model to domain entity
func (r *RefreshTokenModel) ToEntity() *entities.RefreshToken {
	return &entities.RefreshToken{
		ID:         r.ID,
		UserID:     r.UserID,
		TokenHash:  r.TokenHash,
		DeviceID:   r.DeviceID,
		DeviceName: r.DeviceName,
		IPAddress:  r.IPAddress,
		UserAgent:  r.UserAgent,
		ExpiresAt:  r.ExpiresAt,
		LastUsedAt: r.LastUsedAt,
		RevokedAt:  r.RevokedAt,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
}

// RefreshTokenFromEntity creates a GORM model from domain entity
func RefreshTokenFromEntity(entity *entities.RefreshToken) *RefreshTokenModel {
	return &RefreshTokenModel{
		BaseModel: database.BaseModel{
			ID:        entity.ID,
			CreatedAt: entity.CreatedAt,
			UpdatedAt: entity.UpdatedAt,
		},
		UserID:     entity.UserID,
		TokenHash:  entity.TokenHash,
		DeviceID:   entity.DeviceID,
		DeviceName: entity.DeviceName,
		IPAddress:  entity.IPAddress,
		UserAgent:  entity.UserAgent,
		ExpiresAt:  entity.ExpiresAt,
		LastUsedAt: entity.LastUsedAt,
		RevokedAt:  entity.RevokedAt,
	}
}

package models

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"github.com/paradoxe35/mytorra/packages/database"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/entities"
)

// RefreshTokenModel represents the Bun model for the refresh_tokens table
type RefreshTokenModel struct {
	database.BaseModel `bun:"table:refresh_tokens,alias:rt"`

	UserID     uuid.UUID  `bun:"user_id,type:uuid,notnull"`
	User       *UserModel `bun:"rel:belongs-to,join:user_id=id"`
	TokenHash  string     `bun:"token_hash,notnull,unique"`
	DeviceID   string     `bun:"device_id"`
	DeviceName string     `bun:"device_name"`
	IPAddress  string     `bun:"ip_address"`
	UserAgent  string     `bun:"user_agent"`
	ExpiresAt  time.Time  `bun:"expires_at,notnull"`
	LastUsedAt *time.Time `bun:"last_used_at"`
	RevokedAt  *time.Time `bun:"revoked_at"`
}

// BeforeAppendModel is called before inserting or updating the model
func (r *RefreshTokenModel) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	// Add any custom logic here if needed
	return nil
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

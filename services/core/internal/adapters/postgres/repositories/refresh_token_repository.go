package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/paradoxe35/torra/packages/database"
	"github.com/paradoxe35/torra/packages/errors"
	"github.com/paradoxe35/torra/services/core/internal/adapters/postgres/models"
	"github.com/paradoxe35/torra/services/core/internal/application/ports"
	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
)

type refreshTokenRepository struct {
	db *database.GormDB
}

// NewRefreshTokenRepository creates a new GORM-based refresh token repository
func NewRefreshTokenRepository(db *database.GormDB) ports.RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(ctx context.Context, token *entities.RefreshToken) error {
	model := models.RefreshTokenFromEntity(token)

	err := r.db.WithContext(ctx).Create(model).Error
	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	// Update the entity with any generated values
	*token = *model.ToEntity()
	return nil
}

func (r *refreshTokenRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*entities.RefreshToken, error) {
	var tokenModel models.RefreshTokenModel

	err := r.db.WithContext(ctx).
		Where("token_hash = ?", tokenHash).
		First(&tokenModel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find refresh token: %w", err)
	}

	return tokenModel.ToEntity(), nil
}

func (r *refreshTokenRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.RefreshToken, error) {
	var tokenModel models.RefreshTokenModel

	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&tokenModel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find refresh token by ID: %w", err)
	}

	return tokenModel.ToEntity(), nil
}

func (r *refreshTokenRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.RefreshToken, error) {
	var tokenModels []models.RefreshTokenModel

	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&tokenModels).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find refresh tokens by user: %w", err)
	}

	tokens := make([]*entities.RefreshToken, len(tokenModels))
	for i, model := range tokenModels {
		tokens[i] = model.ToEntity()
	}

	return tokens, nil
}

func (r *refreshTokenRepository) Update(ctx context.Context, token *entities.RefreshToken) error {
	model := models.RefreshTokenFromEntity(token)

	result := r.db.WithContext(ctx).
		Model(&models.RefreshTokenModel{}).
		Where("id = ?", token.ID).
		Updates(model)

	if result.Error != nil {
		return fmt.Errorf("failed to update refresh token: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}

func (r *refreshTokenRepository) UpdateLastUsed(ctx context.Context, tokenID uuid.UUID) error {
	now := time.Now()

	result := r.db.WithContext(ctx).
		Model(&models.RefreshTokenModel{}).
		Where("id = ?", tokenID).
		Update("last_used_at", now)

	if result.Error != nil {
		return fmt.Errorf("failed to update last used: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}

func (r *refreshTokenRepository) RevokeByID(ctx context.Context, id uuid.UUID) error {
	now := time.Now()

	result := r.db.WithContext(ctx).
		Model(&models.RefreshTokenModel{}).
		Where("id = ?", id).
		Update("revoked_at", now)

	if result.Error != nil {
		return fmt.Errorf("failed to revoke token: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}

func (r *refreshTokenRepository) Revoke(ctx context.Context, tokenID uuid.UUID) error {
	return r.RevokeByID(ctx, tokenID)
}

func (r *refreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()

	err := r.db.WithContext(ctx).
		Model(&models.RefreshTokenModel{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now).Error
	if err != nil {
		return fmt.Errorf("failed to revoke all user tokens: %w", err)
	}

	return nil
}

func (r *refreshTokenRepository) RevokeByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.RevokeAllForUser(ctx, userID)
}

func (r *refreshTokenRepository) RevokeByDeviceID(ctx context.Context, userID uuid.UUID, deviceID string) error {
	now := time.Now()

	err := r.db.WithContext(ctx).
		Model(&models.RefreshTokenModel{}).
		Where("user_id = ? AND device_id = ? AND revoked_at IS NULL", userID, deviceID).
		Update("revoked_at", now).Error
	if err != nil {
		return fmt.Errorf("failed to revoke device tokens: %w", err)
	}

	return nil
}

func (r *refreshTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	now := time.Now()

	result := r.db.WithContext(ctx).
		Where("expires_at < ? OR revoked_at IS NOT NULL", now).
		Delete(&models.RefreshTokenModel{})

	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete expired tokens: %w", result.Error)
	}

	return result.RowsAffected, nil
}

func (r *refreshTokenRepository) CountActiveByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	now := time.Now()

	err := r.db.WithContext(ctx).
		Model(&models.RefreshTokenModel{}).
		Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, now).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count active tokens: %w", err)
	}

	return count, nil
}

func (r *refreshTokenRepository) FindActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.RefreshToken, error) {
	var tokenModels []models.RefreshTokenModel

	now := time.Now()
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, now).
		Order("created_at DESC").
		Find(&tokenModels).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find active refresh tokens: %w", err)
	}

	tokens := make([]*entities.RefreshToken, len(tokenModels))
	for i, model := range tokenModels {
		tokens[i] = model.ToEntity()
	}

	return tokens, nil
}

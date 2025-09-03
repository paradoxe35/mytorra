package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/paradoxe35/mytorra/packages/database"
	"github.com/paradoxe35/mytorra/packages/errors"
	"github.com/paradoxe35/mytorra/services/core/internal/adapters/postgres/models"
	"github.com/paradoxe35/mytorra/services/core/internal/application/ports"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/entities"
)

type refreshTokenRepository struct {
	db *database.DB
}

// NewRefreshTokenRepository creates a new Bun-based refresh token repository
func NewRefreshTokenRepository(db *database.DB) ports.RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(ctx context.Context, token *entities.RefreshToken) error {
	model := models.RefreshTokenFromEntity(token)

	_, err := r.db.NewInsert().
		Model(model).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	// Update the entity with any generated values
	*token = *model.ToEntity()
	return nil
}

func (r *refreshTokenRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*entities.RefreshToken, error) {
	var tokenModel models.RefreshTokenModel

	err := r.db.NewSelect().
		Model(&tokenModel).
		Where("token_hash = ?", tokenHash).
		Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find refresh token: %w", err)
	}

	return tokenModel.ToEntity(), nil
}

func (r *refreshTokenRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.RefreshToken, error) {
	var tokenModel models.RefreshTokenModel

	err := r.db.NewSelect().
		Model(&tokenModel).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find refresh token by ID: %w", err)
	}

	return tokenModel.ToEntity(), nil
}

func (r *refreshTokenRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.RefreshToken, error) {
	var tokenModels []models.RefreshTokenModel

	err := r.db.NewSelect().
		Model(&tokenModels).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Scan(ctx)
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

	result, err := r.db.NewUpdate().
		Model(model).
		Where("id = ?", token.ID).
		OmitZero().
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update refresh token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}

func (r *refreshTokenRepository) UpdateLastUsed(ctx context.Context, tokenID uuid.UUID) error {
	now := time.Now()

	result, err := r.db.NewUpdate().
		Model(&models.RefreshTokenModel{}).
		Where("id = ?", tokenID).
		Set("last_used_at = ?", now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update last used: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}

func (r *refreshTokenRepository) RevokeByID(ctx context.Context, id uuid.UUID) error {
	now := time.Now()

	result, err := r.db.NewUpdate().
		Model(&models.RefreshTokenModel{}).
		Where("id = ?", id).
		Set("revoked_at = ?", now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}

func (r *refreshTokenRepository) Revoke(ctx context.Context, tokenID uuid.UUID) error {
	return r.RevokeByID(ctx, tokenID)
}

func (r *refreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()

	_, err := r.db.NewUpdate().
		Model(&models.RefreshTokenModel{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Set("revoked_at = ?", now).
		Exec(ctx)
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

	_, err := r.db.NewUpdate().
		Model(&models.RefreshTokenModel{}).
		Where("user_id = ? AND device_id = ? AND revoked_at IS NULL", userID, deviceID).
		Set("revoked_at = ?", now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to revoke device tokens: %w", err)
	}

	return nil
}

func (r *refreshTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	now := time.Now()

	result, err := r.db.NewDelete().
		Model(&models.RefreshTokenModel{}).
		Where("expires_at < ? OR revoked_at IS NOT NULL", now).
		Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired tokens: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

func (r *refreshTokenRepository) CountActiveByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	now := time.Now()

	count, err := r.db.NewSelect().
		Model(&models.RefreshTokenModel{}).
		Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, now).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count active tokens: %w", err)
	}

	return int64(count), nil
}

func (r *refreshTokenRepository) FindActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.RefreshToken, error) {
	var tokenModels []models.RefreshTokenModel

	now := time.Now()
	err := r.db.NewSelect().
		Model(&tokenModels).
		Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, now).
		Order("created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find active refresh tokens: %w", err)
	}

	tokens := make([]*entities.RefreshToken, len(tokenModels))
	for i, model := range tokenModels {
		tokens[i] = model.ToEntity()
	}

	return tokens, nil
}

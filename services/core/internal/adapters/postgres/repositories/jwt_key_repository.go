package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/uptrace/bun"

	"github.com/paradoxe35/mytorra/packages/database"
	"github.com/paradoxe35/mytorra/packages/errors"
	"github.com/paradoxe35/mytorra/services/core/internal/adapters/postgres/models"
	"github.com/paradoxe35/mytorra/services/core/internal/application/ports"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/entities"
)

type jwtKeyRepository struct {
	db *database.DB
}

// NewJWTKeyRepository creates a new Bun-based JWT key repository
func NewJWTKeyRepository(db *database.DB) ports.JWTKeyRepository {
	return &jwtKeyRepository{db: db}
}

func (r *jwtKeyRepository) Create(ctx context.Context, key *entities.JWTKey) error {
	model := models.JWTKeyFromEntity(key)

	_, err := r.db.NewInsert().
		Model(model).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create JWT key: %w", err)
	}

	// Update the entity with any generated values
	*key = *model.ToEntity()
	return nil
}

func (r *jwtKeyRepository) FindByKeyID(ctx context.Context, keyID string) (*entities.JWTKey, error) {
	var keyModel models.JWTKeyModel

	err := r.db.NewSelect().
		Model(&keyModel).
		Where("key_id = ?", keyID).
		Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find JWT key: %w", err)
	}

	return keyModel.ToEntity(), nil
}

func (r *jwtKeyRepository) FindActiveKey(ctx context.Context) (*entities.JWTKey, error) {
	var keyModel models.JWTKeyModel

	now := time.Now()
	err := r.db.NewSelect().
		Model(&keyModel).
		Where("is_active = ? AND (expires_at IS NULL OR expires_at > ?)", true, now).
		Order("created_at DESC").
		Limit(1).
		Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find active JWT key: %w", err)
	}

	return keyModel.ToEntity(), nil
}

func (r *jwtKeyRepository) FindAllActive(ctx context.Context) ([]*entities.JWTKey, error) {
	var keyModels []models.JWTKeyModel

	now := time.Now()
	err := r.db.NewSelect().
		Model(&keyModels).
		Where("is_active = ? AND (expires_at IS NULL OR expires_at > ?)", true, now).
		Order("created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find all active JWT keys: %w", err)
	}

	keys := make([]*entities.JWTKey, len(keyModels))
	for i, model := range keyModels {
		keys[i] = model.ToEntity()
	}

	return keys, nil
}

func (r *jwtKeyRepository) FindAll(ctx context.Context) ([]*entities.JWTKey, error) {
	var keyModels []models.JWTKeyModel

	err := r.db.NewSelect().
		Model(&keyModels).
		Order("created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find all JWT keys: %w", err)
	}

	keys := make([]*entities.JWTKey, len(keyModels))
	for i, model := range keyModels {
		keys[i] = model.ToEntity()
	}

	return keys, nil
}

func (r *jwtKeyRepository) Update(ctx context.Context, key *entities.JWTKey) error {
	model := models.JWTKeyFromEntity(key)

	result, err := r.db.NewUpdate().
		Model(model).
		Where("id = ?", key.ID).
		OmitZero().
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update JWT key: %w", err)
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

func (r *jwtKeyRepository) RotateKey(ctx context.Context, newKey *entities.JWTKey) error {
	return r.db.Transaction(ctx, func(ctx context.Context, tx bun.Tx) error {
		// Deactivate all current active keys
		now := time.Now()
		_, err := tx.NewUpdate().
			Model(&models.JWTKeyModel{}).
			Where("is_active = ?", true).
			Set("is_active = ?", false).
			Set("rotated_at = ?", now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to deactivate old keys: %w", err)
		}

		// Create the new active key
		model := models.JWTKeyFromEntity(newKey)
		_, err = tx.NewInsert().
			Model(model).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new JWT key: %w", err)
		}

		// Update the entity with any generated values
		*newKey = *model.ToEntity()
		return nil
	})
}

func (r *jwtKeyRepository) DeactivateExpired(ctx context.Context) (int64, error) {
	now := time.Now()

	result, err := r.db.NewUpdate().
		Model(&models.JWTKeyModel{}).
		Where("is_active = ? AND expires_at IS NOT NULL AND expires_at < ?", true, now).
		Set("is_active = ?", false).
		Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to deactivate expired JWT keys: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

func (r *jwtKeyRepository) DeleteOldKeys(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoffTime := time.Now().Add(-olderThan)

	result, err := r.db.NewDelete().
		Model(&models.JWTKeyModel{}).
		Where("created_at < ? AND is_active = ?", cutoffTime, false).
		Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old JWT keys: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

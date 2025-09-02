package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/paradoxe35/mytorra/packages/database"
	"github.com/paradoxe35/mytorra/packages/errors"
	"github.com/paradoxe35/mytorra/services/core/internal/adapters/postgres/models"
	"github.com/paradoxe35/mytorra/services/core/internal/application/ports"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/entities"
)

type jwtKeyRepository struct {
	db *database.GormDB
}

// NewJWTKeyRepository creates a new GORM-based JWT key repository
func NewJWTKeyRepository(db *database.GormDB) ports.JWTKeyRepository {
	return &jwtKeyRepository{db: db}
}

func (r *jwtKeyRepository) Create(ctx context.Context, key *entities.JWTKey) error {
	model := models.JWTKeyFromEntity(key)

	err := r.db.WithContext(ctx).Create(model).Error
	if err != nil {
		return fmt.Errorf("failed to create JWT key: %w", err)
	}

	// Update the entity with any generated values
	*key = *model.ToEntity()
	return nil
}

func (r *jwtKeyRepository) FindByKeyID(ctx context.Context, keyID string) (*entities.JWTKey, error) {
	var keyModel models.JWTKeyModel

	err := r.db.WithContext(ctx).
		Where("key_id = ?", keyID).
		First(&keyModel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find JWT key: %w", err)
	}

	return keyModel.ToEntity(), nil
}

func (r *jwtKeyRepository) FindActiveKey(ctx context.Context) (*entities.JWTKey, error) {
	var keyModel models.JWTKeyModel

	now := time.Now()
	err := r.db.WithContext(ctx).
		Where("is_active = ? AND (expires_at IS NULL OR expires_at > ?)", true, now).
		Order("created_at DESC").
		First(&keyModel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find active JWT key: %w", err)
	}

	return keyModel.ToEntity(), nil
}

func (r *jwtKeyRepository) FindAllActive(ctx context.Context) ([]*entities.JWTKey, error) {
	var keyModels []models.JWTKeyModel

	now := time.Now()
	err := r.db.WithContext(ctx).
		Where("is_active = ? AND (expires_at IS NULL OR expires_at > ?)", true, now).
		Order("created_at DESC").
		Find(&keyModels).Error
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

	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Find(&keyModels).Error
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

	result := r.db.WithContext(ctx).
		Model(&models.JWTKeyModel{}).
		Where("id = ?", key.ID).
		Updates(model)

	if result.Error != nil {
		return fmt.Errorf("failed to update JWT key: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}

func (r *jwtKeyRepository) RotateKey(ctx context.Context, newKey *entities.JWTKey) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Deactivate all current active keys
		now := time.Now()
		if err := tx.Model(&models.JWTKeyModel{}).
			Where("is_active = ?", true).
			Updates(map[string]interface{}{
				"is_active":  false,
				"rotated_at": now,
			}).Error; err != nil {
			return fmt.Errorf("failed to deactivate old keys: %w", err)
		}

		// Create the new active key
		model := models.JWTKeyFromEntity(newKey)
		if err := tx.Create(model).Error; err != nil {
			return fmt.Errorf("failed to create new JWT key: %w", err)
		}

		// Update the entity with any generated values
		*newKey = *model.ToEntity()
		return nil
	})
}

func (r *jwtKeyRepository) DeactivateExpired(ctx context.Context) (int64, error) {
	now := time.Now()

	result := r.db.WithContext(ctx).
		Model(&models.JWTKeyModel{}).
		Where("is_active = ? AND expires_at IS NOT NULL AND expires_at < ?", true, now).
		Update("is_active", false)

	if result.Error != nil {
		return 0, fmt.Errorf("failed to deactivate expired JWT keys: %w", result.Error)
	}

	return result.RowsAffected, nil
}

func (r *jwtKeyRepository) DeleteOldKeys(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoffTime := time.Now().Add(-olderThan)

	result := r.db.WithContext(ctx).
		Where("created_at < ? AND is_active = ?", cutoffTime, false).
		Delete(&models.JWTKeyModel{})

	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete old JWT keys: %w", result.Error)
	}

	return result.RowsAffected, nil
}

package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/paradoxe35/torra/packages/database"
	"github.com/paradoxe35/torra/packages/errors"
	"github.com/paradoxe35/torra/services/core/internal/adapters/postgres/models"
	"github.com/paradoxe35/torra/services/core/internal/application/ports"
	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
)

type subscriptionTierRepository struct {
	db *database.GormDB
}

// NewSubscriptionTierRepository creates a new subscription tier repository
func NewSubscriptionTierRepository(db *database.GormDB) ports.SubscriptionTierRepository {
	return &subscriptionTierRepository{db: db}
}

func (r *subscriptionTierRepository) Create(ctx context.Context, tier *entities.SubscriptionTier) error {
	model := models.SubscriptionTierFromEntity(tier)

	err := r.db.WithContext(ctx).Create(model).Error
	if err != nil {
		return fmt.Errorf("failed to create subscription tier: %w", err)
	}

	return nil
}

func (r *subscriptionTierRepository) Update(ctx context.Context, tier *entities.SubscriptionTier) error {
	model := models.SubscriptionTierFromEntity(tier)

	result := r.db.WithContext(ctx).Model(&models.SubscriptionTierModel{}).Where("id = ?", model.ID).Updates(model)
	if result.Error != nil {
		return fmt.Errorf("failed to update subscription tier: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}

func (r *subscriptionTierRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.SubscriptionTier, error) {
	var model models.SubscriptionTierModel

	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find subscription tier: %w", err)
	}

	entity, err := model.ToEntity()
	if err != nil {
		return nil, fmt.Errorf("failed to convert model to entity: %w", err)
	}

	return entity, nil
}

func (r *subscriptionTierRepository) FindByCode(ctx context.Context, code string) (*entities.SubscriptionTier, error) {
	var model models.SubscriptionTierModel

	err := r.db.WithContext(ctx).Where("code = ? AND is_active = ?", code, true).First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find subscription tier by code: %w", err)
	}

	entity, err := model.ToEntity()
	if err != nil {
		return nil, fmt.Errorf("failed to convert model to entity: %w", err)
	}

	return entity, nil
}

func (r *subscriptionTierRepository) FindAll(ctx context.Context) ([]*entities.SubscriptionTier, error) {
	var models []models.SubscriptionTierModel

	err := r.db.WithContext(ctx).Order("sort_order ASC, created_at ASC").Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find subscription tiers: %w", err)
	}

	tiers := make([]*entities.SubscriptionTier, len(models))
	for i, model := range models {
		entity, err := model.ToEntity()
		if err != nil {
			return nil, fmt.Errorf("failed to convert model to entity: %w", err)
		}
		tiers[i] = entity
	}

	return tiers, nil
}

func (r *subscriptionTierRepository) FindActive(ctx context.Context) ([]*entities.SubscriptionTier, error) {
	var models []models.SubscriptionTierModel

	err := r.db.WithContext(ctx).Where("is_active = ?", true).Order("sort_order ASC, created_at ASC").Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find active subscription tiers: %w", err)
	}

	tiers := make([]*entities.SubscriptionTier, len(models))
	for i, model := range models {
		entity, err := model.ToEntity()
		if err != nil {
			return nil, fmt.Errorf("failed to convert model to entity: %w", err)
		}
		tiers[i] = entity
	}

	return tiers, nil
}

func (r *subscriptionTierRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.SubscriptionTierModel{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete subscription tier: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}

package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"github.com/paradoxe35/mytorra/packages/database"
	"github.com/paradoxe35/mytorra/packages/errors"
	"github.com/paradoxe35/mytorra/services/core/internal/adapters/postgres/models"
	"github.com/paradoxe35/mytorra/services/core/internal/application/ports"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/entities"
)

type subscriptionTierRepository struct {
	db *database.DB
}

// NewSubscriptionTierRepository creates a new subscription tier repository
func NewSubscriptionTierRepository(db *database.DB) ports.SubscriptionTierRepository {
	return &subscriptionTierRepository{db: db}
}

func (r *subscriptionTierRepository) Create(ctx context.Context, tier *entities.SubscriptionTier) error {
	model := models.SubscriptionTierFromEntity(tier)

	_, err := r.db.NewInsert().
		Model(model).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create subscription tier: %w", err)
	}

	return nil
}

func (r *subscriptionTierRepository) Update(ctx context.Context, tier *entities.SubscriptionTier) error {
	model := models.SubscriptionTierFromEntity(tier)

	result, err := r.db.NewUpdate().
		Model(model).
		Where("id = ?", model.ID).
		OmitZero().
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update subscription tier: %w", err)
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

func (r *subscriptionTierRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.SubscriptionTier, error) {
	var model models.SubscriptionTierModel

	err := r.db.NewSelect().
		Model(&model).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
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

	err := r.db.NewSelect().
		Model(&model).
		Where("code = ? AND is_active = ?", code, true).
		Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
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

	err := r.db.NewSelect().
		Model(&models).
		Order("sort_order ASC", "created_at ASC").
		Scan(ctx)
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

	err := r.db.NewSelect().
		Model(&models).
		Where("is_active = ?", true).
		Order("sort_order ASC", "created_at ASC").
		Scan(ctx)
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
	result, err := r.db.NewDelete().
		Model(&models.SubscriptionTierModel{}).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete subscription tier: %w", err)
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

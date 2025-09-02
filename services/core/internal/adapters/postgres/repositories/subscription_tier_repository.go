package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/paradoxe35/torra/packages/database"
	"github.com/paradoxe35/torra/packages/errors"
	"github.com/paradoxe35/torra/services/core/internal/adapters/postgres/models"
	"github.com/paradoxe35/torra/services/core/internal/application/ports"
	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
)

type subscriptionTierRepository struct {
	db *database.DB
}

// NewSubscriptionTierRepository creates a new subscription tier repository
func NewSubscriptionTierRepository(db *database.DB) ports.SubscriptionTierRepository {
	return &subscriptionTierRepository{db: db}
}

func (r *subscriptionTierRepository) Create(ctx context.Context, tier *entities.SubscriptionTier) error {
	model := models.FromSubscriptionTierEntity(tier)

	query := `
		INSERT INTO subscription_tiers (
			id, name, code, price, currency, features, limits,
			trial_days, sort_order, is_active, created_at, updated_at
		) VALUES (
			:id, :name, :code, :price, :currency, :features, :limits,
			:trial_days, :sort_order, :is_active, :created_at, :updated_at
		)
	`

	_, err := r.db.NamedExecContext(ctx, query, model)
	if err != nil {
		return fmt.Errorf("failed to create subscription tier: %w", err)
	}

	return nil
}

func (r *subscriptionTierRepository) Update(ctx context.Context, tier *entities.SubscriptionTier) error {
	model := models.FromSubscriptionTierEntity(tier)

	query := `
		UPDATE subscription_tiers SET
			name = :name,
			code = :code,
			price = :price,
			currency = :currency,
			features = :features,
			limits = :limits,
			trial_days = :trial_days,
			sort_order = :sort_order,
			is_active = :is_active,
			updated_at = :updated_at
		WHERE id = :id
	`

	result, err := r.db.NamedExecContext(ctx, query, model)
	if err != nil {
		return fmt.Errorf("failed to update subscription tier: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}

func (r *subscriptionTierRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.SubscriptionTier, error) {
	var model models.SubscriptionTierModel

	query := `
		SELECT * FROM subscription_tiers
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &model, query, id)
	if err != nil {
		if database.IsNoRowsError(err) {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find subscription tier: %w", err)
	}

	return model.ToEntity(), nil
}

func (r *subscriptionTierRepository) FindByCode(ctx context.Context, code string) (*entities.SubscriptionTier, error) {
	var model models.SubscriptionTierModel

	query := `
		SELECT * FROM subscription_tiers
		WHERE code = $1 AND is_active = true
	`

	err := r.db.GetContext(ctx, &model, query, code)
	if err != nil {
		if database.IsNoRowsError(err) {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find subscription tier by code: %w", err)
	}

	return model.ToEntity(), nil
}

func (r *subscriptionTierRepository) FindAll(ctx context.Context) ([]*entities.SubscriptionTier, error) {
	var models []models.SubscriptionTierModel

	query := `
		SELECT * FROM subscription_tiers
		ORDER BY sort_order ASC, created_at ASC
	`

	err := r.db.SelectContext(ctx, &models, query)
	if err != nil {
		return nil, fmt.Errorf("failed to find subscription tiers: %w", err)
	}

	tiers := make([]*entities.SubscriptionTier, len(models))
	for i, model := range models {
		tiers[i] = model.ToEntity()
	}

	return tiers, nil
}

func (r *subscriptionTierRepository) FindActive(ctx context.Context) ([]*entities.SubscriptionTier, error) {
	var models []models.SubscriptionTierModel

	query := `
		SELECT * FROM subscription_tiers
		WHERE is_active = true
		ORDER BY sort_order ASC, created_at ASC
	`

	err := r.db.SelectContext(ctx, &models, query)
	if err != nil {
		return nil, fmt.Errorf("failed to find active subscription tiers: %w", err)
	}

	tiers := make([]*entities.SubscriptionTier, len(models))
	for i, model := range models {
		tiers[i] = model.ToEntity()
	}

	return tiers, nil
}

func (r *subscriptionTierRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM subscription_tiers WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete subscription tier: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}

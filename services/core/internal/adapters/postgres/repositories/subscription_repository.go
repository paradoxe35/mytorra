package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/paradoxe35/torra/packages/database"
	"github.com/paradoxe35/torra/packages/errors"
	"github.com/paradoxe35/torra/services/core/internal/adapters/postgres/models"
	"github.com/paradoxe35/torra/services/core/internal/application/ports"
	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
)

type subscriptionRepository struct {
	db *database.DB
}

// NewSubscriptionRepository creates a new subscription repository
func NewSubscriptionRepository(db *database.DB) ports.SubscriptionRepository {
	return &subscriptionRepository{db: db}
}

func (r *subscriptionRepository) Create(ctx context.Context, subscription *entities.Subscription) error {
	model := models.FromSubscriptionEntity(subscription)

	query := `
		INSERT INTO subscriptions (
			id, user_id, entity_id, entity_type, tier_id, tier, plan_name,
			price, currency, billing_cycle, status, current_period_start,
			current_period_end, trial_ends_at, is_trialing, next_billing_date,
			stripe_customer_id, stripe_subscription_id, usage_limits, current_usage,
			metadata, paused_at, resume_at, canceled_at, cancel_reason,
			activated_at, created_at, updated_at
		) VALUES (
			:id, :user_id, :entity_id, :entity_type, :tier_id, :tier, :plan_name,
			:price, :currency, :billing_cycle, :status, :current_period_start,
			:current_period_end, :trial_ends_at, :is_trialing, :next_billing_date,
			:stripe_customer_id, :stripe_subscription_id, :usage_limits, :current_usage,
			:metadata, :paused_at, :resume_at, :canceled_at, :cancel_reason,
			:activated_at, :created_at, :updated_at
		)
	`

	_, err := r.db.NamedExecContext(ctx, query, model)
	return err
}

func (r *subscriptionRepository) Update(ctx context.Context, subscription *entities.Subscription) error {
	model := models.FromSubscriptionEntity(subscription)

	query := `
		UPDATE subscriptions SET
			tier_id = :tier_id,
			tier = :tier,
			plan_name = :plan_name,
			price = :price,
			currency = :currency,
			billing_cycle = :billing_cycle,
			status = :status,
			current_period_start = :current_period_start,
			current_period_end = :current_period_end,
			trial_ends_at = :trial_ends_at,
			is_trialing = :is_trialing,
			next_billing_date = :next_billing_date,
			stripe_customer_id = :stripe_customer_id,
			stripe_subscription_id = :stripe_subscription_id,
			usage_limits = :usage_limits,
			current_usage = :current_usage,
			metadata = :metadata,
			paused_at = :paused_at,
			resume_at = :resume_at,
			canceled_at = :canceled_at,
			cancel_reason = :cancel_reason,
			activated_at = :activated_at,
			updated_at = :updated_at
		WHERE id = :id
	`

	result, err := r.db.NamedExecContext(ctx, query, model)
	if err != nil {
		return err
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

func (r *subscriptionRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Subscription, error) {
	var model models.SubscriptionModel

	query := `
		SELECT * FROM subscriptions
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &model, query, id)
	if err != nil {
		if database.IsNoRowsError(err) {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find subscription: %w", err)
	}

	return model.ToEntity(), nil
}

func (r *subscriptionRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.Subscription, error) {
	var models []models.SubscriptionModel

	query := `
		SELECT * FROM subscriptions
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	err := r.db.SelectContext(ctx, &models, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find subscriptions: %w", err)
	}

	subscriptions := make([]*entities.Subscription, len(models))
	for i, model := range models {
		subscriptions[i] = model.ToEntity()
	}

	return subscriptions, nil
}

func (r *subscriptionRepository) FindActiveByUserID(ctx context.Context, userID uuid.UUID) (*entities.Subscription, error) {
	var model models.SubscriptionModel

	query := `
		SELECT * FROM subscriptions
		WHERE user_id = $1
		AND status IN ('active', 'trialing')
		LIMIT 1
	`

	err := r.db.GetContext(ctx, &model, query, userID)
	if err != nil {
		if database.IsNoRowsError(err) {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find active subscription: %w", err)
	}

	return model.ToEntity(), nil
}

func (r *subscriptionRepository) FindExpiring(ctx context.Context, beforeDate time.Time) ([]*entities.Subscription, error) {
	var models []models.SubscriptionModel

	query := `
		SELECT * FROM subscriptions
		WHERE (trial_ends_at IS NOT NULL AND trial_ends_at <= $1)
		OR (next_billing_date IS NOT NULL AND next_billing_date <= $1)
		AND status IN ('active', 'trialing')
	`

	err := r.db.SelectContext(ctx, &models, query, beforeDate)
	if err != nil {
		return nil, fmt.Errorf("failed to find expiring subscriptions: %w", err)
	}

	subscriptions := make([]*entities.Subscription, len(models))
	for i, model := range models {
		subscriptions[i] = model.ToEntity()
	}

	return subscriptions, nil
}

func (r *subscriptionRepository) CountByTier(ctx context.Context, tierID uuid.UUID) (int64, error) {
	var count int64

	query := `
		SELECT COUNT(*) FROM subscriptions
		WHERE tier_id = $1
		AND status IN ('active', 'trialing')
	`

	err := r.db.GetContext(ctx, &count, query, tierID)
	if err != nil {
		return 0, fmt.Errorf("failed to count subscriptions by tier: %w", err)
	}

	return count, nil
}

func (r *subscriptionRepository) FindActiveByUserAndEntity(ctx context.Context, userID, entityID uuid.UUID) (*entities.Subscription, error) {
	var model models.SubscriptionModel

	query := `
		SELECT * FROM subscriptions
		WHERE user_id = $1
		AND entity_id = $2
		AND status IN ('active', 'trialing')
		LIMIT 1
	`

	err := r.db.GetContext(ctx, &model, query, userID, entityID)
	if err != nil {
		if database.IsNoRowsError(err) {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find subscription: %w", err)
	}

	return model.ToEntity(), nil
}

func (r *subscriptionRepository) FindByEntityID(ctx context.Context, entityID uuid.UUID) ([]*entities.Subscription, error) {
	var models []models.SubscriptionModel

	query := `
		SELECT * FROM subscriptions
		WHERE entity_id = $1
		ORDER BY created_at DESC
	`

	err := r.db.SelectContext(ctx, &models, query, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to find subscriptions: %w", err)
	}

	subscriptions := make([]*entities.Subscription, len(models))
	for i, model := range models {
		subscriptions[i] = model.ToEntity()
	}

	return subscriptions, nil
}

func (r *subscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM subscriptions WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
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

func (r *subscriptionRepository) Cancel(ctx context.Context, id uuid.UUID, reason string) error {
	query := `
		UPDATE subscriptions
		SET status = 'canceled',
		    canceled_at = NOW(),
		    cancel_reason = $2,
		    updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id, reason)
	if err != nil {
		return fmt.Errorf("failed to cancel subscription: %w", err)
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

func (r *subscriptionRepository) Pause(ctx context.Context, id uuid.UUID, until *time.Time) error {
	query := `
		UPDATE subscriptions
		SET status = 'paused',
		    paused_at = NOW(),
		    resume_at = $2,
		    updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id, until)
	if err != nil {
		return fmt.Errorf("failed to pause subscription: %w", err)
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

func (r *subscriptionRepository) Resume(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE subscriptions
		SET status = 'active',
		    paused_at = NULL,
		    resume_at = NULL,
		    updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to resume subscription: %w", err)
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

func (r *subscriptionRepository) GetExpiringSubscriptions(ctx context.Context, days int) ([]*entities.Subscription, error) {
	var models []models.SubscriptionModel

	query := `
		SELECT * FROM subscriptions
		WHERE status IN ('active', 'trialing')
		AND (
			(trial_ends_at IS NOT NULL AND trial_ends_at BETWEEN NOW() AND NOW() + INTERVAL '%d days')
			OR (next_billing_date IS NOT NULL AND next_billing_date BETWEEN NOW() AND NOW() + INTERVAL '%d days')
		)
	`

	err := r.db.SelectContext(ctx, &models, fmt.Sprintf(query, days, days))
	if err != nil {
		return nil, fmt.Errorf("failed to find expiring subscriptions: %w", err)
	}

	subscriptions := make([]*entities.Subscription, len(models))
	for i, model := range models {
		subscriptions[i] = model.ToEntity()
	}

	return subscriptions, nil
}

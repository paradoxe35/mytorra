package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"github.com/paradoxe35/mytorra/packages/database"
	"github.com/paradoxe35/mytorra/packages/errors"
	"github.com/paradoxe35/mytorra/services/core/internal/adapters/postgres/models"
	"github.com/paradoxe35/mytorra/services/core/internal/application/ports"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/entities"
)

type subscriptionRepository struct {
	db *database.DB
}

// NewSubscriptionRepository creates a new subscription repository
func NewSubscriptionRepository(db *database.DB) ports.SubscriptionRepository {
	return &subscriptionRepository{db: db}
}

func (r *subscriptionRepository) Create(ctx context.Context, subscription *entities.Subscription) error {
	model := models.SubscriptionFromEntity(subscription)

	_, err := r.db.NewInsert().
		Model(model).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	return nil
}

func (r *subscriptionRepository) Update(ctx context.Context, subscription *entities.Subscription) error {
	model := models.SubscriptionFromEntity(subscription)

	result, err := r.db.NewUpdate().
		Model(model).
		Where("id = ?", model.ID).
		OmitZero().
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
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

func (r *subscriptionRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Subscription, error) {
	var model models.SubscriptionModel

	err := r.db.NewSelect().
		Model(&model).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find subscription: %w", err)
	}

	entity, err := model.ToEntity()
	if err != nil {
		return nil, fmt.Errorf("failed to convert model to entity: %w", err)
	}

	return entity, nil
}

func (r *subscriptionRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.Subscription, error) {
	var models []models.SubscriptionModel

	err := r.db.NewSelect().
		Model(&models).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find subscriptions: %w", err)
	}

	subscriptions := make([]*entities.Subscription, len(models))
	for i, model := range models {
		entity, err := model.ToEntity()
		if err != nil {
			return nil, fmt.Errorf("failed to convert model to entity: %w", err)
		}
		subscriptions[i] = entity
	}

	return subscriptions, nil
}

func (r *subscriptionRepository) FindActiveByUserID(ctx context.Context, userID uuid.UUID) (*entities.Subscription, error) {
	var model models.SubscriptionModel

	err := r.db.NewSelect().
		Model(&model).
		Where("user_id = ?", userID).
		Where("status IN (?)", bun.In([]string{"active", "trialing"})).
		Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find active subscription: %w", err)
	}

	entity, err := model.ToEntity()
	if err != nil {
		return nil, fmt.Errorf("failed to convert model to entity: %w", err)
	}

	return entity, nil
}

func (r *subscriptionRepository) FindExpiring(ctx context.Context, beforeDate time.Time) ([]*entities.Subscription, error) {
	var models []models.SubscriptionModel

	err := r.db.NewSelect().
		Model(&models).
		Where("((trial_ends_at IS NOT NULL AND trial_ends_at <= ?) OR (next_billing_date IS NOT NULL AND next_billing_date <= ?))", beforeDate, beforeDate).
		Where("status IN (?)", bun.In([]string{"active", "trialing"})).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find expiring subscriptions: %w", err)
	}

	subscriptions := make([]*entities.Subscription, len(models))
	for i, model := range models {
		entity, err := model.ToEntity()
		if err != nil {
			return nil, fmt.Errorf("failed to convert model to entity: %w", err)
		}
		subscriptions[i] = entity
	}

	return subscriptions, nil
}

func (r *subscriptionRepository) CountByTier(ctx context.Context, tierID uuid.UUID) (int64, error) {
	count, err := r.db.NewSelect().
		Model(&models.SubscriptionModel{}).
		Where("tier_id = ?", tierID).
		Where("status IN (?)", bun.In([]string{"active", "trialing"})).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count subscriptions by tier: %w", err)
	}

	return int64(count), nil
}

func (r *subscriptionRepository) FindActiveByUserAndEntity(ctx context.Context, userID, entityID uuid.UUID) (*entities.Subscription, error) {
	var model models.SubscriptionModel

	err := r.db.NewSelect().
		Model(&model).
		Where("user_id = ?", userID).
		Where("entity_id = ?", entityID).
		Where("status IN (?)", bun.In([]string{"active", "trialing"})).
		Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find subscription: %w", err)
	}

	entity, err := model.ToEntity()
	if err != nil {
		return nil, fmt.Errorf("failed to convert model to entity: %w", err)
	}

	return entity, nil
}

func (r *subscriptionRepository) FindByEntityID(ctx context.Context, entityID uuid.UUID) ([]*entities.Subscription, error) {
	var models []models.SubscriptionModel

	err := r.db.NewSelect().
		Model(&models).
		Where("entity_id = ?", entityID).
		Order("created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find subscriptions: %w", err)
	}

	subscriptions := make([]*entities.Subscription, len(models))
	for i, model := range models {
		entity, err := model.ToEntity()
		if err != nil {
			return nil, fmt.Errorf("failed to convert model to entity: %w", err)
		}
		subscriptions[i] = entity
	}

	return subscriptions, nil
}

func (r *subscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.NewDelete().
		Model(&models.SubscriptionModel{}).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
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

func (r *subscriptionRepository) Cancel(ctx context.Context, id uuid.UUID, reason string) error {
	updates := map[string]interface{}{
		"status":        "canceled",
		"canceled_at":   time.Now(),
		"cancel_reason": reason,
		"updated_at":    time.Now(),
	}

	query := r.db.NewUpdate().
		Model(&models.SubscriptionModel{}).
		Where("id = ?", id)

	for key, value := range updates {
		query = query.Set("? = ?", bun.Ident(key), value)
	}

	result, err := query.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to cancel subscription: %w", err)
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

func (r *subscriptionRepository) Pause(ctx context.Context, id uuid.UUID, until *time.Time) error {
	updates := map[string]interface{}{
		"status":     "paused",
		"paused_at":  time.Now(),
		"resume_at":  until,
		"updated_at": time.Now(),
	}

	query := r.db.NewUpdate().
		Model(&models.SubscriptionModel{}).
		Where("id = ?", id)

	for key, value := range updates {
		query = query.Set("? = ?", bun.Ident(key), value)
	}

	result, err := query.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to pause subscription: %w", err)
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

func (r *subscriptionRepository) Resume(ctx context.Context, id uuid.UUID) error {
	updates := map[string]interface{}{
		"status":     "active",
		"paused_at":  nil,
		"resume_at":  nil,
		"updated_at": time.Now(),
	}

	query := r.db.NewUpdate().
		Model(&models.SubscriptionModel{}).
		Where("id = ?", id)

	for key, value := range updates {
		query = query.Set("? = ?", bun.Ident(key), value)
	}

	result, err := query.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to resume subscription: %w", err)
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

func (r *subscriptionRepository) GetExpiringSubscriptions(ctx context.Context, days int) ([]*entities.Subscription, error) {
	var models []models.SubscriptionModel

	now := time.Now()
	futureDate := now.AddDate(0, 0, days)

	err := r.db.NewSelect().
		Model(&models).
		Where("status IN (?)", bun.In([]string{"active", "trialing"})).
		Where("((trial_ends_at IS NOT NULL AND trial_ends_at BETWEEN ? AND ?) OR (next_billing_date IS NOT NULL AND next_billing_date BETWEEN ? AND ?))",
			now, futureDate, now, futureDate).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find expiring subscriptions: %w", err)
	}

	subscriptions := make([]*entities.Subscription, len(models))
	for i, model := range models {
		entity, err := model.ToEntity()
		if err != nil {
			return nil, fmt.Errorf("failed to convert model to entity: %w", err)
		}
		subscriptions[i] = entity
	}

	return subscriptions, nil
}

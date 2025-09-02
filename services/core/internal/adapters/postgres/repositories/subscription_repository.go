package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/paradoxe35/mytorra/packages/database"
	"github.com/paradoxe35/mytorra/packages/errors"
	"github.com/paradoxe35/mytorra/services/core/internal/adapters/postgres/models"
	"github.com/paradoxe35/mytorra/services/core/internal/application/ports"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/entities"
)

type subscriptionRepository struct {
	db *database.GormDB
}

// NewSubscriptionRepository creates a new subscription repository
func NewSubscriptionRepository(db *database.GormDB) ports.SubscriptionRepository {
	return &subscriptionRepository{db: db}
}

func (r *subscriptionRepository) Create(ctx context.Context, subscription *entities.Subscription) error {
	model := models.SubscriptionFromEntity(subscription)

	err := r.db.WithContext(ctx).Create(model).Error
	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	return nil
}

func (r *subscriptionRepository) Update(ctx context.Context, subscription *entities.Subscription) error {
	model := models.SubscriptionFromEntity(subscription)

	result := r.db.WithContext(ctx).Model(&models.SubscriptionModel{}).Where("id = ?", model.ID).Updates(model)
	if result.Error != nil {
		return fmt.Errorf("failed to update subscription: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}

func (r *subscriptionRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Subscription, error) {
	var model models.SubscriptionModel

	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
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

	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&models).Error
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

	err := r.db.WithContext(ctx).Where("user_id = ? AND status IN (?)", userID, []string{"active", "trialing"}).First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
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

	err := r.db.WithContext(ctx).Where(
		"((trial_ends_at IS NOT NULL AND trial_ends_at <= ?) OR (next_billing_date IS NOT NULL AND next_billing_date <= ?)) AND status IN (?)",
		beforeDate, beforeDate, []string{"active", "trialing"},
	).Find(&models).Error
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
	var count int64

	err := r.db.WithContext(ctx).Model(&models.SubscriptionModel{}).Where("tier_id = ? AND status IN (?)", tierID, []string{"active", "trialing"}).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count subscriptions by tier: %w", err)
	}

	return count, nil
}

func (r *subscriptionRepository) FindActiveByUserAndEntity(ctx context.Context, userID, entityID uuid.UUID) (*entities.Subscription, error) {
	var model models.SubscriptionModel

	err := r.db.WithContext(ctx).Where("user_id = ? AND entity_id = ? AND status IN (?)", userID, entityID, []string{"active", "trialing"}).First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
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

	err := r.db.WithContext(ctx).Where("entity_id = ?", entityID).Order("created_at DESC").Find(&models).Error
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
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.SubscriptionModel{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete subscription: %w", result.Error)
	}

	if result.RowsAffected == 0 {
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

	result := r.db.WithContext(ctx).Model(&models.SubscriptionModel{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to cancel subscription: %w", result.Error)
	}

	if result.RowsAffected == 0 {
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

	result := r.db.WithContext(ctx).Model(&models.SubscriptionModel{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to pause subscription: %w", result.Error)
	}

	if result.RowsAffected == 0 {
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

	result := r.db.WithContext(ctx).Model(&models.SubscriptionModel{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to resume subscription: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}

func (r *subscriptionRepository) GetExpiringSubscriptions(ctx context.Context, days int) ([]*entities.Subscription, error) {
	var models []models.SubscriptionModel

	err := r.db.WithContext(ctx).Where(
		"status IN (?) AND ((trial_ends_at IS NOT NULL AND trial_ends_at BETWEEN ? AND ?) OR (next_billing_date IS NOT NULL AND next_billing_date BETWEEN ? AND ?))",
		[]string{"active", "trialing"},
		time.Now(),
		time.Now().AddDate(0, 0, days),
		time.Now(),
		time.Now().AddDate(0, 0, days),
	).Find(&models).Error
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

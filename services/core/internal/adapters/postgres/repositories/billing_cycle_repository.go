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

type billingCycleRepository struct {
	db *database.GormDB
}

// NewBillingCycleRepository creates a new billing cycle repository
func NewBillingCycleRepository(db *database.GormDB) ports.BillingCycleRepository {
	return &billingCycleRepository{db: db}
}

func (r *billingCycleRepository) Create(ctx context.Context, cycle *entities.BillingCycle) error {
	model := models.BillingCycleFromEntity(cycle)

	err := r.db.WithContext(ctx).Create(model).Error
	if err != nil {
		return fmt.Errorf("failed to create billing cycle: %w", err)
	}

	return nil
}

func (r *billingCycleRepository) Update(ctx context.Context, cycle *entities.BillingCycle) error {
	model := models.BillingCycleFromEntity(cycle)

	result := r.db.WithContext(ctx).Model(&models.BillingCycleModel{}).Where("id = ?", model.ID).Updates(model)
	if result.Error != nil {
		return fmt.Errorf("failed to update billing cycle: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}

func (r *billingCycleRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.BillingCycle, error) {
	var model models.BillingCycleModel

	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find billing cycle: %w", err)
	}

	entity, err := model.ToEntity()
	if err != nil {
		return nil, fmt.Errorf("failed to convert model to entity: %w", err)
	}

	return entity, nil
}

func (r *billingCycleRepository) FindBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) ([]*entities.BillingCycle, error) {
	var models []models.BillingCycleModel

	err := r.db.WithContext(ctx).Where("subscription_id = ?", subscriptionID).Order("start_date DESC").Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find billing cycles: %w", err)
	}

	cycles := make([]*entities.BillingCycle, len(models))
	for i, model := range models {
		entity, err := model.ToEntity()
		if err != nil {
			return nil, fmt.Errorf("failed to convert model to entity: %w", err)
		}
		cycles[i] = entity
	}

	return cycles, nil
}

func (r *billingCycleRepository) FindPending(ctx context.Context) ([]*entities.BillingCycle, error) {
	var models []models.BillingCycleModel

	err := r.db.WithContext(ctx).Where("status = ? AND end_date <= ?", "pending", time.Now()).Order("end_date ASC").Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find pending billing cycles: %w", err)
	}

	cycles := make([]*entities.BillingCycle, len(models))
	for i, model := range models {
		entity, err := model.ToEntity()
		if err != nil {
			return nil, fmt.Errorf("failed to convert model to entity: %w", err)
		}
		cycles[i] = entity
	}

	return cycles, nil
}

func (r *billingCycleRepository) FindOverdue(ctx context.Context) ([]*entities.BillingCycle, error) {
	var models []models.BillingCycleModel

	threeDaysAgo := time.Now().AddDate(0, 0, -3)
	err := r.db.WithContext(ctx).Where(
		"status IN (?) AND end_date < ? AND paid_at IS NULL",
		[]string{"pending", "failed"}, threeDaysAgo,
	).Order("end_date ASC").Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find overdue billing cycles: %w", err)
	}

	cycles := make([]*entities.BillingCycle, len(models))
	for i, model := range models {
		entity, err := model.ToEntity()
		if err != nil {
			return nil, fmt.Errorf("failed to convert model to entity: %w", err)
		}
		cycles[i] = entity
	}

	return cycles, nil
}

func (r *billingCycleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.BillingCycleModel{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete billing cycle: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}

func (r *billingCycleRepository) FindCurrentBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) (*entities.BillingCycle, error) {
	var model models.BillingCycleModel

	now := time.Now()
	err := r.db.WithContext(ctx).Where(
		"subscription_id = ? AND start_date <= ? AND end_date >= ?",
		subscriptionID, now, now,
	).First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find current billing cycle: %w", err)
	}

	entity, err := model.ToEntity()
	if err != nil {
		return nil, fmt.Errorf("failed to convert model to entity: %w", err)
	}

	return entity, nil
}

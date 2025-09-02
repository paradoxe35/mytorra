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

type billingCycleRepository struct {
	db *database.DB
}

// NewBillingCycleRepository creates a new billing cycle repository
func NewBillingCycleRepository(db *database.DB) ports.BillingCycleRepository {
	return &billingCycleRepository{db: db}
}

func (r *billingCycleRepository) Create(ctx context.Context, cycle *entities.BillingCycle) error {
	model := models.BillingCycleFromEntity(cycle)

	_, err := r.db.NewInsert().
		Model(model).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create billing cycle: %w", err)
	}

	return nil
}

func (r *billingCycleRepository) Update(ctx context.Context, cycle *entities.BillingCycle) error {
	model := models.BillingCycleFromEntity(cycle)

	result, err := r.db.NewUpdate().
		Model(model).
		Where("id = ?", model.ID).
		OmitZero().
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update billing cycle: %w", err)
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

func (r *billingCycleRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.BillingCycle, error) {
	var model models.BillingCycleModel

	err := r.db.NewSelect().
		Model(&model).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
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

	err := r.db.NewSelect().
		Model(&models).
		Where("subscription_id = ?", subscriptionID).
		Order("start_date DESC").
		Scan(ctx)
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

	err := r.db.NewSelect().
		Model(&models).
		Where("status = ? AND end_date <= ?", "pending", time.Now()).
		Order("end_date ASC").
		Scan(ctx)
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
	err := r.db.NewSelect().
		Model(&models).
		Where("status IN (?)", []string{"pending", "failed"}).
		Where("end_date < ?", threeDaysAgo).
		Where("paid_at IS NULL").
		Order("end_date ASC").
		Scan(ctx)
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
	result, err := r.db.NewDelete().
		Model(&models.BillingCycleModel{}).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete billing cycle: %w", err)
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

func (r *billingCycleRepository) FindCurrentBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) (*entities.BillingCycle, error) {
	var model models.BillingCycleModel

	now := time.Now()
	err := r.db.NewSelect().
		Model(&model).
		Where("subscription_id = ?", subscriptionID).
		Where("start_date <= ?", now).
		Where("end_date >= ?", now).
		Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
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

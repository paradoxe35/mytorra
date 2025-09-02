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

type billingCycleRepository struct {
	db *database.DB
}

// NewBillingCycleRepository creates a new billing cycle repository
func NewBillingCycleRepository(db *database.DB) ports.BillingCycleRepository {
	return &billingCycleRepository{db: db}
}

func (r *billingCycleRepository) Create(ctx context.Context, cycle *entities.BillingCycle) error {
	model := models.FromBillingCycleEntity(cycle)

	query := `
		INSERT INTO billing_cycles (
			id, subscription_id, start_date, end_date, amount, currency,
			status, invoice_id, paid_at, created_at, updated_at
		) VALUES (
			:id, :subscription_id, :start_date, :end_date, :amount, :currency,
			:status, :invoice_id, :paid_at, :created_at, :updated_at
		)
	`

	_, err := r.db.NamedExecContext(ctx, query, model)
	if err != nil {
		return fmt.Errorf("failed to create billing cycle: %w", err)
	}

	return nil
}

func (r *billingCycleRepository) Update(ctx context.Context, cycle *entities.BillingCycle) error {
	model := models.FromBillingCycleEntity(cycle)

	query := `
		UPDATE billing_cycles SET
			start_date = :start_date,
			end_date = :end_date,
			amount = :amount,
			currency = :currency,
			status = :status,
			invoice_id = :invoice_id,
			paid_at = :paid_at,
			updated_at = :updated_at
		WHERE id = :id
	`

	result, err := r.db.NamedExecContext(ctx, query, model)
	if err != nil {
		return fmt.Errorf("failed to update billing cycle: %w", err)
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

func (r *billingCycleRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.BillingCycle, error) {
	var model models.BillingCycleModel

	query := `
		SELECT * FROM billing_cycles
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &model, query, id)
	if err != nil {
		if database.IsNoRowsError(err) {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find billing cycle: %w", err)
	}

	return model.ToEntity(), nil
}

func (r *billingCycleRepository) FindBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) ([]*entities.BillingCycle, error) {
	var models []models.BillingCycleModel

	query := `
		SELECT * FROM billing_cycles
		WHERE subscription_id = $1
		ORDER BY start_date DESC
	`

	err := r.db.SelectContext(ctx, &models, query, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("failed to find billing cycles: %w", err)
	}

	cycles := make([]*entities.BillingCycle, len(models))
	for i, model := range models {
		cycles[i] = model.ToEntity()
	}

	return cycles, nil
}

func (r *billingCycleRepository) FindPending(ctx context.Context) ([]*entities.BillingCycle, error) {
	var models []models.BillingCycleModel

	query := `
		SELECT * FROM billing_cycles
		WHERE status = 'pending'
		AND end_date <= NOW()
		ORDER BY end_date ASC
	`

	err := r.db.SelectContext(ctx, &models, query)
	if err != nil {
		return nil, fmt.Errorf("failed to find pending billing cycles: %w", err)
	}

	cycles := make([]*entities.BillingCycle, len(models))
	for i, model := range models {
		cycles[i] = model.ToEntity()
	}

	return cycles, nil
}

func (r *billingCycleRepository) FindOverdue(ctx context.Context) ([]*entities.BillingCycle, error) {
	var models []models.BillingCycleModel

	query := `
		SELECT * FROM billing_cycles
		WHERE status IN ('pending', 'failed')
		AND end_date < NOW() - INTERVAL '3 days'
		AND paid_at IS NULL
		ORDER BY end_date ASC
	`

	err := r.db.SelectContext(ctx, &models, query)
	if err != nil {
		return nil, fmt.Errorf("failed to find overdue billing cycles: %w", err)
	}

	cycles := make([]*entities.BillingCycle, len(models))
	for i, model := range models {
		cycles[i] = model.ToEntity()
	}

	return cycles, nil
}

func (r *billingCycleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM billing_cycles WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete billing cycle: %w", err)
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

func (r *billingCycleRepository) FindCurrentBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) (*entities.BillingCycle, error) {
	var model models.BillingCycleModel

	query := `
		SELECT * FROM billing_cycles
		WHERE subscription_id = $1
		AND start_date <= NOW()
		AND end_date >= NOW()
		LIMIT 1
	`

	err := r.db.GetContext(ctx, &model, query, subscriptionID)
	if err != nil {
		if database.IsNoRowsError(err) {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find current billing cycle: %w", err)
	}

	return model.ToEntity(), nil
}

package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/paradoxe35/torra/packages/database"
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
	// TODO: Implement
	return nil
}

func (r *billingCycleRepository) Update(ctx context.Context, cycle *entities.BillingCycle) error {
	// TODO: Implement
	return nil
}

func (r *billingCycleRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.BillingCycle, error) {
	// TODO: Implement
	return &entities.BillingCycle{
		ID:             id,
		SubscriptionID: uuid.New(),
		StartDate:      time.Now(),
		EndDate:        time.Now().AddDate(0, 1, 0),
		Status:         "pending",
	}, nil
}

func (r *billingCycleRepository) FindBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) ([]*entities.BillingCycle, error) {
	// TODO: Implement
	return []*entities.BillingCycle{}, nil
}

func (r *billingCycleRepository) FindPending(ctx context.Context) ([]*entities.BillingCycle, error) {
	// TODO: Implement
	return []*entities.BillingCycle{}, nil
}

func (r *billingCycleRepository) FindOverdue(ctx context.Context) ([]*entities.BillingCycle, error) {
	// TODO: Implement
	return []*entities.BillingCycle{}, nil
}

func (r *billingCycleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// TODO: Implement
	return nil
}

func (r *billingCycleRepository) FindCurrentBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) (*entities.BillingCycle, error) {
	// TODO: Implement
	return &entities.BillingCycle{
		ID:             uuid.New(),
		SubscriptionID: subscriptionID,
		StartDate:      time.Now().AddDate(0, -1, 0),
		EndDate:        time.Now(),
		Status:         "active",
	}, nil
}

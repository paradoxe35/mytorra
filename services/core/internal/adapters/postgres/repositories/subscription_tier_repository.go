package repositories

import (
	"context"

	"github.com/google/uuid"

	"github.com/paradoxe35/torra/packages/database"
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
	// TODO: Implement
	return nil
}

func (r *subscriptionTierRepository) Update(ctx context.Context, tier *entities.SubscriptionTier) error {
	// TODO: Implement
	return nil
}

func (r *subscriptionTierRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.SubscriptionTier, error) {
	// TODO: Implement
	return &entities.SubscriptionTier{
		ID:   id,
		Name: "Basic",
		Code: "basic",
	}, nil
}

func (r *subscriptionTierRepository) FindByCode(ctx context.Context, code string) (*entities.SubscriptionTier, error) {
	// TODO: Implement
	return &entities.SubscriptionTier{
		ID:   uuid.New(),
		Name: "Basic",
		Code: code,
	}, nil
}

func (r *subscriptionTierRepository) FindAll(ctx context.Context) ([]*entities.SubscriptionTier, error) {
	// TODO: Implement
	return []*entities.SubscriptionTier{}, nil
}

func (r *subscriptionTierRepository) FindActive(ctx context.Context) ([]*entities.SubscriptionTier, error) {
	// TODO: Implement
	return []*entities.SubscriptionTier{}, nil
}

func (r *subscriptionTierRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// TODO: Implement
	return nil
}

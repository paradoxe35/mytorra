package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"github.com/paradoxe35/torra/packages/database"
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
	// TODO: Implement
	return nil
}

func (r *subscriptionRepository) Update(ctx context.Context, subscription *entities.Subscription) error {
	// TODO: Implement
	return nil
}

func (r *subscriptionRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Subscription, error) {
	// TODO: Implement
	return &entities.Subscription{
		ID:     id,
		UserID: uuid.New(),
		Status: entities.SubscriptionStatusActive,
	}, nil
}

func (r *subscriptionRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.Subscription, error) {
	// TODO: Implement
	return []*entities.Subscription{}, nil
}

func (r *subscriptionRepository) FindActiveByUserID(ctx context.Context, userID uuid.UUID) (*entities.Subscription, error) {
	// TODO: Implement
	return nil, sql.ErrNoRows
}

func (r *subscriptionRepository) FindExpiring(ctx context.Context, beforeDate time.Time) ([]*entities.Subscription, error) {
	// TODO: Implement
	return []*entities.Subscription{}, nil
}

func (r *subscriptionRepository) CountByTier(ctx context.Context, tierID uuid.UUID) (int64, error) {
	// TODO: Implement
	return 0, nil
}

func (r *subscriptionRepository) FindActiveByUserAndEntity(ctx context.Context, userID, entityID uuid.UUID) (*entities.Subscription, error) {
	// TODO: Implement
	return nil, sql.ErrNoRows
}

func (r *subscriptionRepository) FindByEntityID(ctx context.Context, entityID uuid.UUID) ([]*entities.Subscription, error) {
	// TODO: Implement
	return []*entities.Subscription{}, nil
}

func (r *subscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// TODO: Implement
	return nil
}

func (r *subscriptionRepository) Cancel(ctx context.Context, id uuid.UUID, reason string) error {
	// TODO: Implement
	return nil
}

func (r *subscriptionRepository) Pause(ctx context.Context, id uuid.UUID, until *time.Time) error {
	// TODO: Implement
	return nil
}

func (r *subscriptionRepository) Resume(ctx context.Context, id uuid.UUID) error {
	// TODO: Implement
	return nil
}

func (r *subscriptionRepository) GetExpiringSubscriptions(ctx context.Context, days int) ([]*entities.Subscription, error) {
	// TODO: Implement
	return []*entities.Subscription{}, nil
}

package ports

import (
	"context"

	"github.com/google/uuid"

	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
	"github.com/paradoxe35/torra/services/core/internal/domain/valueobjects"
)

// AreaRepository defines the interface for area persistence
type AreaRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Area, error)
	FindByCode(ctx context.Context, code string) (*entities.Area, error)
	FindByParentID(ctx context.Context, parentID *uuid.UUID) ([]*entities.Area, error)
	FindByType(ctx context.Context, areaType valueobjects.AreaType) ([]*entities.Area, error)
	FindActive(ctx context.Context) ([]*entities.Area, error)
	FindByCoordinate(ctx context.Context, lat, lng float64) ([]*entities.Area, error)
	Save(ctx context.Context, area *entities.Area) error
	Update(ctx context.Context, area *entities.Area) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetHierarchy(ctx context.Context, areaID uuid.UUID) ([]*entities.Area, error)
}

// UserRepository defines the interface for user persistence
type UserRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*entities.User, error)
	FindByEmail(ctx context.Context, email valueobjects.Email) (*entities.User, error)
	FindByPhone(ctx context.Context, phone valueobjects.Phone) (*entities.User, error)
	Save(ctx context.Context, user *entities.User) error
	Update(ctx context.Context, user *entities.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByArea(ctx context.Context, area string, limit, offset int) ([]*entities.User, error)
	Count(ctx context.Context) (int64, error)
}

// PaymentRepository defines the interface for payment persistence
type PaymentRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Payment, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Payment, error)
	FindByProviderID(ctx context.Context, providerID string) (*entities.Payment, error)
	Save(ctx context.Context, payment *entities.Payment) error
	Update(ctx context.Context, payment *entities.Payment) error
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
	GetTotalAmountByUserID(ctx context.Context, userID uuid.UUID) (float64, error)
}

// SubscriptionRepository defines the interface for subscription persistence
type SubscriptionRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Subscription, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Subscription, error)
	FindByEntityID(ctx context.Context, entityID uuid.UUID) ([]*entities.Subscription, error)
	FindActiveByUserAndEntity(ctx context.Context, userID, entityID uuid.UUID) (*entities.Subscription, error)
	Save(ctx context.Context, subscription *entities.Subscription) error
	Update(ctx context.Context, subscription *entities.Subscription) error
	Cancel(ctx context.Context, id uuid.UUID, reason string) error
	GetExpiringSubscriptions(ctx context.Context, days int) ([]*entities.Subscription, error)
}

// WalletRepository defines the interface for wallet persistence
type WalletRepository interface {
	FindByUserID(ctx context.Context, userID uuid.UUID) (*entities.Wallet, error)
	Create(ctx context.Context, wallet *entities.Wallet) error
	UpdateBalance(ctx context.Context, userID uuid.UUID, amount float64) error
	AddTransaction(ctx context.Context, transaction *entities.WalletTransaction) error
	GetTransactionHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.WalletTransaction, error)
}

// NotificationRepository defines the interface for notification persistence
type NotificationRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Notification, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Notification, error)
	Save(ctx context.Context, notification *entities.Notification) error
	MarkAsRead(ctx context.Context, id uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	CountUnread(ctx context.Context, userID uuid.UUID) (int64, error)
}

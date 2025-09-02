package ports

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/paradoxe35/mytorra/services/core/internal/domain/entities"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/valueobjects"
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
	Create(ctx context.Context, subscription *entities.Subscription) error
	Update(ctx context.Context, subscription *entities.Subscription) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Subscription, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.Subscription, error)
	FindActiveByUserID(ctx context.Context, userID uuid.UUID) (*entities.Subscription, error)
	FindByEntityID(ctx context.Context, entityID uuid.UUID) ([]*entities.Subscription, error)
	FindActiveByUserAndEntity(ctx context.Context, userID, entityID uuid.UUID) (*entities.Subscription, error)
	FindExpiring(ctx context.Context, beforeDate time.Time) ([]*entities.Subscription, error)
	CountByTier(ctx context.Context, tierID uuid.UUID) (int64, error)
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

// SubscriptionTierRepository defines the interface for subscription tier persistence
type SubscriptionTierRepository interface {
	Create(ctx context.Context, tier *entities.SubscriptionTier) error
	Update(ctx context.Context, tier *entities.SubscriptionTier) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.SubscriptionTier, error)
	FindByCode(ctx context.Context, code string) (*entities.SubscriptionTier, error)
	FindActive(ctx context.Context) ([]*entities.SubscriptionTier, error)
	FindAll(ctx context.Context) ([]*entities.SubscriptionTier, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// BillingCycleRepository defines the interface for billing cycle persistence
type BillingCycleRepository interface {
	Create(ctx context.Context, cycle *entities.BillingCycle) error
	Update(ctx context.Context, cycle *entities.BillingCycle) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.BillingCycle, error)
	FindBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) ([]*entities.BillingCycle, error)
	FindCurrentBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) (*entities.BillingCycle, error)
	FindPending(ctx context.Context) ([]*entities.BillingCycle, error)
	FindOverdue(ctx context.Context) ([]*entities.BillingCycle, error)
}

// PaymentTransactionRepository defines the interface for payment transaction persistence
type PaymentTransactionRepository interface {
	Create(ctx context.Context, transaction *entities.PaymentTransaction) error
	Update(ctx context.Context, transaction *entities.PaymentTransaction) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.PaymentTransaction, error)
	FindBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) ([]*entities.PaymentTransaction, error)
	FindByGatewayTransactionID(ctx context.Context, gatewayTransactionID string) (*entities.PaymentTransaction, error)
	FindFailedForRetry(ctx context.Context, maxRetries int) ([]*entities.PaymentTransaction, error)
	GetTotalRevenue(ctx context.Context, startDate, endDate time.Time) (float64, error)
}

// NotificationService defines the interface for notification services
type NotificationService interface {
	Send(ctx context.Context, notification *entities.Notification) error
	SendBulk(ctx context.Context, notifications []*entities.Notification) error
	SendEmail(ctx context.Context, to, subject, body string) error
	SendSMS(ctx context.Context, to, message string) error
}

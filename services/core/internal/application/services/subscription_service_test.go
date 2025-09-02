package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/paradoxe35/torra/packages/events"
	"github.com/paradoxe35/torra/packages/i18n"
	"github.com/paradoxe35/torra/packages/logger"
	"github.com/paradoxe35/torra/services/core/internal/application/dtos"
	"github.com/paradoxe35/torra/services/core/internal/application/ports"
	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
	"github.com/paradoxe35/torra/services/core/internal/domain/valueobjects"
)

// Mock implementations
type MockSubscriptionRepository struct {
	mock.Mock
}

func (m *MockSubscriptionRepository) Create(ctx context.Context, subscription *entities.Subscription) error {
	args := m.Called(ctx, subscription)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) Update(ctx context.Context, subscription *entities.Subscription) error {
	args := m.Called(ctx, subscription)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Subscription, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.Subscription, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) FindActiveByUserID(ctx context.Context, userID uuid.UUID) (*entities.Subscription, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) FindByEntityID(ctx context.Context, entityID uuid.UUID) ([]*entities.Subscription, error) {
	args := m.Called(ctx, entityID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) FindActiveByUserAndEntity(ctx context.Context, userID, entityID uuid.UUID) (*entities.Subscription, error) {
	args := m.Called(ctx, userID, entityID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) FindExpiring(ctx context.Context, beforeDate time.Time) ([]*entities.Subscription, error) {
	args := m.Called(ctx, beforeDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) CountByTier(ctx context.Context, tierID uuid.UUID) (int64, error) {
	args := m.Called(ctx, tierID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockSubscriptionRepository) Cancel(ctx context.Context, id uuid.UUID, reason string) error {
	args := m.Called(ctx, id, reason)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) GetExpiringSubscriptions(ctx context.Context, days int) ([]*entities.Subscription, error) {
	args := m.Called(ctx, days)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Subscription), args.Error(1)
}

type MockTierRepository struct {
	mock.Mock
}

func (m *MockTierRepository) Create(ctx context.Context, tier *entities.SubscriptionTier) error {
	args := m.Called(ctx, tier)
	return args.Error(0)
}

func (m *MockTierRepository) Update(ctx context.Context, tier *entities.SubscriptionTier) error {
	args := m.Called(ctx, tier)
	return args.Error(0)
}

func (m *MockTierRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.SubscriptionTier, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.SubscriptionTier), args.Error(1)
}

func (m *MockTierRepository) FindByCode(ctx context.Context, code string) (*entities.SubscriptionTier, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.SubscriptionTier), args.Error(1)
}

func (m *MockTierRepository) FindActive(ctx context.Context) ([]*entities.SubscriptionTier, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.SubscriptionTier), args.Error(1)
}

func (m *MockTierRepository) FindAll(ctx context.Context) ([]*entities.SubscriptionTier, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.SubscriptionTier), args.Error(1)
}

func (m *MockTierRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockBillingCycleRepository struct {
	mock.Mock
}

func (m *MockBillingCycleRepository) Create(ctx context.Context, cycle *entities.BillingCycle) error {
	args := m.Called(ctx, cycle)
	return args.Error(0)
}

func (m *MockBillingCycleRepository) Update(ctx context.Context, cycle *entities.BillingCycle) error {
	args := m.Called(ctx, cycle)
	return args.Error(0)
}

func (m *MockBillingCycleRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.BillingCycle, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.BillingCycle), args.Error(1)
}

func (m *MockBillingCycleRepository) FindBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) ([]*entities.BillingCycle, error) {
	args := m.Called(ctx, subscriptionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.BillingCycle), args.Error(1)
}

func (m *MockBillingCycleRepository) FindCurrentBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) (*entities.BillingCycle, error) {
	args := m.Called(ctx, subscriptionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.BillingCycle), args.Error(1)
}

func (m *MockBillingCycleRepository) FindPending(ctx context.Context) ([]*entities.BillingCycle, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.BillingCycle), args.Error(1)
}

func (m *MockBillingCycleRepository) FindOverdue(ctx context.Context) ([]*entities.BillingCycle, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.BillingCycle), args.Error(1)
}

type MockPaymentGateway struct {
	mock.Mock
}

func (m *MockPaymentGateway) CreatePaymentIntent(ctx context.Context, req ports.PaymentIntentRequest) (*ports.PaymentIntentResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ports.PaymentIntentResponse), args.Error(1)
}

func (m *MockPaymentGateway) ConfirmPayment(ctx context.Context, paymentIntentID string) (*ports.PaymentConfirmation, error) {
	args := m.Called(ctx, paymentIntentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ports.PaymentConfirmation), args.Error(1)
}

func (m *MockPaymentGateway) CancelPayment(ctx context.Context, paymentIntentID string) error {
	args := m.Called(ctx, paymentIntentID)
	return args.Error(0)
}

func (m *MockPaymentGateway) RefundPayment(ctx context.Context, req ports.RefundPaymentRequest) (*ports.RefundResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ports.RefundResponse), args.Error(1)
}

func (m *MockPaymentGateway) CreatePaymentMethod(ctx context.Context, req ports.CreatePaymentMethodRequest) (*ports.PaymentMethodResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ports.PaymentMethodResponse), args.Error(1)
}

func (m *MockPaymentGateway) AttachPaymentMethod(ctx context.Context, paymentMethodID, customerID string) error {
	args := m.Called(ctx, paymentMethodID, customerID)
	return args.Error(0)
}

func (m *MockPaymentGateway) DetachPaymentMethod(ctx context.Context, paymentMethodID string) error {
	args := m.Called(ctx, paymentMethodID)
	return args.Error(0)
}

func (m *MockPaymentGateway) ListPaymentMethods(ctx context.Context, customerID string) ([]*ports.PaymentMethod, error) {
	args := m.Called(ctx, customerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*ports.PaymentMethod), args.Error(1)
}

func (m *MockPaymentGateway) CreateCustomer(ctx context.Context, user *entities.User) (string, error) {
	args := m.Called(ctx, user)
	return args.String(0), args.Error(1)
}

func (m *MockPaymentGateway) UpdateCustomer(ctx context.Context, customerID string, req ports.CustomerRequest) (*ports.Customer, error) {
	args := m.Called(ctx, customerID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ports.Customer), args.Error(1)
}

func (m *MockPaymentGateway) DeleteCustomer(ctx context.Context, customerID string) error {
	args := m.Called(ctx, customerID)
	return args.Error(0)
}

func (m *MockPaymentGateway) ValidateWebhook(ctx context.Context, payload []byte, signature string) (bool, error) {
	args := m.Called(ctx, payload, signature)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockPaymentGateway) CreateSubscription(ctx context.Context, req ports.CreateSubscriptionRequest) (*ports.SubscriptionResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ports.SubscriptionResponse), args.Error(1)
}

func (m *MockPaymentGateway) UpdateSubscription(ctx context.Context, req ports.UpdateSubscriptionRequest) (*ports.SubscriptionResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ports.SubscriptionResponse), args.Error(1)
}

func (m *MockPaymentGateway) CancelSubscription(ctx context.Context, subscriptionID string, immediate bool) error {
	args := m.Called(ctx, subscriptionID, immediate)
	return args.Error(0)
}

func (m *MockPaymentGateway) PauseSubscription(ctx context.Context, subscriptionID string) error {
	args := m.Called(ctx, subscriptionID)
	return args.Error(0)
}

func (m *MockPaymentGateway) ResumeSubscription(ctx context.Context, subscriptionID string) error {
	args := m.Called(ctx, subscriptionID)
	return args.Error(0)
}

func (m *MockPaymentGateway) GetSubscription(ctx context.Context, subscriptionID string) (*ports.SubscriptionResponse, error) {
	args := m.Called(ctx, subscriptionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ports.SubscriptionResponse), args.Error(1)
}

func (m *MockPaymentGateway) ChargePayment(ctx context.Context, req ports.ChargePaymentRequest) (*ports.PaymentResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ports.PaymentResponse), args.Error(1)
}

func (m *MockPaymentGateway) DeletePaymentMethod(ctx context.Context, methodID string) error {
	args := m.Called(ctx, methodID)
	return args.Error(0)
}

func (m *MockPaymentGateway) GetPaymentMethods(ctx context.Context, customerID string) ([]*ports.PaymentMethodResponse, error) {
	args := m.Called(ctx, customerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*ports.PaymentMethodResponse), args.Error(1)
}

func (m *MockPaymentGateway) GetInvoice(ctx context.Context, invoiceID string) (*ports.InvoiceResponse, error) {
	args := m.Called(ctx, invoiceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ports.InvoiceResponse), args.Error(1)
}

func (m *MockPaymentGateway) ProcessWebhook(ctx context.Context, payload []byte, signature string) (*ports.WebhookEvent, error) {
	args := m.Called(ctx, payload, signature)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ports.WebhookEvent), args.Error(1)
}

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email valueobjects.Email) (*entities.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) FindByPhone(ctx context.Context, phone valueobjects.Phone) (*entities.User, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) Save(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Update(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) ListByArea(ctx context.Context, area string, limit, offset int) ([]*entities.User, error) {
	args := m.Called(ctx, area, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.User), args.Error(1)
}

func (m *MockUserRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// Test functions
func TestSubscriptionService_CreateSubscription(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	mockSubRepo := new(MockSubscriptionRepository)
	mockTierRepo := new(MockTierRepository)
	mockBillingRepo := new(MockBillingCycleRepository)
	mockPaymentGateway := new(MockPaymentGateway)
	mockUserRepo := new(MockUserRepository)
	mockEventBus := new(events.MockEventPublisher)

	translator, _ := i18n.New("en", "en", "fr")
	testLogger := logger.New("test", "test")

	service := NewSubscriptionService(
		mockSubRepo,
		mockTierRepo,
		mockBillingRepo,
		mockPaymentGateway,
		mockUserRepo,
		mockEventBus,
		testLogger,
		translator,
	)

	userID := uuid.New()
	tierID := uuid.New()

	t.Run("Create subscription successfully", func(t *testing.T) {
		// Setup test data
		dto := dtos.CreateSubscriptionDTO{
			UserID:        userID,
			TierID:        tierID,
			PaymentMethod: "pm_test",
			StartTrial:    true,
		}

		tier := &entities.SubscriptionTier{
			ID:        tierID,
			Name:      "Professional",
			Code:      "professional",
			Price:     decimal.NewFromFloat(29.99),
			Currency:  "USD",
			TrialDays: 14,
			IsActive:  true,
			Limits:    map[string]int{"api_calls": 1000},
		}

		email, _ := valueobjects.NewEmail("test@example.com")
		phone, _ := valueobjects.NewPhone("+1234567890")
		user := &entities.User{
			ID:    userID,
			Email: email,
			Name:  "Test User",
			Phone: phone,
		}

		// Setup expectations
		mockSubRepo.On("FindActiveByUserID", ctx, userID).Return(nil, errors.New("not found")).Once()
		mockTierRepo.On("FindByID", ctx, tierID).Return(tier, nil).Once()
		mockUserRepo.On("FindByID", ctx, userID).Return(user, nil).Once()
		mockPaymentGateway.On("CreateCustomer", ctx, user).Return("cus_test", nil).Once()
		mockSubRepo.On("Create", ctx, mock.AnythingOfType("*entities.Subscription")).Return(nil).Once()
		mockBillingRepo.On("Create", ctx, mock.AnythingOfType("*entities.BillingCycle")).Return(nil).Once()
		mockEventBus.On("Publish", ctx, mock.Anything).Return(nil).Once()

		// Execute
		subscription, err := service.CreateSubscription(ctx, dto)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, subscription)
		assert.Equal(t, userID, subscription.UserID)
		assert.Equal(t, entities.SubscriptionStatusTrialing, subscription.Status)
		assert.True(t, subscription.IsTrialing)
		assert.NotNil(t, subscription.TrialEndsAt)

		mockSubRepo.AssertExpectations(t)
		mockTierRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockPaymentGateway.AssertExpectations(t)
		mockBillingRepo.AssertExpectations(t)
		mockEventBus.AssertExpectations(t)
	})

	t.Run("Fail when active subscription exists", func(t *testing.T) {
		dto := dtos.CreateSubscriptionDTO{
			UserID:        userID,
			TierID:        tierID,
			PaymentMethod: "pm_test",
		}

		existingSub := &entities.Subscription{
			ID:     uuid.New(),
			UserID: userID,
			Status: entities.SubscriptionStatusActive,
		}

		mockSubRepo.On("FindActiveByUserID", ctx, userID).Return(existingSub, nil).Once()

		// Execute
		subscription, err := service.CreateSubscription(ctx, dto)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, subscription)
		assert.Contains(t, err.Error(), "already exists")

		mockSubRepo.AssertExpectations(t)
	})
}

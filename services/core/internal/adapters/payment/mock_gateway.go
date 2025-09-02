package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/paradoxe35/torra/packages/logger"
	"github.com/paradoxe35/torra/services/core/internal/application/ports"
	"github.com/paradoxe35/torra/services/core/internal/domain/valueobjects"
)

// MockGateway is a mock implementation of the PaymentGateway interface for testing
type MockGateway struct {
	customers      map[string]*ports.Customer
	paymentMethods map[string]*ports.PaymentMethod
	subscriptions  map[string]*ports.SubscriptionResponse
	logger         logger.Interface
}

// NewMockGateway creates a new mock payment gateway
func NewMockGateway(logger logger.Interface) *MockGateway {
	return &MockGateway{
		customers:      make(map[string]*ports.Customer),
		paymentMethods: make(map[string]*ports.PaymentMethod),
		subscriptions:  make(map[string]*ports.SubscriptionResponse),
		logger:         logger,
	}
}

// CreatePaymentIntent creates a mock payment intent
func (m *MockGateway) CreatePaymentIntent(ctx context.Context, req ports.PaymentIntentRequest) (*ports.PaymentIntentResponse, error) {
	m.logger.Info("Mock: Creating payment intent")

	return &ports.PaymentIntentResponse{
		ID:           fmt.Sprintf("pi_%s", uuid.New().String()[:8]),
		ClientSecret: fmt.Sprintf("pi_%s_secret", uuid.New().String()[:8]),
		Status:       "requires_confirmation",
		Amount:       req.Amount,
		CreatedAt:    time.Now(),
	}, nil
}

// ConfirmPayment confirms a mock payment
func (m *MockGateway) ConfirmPayment(ctx context.Context, paymentIntentID string) (*ports.PaymentConfirmation, error) {
	m.logger.Info("Mock: Confirming payment")

	// Simulate random payment failures (10% failure rate)
	if time.Now().Unix()%10 == 0 {
		return nil, fmt.Errorf("mock payment failed: insufficient funds")
	}

	return &ports.PaymentConfirmation{
		ID:        paymentIntentID,
		Status:    "succeeded",
		Amount:    valueobjects.Money{Amount: decimal.NewFromFloat(100), Currency: "USD"},
		ChargeID:  fmt.Sprintf("ch_%s", uuid.New().String()[:8]),
		CreatedAt: time.Now(),
	}, nil
}

// CancelPayment cancels a mock payment
func (m *MockGateway) CancelPayment(ctx context.Context, paymentIntentID string) error {
	m.logger.Info("Mock: Canceling payment")
	return nil
}

// RefundPayment processes a mock refund
func (m *MockGateway) RefundPayment(ctx context.Context, paymentID string, amount *valueobjects.Money) (*ports.RefundResponse, error) {
	m.logger.Info("Mock: Processing refund")

	return &ports.RefundResponse{
		RefundID:      fmt.Sprintf("re_%s", uuid.New().String()[:8]),
		Amount:        amount.Amount,
		Currency:      amount.Currency,
		Status:        "succeeded",
		TransactionID: fmt.Sprintf("txn_%s", uuid.New().String()[:8]),
	}, nil
}

// CreatePaymentMethod creates a mock payment method
func (m *MockGateway) CreatePaymentMethod(ctx context.Context, req ports.PaymentMethodRequest) (*ports.PaymentMethod, error) {
	m.logger.Info("Mock: Creating payment method")

	pm := &ports.PaymentMethod{
		ID:   fmt.Sprintf("pm_%s", uuid.New().String()[:8]),
		Type: req.Type,
		Card: &ports.CardInfo{
			Last4:    "4242",
			Brand:    "visa",
			ExpMonth: 12,
			ExpYear:  2025,
		},
		Created: time.Now(),
	}

	m.paymentMethods[pm.ID] = pm
	return pm, nil
}

// AttachPaymentMethod attaches a payment method to a customer
func (m *MockGateway) AttachPaymentMethod(ctx context.Context, paymentMethodID, customerID string) error {
	m.logger.Info("Mock: Attaching payment method to customer")
	return nil
}

// DetachPaymentMethod detaches a payment method
func (m *MockGateway) DetachPaymentMethod(ctx context.Context, paymentMethodID string) error {
	m.logger.Info("Mock: Detaching payment method")
	delete(m.paymentMethods, paymentMethodID)
	return nil
}

// ListPaymentMethods lists payment methods for a customer
func (m *MockGateway) ListPaymentMethods(ctx context.Context, customerID string) ([]*ports.PaymentMethod, error) {
	m.logger.Info("Mock: Listing payment methods")

	methods := make([]*ports.PaymentMethod, 0, len(m.paymentMethods))
	for _, pm := range m.paymentMethods {
		methods = append(methods, pm)
	}
	return methods, nil
}

// CreateCustomer creates a mock customer
func (m *MockGateway) CreateCustomer(ctx context.Context, req ports.CustomerRequest) (*ports.Customer, error) {
	m.logger.Info("Mock: Creating customer")

	customer := &ports.Customer{
		ID:                   fmt.Sprintf("cus_%s", uuid.New().String()[:8]),
		Email:                req.Email,
		Name:                 req.Name,
		DefaultPaymentMethod: "",
		Created:              time.Now(),
	}

	m.customers[customer.ID] = customer
	return customer, nil
}

// UpdateCustomer updates a mock customer
func (m *MockGateway) UpdateCustomer(ctx context.Context, customerID string, req ports.CustomerRequest) (*ports.Customer, error) {
	m.logger.Info("Mock: Updating customer")

	if customer, exists := m.customers[customerID]; exists {
		customer.Email = req.Email
		customer.Name = req.Name
		return customer, nil
	}
	return nil, fmt.Errorf("customer not found")
}

// DeleteCustomer deletes a mock customer
func (m *MockGateway) DeleteCustomer(ctx context.Context, customerID string) error {
	m.logger.Info("Mock: Deleting customer")
	delete(m.customers, customerID)
	return nil
}

// ValidateWebhook validates a mock webhook
func (m *MockGateway) ValidateWebhook(ctx context.Context, payload []byte, signature string) (bool, error) {
	m.logger.Info("Mock: Validating webhook")
	// Always return true for mock
	return true, nil
}

// CreateSubscription creates a mock subscription
func (m *MockGateway) CreateSubscription(ctx context.Context, req ports.CreateSubscriptionRequest) (*ports.SubscriptionResponse, error) {
	m.logger.Info("Mock: Creating subscription")

	now := time.Now()
	subscription := &ports.SubscriptionResponse{
		ProviderSubscriptionID: fmt.Sprintf("sub_%s", uuid.New().String()[:8]),
		Status:                 "active",
		CurrentPeriodStart:     now,
		CurrentPeriodEnd:       now.AddDate(0, 1, 0),
	}

	if req.TrialDays > 0 {
		trialEnd := now.AddDate(0, 0, req.TrialDays)
		subscription.TrialEnd = &trialEnd
		subscription.Status = "trialing"
	}

	m.subscriptions[subscription.ProviderSubscriptionID] = subscription
	return subscription, nil
}

// UpdateSubscription updates a mock subscription
func (m *MockGateway) UpdateSubscription(ctx context.Context, req ports.UpdateSubscriptionRequest) (*ports.SubscriptionResponse, error) {
	m.logger.Info("Mock: Updating subscription")

	if subscription, exists := m.subscriptions[req.ProviderSubscriptionID]; exists {
		// Update subscription based on request
		return subscription, nil
	}
	return nil, fmt.Errorf("subscription not found")
}

// CancelSubscription cancels a mock subscription
func (m *MockGateway) CancelSubscription(ctx context.Context, subscriptionID string, immediate bool) error {
	m.logger.Info("Mock: Canceling subscription")

	if subscription, exists := m.subscriptions[subscriptionID]; exists {
		if immediate {
			subscription.Status = "canceled"
		} else {
			subscription.Status = "active" // Will cancel at period end
		}
		return nil
	}
	return fmt.Errorf("subscription not found")
}

// PauseSubscription pauses a mock subscription
func (m *MockGateway) PauseSubscription(ctx context.Context, subscriptionID string) error {
	m.logger.Info("Mock: Pausing subscription")

	if subscription, exists := m.subscriptions[subscriptionID]; exists {
		subscription.Status = "paused"
		return nil
	}
	return fmt.Errorf("subscription not found")
}

// ResumeSubscription resumes a mock subscription
func (m *MockGateway) ResumeSubscription(ctx context.Context, subscriptionID string) error {
	m.logger.Info("Mock: Resuming subscription")

	if subscription, exists := m.subscriptions[subscriptionID]; exists {
		subscription.Status = "active"
		return nil
	}
	return fmt.Errorf("subscription not found")
}

// GetSubscription gets a mock subscription
func (m *MockGateway) GetSubscription(ctx context.Context, subscriptionID string) (*ports.SubscriptionResponse, error) {
	m.logger.Info("Mock: Getting subscription")

	if subscription, exists := m.subscriptions[subscriptionID]; exists {
		return subscription, nil
	}
	return nil, fmt.Errorf("subscription not found")
}

package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/paradoxe35/mytorra/packages/logger"
	"github.com/paradoxe35/mytorra/services/core/internal/application/ports"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/entities"
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

// CreateCustomer creates a mock customer from user entity
func (m *MockGateway) CreateCustomer(ctx context.Context, user *entities.User) (string, error) {
	m.logger.Info("Mock: Creating customer from user")

	customerID := fmt.Sprintf("cus_%s", uuid.New().String()[:8])
	customer := &ports.Customer{
		ID:                   customerID,
		Email:                user.Email.String(),
		Name:                 user.Name,
		DefaultPaymentMethod: "",
		Created:              time.Now(),
	}

	m.customers[customerID] = customer
	return customerID, nil
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

// ChargePayment charges a mock payment
func (m *MockGateway) ChargePayment(ctx context.Context, req ports.ChargePaymentRequest) (*ports.PaymentResponse, error) {
	m.logger.Info("Mock: Charging payment")

	// Simulate random payment failures (10% failure rate)
	if time.Now().Unix()%10 == 0 {
		return nil, fmt.Errorf("mock payment failed: insufficient funds")
	}

	return &ports.PaymentResponse{
		PaymentID:     fmt.Sprintf("pi_%s", uuid.New().String()[:8]),
		Amount:        req.Amount,
		Currency:      req.Currency,
		Status:        "succeeded",
		TransactionID: fmt.Sprintf("txn_%s", uuid.New().String()[:8]),
	}, nil
}

// RefundPayment processes a mock refund
func (m *MockGateway) RefundPayment(ctx context.Context, req ports.RefundPaymentRequest) (*ports.RefundResponse, error) {
	m.logger.Info("Mock: Processing refund")

	return &ports.RefundResponse{
		RefundID:      fmt.Sprintf("re_%s", uuid.New().String()[:8]),
		Amount:        req.Amount,
		Currency:      "USD", // Default currency for mock
		Status:        "succeeded",
		TransactionID: fmt.Sprintf("txn_%s", uuid.New().String()[:8]),
	}, nil
}

// CreatePaymentMethod creates a mock payment method
func (m *MockGateway) CreatePaymentMethod(ctx context.Context, req ports.CreatePaymentMethodRequest) (*ports.PaymentMethodResponse, error) {
	m.logger.Info("Mock: Creating payment method")

	pm := &ports.PaymentMethodResponse{
		MethodID:  fmt.Sprintf("pm_%s", uuid.New().String()[:8]),
		Type:      "card",
		Last4:     "4242",
		Brand:     "visa",
		ExpMonth:  12,
		ExpYear:   2025,
		IsDefault: req.SetDefault,
	}

	return pm, nil
}

// DeletePaymentMethod deletes a mock payment method
func (m *MockGateway) DeletePaymentMethod(ctx context.Context, methodID string) error {
	m.logger.Info("Mock: Deleting payment method")
	return nil
}

// GetPaymentMethods gets payment methods for a customer
func (m *MockGateway) GetPaymentMethods(ctx context.Context, customerID string) ([]*ports.PaymentMethodResponse, error) {
	m.logger.Info("Mock: Getting payment methods")

	// Return some mock payment methods
	methods := []*ports.PaymentMethodResponse{
		{
			MethodID:  fmt.Sprintf("pm_%s", uuid.New().String()[:8]),
			Type:      "card",
			Last4:     "4242",
			Brand:     "visa",
			ExpMonth:  12,
			ExpYear:   2025,
			IsDefault: true,
		},
	}

	return methods, nil
}

// GetInvoice gets a mock invoice
func (m *MockGateway) GetInvoice(ctx context.Context, invoiceID string) (*ports.InvoiceResponse, error) {
	m.logger.Info("Mock: Getting invoice")

	return &ports.InvoiceResponse{
		InvoiceID: invoiceID,
		Number:    fmt.Sprintf("INV-%s", uuid.New().String()[:8]),
		Amount:    decimal.NewFromFloat(25.00),
		Currency:  "USD",
		Status:    "paid",
		DueDate:   time.Now().AddDate(0, 0, 30),
		Items: []ports.InvoiceItem{
			{
				Description: "Professional Plan - Monthly",
				Amount:      decimal.NewFromFloat(25.00),
				Quantity:    1,
			},
		},
		DownloadURL: fmt.Sprintf("https://mock-gateway.com/invoices/%s.pdf", invoiceID),
	}, nil
}

// ProcessWebhook processes a mock webhook
func (m *MockGateway) ProcessWebhook(ctx context.Context, payload []byte, signature string) (*ports.WebhookEvent, error) {
	m.logger.Info("Mock: Processing webhook")

	return &ports.WebhookEvent{
		ID:        fmt.Sprintf("evt_%s", uuid.New().String()[:8]),
		Type:      "invoice.payment_succeeded",
		Data:      payload,
		Timestamp: time.Now(),
	}, nil
}

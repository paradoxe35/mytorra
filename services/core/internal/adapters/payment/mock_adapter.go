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

// MockAdapter is a mock implementation of the PaymentGateway interface for testing
type MockAdapter struct {
	logger logger.Interface
	// In-memory storage for testing
	customers     map[string]*mockCustomer
	subscriptions map[string]*mockSubscription
	payments      map[string]*mockPayment
	methods       map[string]*mockPaymentMethod
}

type mockCustomer struct {
	ID    string
	Email string
	Name  string
}

type mockSubscription struct {
	ID                 string
	CustomerID         string
	Status             string
	CurrentPeriodStart time.Time
	CurrentPeriodEnd   time.Time
	TrialEnd           *time.Time
}

type mockPayment struct {
	ID            string
	Amount        decimal.Decimal
	Currency      string
	Status        string
	TransactionID string
}

type mockPaymentMethod struct {
	ID       string
	Customer string
	Type     string
	Last4    string
	Brand    string
	ExpMonth int
	ExpYear  int
}

func NewMockAdapter(logger logger.Interface) *MockAdapter {
	return &MockAdapter{
		logger:        logger,
		customers:     make(map[string]*mockCustomer),
		subscriptions: make(map[string]*mockSubscription),
		payments:      make(map[string]*mockPayment),
		methods:       make(map[string]*mockPaymentMethod),
	}
}

func (m *MockAdapter) CreateCustomer(ctx context.Context, user *entities.User) (string, error) {
	customerID := "cus_" + uuid.New().String()[:14]
	m.customers[customerID] = &mockCustomer{
		ID:    customerID,
		Email: user.Email.String(),
		Name:  user.Name,
	}

	m.logger.Info("Mock: Created customer", "customer_id", customerID)
	return customerID, nil
}

func (m *MockAdapter) CreateSubscription(ctx context.Context, req ports.CreateSubscriptionRequest) (*ports.SubscriptionResponse, error) {
	subscriptionID := "sub_" + uuid.New().String()[:14]

	now := time.Now()
	var trialEnd *time.Time
	if req.TrialDays > 0 {
		t := now.AddDate(0, 0, req.TrialDays)
		trialEnd = &t
	}

	var periodEnd time.Time
	switch req.BillingCycle {
	case entities.BillingCycleMonthly:
		periodEnd = now.AddDate(0, 1, 0)
	case entities.BillingCycleAnnual:
		periodEnd = now.AddDate(1, 0, 0)
	default:
		periodEnd = now.AddDate(0, 1, 0)
	}

	status := "active"
	if trialEnd != nil {
		status = "trialing"
	}

	m.subscriptions[subscriptionID] = &mockSubscription{
		ID:                 subscriptionID,
		CustomerID:         req.CustomerID,
		Status:             status,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   periodEnd,
		TrialEnd:           trialEnd,
	}

	m.logger.Info("Mock: Created subscription",
		"subscription_id", subscriptionID,
		"customer_id", req.CustomerID,
		"tier", req.TierCode,
	)

	return &ports.SubscriptionResponse{
		ProviderSubscriptionID: subscriptionID,
		Status:                 status,
		CurrentPeriodStart:     now,
		CurrentPeriodEnd:       periodEnd,
		TrialEnd:               trialEnd,
	}, nil
}

func (m *MockAdapter) UpdateSubscription(ctx context.Context, req ports.UpdateSubscriptionRequest) (*ports.SubscriptionResponse, error) {
	sub, exists := m.subscriptions[req.ProviderSubscriptionID]
	if !exists {
		return nil, fmt.Errorf("subscription not found: %s", req.ProviderSubscriptionID)
	}

	m.logger.Info("Mock: Updated subscription",
		"subscription_id", req.ProviderSubscriptionID,
		"new_tier", req.NewTierCode,
	)

	return &ports.SubscriptionResponse{
		ProviderSubscriptionID: sub.ID,
		Status:                 sub.Status,
		CurrentPeriodStart:     sub.CurrentPeriodStart,
		CurrentPeriodEnd:       sub.CurrentPeriodEnd,
		TrialEnd:               sub.TrialEnd,
	}, nil
}

func (m *MockAdapter) CancelSubscription(ctx context.Context, providerSubscriptionID string, immediate bool) error {
	sub, exists := m.subscriptions[providerSubscriptionID]
	if !exists {
		return fmt.Errorf("subscription not found: %s", providerSubscriptionID)
	}

	if immediate {
		sub.Status = "canceled"
	} else {
		sub.Status = "cancel_at_period_end"
	}

	m.logger.Info("Mock: Canceled subscription",
		"subscription_id", providerSubscriptionID,
		"immediate", immediate,
	)

	return nil
}

func (m *MockAdapter) PauseSubscription(ctx context.Context, providerSubscriptionID string) error {
	sub, exists := m.subscriptions[providerSubscriptionID]
	if !exists {
		return fmt.Errorf("subscription not found: %s", providerSubscriptionID)
	}

	sub.Status = "paused"

	m.logger.Info("Mock: Paused subscription", "subscription_id", providerSubscriptionID)
	return nil
}

func (m *MockAdapter) ResumeSubscription(ctx context.Context, providerSubscriptionID string) error {
	sub, exists := m.subscriptions[providerSubscriptionID]
	if !exists {
		return fmt.Errorf("subscription not found: %s", providerSubscriptionID)
	}

	sub.Status = "active"

	m.logger.Info("Mock: Resumed subscription", "subscription_id", providerSubscriptionID)
	return nil
}

func (m *MockAdapter) ChargePayment(ctx context.Context, req ports.ChargePaymentRequest) (*ports.PaymentResponse, error) {
	paymentID := "pi_" + uuid.New().String()[:14]
	transactionID := "txn_" + uuid.New().String()[:14]

	m.payments[paymentID] = &mockPayment{
		ID:            paymentID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		Status:        "succeeded",
		TransactionID: transactionID,
	}

	m.logger.Info("Mock: Charged payment",
		"payment_id", paymentID,
		"amount", req.Amount.String(),
		"currency", req.Currency,
	)

	return &ports.PaymentResponse{
		PaymentID:     paymentID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		Status:        "succeeded",
		TransactionID: transactionID,
	}, nil
}

func (m *MockAdapter) RefundPayment(ctx context.Context, req ports.RefundPaymentRequest) (*ports.RefundResponse, error) {
	payment, exists := m.payments[req.PaymentID]
	if !exists {
		return nil, fmt.Errorf("payment not found: %s", req.PaymentID)
	}

	refundID := "re_" + uuid.New().String()[:14]

	m.logger.Info("Mock: Refunded payment",
		"refund_id", refundID,
		"payment_id", req.PaymentID,
		"amount", req.Amount.String(),
	)

	return &ports.RefundResponse{
		RefundID:      refundID,
		Amount:        req.Amount,
		Currency:      payment.Currency,
		Status:        "succeeded",
		TransactionID: "txn_" + uuid.New().String()[:14],
	}, nil
}

func (m *MockAdapter) CreatePaymentMethod(ctx context.Context, req ports.CreatePaymentMethodRequest) (*ports.PaymentMethodResponse, error) {
	methodID := "pm_" + uuid.New().String()[:14]

	method := &mockPaymentMethod{
		ID:       methodID,
		Customer: req.CustomerID,
		Type:     "card",
		Last4:    "4242",
		Brand:    "visa",
		ExpMonth: 12,
		ExpYear:  2025,
	}

	m.methods[methodID] = method

	m.logger.Info("Mock: Created payment method",
		"method_id", methodID,
		"customer_id", req.CustomerID,
	)

	return &ports.PaymentMethodResponse{
		MethodID:  methodID,
		Type:      method.Type,
		Last4:     method.Last4,
		Brand:     method.Brand,
		ExpMonth:  method.ExpMonth,
		ExpYear:   method.ExpYear,
		IsDefault: req.SetDefault,
	}, nil
}

func (m *MockAdapter) DeletePaymentMethod(ctx context.Context, methodID string) error {
	if _, exists := m.methods[methodID]; !exists {
		return fmt.Errorf("payment method not found: %s", methodID)
	}

	delete(m.methods, methodID)

	m.logger.Info("Mock: Deleted payment method", "method_id", methodID)
	return nil
}

func (m *MockAdapter) GetPaymentMethods(ctx context.Context, customerID string) ([]*ports.PaymentMethodResponse, error) {
	var methods []*ports.PaymentMethodResponse

	for _, method := range m.methods {
		if method.Customer == customerID {
			methods = append(methods, &ports.PaymentMethodResponse{
				MethodID: method.ID,
				Type:     method.Type,
				Last4:    method.Last4,
				Brand:    method.Brand,
				ExpMonth: method.ExpMonth,
				ExpYear:  method.ExpYear,
			})
		}
	}

	m.logger.Info("Mock: Listed payment methods",
		"customer_id", customerID,
		"count", len(methods),
	)

	return methods, nil
}

func (m *MockAdapter) GetInvoice(ctx context.Context, invoiceID string) (*ports.InvoiceResponse, error) {
	// Mock invoice data
	invoice := &ports.InvoiceResponse{
		InvoiceID: invoiceID,
		Number:    "INV-" + invoiceID[:8],
		Amount:    decimal.NewFromFloat(99.99),
		Currency:  "USD",
		Status:    "paid",
		DueDate:   time.Now().AddDate(0, 0, 30),
		Items: []ports.InvoiceItem{
			{
				Description: "Monthly Subscription",
				Amount:      decimal.NewFromFloat(99.99),
				Quantity:    1,
			},
		},
		DownloadURL: fmt.Sprintf("https://example.com/invoices/%s.pdf", invoiceID),
	}

	m.logger.Info("Mock: Retrieved invoice", "invoice_id", invoiceID)
	return invoice, nil
}

func (m *MockAdapter) ProcessWebhook(ctx context.Context, payload []byte, signature string) (*ports.WebhookEvent, error) {
	// Mock webhook processing
	eventID := "evt_" + uuid.New().String()[:14]

	m.logger.Info("Mock: Processed webhook",
		"event_id", eventID,
		"signature", signature[:10]+"...",
	)

	return &ports.WebhookEvent{
		ID:        eventID,
		Type:      "payment_intent.succeeded",
		Data:      payload,
		Timestamp: time.Now(),
	}, nil
}

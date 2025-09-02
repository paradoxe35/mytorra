package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
	"github.com/paradoxe35/torra/services/core/internal/domain/valueobjects"
)

// PaymentGateway defines the interface for payment processing
type PaymentGateway interface {
	// Customer management
	CreateCustomer(ctx context.Context, user *entities.User) (string, error)

	// Subscription management
	CreateSubscription(ctx context.Context, req CreateSubscriptionRequest) (*SubscriptionResponse, error)
	UpdateSubscription(ctx context.Context, req UpdateSubscriptionRequest) (*SubscriptionResponse, error)
	CancelSubscription(ctx context.Context, providerSubscriptionID string, immediate bool) error
	PauseSubscription(ctx context.Context, providerSubscriptionID string) error
	ResumeSubscription(ctx context.Context, providerSubscriptionID string) error

	// Payments
	ChargePayment(ctx context.Context, req ChargePaymentRequest) (*PaymentResponse, error)
	RefundPayment(ctx context.Context, req RefundPaymentRequest) (*RefundResponse, error)

	// Payment methods
	CreatePaymentMethod(ctx context.Context, req CreatePaymentMethodRequest) (*PaymentMethodResponse, error)
	DeletePaymentMethod(ctx context.Context, methodID string) error
	GetPaymentMethods(ctx context.Context, customerID string) ([]*PaymentMethodResponse, error)

	// Invoices
	GetInvoice(ctx context.Context, invoiceID string) (*InvoiceResponse, error)

	// Webhooks
	ProcessWebhook(ctx context.Context, payload []byte, signature string) (*WebhookEvent, error)
}

// PaymentIntentRequest represents a payment intent creation request
type PaymentIntentRequest struct {
	Amount      valueobjects.Money
	CustomerID  string
	Description string
	Metadata    map[string]string
}

// PaymentIntentResponse represents a payment intent response
type PaymentIntentResponse struct {
	ID           string
	ClientSecret string
	Status       string
	Amount       valueobjects.Money
	CreatedAt    time.Time
}

// PaymentConfirmation represents a payment confirmation
type PaymentConfirmation struct {
	ID        string
	Status    string
	Amount    valueobjects.Money
	ChargeID  string
	CreatedAt time.Time
}

// RefundResponse represents a refund response
type RefundResponse struct {
	RefundID      string
	Amount        decimal.Decimal
	Currency      string
	Status        string
	TransactionID string
}

// PaymentMethodRequest represents a payment method creation request
type PaymentMethodRequest struct {
	Type       string
	Card       *CardDetails
	CustomerID string
}

// CardDetails represents credit card information
type CardDetails struct {
	Number   string
	ExpMonth int
	ExpYear  int
	CVC      string
}

// PaymentMethod represents a payment method
type PaymentMethod struct {
	ID      string
	Type    string
	Card    *CardInfo
	Created time.Time
}

// CardInfo represents stored card information
type CardInfo struct {
	Last4    string
	Brand    string
	ExpMonth int
	ExpYear  int
}

// CustomerRequest represents a customer creation/update request
type CustomerRequest struct {
	ID          uuid.UUID
	Email       string
	Name        string
	Phone       string
	Description string
	Metadata    map[string]string
}

// Customer represents a payment gateway customer
type Customer struct {
	ID                   string
	Email                string
	Name                 string
	DefaultPaymentMethod string
	Created              time.Time
}

// CreateSubscriptionRequest represents a subscription creation request
type CreateSubscriptionRequest struct {
	CustomerID     string
	SubscriptionID string
	TierCode       string
	TierName       string
	Amount         decimal.Decimal
	Currency       string
	BillingCycle   entities.SubscriptionBillingCycle
	TrialDays      int
	Metadata       map[string]string
}

// UpdateSubscriptionRequest represents a subscription update request
type UpdateSubscriptionRequest struct {
	ProviderSubscriptionID string
	NewTierCode            string
	NewTierName            string
	NewAmount              decimal.Decimal
	Currency               string
	BillingCycle           entities.SubscriptionBillingCycle
	ProrationBehavior      string
}

// SubscriptionResponse represents a subscription response from payment gateway
type SubscriptionResponse struct {
	ProviderSubscriptionID string
	Status                 string
	CurrentPeriodStart     time.Time
	CurrentPeriodEnd       time.Time
	TrialEnd               *time.Time
}

// PaymentResponse represents a generic payment response
type PaymentResponse struct {
	PaymentID     string
	Amount        decimal.Decimal
	Currency      string
	Status        string
	TransactionID string
}

// ChargePaymentRequest represents a payment charge request
type ChargePaymentRequest struct {
	CustomerID      string
	Amount          decimal.Decimal
	Currency        string
	Description     string
	PaymentMethodID string
	ReferenceID     string
}

// RefundPaymentRequest represents a refund request
type RefundPaymentRequest struct {
	PaymentID string
	Amount    decimal.Decimal
	Reason    string
}

// CreatePaymentMethodRequest represents a payment method creation request
type CreatePaymentMethodRequest struct {
	CustomerID string
	Token      string
	SetDefault bool
}

// PaymentMethodResponse represents a payment method response
type PaymentMethodResponse struct {
	MethodID  string
	Type      string
	Last4     string
	Brand     string
	ExpMonth  int
	ExpYear   int
	IsDefault bool
}

// InvoiceResponse represents an invoice response
type InvoiceResponse struct {
	InvoiceID   string
	Number      string
	Amount      decimal.Decimal
	Currency    string
	Status      string
	DueDate     time.Time
	Items       []InvoiceItem
	DownloadURL string
}

// InvoiceItem represents an item in an invoice
type InvoiceItem struct {
	Description string
	Amount      decimal.Decimal
	Quantity    int
}

// WebhookEvent represents a webhook event from payment gateway
type WebhookEvent struct {
	ID        string
	Type      string
	Data      []byte
	Timestamp time.Time
}

package ports

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/paradoxe35/torra/services/core/internal/domain/valueobjects"
)

// PaymentGateway defines the interface for payment processing
type PaymentGateway interface {
	// Card payments
	CreatePaymentIntent(ctx context.Context, req PaymentIntentRequest) (*PaymentIntentResponse, error)
	ConfirmPayment(ctx context.Context, paymentIntentID string) (*PaymentConfirmation, error)
	CancelPayment(ctx context.Context, paymentIntentID string) error
	RefundPayment(ctx context.Context, paymentID string, amount *valueobjects.Money) (*RefundResponse, error)

	// Payment methods
	CreatePaymentMethod(ctx context.Context, req PaymentMethodRequest) (*PaymentMethod, error)
	AttachPaymentMethod(ctx context.Context, paymentMethodID, customerID string) error
	DetachPaymentMethod(ctx context.Context, paymentMethodID string) error
	ListPaymentMethods(ctx context.Context, customerID string) ([]*PaymentMethod, error)

	// Customer management
	CreateCustomer(ctx context.Context, req CustomerRequest) (*Customer, error)
	UpdateCustomer(ctx context.Context, customerID string, req CustomerRequest) (*Customer, error)
	DeleteCustomer(ctx context.Context, customerID string) error

	// Webhooks
	ValidateWebhook(ctx context.Context, payload []byte, signature string) (bool, error)
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
	ID        string
	Amount    valueobjects.Money
	Status    string
	Reason    string
	CreatedAt time.Time
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

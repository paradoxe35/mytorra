package events

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// BillingPaymentSuccessful event is published when a payment is successful
type BillingPaymentSuccessful struct {
	SubscriptionID uuid.UUID       `json:"subscription_id"`
	BillingCycleID uuid.UUID       `json:"billing_cycle_id"`
	TransactionID  uuid.UUID       `json:"transaction_id"`
	Amount         decimal.Decimal `json:"amount"`
	Currency       string          `json:"currency"`
	ProcessedAt    time.Time       `json:"processed_at"`
}

func (e BillingPaymentSuccessful) EventType() string {
	return "billing.payment_successful"
}

func (e BillingPaymentSuccessful) EventData() interface{} {
	return e
}

// BillingPaymentFailed event is published when a payment fails
type BillingPaymentFailed struct {
	SubscriptionID uuid.UUID       `json:"subscription_id"`
	BillingCycleID uuid.UUID       `json:"billing_cycle_id"`
	TransactionID  uuid.UUID       `json:"transaction_id"`
	Amount         decimal.Decimal `json:"amount"`
	Currency       string          `json:"currency"`
	FailureReason  string          `json:"failure_reason"`
	FailedAt       time.Time       `json:"failed_at"`
}

func (e BillingPaymentFailed) EventType() string {
	return "billing.payment_failed"
}

func (e BillingPaymentFailed) EventData() interface{} {
	return e
}

// BillingPaymentRetryScheduled event is published when a payment retry is scheduled
type BillingPaymentRetryScheduled struct {
	SubscriptionID uuid.UUID `json:"subscription_id"`
	BillingCycleID uuid.UUID `json:"billing_cycle_id"`
	TransactionID  uuid.UUID `json:"transaction_id"`
	RetryCount     int       `json:"retry_count"`
	NextRetryAt    time.Time `json:"next_retry_at"`
}

func (e BillingPaymentRetryScheduled) EventType() string {
	return "billing.payment_retry_scheduled"
}

func (e BillingPaymentRetryScheduled) EventData() interface{} {
	return e
}

// BillingInvoiceGenerated event is published when an invoice is generated
type BillingInvoiceGenerated struct {
	SubscriptionID uuid.UUID       `json:"subscription_id"`
	BillingCycleID uuid.UUID       `json:"billing_cycle_id"`
	InvoiceID      string          `json:"invoice_id"`
	Amount         decimal.Decimal `json:"amount"`
	Currency       string          `json:"currency"`
	DueDate        time.Time       `json:"due_date"`
	GeneratedAt    time.Time       `json:"generated_at"`
}

func (e BillingInvoiceGenerated) EventType() string {
	return "billing.invoice_generated"
}

func (e BillingInvoiceGenerated) EventData() interface{} {
	return e
}

// BillingRefundProcessed event is published when a refund is processed
type BillingRefundProcessed struct {
	SubscriptionID uuid.UUID       `json:"subscription_id"`
	BillingCycleID uuid.UUID       `json:"billing_cycle_id"`
	TransactionID  uuid.UUID       `json:"transaction_id"`
	RefundID       string          `json:"refund_id"`
	Amount         decimal.Decimal `json:"amount"`
	Currency       string          `json:"currency"`
	Reason         string          `json:"reason"`
	ProcessedAt    time.Time       `json:"processed_at"`
}

func (e BillingRefundProcessed) EventType() string {
	return "billing.refund_processed"
}

func (e BillingRefundProcessed) EventData() interface{} {
	return e
}

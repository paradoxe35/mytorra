package entities

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// PaymentTransaction represents a payment transaction for a subscription
type PaymentTransaction struct {
	ID                   uuid.UUID       `db:"id" json:"id"`
	SubscriptionID       uuid.UUID       `db:"subscription_id" json:"subscription_id"`
	BillingCycleID       *uuid.UUID      `db:"billing_cycle_id" json:"billing_cycle_id,omitempty"`
	Amount               decimal.Decimal `db:"amount" json:"amount"`
	Currency             string          `db:"currency" json:"currency"`
	Type                 string          `db:"type" json:"type"` // "subscription", "upgrade", "downgrade", "refund"
	Status               string          `db:"status" json:"status"`
	GatewayTransactionID string          `db:"gateway_transaction_id" json:"gateway_transaction_id"`
	GatewayResponse      json.RawMessage `db:"gateway_response" json:"gateway_response,omitempty"`
	ProcessedAt          *time.Time      `db:"processed_at" json:"processed_at,omitempty"`
	FailureReason        *string         `db:"failure_reason" json:"failure_reason,omitempty"`
	RetryCount           int             `db:"retry_count" json:"retry_count"`
	CreatedAt            time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time       `db:"updated_at" json:"updated_at"`
}

// PaymentTransactionType represents the type of payment transaction
type PaymentTransactionType string

const (
	PaymentTransactionTypeSubscription PaymentTransactionType = "subscription"
	PaymentTransactionTypeUpgrade      PaymentTransactionType = "upgrade"
	PaymentTransactionTypeDowngrade    PaymentTransactionType = "downgrade"
	PaymentTransactionTypeRefund       PaymentTransactionType = "refund"
	PaymentTransactionTypeProration    PaymentTransactionType = "proration"
)

// PaymentTransactionStatus represents the status of a payment transaction
type PaymentTransactionStatus string

const (
	PaymentTransactionStatusPending    PaymentTransactionStatus = "pending"
	PaymentTransactionStatusProcessing PaymentTransactionStatus = "processing"
	PaymentTransactionStatusSucceeded  PaymentTransactionStatus = "succeeded"
	PaymentTransactionStatusFailed     PaymentTransactionStatus = "failed"
	PaymentTransactionStatusCanceled   PaymentTransactionStatus = "canceled"
	PaymentTransactionStatusRefunded   PaymentTransactionStatus = "refunded"
)

// IsSuccessful checks if the transaction was successful
func (p *PaymentTransaction) IsSuccessful() bool {
	return p.Status == string(PaymentTransactionStatusSucceeded)
}

// IsFailed checks if the transaction failed
func (p *PaymentTransaction) IsFailed() bool {
	return p.Status == string(PaymentTransactionStatusFailed)
}

// IsPending checks if the transaction is pending
func (p *PaymentTransaction) IsPending() bool {
	return p.Status == string(PaymentTransactionStatusPending) ||
		p.Status == string(PaymentTransactionStatusProcessing)
}

// CanRetry checks if the transaction can be retried
func (p *PaymentTransaction) CanRetry() bool {
	maxRetries := 3
	return p.IsFailed() && p.RetryCount < maxRetries
}

// MarkAsSuccessful marks the transaction as successful
func (p *PaymentTransaction) MarkAsSuccessful(gatewayTransactionID string) error {
	if p.IsSuccessful() {
		return fmt.Errorf("transaction is already successful")
	}

	now := time.Now()
	p.Status = string(PaymentTransactionStatusSucceeded)
	p.GatewayTransactionID = gatewayTransactionID
	p.ProcessedAt = &now
	p.UpdatedAt = now

	return nil
}

// MarkAsFailed marks the transaction as failed
func (p *PaymentTransaction) MarkAsFailed(reason string) error {
	if p.IsSuccessful() {
		return fmt.Errorf("cannot mark successful transaction as failed")
	}

	p.Status = string(PaymentTransactionStatusFailed)
	p.FailureReason = &reason
	p.UpdatedAt = time.Now()

	return nil
}

// IncrementRetryCount increments the retry count
func (p *PaymentTransaction) IncrementRetryCount() {
	p.RetryCount++
	p.UpdatedAt = time.Now()
}

// SetGatewayResponse sets the gateway response
func (p *PaymentTransaction) SetGatewayResponse(response interface{}) error {
	responseBytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal gateway response: %w", err)
	}

	p.GatewayResponse = responseBytes
	p.UpdatedAt = time.Now()

	return nil
}

// Validate validates the payment transaction
func (p *PaymentTransaction) Validate() error {
	if p.SubscriptionID == uuid.Nil {
		return fmt.Errorf("subscription ID is required")
	}

	if p.Amount.LessThan(decimal.Zero) {
		return fmt.Errorf("amount cannot be negative")
	}

	if p.Currency == "" {
		return fmt.Errorf("currency is required")
	}

	if p.Type == "" {
		return fmt.Errorf("transaction type is required")
	}

	if p.Status == "" {
		return fmt.Errorf("transaction status is required")
	}

	// Validate transaction type
	validTypes := map[string]bool{
		string(PaymentTransactionTypeSubscription): true,
		string(PaymentTransactionTypeUpgrade):      true,
		string(PaymentTransactionTypeDowngrade):    true,
		string(PaymentTransactionTypeRefund):       true,
		string(PaymentTransactionTypeProration):    true,
	}

	if !validTypes[p.Type] {
		return fmt.Errorf("invalid transaction type: %s", p.Type)
	}

	// Validate status
	validStatuses := map[string]bool{
		string(PaymentTransactionStatusPending):    true,
		string(PaymentTransactionStatusProcessing): true,
		string(PaymentTransactionStatusSucceeded):  true,
		string(PaymentTransactionStatusFailed):     true,
		string(PaymentTransactionStatusCanceled):   true,
		string(PaymentTransactionStatusRefunded):   true,
	}

	if !validStatuses[p.Status] {
		return fmt.Errorf("invalid transaction status: %s", p.Status)
	}

	return nil
}

package entities

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// BillingCycle represents a billing period for a subscription
type BillingCycle struct {
	ID             uuid.UUID       `db:"id" json:"id"`
	SubscriptionID uuid.UUID       `db:"subscription_id" json:"subscription_id"`
	StartDate      time.Time       `db:"start_date" json:"start_date"`
	EndDate        time.Time       `db:"end_date" json:"end_date"`
	Amount         decimal.Decimal `db:"amount" json:"amount"`
	Currency       string          `db:"currency" json:"currency"`
	Status         string          `db:"status" json:"status"`
	InvoiceID      *string         `db:"invoice_id" json:"invoice_id,omitempty"`
	PaidAt         *time.Time      `db:"paid_at" json:"paid_at,omitempty"`
	CreatedAt      time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time       `db:"updated_at" json:"updated_at"`
}

// BillingCycleStatus represents the status of a billing cycle
type BillingCycleStatus string

const (
	BillingCycleStatusPending    BillingCycleStatus = "pending"
	BillingCycleStatusProcessing BillingCycleStatus = "processing"
	BillingCycleStatusPaid       BillingCycleStatus = "paid"
	BillingCycleStatusFailed     BillingCycleStatus = "failed"
	BillingCycleStatusCanceled   BillingCycleStatus = "canceled"
	BillingCycleStatusRefunded   BillingCycleStatus = "refunded"
)

// IsCurrent checks if the billing cycle is the current one
func (b *BillingCycle) IsCurrent() bool {
	now := time.Now()
	return now.After(b.StartDate) && now.Before(b.EndDate)
}

// IsOverdue checks if the billing cycle is overdue for payment
func (b *BillingCycle) IsOverdue() bool {
	if b.Status == string(BillingCycleStatusPaid) {
		return false
	}

	// Consider overdue if end date has passed and not paid
	return time.Now().After(b.EndDate) && b.PaidAt == nil
}

// MarkAsPaid marks the billing cycle as paid
func (b *BillingCycle) MarkAsPaid(transactionID string) error {
	if b.Status == string(BillingCycleStatusPaid) {
		return fmt.Errorf("billing cycle is already marked as paid")
	}

	if b.Status == string(BillingCycleStatusCanceled) {
		return fmt.Errorf("cannot mark canceled billing cycle as paid")
	}

	now := time.Now()
	b.Status = string(BillingCycleStatusPaid)
	b.PaidAt = &now
	if transactionID != "" {
		b.InvoiceID = &transactionID
	}
	b.UpdatedAt = now

	return nil
}

// MarkAsFailed marks the billing cycle as failed
func (b *BillingCycle) MarkAsFailed() error {
	if b.Status == string(BillingCycleStatusPaid) {
		return fmt.Errorf("cannot mark paid billing cycle as failed")
	}

	b.Status = string(BillingCycleStatusFailed)
	b.UpdatedAt = time.Now()

	return nil
}

// Cancel cancels the billing cycle
func (b *BillingCycle) Cancel() error {
	if b.Status == string(BillingCycleStatusPaid) {
		return fmt.Errorf("cannot cancel paid billing cycle")
	}

	b.Status = string(BillingCycleStatusCanceled)
	b.UpdatedAt = time.Now()

	return nil
}

// Refund marks the billing cycle as refunded
func (b *BillingCycle) Refund() error {
	if b.Status != string(BillingCycleStatusPaid) {
		return fmt.Errorf("can only refund paid billing cycles")
	}

	b.Status = string(BillingCycleStatusRefunded)
	b.UpdatedAt = time.Now()

	return nil
}

// GetDaysInCycle returns the number of days in the billing cycle
func (b *BillingCycle) GetDaysInCycle() int {
	return int(b.EndDate.Sub(b.StartDate).Hours() / 24)
}

// GetDaysRemaining returns the number of days remaining in the billing cycle
func (b *BillingCycle) GetDaysRemaining() int {
	if time.Now().After(b.EndDate) {
		return 0
	}

	return int(time.Until(b.EndDate).Hours() / 24)
}

// Validate validates the billing cycle
func (b *BillingCycle) Validate() error {
	if b.SubscriptionID == uuid.Nil {
		return fmt.Errorf("subscription ID is required")
	}

	if b.StartDate.After(b.EndDate) {
		return fmt.Errorf("start date must be before end date")
	}

	if b.Amount.LessThan(decimal.Zero) {
		return fmt.Errorf("amount cannot be negative")
	}

	if b.Currency == "" {
		return fmt.Errorf("currency is required")
	}

	if b.Status == "" {
		return fmt.Errorf("status is required")
	}

	return nil
}

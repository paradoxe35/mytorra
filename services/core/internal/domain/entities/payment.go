package entities

import (
	"time"

	"github.com/google/uuid"

	"github.com/paradoxe35/mytorra/services/core/internal/domain/valueobjects"
)

type Payment struct {
	ID     uuid.UUID          `db:"id" json:"id"`
	UserID uuid.UUID          `db:"user_id" json:"user_id"`
	Amount valueobjects.Money `db:"-" json:"amount"`
	Type   PaymentType        `db:"type" json:"type"`
	Method PaymentMethod      `db:"method" json:"method"`
	Status PaymentStatus      `db:"status" json:"status"`

	// Provider details
	ProviderID   string `db:"provider_id" json:"provider_id"`
	ProviderName string `db:"provider_name" json:"provider_name"`

	// Transaction details
	Description string                 `db:"description" json:"description"`
	Metadata    map[string]interface{} `db:"metadata" json:"metadata"`

	// Refund information
	RefundedAmount *valueobjects.Money `db:"-" json:"refunded_amount,omitempty"`
	RefundReason   *string             `db:"refund_reason" json:"refund_reason,omitempty"`

	// Commission
	Commission     *valueobjects.Money `db:"-" json:"commission,omitempty"`
	CommissionRate float64             `db:"commission_rate" json:"commission_rate"`

	// Timestamps
	ProcessedAt *time.Time `db:"processed_at" json:"processed_at,omitempty"`
	FailedAt    *time.Time `db:"failed_at" json:"failed_at,omitempty"`
	RefundedAt  *time.Time `db:"refunded_at" json:"refunded_at,omitempty"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}

type (
	PaymentType   string
	PaymentMethod string
	PaymentStatus string
)

const (
	PaymentTypePurchase     PaymentType = "purchase"
	PaymentTypeSubscription PaymentType = "subscription"
	PaymentTypeWalletTopUp  PaymentType = "wallet_topup"
	PaymentTypeCommission   PaymentType = "commission"

	PaymentMethodCard         PaymentMethod = "card"
	PaymentMethodWallet       PaymentMethod = "wallet"
	PaymentMethodMobileMoney  PaymentMethod = "mobile_money"
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"

	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusProcessing PaymentStatus = "processing"
	PaymentStatusCompleted  PaymentStatus = "completed"
	PaymentStatusFailed     PaymentStatus = "failed"
	PaymentStatusRefunded   PaymentStatus = "refunded"
)

// Domain methods
func (p *Payment) CanBeRefunded() bool {
	return p.Status == PaymentStatusCompleted && p.RefundedAmount == nil
}

func (p *Payment) CalculateCommission(rate float64) valueobjects.Money {
	return p.Amount.Multiply(rate)
}

func (p *Payment) MarkAsCompleted() {
	now := time.Now()
	p.Status = PaymentStatusCompleted
	p.ProcessedAt = &now
	p.UpdatedAt = now
}

func (p *Payment) MarkAsFailed(reason string) {
	now := time.Now()
	p.Status = PaymentStatusFailed
	p.FailedAt = &now
	p.UpdatedAt = now
	if p.Metadata == nil {
		p.Metadata = make(map[string]interface{})
	}
	p.Metadata["failure_reason"] = reason
}

package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
)

// BillingCycleModel represents the database model for billing cycles
type BillingCycleModel struct {
	ID             uuid.UUID       `db:"id"`
	SubscriptionID uuid.UUID       `db:"subscription_id"`
	StartDate      time.Time       `db:"start_date"`
	EndDate        time.Time       `db:"end_date"`
	Amount         decimal.Decimal `db:"amount"`
	Currency       string          `db:"currency"`
	Status         string          `db:"status"`
	InvoiceID      *string         `db:"invoice_id"`
	PaidAt         *time.Time      `db:"paid_at"`
	CreatedAt      time.Time       `db:"created_at"`
	UpdatedAt      time.Time       `db:"updated_at"`
}

// ToEntity converts the database model to a domain entity
func (m *BillingCycleModel) ToEntity() *entities.BillingCycle {
	return &entities.BillingCycle{
		ID:             m.ID,
		SubscriptionID: m.SubscriptionID,
		StartDate:      m.StartDate,
		EndDate:        m.EndDate,
		Amount:         m.Amount,
		Currency:       m.Currency,
		Status:         m.Status,
		InvoiceID:      m.InvoiceID,
		PaidAt:         m.PaidAt,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

// FromEntity converts a domain entity to a database model
func FromBillingCycleEntity(e *entities.BillingCycle) *BillingCycleModel {
	return &BillingCycleModel{
		ID:             e.ID,
		SubscriptionID: e.SubscriptionID,
		StartDate:      e.StartDate,
		EndDate:        e.EndDate,
		Amount:         e.Amount,
		Currency:       e.Currency,
		Status:         e.Status,
		InvoiceID:      e.InvoiceID,
		PaidAt:         e.PaidAt,
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
	}
}

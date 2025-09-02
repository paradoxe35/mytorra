package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/paradoxe35/torra/services/core/internal/domain/valueobjects"
)

type Wallet struct {
	ID       uuid.UUID          `db:"id" json:"id"`
	UserID   uuid.UUID          `db:"user_id" json:"user_id"`
	Balance  valueobjects.Money `db:"-" json:"balance"`
	Currency string             `db:"currency" json:"currency"`
	Status   WalletStatus       `db:"status" json:"status"`

	// Limits
	DailyLimit   float64 `db:"daily_limit" json:"daily_limit"`
	MonthlyLimit float64 `db:"monthly_limit" json:"monthly_limit"`

	// Timestamps
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type WalletStatus string

const (
	WalletStatusActive  WalletStatus = "active"
	WalletStatusFrozen  WalletStatus = "frozen"
	WalletStatusBlocked WalletStatus = "blocked"
)

type WalletTransaction struct {
	ID            uuid.UUID              `db:"id" json:"id"`
	WalletID      uuid.UUID              `db:"wallet_id" json:"wallet_id"`
	Type          TransactionType        `db:"type" json:"type"`
	Amount        valueobjects.Money     `db:"-" json:"amount"`
	BalanceBefore valueobjects.Money     `db:"-" json:"balance_before"`
	BalanceAfter  valueobjects.Money     `db:"-" json:"balance_after"`
	Description   string                 `db:"description" json:"description"`
	ReferenceID   *uuid.UUID             `db:"reference_id" json:"reference_id,omitempty"`
	ReferenceType *string                `db:"reference_type" json:"reference_type,omitempty"`
	Metadata      map[string]interface{} `db:"metadata" json:"metadata"`
	CreatedAt     time.Time              `db:"created_at" json:"created_at"`
}

type TransactionType string

const (
	TransactionTypeCredit     TransactionType = "credit"
	TransactionTypeDebit      TransactionType = "debit"
	TransactionTypeTopUp      TransactionType = "topup"
	TransactionTypeWithdrawal TransactionType = "withdrawal"
	TransactionTypeRefund     TransactionType = "refund"
	TransactionTypeCommission TransactionType = "commission"
)

// Domain methods
func (w *Wallet) CanWithdraw(amount float64) bool {
	return w.Status == WalletStatusActive && w.Balance.Amount.GreaterThanOrEqual(decimal.NewFromFloat(amount))
}

func (w *Wallet) IsActive() bool {
	return w.Status == WalletStatusActive
}

func (w *Wallet) AddFunds(amount valueobjects.Money) error {
	newBalance, err := w.Balance.Add(amount)
	if err != nil {
		return err
	}
	w.Balance = newBalance
	w.UpdatedAt = time.Now()
	return nil
}

func (w *Wallet) DeductFunds(amount valueobjects.Money) error {
	newBalance, err := w.Balance.Subtract(amount)
	if err != nil {
		return err
	}
	w.Balance = newBalance
	w.UpdatedAt = time.Now()
	return nil
}

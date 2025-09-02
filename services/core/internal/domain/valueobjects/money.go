package valueobjects

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

type Money struct {
	Amount   decimal.Decimal
	Currency string
}

func NewMoney(amount float64, currency string) (Money, error) {
	decAmount := decimal.NewFromFloat(amount)
	if decAmount.LessThan(decimal.Zero) {
		return Money{}, errors.New("amount cannot be negative")
	}
	if currency == "" {
		return Money{}, errors.New("currency is required")
	}
	return Money{
		Amount:   decAmount,
		Currency: currency,
	}, nil
}

func NewMoneyFromDecimal(amount decimal.Decimal, currency string) (Money, error) {
	if amount.LessThan(decimal.Zero) {
		return Money{}, errors.New("amount cannot be negative")
	}
	if currency == "" {
		return Money{}, errors.New("currency is required")
	}
	return Money{
		Amount:   amount,
		Currency: currency,
	}, nil
}

func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("cannot add different currencies: %s and %s", m.Currency, other.Currency)
	}
	return Money{
		Amount:   m.Amount.Add(other.Amount),
		Currency: m.Currency,
	}, nil
}

func (m Money) Subtract(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("cannot subtract different currencies: %s and %s", m.Currency, other.Currency)
	}
	if m.Amount.LessThan(other.Amount) {
		return Money{}, errors.New("insufficient amount")
	}
	return Money{
		Amount:   m.Amount.Sub(other.Amount),
		Currency: m.Currency,
	}, nil
}

func (m Money) Multiply(factor float64) Money {
	return Money{
		Amount:   m.Amount.Mul(decimal.NewFromFloat(factor)),
		Currency: m.Currency,
	}
}

func (m Money) MultiplyDecimal(factor decimal.Decimal) Money {
	return Money{
		Amount:   m.Amount.Mul(factor),
		Currency: m.Currency,
	}
}

func (m Money) String() string {
	return fmt.Sprintf("%s %s", m.Amount.StringFixed(2), m.Currency)
}

func (m Money) Equal(other Money) bool {
	return m.Amount.Equal(other.Amount) && m.Currency == other.Currency
}

func (m Money) IsZero() bool {
	return m.Amount.IsZero()
}

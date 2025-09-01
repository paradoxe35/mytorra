package valueobjects

import (
	"errors"
	"fmt"
)

type Money struct {
	Amount   float64
	Currency string
}

func NewMoney(amount float64, currency string) (Money, error) {
	if amount < 0 {
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
		Amount:   m.Amount + other.Amount,
		Currency: m.Currency,
	}, nil
}

func (m Money) Subtract(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("cannot subtract different currencies: %s and %s", m.Currency, other.Currency)
	}
	if m.Amount < other.Amount {
		return Money{}, errors.New("insufficient amount")
	}
	return Money{
		Amount:   m.Amount - other.Amount,
		Currency: m.Currency,
	}, nil
}

func (m Money) Multiply(factor float64) Money {
	return Money{
		Amount:   m.Amount * factor,
		Currency: m.Currency,
	}
}

func (m Money) String() string {
	return fmt.Sprintf("%.2f %s", m.Amount, m.Currency)
}

func (m Money) Equal(other Money) bool {
	return m.Amount == other.Amount && m.Currency == other.Currency
}

func (m Money) IsZero() bool {
	return m.Amount == 0
}

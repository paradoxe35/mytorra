package entities

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBillingCycle_IsCurrent(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		startDate time.Time
		endDate   time.Time
		want      bool
	}{
		{
			name:      "Current cycle",
			startDate: now.AddDate(0, 0, -15),
			endDate:   now.AddDate(0, 0, 15),
			want:      true,
		},
		{
			name:      "Past cycle",
			startDate: now.AddDate(0, -2, 0),
			endDate:   now.AddDate(0, -1, 0),
			want:      false,
		},
		{
			name:      "Future cycle",
			startDate: now.AddDate(0, 1, 0),
			endDate:   now.AddDate(0, 2, 0),
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BillingCycle{
				StartDate: tt.startDate,
				EndDate:   tt.endDate,
			}
			assert.Equal(t, tt.want, b.IsCurrent())
		})
	}
}

func TestBillingCycle_IsOverdue(t *testing.T) {
	now := time.Now()
	past := now.AddDate(0, 0, -10)
	future := now.AddDate(0, 0, 10)

	tests := []struct {
		name    string
		status  string
		endDate time.Time
		paidAt  *time.Time
		want    bool
	}{
		{
			name:    "Paid cycle",
			status:  string(BillingCycleStatusPaid),
			endDate: past,
			paidAt:  &now,
			want:    false,
		},
		{
			name:    "Unpaid past cycle",
			status:  string(BillingCycleStatusPending),
			endDate: past,
			paidAt:  nil,
			want:    true,
		},
		{
			name:    "Unpaid future cycle",
			status:  string(BillingCycleStatusPending),
			endDate: future,
			paidAt:  nil,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BillingCycle{
				Status:  tt.status,
				EndDate: tt.endDate,
				PaidAt:  tt.paidAt,
			}
			assert.Equal(t, tt.want, b.IsOverdue())
		})
	}
}

func TestBillingCycle_MarkAsPaid(t *testing.T) {
	t.Run("Mark pending cycle as paid", func(t *testing.T) {
		b := &BillingCycle{
			Status: string(BillingCycleStatusPending),
		}

		err := b.MarkAsPaid("ch_123456")
		require.NoError(t, err)
		assert.Equal(t, string(BillingCycleStatusPaid), b.Status)
		assert.NotNil(t, b.PaidAt)
		assert.NotNil(t, b.InvoiceID)
		assert.Equal(t, "ch_123456", *b.InvoiceID)
	})

	t.Run("Mark already paid cycle", func(t *testing.T) {
		b := &BillingCycle{
			Status: string(BillingCycleStatusPaid),
		}

		err := b.MarkAsPaid("ch_123456")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already marked as paid")
	})

	t.Run("Mark canceled cycle as paid", func(t *testing.T) {
		b := &BillingCycle{
			Status: string(BillingCycleStatusCanceled),
		}

		err := b.MarkAsPaid("ch_123456")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot mark canceled")
	})
}

func TestBillingCycle_MarkAsFailed(t *testing.T) {
	t.Run("Mark pending cycle as failed", func(t *testing.T) {
		b := &BillingCycle{
			Status: string(BillingCycleStatusPending),
		}

		err := b.MarkAsFailed()
		require.NoError(t, err)
		assert.Equal(t, string(BillingCycleStatusFailed), b.Status)
	})

	t.Run("Cannot mark paid cycle as failed", func(t *testing.T) {
		b := &BillingCycle{
			Status: string(BillingCycleStatusPaid),
		}

		err := b.MarkAsFailed()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot mark paid")
	})
}

func TestBillingCycle_Cancel(t *testing.T) {
	t.Run("Cancel pending cycle", func(t *testing.T) {
		b := &BillingCycle{
			Status: string(BillingCycleStatusPending),
		}

		err := b.Cancel()
		require.NoError(t, err)
		assert.Equal(t, string(BillingCycleStatusCanceled), b.Status)
	})

	t.Run("Cannot cancel paid cycle", func(t *testing.T) {
		b := &BillingCycle{
			Status: string(BillingCycleStatusPaid),
		}

		err := b.Cancel()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot cancel paid")
	})
}

func TestBillingCycle_Refund(t *testing.T) {
	t.Run("Refund paid cycle", func(t *testing.T) {
		b := &BillingCycle{
			Status: string(BillingCycleStatusPaid),
		}

		err := b.Refund()
		require.NoError(t, err)
		assert.Equal(t, string(BillingCycleStatusRefunded), b.Status)
	})

	t.Run("Cannot refund unpaid cycle", func(t *testing.T) {
		b := &BillingCycle{
			Status: string(BillingCycleStatusPending),
		}

		err := b.Refund()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can only refund paid")
	})
}

func TestBillingCycle_GetDaysInCycle(t *testing.T) {
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	b := &BillingCycle{
		StartDate: startDate,
		EndDate:   endDate,
	}

	days := b.GetDaysInCycle()
	assert.Equal(t, 30, days)
}

func TestBillingCycle_GetDaysRemaining(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		endDate time.Time
		want    int
	}{
		{
			name:    "Future end date",
			endDate: now.AddDate(0, 0, 10),
			want:    10,
		},
		{
			name:    "Past end date",
			endDate: now.AddDate(0, 0, -10),
			want:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BillingCycle{
				EndDate: tt.endDate,
			}

			days := b.GetDaysRemaining()
			// Allow for some variance due to time passing during test
			assert.True(t, days >= tt.want-1 && days <= tt.want+1)
		})
	}
}

func TestBillingCycle_Validate(t *testing.T) {
	validDate := time.Now()

	tests := []struct {
		name    string
		cycle   *BillingCycle
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid billing cycle",
			cycle: &BillingCycle{
				SubscriptionID: uuid.New(),
				StartDate:      validDate,
				EndDate:        validDate.AddDate(0, 1, 0),
				Amount:         decimal.NewFromFloat(29.99),
				Currency:       "USD",
				Status:         string(BillingCycleStatusPending),
			},
			wantErr: false,
		},
		{
			name: "Missing subscription ID",
			cycle: &BillingCycle{
				StartDate: validDate,
				EndDate:   validDate.AddDate(0, 1, 0),
				Amount:    decimal.NewFromFloat(29.99),
				Currency:  "USD",
				Status:    string(BillingCycleStatusPending),
			},
			wantErr: true,
			errMsg:  "subscription ID is required",
		},
		{
			name: "Invalid date range",
			cycle: &BillingCycle{
				SubscriptionID: uuid.New(),
				StartDate:      validDate.AddDate(0, 1, 0),
				EndDate:        validDate,
				Amount:         decimal.NewFromFloat(29.99),
				Currency:       "USD",
				Status:         string(BillingCycleStatusPending),
			},
			wantErr: true,
			errMsg:  "start date must be before end date",
		},
		{
			name: "Negative amount",
			cycle: &BillingCycle{
				SubscriptionID: uuid.New(),
				StartDate:      validDate,
				EndDate:        validDate.AddDate(0, 1, 0),
				Amount:         decimal.NewFromFloat(-29.99),
				Currency:       "USD",
				Status:         string(BillingCycleStatusPending),
			},
			wantErr: true,
			errMsg:  "amount cannot be negative",
		},
		{
			name: "Missing currency",
			cycle: &BillingCycle{
				SubscriptionID: uuid.New(),
				StartDate:      validDate,
				EndDate:        validDate.AddDate(0, 1, 0),
				Amount:         decimal.NewFromFloat(29.99),
				Status:         string(BillingCycleStatusPending),
			},
			wantErr: true,
			errMsg:  "currency is required",
		},
		{
			name: "Missing status",
			cycle: &BillingCycle{
				SubscriptionID: uuid.New(),
				StartDate:      validDate,
				EndDate:        validDate.AddDate(0, 1, 0),
				Amount:         decimal.NewFromFloat(29.99),
				Currency:       "USD",
			},
			wantErr: true,
			errMsg:  "status is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cycle.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

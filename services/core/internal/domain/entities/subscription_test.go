package entities

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/paradoxe35/mytorra/services/core/internal/domain/valueobjects"
)

func TestSubscription_IsActive(t *testing.T) {
	tests := []struct {
		name   string
		status SubscriptionStatus
		want   bool
	}{
		{"Active subscription", SubscriptionStatusActive, true},
		{"Trialing subscription", SubscriptionStatusTrialing, true},
		{"Paused subscription", SubscriptionStatusPaused, false},
		{"Canceled subscription", SubscriptionStatusCanceled, false},
		{"Expired subscription", SubscriptionStatusExpired, false},
		{"Past due subscription", SubscriptionStatusPastDue, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Subscription{
				Status: tt.status,
			}
			assert.Equal(t, tt.want, s.IsActive())
		})
	}
}

func TestSubscription_IsInTrial(t *testing.T) {
	now := time.Now()
	future := now.AddDate(0, 0, 7)
	past := now.AddDate(0, 0, -7)

	tests := []struct {
		name        string
		isTrialing  bool
		trialEndsAt *time.Time
		want        bool
	}{
		{"In trial with future end date", true, &future, true},
		{"In trial with past end date", true, &past, false},
		{"Not in trial", false, nil, false},
		{"Not in trial with end date", false, &future, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Subscription{
				IsTrialing:  tt.isTrialing,
				TrialEndsAt: tt.trialEndsAt,
			}
			assert.Equal(t, tt.want, s.IsInTrial())
		})
	}
}

func TestSubscription_Cancel(t *testing.T) {
	t.Run("Cancel active subscription", func(t *testing.T) {
		s := &Subscription{
			Status: SubscriptionStatusActive,
		}

		err := s.Cancel("Customer request")
		require.NoError(t, err)
		assert.Equal(t, SubscriptionStatusCanceled, s.Status)
		assert.NotNil(t, s.CanceledAt)
		assert.NotNil(t, s.CancelReason)
		assert.Equal(t, "Customer request", *s.CancelReason)
	})

	t.Run("Cancel already canceled subscription", func(t *testing.T) {
		s := &Subscription{
			Status: SubscriptionStatusCanceled,
		}

		err := s.Cancel("Customer request")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already canceled")
	})
}

func TestSubscription_ExtendTrial(t *testing.T) {
	t.Run("Extend valid trial", func(t *testing.T) {
		trialEnd := time.Now().AddDate(0, 0, 7)
		s := &Subscription{
			IsTrialing:  true,
			TrialEndsAt: &trialEnd,
		}

		err := s.ExtendTrial(7)
		require.NoError(t, err)
		assert.True(t, s.TrialEndsAt.After(trialEnd))
	})

	t.Run("Extend non-trial subscription", func(t *testing.T) {
		s := &Subscription{
			IsTrialing: false,
		}

		err := s.ExtendTrial(7)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not in trial")
	})

	t.Run("Extend trial without end date", func(t *testing.T) {
		s := &Subscription{
			IsTrialing:  true,
			TrialEndsAt: nil,
		}

		err := s.ExtendTrial(7)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "trial end date not set")
	})
}

func TestSubscription_ResetUsage(t *testing.T) {
	s := &Subscription{
		CurrentUsage: map[string]int{
			"api_calls": 500,
			"storage":   10,
		},
	}

	err := s.ResetUsage()
	require.NoError(t, err)
	assert.Empty(t, s.CurrentUsage)
}

func TestSubscription_IncrementUsage(t *testing.T) {
	s := &Subscription{
		CurrentUsage: map[string]int{
			"api_calls": 100,
		},
	}

	s.IncrementUsage("api_calls")
	assert.Equal(t, 101, s.CurrentUsage["api_calls"])

	s.IncrementUsage("new_feature")
	assert.Equal(t, 1, s.CurrentUsage["new_feature"])
}

func TestSubscription_IsWithinUsageLimits(t *testing.T) {
	s := &Subscription{
		UsageLimits: map[string]int{
			"api_calls": 1000,
			"storage":   -1, // Unlimited
		},
		CurrentUsage: map[string]int{
			"api_calls": 500,
			"storage":   1000,
		},
	}

	assert.True(t, s.IsWithinUsageLimits("api_calls"))
	assert.True(t, s.IsWithinUsageLimits("storage")) // Unlimited
	assert.True(t, s.IsWithinUsageLimits("unknown")) // No limit set

	s.CurrentUsage["api_calls"] = 1000
	assert.False(t, s.IsWithinUsageLimits("api_calls"))
}

func TestSubscription_DaysUntilRenewal(t *testing.T) {
	t.Run("With next billing date", func(t *testing.T) {
		nextBilling := time.Now().AddDate(0, 0, 15)
		s := &Subscription{
			NextBillingDate: &nextBilling,
		}

		days := s.DaysUntilRenewal()
		assert.True(t, days >= 14 && days <= 15)
	})

	t.Run("Without next billing date", func(t *testing.T) {
		s := &Subscription{
			NextBillingDate: nil,
		}

		days := s.DaysUntilRenewal()
		assert.Equal(t, -1, days)
	})
}

func TestSubscription_Pause(t *testing.T) {
	s := &Subscription{
		Status: SubscriptionStatusActive,
	}

	s.Pause()
	assert.Equal(t, SubscriptionStatusPaused, s.Status)
	assert.NotNil(t, s.PausedAt)
}

func TestSubscription_Resume(t *testing.T) {
	pausedAt := time.Now()
	s := &Subscription{
		Status:   SubscriptionStatusPaused,
		PausedAt: &pausedAt,
	}

	s.Resume()
	assert.Equal(t, SubscriptionStatusActive, s.Status)
	assert.Nil(t, s.PausedAt)
	assert.Nil(t, s.ResumeAt)
}

func TestSubscription_Validate(t *testing.T) {
	tests := []struct {
		name    string
		sub     *Subscription
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid subscription",
			sub: &Subscription{
				UserID:             uuid.New(),
				EntityID:           uuid.New(),
				EntityType:         "service",
				PlanName:           "Basic",
				CurrentPeriodStart: time.Now(),
				CurrentPeriodEnd:   time.Now().AddDate(0, 1, 0),
			},
			wantErr: false,
		},
		{
			name: "Missing user ID",
			sub: &Subscription{
				EntityID:           uuid.New(),
				EntityType:         "service",
				PlanName:           "Basic",
				CurrentPeriodStart: time.Now(),
				CurrentPeriodEnd:   time.Now().AddDate(0, 1, 0),
			},
			wantErr: true,
			errMsg:  "user ID is required",
		},
		{
			name: "Invalid period dates",
			sub: &Subscription{
				UserID:             uuid.New(),
				EntityID:           uuid.New(),
				EntityType:         "service",
				PlanName:           "Basic",
				CurrentPeriodStart: time.Now().AddDate(0, 1, 0),
				CurrentPeriodEnd:   time.Now(),
			},
			wantErr: true,
			errMsg:  "current period start must be before end",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.sub.Validate()
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

func TestSubscription_CalculateProration(t *testing.T) {
	now := time.Now()
	s := &Subscription{
		Status: SubscriptionStatusActive,
		Price: valueobjects.Money{
			Amount:   decimal.NewFromFloat(100),
			Currency: "USD",
		},
		CurrentPeriodStart: now.AddDate(0, 0, -15),
		CurrentPeriodEnd:   now.AddDate(0, 0, 15),
	}

	newPrice := decimal.NewFromFloat(200)
	proration := s.CalculateProration(newPrice)

	// Should be approximately 50 (half of the difference between 200 and 100)
	assert.True(t, proration.GreaterThan(decimal.NewFromFloat(45)))
	assert.True(t, proration.LessThan(decimal.NewFromFloat(55)))
}

func TestSubscription_CanChangeTier(t *testing.T) {
	tests := []struct {
		name    string
		sub     *Subscription
		newTier valueobjects.SubscriptionTier
		wantErr bool
	}{
		{
			name: "Can change tier",
			sub: &Subscription{
				Status:     SubscriptionStatusActive,
				Tier:       valueobjects.TierBasic,
				IsTrialing: false,
			},
			newTier: valueobjects.TierProfessional,
			wantErr: false,
		},
		{
			name: "Cannot change canceled subscription",
			sub: &Subscription{
				Status: SubscriptionStatusCanceled,
				Tier:   valueobjects.TierBasic,
			},
			newTier: valueobjects.TierProfessional,
			wantErr: true,
		},
		{
			name: "Cannot change during trial",
			sub: &Subscription{
				Status:     SubscriptionStatusActive,
				Tier:       valueobjects.TierBasic,
				IsTrialing: true,
			},
			newTier: valueobjects.TierProfessional,
			wantErr: true,
		},
		{
			name: "Already on same tier",
			sub: &Subscription{
				Status:     SubscriptionStatusActive,
				Tier:       valueobjects.TierBasic,
				IsTrialing: false,
			},
			newTier: valueobjects.TierBasic,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.sub.CanChangeTier(tt.newTier)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSubscription_CanUpgrade(t *testing.T) {
	tests := []struct {
		name string
		sub  *Subscription
		want bool
	}{
		{
			name: "Can upgrade active subscription",
			sub: &Subscription{
				Status:     SubscriptionStatusActive,
				IsTrialing: false,
			},
			want: true,
		},
		{
			name: "Cannot upgrade during trial",
			sub: &Subscription{
				Status:     SubscriptionStatusActive,
				IsTrialing: true,
			},
			want: false,
		},
		{
			name: "Cannot upgrade paused subscription",
			sub: &Subscription{
				Status:     SubscriptionStatusPaused,
				IsTrialing: false,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.sub.CanUpgrade())
		})
	}
}

func TestSubscription_CanDowngrade(t *testing.T) {
	tests := []struct {
		name string
		sub  *Subscription
		want bool
	}{
		{
			name: "Can downgrade active subscription",
			sub: &Subscription{
				Status:     SubscriptionStatusActive,
				IsTrialing: false,
			},
			want: true,
		},
		{
			name: "Cannot downgrade during trial",
			sub: &Subscription{
				Status:     SubscriptionStatusActive,
				IsTrialing: true,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.sub.CanDowngrade())
		})
	}
}

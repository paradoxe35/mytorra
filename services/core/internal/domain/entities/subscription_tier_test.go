package entities

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestSubscriptionTier_Validate(t *testing.T) {
	tests := []struct {
		name    string
		tier    *SubscriptionTier
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid tier",
			tier: &SubscriptionTier{
				ID:        uuid.New(),
				Name:      "Professional",
				Code:      "professional",
				Price:     decimal.NewFromFloat(29.99),
				Currency:  "USD",
				TrialDays: 14,
			},
			wantErr: false,
		},
		{
			name: "Missing name",
			tier: &SubscriptionTier{
				Code:     "professional",
				Price:    decimal.NewFromFloat(29.99),
				Currency: "USD",
			},
			wantErr: true,
			errMsg:  "tier name is required",
		},
		{
			name: "Missing code",
			tier: &SubscriptionTier{
				Name:     "Professional",
				Price:    decimal.NewFromFloat(29.99),
				Currency: "USD",
			},
			wantErr: true,
			errMsg:  "tier code is required",
		},
		{
			name: "Negative price",
			tier: &SubscriptionTier{
				Name:     "Professional",
				Code:     "professional",
				Price:    decimal.NewFromFloat(-29.99),
				Currency: "USD",
			},
			wantErr: true,
			errMsg:  "tier price cannot be negative",
		},
		{
			name: "Missing currency",
			tier: &SubscriptionTier{
				Name:  "Professional",
				Code:  "professional",
				Price: decimal.NewFromFloat(29.99),
			},
			wantErr: true,
			errMsg:  "tier currency is required",
		},
		{
			name: "Negative trial days",
			tier: &SubscriptionTier{
				Name:      "Professional",
				Code:      "professional",
				Price:     decimal.NewFromFloat(29.99),
				Currency:  "USD",
				TrialDays: -5,
			},
			wantErr: true,
			errMsg:  "trial days cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tier.Validate()
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

func TestSubscriptionTier_HasFeature(t *testing.T) {
	tier := &SubscriptionTier{
		Features: map[string]interface{}{
			"api_access":    true,
			"support":       "priority",
			"disabled":      false,
			"custom_domain": "enabled",
			"max_users":     10,
		},
	}

	tests := []struct {
		name    string
		feature string
		want    bool
	}{
		{"Has boolean true feature", "api_access", true},
		{"Has boolean false feature", "disabled", false},
		{"Has string feature", "support", true},
		{"Has string enabled feature", "custom_domain", true},
		{"Has numeric feature", "max_users", true},
		{"Does not have feature", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tier.HasFeature(tt.feature))
		})
	}

	// Test nil features map
	tierNil := &SubscriptionTier{Features: nil}
	assert.False(t, tierNil.HasFeature("any"))
}

func TestSubscriptionTier_GetLimit(t *testing.T) {
	tier := &SubscriptionTier{
		Limits: map[string]int{
			"api_calls":  1000,
			"storage_gb": 100,
			"unlimited":  -1,
		},
	}

	tests := []struct {
		name     string
		resource string
		want     int
	}{
		{"Get existing limit", "api_calls", 1000},
		{"Get unlimited resource", "unlimited", -1},
		{"Get non-existent limit", "unknown", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tier.GetLimit(tt.resource))
		})
	}

	// Test nil limits map
	tierNil := &SubscriptionTier{Limits: nil}
	assert.Equal(t, -1, tierNil.GetLimit("any"))
}

func TestSubscriptionTier_IsUnlimited(t *testing.T) {
	tier := &SubscriptionTier{
		Limits: map[string]int{
			"api_calls":  1000,
			"storage_gb": -1,
		},
	}

	assert.False(t, tier.IsUnlimited("api_calls"))
	assert.True(t, tier.IsUnlimited("storage_gb"))
	assert.True(t, tier.IsUnlimited("unknown"))
}

func TestSubscriptionTier_CompareTo(t *testing.T) {
	tier1 := &SubscriptionTier{
		SortOrder: 1,
		Price:     decimal.NewFromFloat(10),
	}

	tier2 := &SubscriptionTier{
		SortOrder: 2,
		Price:     decimal.NewFromFloat(20),
	}

	tier3 := &SubscriptionTier{
		SortOrder: 1,
		Price:     decimal.NewFromFloat(15),
	}

	tests := []struct {
		name  string
		tier1 *SubscriptionTier
		tier2 *SubscriptionTier
		want  int
	}{
		{"Lower sort order", tier1, tier2, -1},
		{"Higher sort order", tier2, tier1, 1},
		{"Same sort order, lower price", tier1, tier3, -1},
		{"Same sort order, higher price", tier3, tier1, 1},
		{"Identical tiers", tier1, tier1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.tier1.CompareTo(tt.tier2))
		})
	}
}

func TestSubscriptionTier_IsUpgradeFrom(t *testing.T) {
	basic := &SubscriptionTier{
		SortOrder: 1,
		Price:     decimal.NewFromFloat(10),
	}

	professional := &SubscriptionTier{
		SortOrder: 2,
		Price:     decimal.NewFromFloat(30),
	}

	assert.True(t, professional.IsUpgradeFrom(basic))
	assert.False(t, basic.IsUpgradeFrom(professional))
	assert.False(t, basic.IsUpgradeFrom(basic))
}

func TestSubscriptionTier_IsDowngradeFrom(t *testing.T) {
	basic := &SubscriptionTier{
		SortOrder: 1,
		Price:     decimal.NewFromFloat(10),
	}

	professional := &SubscriptionTier{
		SortOrder: 2,
		Price:     decimal.NewFromFloat(30),
	}

	assert.True(t, basic.IsDowngradeFrom(professional))
	assert.False(t, professional.IsDowngradeFrom(basic))
	assert.False(t, basic.IsDowngradeFrom(basic))
}

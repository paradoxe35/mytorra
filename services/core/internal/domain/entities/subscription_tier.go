package entities

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// SubscriptionTier represents a subscription tier with pricing and features
type SubscriptionTier struct {
	ID        uuid.UUID              `db:"id" json:"id"`
	Name      string                 `db:"name" json:"name"`
	Code      string                 `db:"code" json:"code"` // "basic", "professional", "premium"
	Price     decimal.Decimal        `db:"price" json:"price"`
	Currency  string                 `db:"currency" json:"currency"`
	Features  map[string]interface{} `db:"features" json:"features"`
	Limits    map[string]int         `db:"limits" json:"limits"`
	TrialDays int                    `db:"trial_days" json:"trial_days"`
	SortOrder int                    `db:"sort_order" json:"sort_order"`
	IsActive  bool                   `db:"is_active" json:"is_active"`
	CreatedAt time.Time              `db:"created_at" json:"created_at"`
	UpdatedAt time.Time              `db:"updated_at" json:"updated_at"`
}

// Validate validates the subscription tier
func (t *SubscriptionTier) Validate() error {
	if t.Name == "" {
		return fmt.Errorf("tier name is required")
	}

	if t.Code == "" {
		return fmt.Errorf("tier code is required")
	}

	if t.Price.LessThan(decimal.Zero) {
		return fmt.Errorf("tier price cannot be negative")
	}

	if t.Currency == "" {
		return fmt.Errorf("tier currency is required")
	}

	if t.TrialDays < 0 {
		return fmt.Errorf("trial days cannot be negative")
	}

	return nil
}

// HasFeature checks if the tier has a specific feature
func (t *SubscriptionTier) HasFeature(feature string) bool {
	if t.Features == nil {
		return false
	}

	value, exists := t.Features[feature]
	if !exists {
		return false
	}

	// Check if feature is enabled (could be bool or string)
	switch v := value.(type) {
	case bool:
		return v
	case string:
		return v != "" && v != "false" && v != "disabled"
	default:
		return true // Feature exists with some value
	}
}

// GetLimit gets the limit for a specific resource
func (t *SubscriptionTier) GetLimit(resource string) int {
	if t.Limits == nil {
		return -1 // No limit
	}

	limit, exists := t.Limits[resource]
	if !exists {
		return -1 // No limit
	}

	return limit
}

// IsUnlimited checks if a resource has no limit
func (t *SubscriptionTier) IsUnlimited(resource string) bool {
	return t.GetLimit(resource) == -1
}

// CompareTo compares this tier with another tier
// Returns: 1 if this tier is higher, -1 if lower, 0 if equal
func (t *SubscriptionTier) CompareTo(other *SubscriptionTier) int {
	if t.SortOrder > other.SortOrder {
		return 1
	} else if t.SortOrder < other.SortOrder {
		return -1
	}

	// If sort order is the same, compare by price
	if t.Price.GreaterThan(other.Price) {
		return 1
	} else if t.Price.LessThan(other.Price) {
		return -1
	}

	return 0
}

// IsUpgradeFrom checks if this tier is an upgrade from another tier
func (t *SubscriptionTier) IsUpgradeFrom(other *SubscriptionTier) bool {
	return t.CompareTo(other) > 0
}

// IsDowngradeFrom checks if this tier is a downgrade from another tier
func (t *SubscriptionTier) IsDowngradeFrom(other *SubscriptionTier) bool {
	return t.CompareTo(other) < 0
}

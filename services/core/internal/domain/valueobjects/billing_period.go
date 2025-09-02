package valueobjects

import (
	"fmt"
	"time"
)

// BillingPeriod represents the billing period for a subscription
type BillingPeriod string

const (
	BillingPeriodMonthly   BillingPeriod = "monthly"
	BillingPeriodQuarterly BillingPeriod = "quarterly"
	BillingPeriodAnnual    BillingPeriod = "annual"
	BillingPeriodCustom    BillingPeriod = "custom"
)

// String returns the string representation of the billing period
func (b BillingPeriod) String() string {
	return string(b)
}

// IsValid checks if the billing period is valid
func (b BillingPeriod) IsValid() bool {
	switch b {
	case BillingPeriodMonthly, BillingPeriodQuarterly, BillingPeriodAnnual, BillingPeriodCustom:
		return true
	default:
		return false
	}
}

// GetNextBillingDate calculates the next billing date from a given date
func (b BillingPeriod) GetNextBillingDate(from time.Time) time.Time {
	switch b {
	case BillingPeriodMonthly:
		return from.AddDate(0, 1, 0)
	case BillingPeriodQuarterly:
		return from.AddDate(0, 3, 0)
	case BillingPeriodAnnual:
		return from.AddDate(1, 0, 0)
	default:
		// For custom billing period, default to monthly
		return from.AddDate(0, 1, 0)
	}
}

// GetPeriodDays returns the number of days in the billing period
func (b BillingPeriod) GetPeriodDays() int {
	switch b {
	case BillingPeriodMonthly:
		return 30
	case BillingPeriodQuarterly:
		return 90
	case BillingPeriodAnnual:
		return 365
	default:
		return 30 // Default to monthly
	}
}

// GetPeriodMonths returns the number of months in the billing period
func (b BillingPeriod) GetPeriodMonths() int {
	switch b {
	case BillingPeriodMonthly:
		return 1
	case BillingPeriodQuarterly:
		return 3
	case BillingPeriodAnnual:
		return 12
	default:
		return 1 // Default to monthly
	}
}

// CalculatePeriodEnd calculates the end date of a billing period from a start date
func (b BillingPeriod) CalculatePeriodEnd(startDate time.Time) time.Time {
	return b.GetNextBillingDate(startDate).AddDate(0, 0, -1)
}

// GetDisplayName returns the display name for the billing period
func (b BillingPeriod) GetDisplayName() string {
	switch b {
	case BillingPeriodMonthly:
		return "Monthly"
	case BillingPeriodQuarterly:
		return "Quarterly"
	case BillingPeriodAnnual:
		return "Annual"
	case BillingPeriodCustom:
		return "Custom"
	default:
		return string(b)
	}
}

// ParseBillingPeriod parses a string into a BillingPeriod
func ParseBillingPeriod(s string) (BillingPeriod, error) {
	period := BillingPeriod(s)
	if !period.IsValid() {
		return "", fmt.Errorf("invalid billing period: %s", s)
	}
	return period, nil
}

// GetDiscountPercentage returns the typical discount percentage for the billing period
// compared to monthly billing
func (b BillingPeriod) GetDiscountPercentage() float64 {
	switch b {
	case BillingPeriodMonthly:
		return 0.0
	case BillingPeriodQuarterly:
		return 5.0 // 5% discount for quarterly
	case BillingPeriodAnnual:
		return 15.0 // 15% discount for annual
	default:
		return 0.0
	}
}

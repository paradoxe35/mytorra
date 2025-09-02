package dtos

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// CreateSubscriptionDTO represents the request to create a subscription
type CreateSubscriptionDTO struct {
	UserID        uuid.UUID `json:"user_id" validate:"required"`
	TierID        uuid.UUID `json:"tier_id" validate:"required"`
	PaymentMethod string    `json:"payment_method" validate:"required"`
	StartTrial    bool      `json:"start_trial"`
	EntityID      uuid.UUID `json:"entity_id,omitempty"`
	EntityType    string    `json:"entity_type,omitempty"`
}

// UpgradeSubscriptionDTO represents the request to upgrade a subscription
type UpgradeSubscriptionDTO struct {
	NewTierID uuid.UUID `json:"new_tier_id" validate:"required"`
	Immediate bool      `json:"immediate"`
}

// DowngradeSubscriptionDTO represents the request to downgrade a subscription
type DowngradeSubscriptionDTO struct {
	NewTierID          uuid.UUID `json:"new_tier_id" validate:"required"`
	EffectiveNextCycle bool      `json:"effective_next_cycle"`
}

// CancelSubscriptionDTO represents the request to cancel a subscription
type CancelSubscriptionDTO struct {
	Reason            string `json:"reason" validate:"required,min=3,max=500"`
	CancelImmediately bool   `json:"cancel_immediately"`
}

// PauseSubscriptionDTO represents the request to pause a subscription
type PauseSubscriptionDTO struct {
	PausedUntil *time.Time `json:"paused_until,omitempty"`
	Reason      string     `json:"reason,omitempty"`
}

// ExtendTrialDTO represents the request to extend trial period
type ExtendTrialDTO struct {
	ExtensionDays int    `json:"extension_days" validate:"required,min=1,max=30"`
	Reason        string `json:"reason" validate:"required"`
}

// SubscriptionResponseDTO represents the subscription response
type SubscriptionResponseDTO struct {
	ID            uuid.UUID      `json:"id"`
	User          UserSummaryDTO `json:"user"`
	Tier          TierDTO        `json:"tier"`
	Status        string         `json:"status"`
	CurrentPeriod PeriodDTO      `json:"current_period"`
	NextBilling   *time.Time     `json:"next_billing,omitempty"`
	TrialEndsAt   *time.Time     `json:"trial_ends_at,omitempty"`
	PausedAt      *time.Time     `json:"paused_at,omitempty"`
	CanceledAt    *time.Time     `json:"canceled_at,omitempty"`
	UsageLimits   map[string]int `json:"usage_limits"`
	CurrentUsage  map[string]int `json:"current_usage"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// UserSummaryDTO represents a summary of user information
type UserSummaryDTO struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	Name  string    `json:"name"`
}

// TierDTO represents subscription tier information
type TierDTO struct {
	ID        uuid.UUID              `json:"id"`
	Name      string                 `json:"name"`
	Code      string                 `json:"code"`
	Price     decimal.Decimal        `json:"price"`
	Currency  string                 `json:"currency"`
	Features  map[string]interface{} `json:"features"`
	Limits    map[string]int         `json:"limits"`
	TrialDays int                    `json:"trial_days"`
}

// PeriodDTO represents a billing period
type PeriodDTO struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	DaysLeft  int       `json:"days_left"`
}

// SubscriptionListDTO represents a list of subscriptions
type SubscriptionListDTO struct {
	Subscriptions []SubscriptionResponseDTO `json:"subscriptions"`
	Total         int64                     `json:"total"`
	Page          int                       `json:"page"`
	PageSize      int                       `json:"page_size"`
}

// SubscriptionUsageDTO represents subscription usage information
type SubscriptionUsageDTO struct {
	SubscriptionID uuid.UUID `json:"subscription_id"`
	Feature        string    `json:"feature"`
	CurrentUsage   int       `json:"current_usage"`
	Limit          int       `json:"limit"`
	Percentage     float64   `json:"percentage"`
	IsUnlimited    bool      `json:"is_unlimited"`
}

// SubscriptionUpgradePreviewDTO shows what will happen on upgrade
type SubscriptionUpgradePreviewDTO struct {
	CurrentTier       TierDTO         `json:"current_tier"`
	NewTier           TierDTO         `json:"new_tier"`
	ProrationAmount   decimal.Decimal `json:"proration_amount"`
	ImmediateCharge   decimal.Decimal `json:"immediate_charge"`
	NextBillingAmount decimal.Decimal `json:"next_billing_amount"`
	EffectiveDate     time.Time       `json:"effective_date"`
}

// SubscriptionStatisticsDTO represents subscription statistics
type SubscriptionStatisticsDTO struct {
	TotalActive      int64           `json:"total_active"`
	TotalTrialing    int64           `json:"total_trialing"`
	TotalPaused      int64           `json:"total_paused"`
	TotalCanceled    int64           `json:"total_canceled"`
	RevenueThisMonth decimal.Decimal `json:"revenue_this_month"`
	ChurnRate        float64         `json:"churn_rate"`
	TrialConversion  float64         `json:"trial_conversion_rate"`
}

// PaymentMethodDTO represents a payment method
type PaymentMethodDTO struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Last4     string    `json:"last4,omitempty"`
	Brand     string    `json:"brand,omitempty"`
	ExpMonth  int       `json:"exp_month,omitempty"`
	ExpYear   int       `json:"exp_year,omitempty"`
	IsDefault bool      `json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
}

// AddPaymentMethodDTO represents request to add a payment method
type AddPaymentMethodDTO struct {
	Type       string `json:"type" validate:"required,oneof=card bank_account"`
	Token      string `json:"token" validate:"required"`
	SetDefault bool   `json:"set_default"`
}

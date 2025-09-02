package entities

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/paradoxe35/torra/services/core/internal/domain/valueobjects"
)

type Subscription struct {
	ID         uuid.UUID `db:"id" json:"id"`
	UserID     uuid.UUID `db:"user_id" json:"user_id"`
	EntityID   uuid.UUID `db:"entity_id" json:"entity_id"`
	EntityType string    `db:"entity_type" json:"entity_type"`

	// Plan details
	Tier         valueobjects.SubscriptionTier `db:"tier" json:"tier"`
	PlanName     string                        `db:"plan_name" json:"plan_name"`
	Price        valueobjects.Money            `db:"-" json:"price"`
	BillingCycle SubscriptionBillingCycle      `db:"billing_cycle" json:"billing_cycle"`

	// Status
	Status SubscriptionStatus `db:"status" json:"status"`

	// Trial
	TrialEndsAt *time.Time `db:"trial_ends_at" json:"trial_ends_at,omitempty"`
	IsTrialing  bool       `db:"is_trialing" json:"is_trialing"`

	// Billing
	CurrentPeriodStart time.Time  `db:"current_period_start" json:"current_period_start"`
	CurrentPeriodEnd   time.Time  `db:"current_period_end" json:"current_period_end"`
	NextBillingDate    *time.Time `db:"next_billing_date" json:"next_billing_date,omitempty"`

	// Payment provider integration
	ProviderCustomerID     *string `db:"provider_customer_id" json:"provider_customer_id,omitempty"`
	ProviderSubscriptionID *string `db:"provider_subscription_id" json:"provider_subscription_id,omitempty"`

	// Metadata for storing additional information
	Metadata map[string]interface{} `db:"metadata" json:"metadata,omitempty"`

	// Usage limits
	UsageLimits  map[string]int `db:"usage_limits" json:"usage_limits"`
	CurrentUsage map[string]int `db:"current_usage" json:"current_usage"`

	// Pause/Resume
	PausedAt *time.Time `db:"paused_at" json:"paused_at,omitempty"`
	ResumeAt *time.Time `db:"resume_at" json:"resume_at,omitempty"`

	// Cancellation
	CanceledAt   *time.Time `db:"canceled_at" json:"canceled_at,omitempty"`
	CancelReason *string    `db:"cancel_reason" json:"cancel_reason,omitempty"`

	// Timestamps
	ActivatedAt *time.Time `db:"activated_at" json:"activated_at,omitempty"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}

type (
	SubscriptionBillingCycle string
	SubscriptionStatus       string
)

const (
	BillingCycleMonthly SubscriptionBillingCycle = "monthly"
	BillingCycleAnnual  SubscriptionBillingCycle = "annual"
	BillingCycleCustom  SubscriptionBillingCycle = "custom"

	SubscriptionStatusActive   SubscriptionStatus = "active"
	SubscriptionStatusTrialing SubscriptionStatus = "trialing"
	SubscriptionStatusPastDue  SubscriptionStatus = "past_due"
	SubscriptionStatusPaused   SubscriptionStatus = "paused"
	SubscriptionStatusCanceled SubscriptionStatus = "canceled"
	SubscriptionStatusExpired  SubscriptionStatus = "expired"
)

// Domain methods
func (s *Subscription) CanUpgrade() bool {
	return s.Status == SubscriptionStatusActive && !s.IsTrialing
}

func (s *Subscription) IsWithinUsageLimits(feature string) bool {
	limit, hasLimit := s.UsageLimits[feature]
	if !hasLimit {
		return true // No limit set
	}

	usage := s.CurrentUsage[feature]
	return usage < limit
}

func (s *Subscription) DaysUntilRenewal() int {
	if s.NextBillingDate == nil {
		return -1
	}
	return int(time.Until(*s.NextBillingDate).Hours() / 24)
}

func (s *Subscription) IncrementUsage(feature string) {
	if s.CurrentUsage == nil {
		s.CurrentUsage = make(map[string]int)
	}
	s.CurrentUsage[feature]++
	s.UpdatedAt = time.Now()
}

func (s *Subscription) Pause() {
	now := time.Now()
	s.Status = SubscriptionStatusPaused
	s.PausedAt = &now
	s.UpdatedAt = now
}

func (s *Subscription) Resume() {
	s.Status = SubscriptionStatusActive
	s.PausedAt = nil
	s.ResumeAt = nil
	s.UpdatedAt = time.Now()
}

// IsActive checks if the subscription is currently active
func (s *Subscription) IsActive() bool {
	return s.Status == SubscriptionStatusActive || s.Status == SubscriptionStatusTrialing
}

// IsInTrial checks if the subscription is in trial period
func (s *Subscription) IsInTrial() bool {
	if !s.IsTrialing || s.TrialEndsAt == nil {
		return false
	}
	return time.Now().Before(*s.TrialEndsAt)
}

// CanDowngrade checks if the subscription can be downgraded
func (s *Subscription) CanDowngrade() bool {
	return s.Status == SubscriptionStatusActive && !s.IsTrialing
}

// Cancel cancels the subscription with a reason
func (s *Subscription) Cancel(reason string) error {
	if s.Status == SubscriptionStatusCanceled {
		return fmt.Errorf("subscription is already canceled")
	}

	now := time.Now()
	s.Status = SubscriptionStatusCanceled
	s.CanceledAt = &now
	s.CancelReason = &reason
	s.UpdatedAt = now
	return nil
}

// ExtendTrial extends the trial period by specified days
func (s *Subscription) ExtendTrial(days int) error {
	if !s.IsTrialing {
		return fmt.Errorf("subscription is not in trial")
	}

	if s.TrialEndsAt == nil {
		return fmt.Errorf("trial end date not set")
	}

	newTrialEnd := s.TrialEndsAt.AddDate(0, 0, days)
	s.TrialEndsAt = &newTrialEnd
	s.UpdatedAt = time.Now()
	return nil
}

// ResetUsage resets the current usage for all features
func (s *Subscription) ResetUsage() error {
	s.CurrentUsage = make(map[string]int)
	s.UpdatedAt = time.Now()
	return nil
}

// Validate validates the subscription entity
func (s *Subscription) Validate() error {
	if s.UserID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}

	if s.EntityID == uuid.Nil {
		return fmt.Errorf("entity ID is required")
	}

	if s.EntityType == "" {
		return fmt.Errorf("entity type is required")
	}

	if s.PlanName == "" {
		return fmt.Errorf("plan name is required")
	}

	if s.CurrentPeriodStart.After(s.CurrentPeriodEnd) {
		return fmt.Errorf("current period start must be before end")
	}

	return nil
}

// CalculateProration calculates the proration amount for changing to a new tier
func (s *Subscription) CalculateProration(newTierPrice decimal.Decimal) decimal.Decimal {
	if s.Status != SubscriptionStatusActive {
		return decimal.Zero
	}

	// Calculate days remaining in current period
	now := time.Now()
	if now.After(s.CurrentPeriodEnd) {
		return decimal.Zero
	}

	totalDays := s.CurrentPeriodEnd.Sub(s.CurrentPeriodStart).Hours() / 24
	remainingDays := s.CurrentPeriodEnd.Sub(now).Hours() / 24

	if totalDays <= 0 {
		return decimal.Zero
	}

	// Calculate proration ratio
	prorationRatio := decimal.NewFromFloat(remainingDays / totalDays)

	// Calculate current tier's remaining value
	remainingValue := s.Price.Amount.Mul(prorationRatio)

	// Calculate new tier's prorated cost
	newTierCost := newTierPrice.Mul(prorationRatio)

	// Return the difference (positive means customer owes, negative means credit)
	return newTierCost.Sub(remainingValue)
}

// CanChangeTier checks if the subscription can be changed to a new tier
func (s *Subscription) CanChangeTier(newTier valueobjects.SubscriptionTier) error {
	if s.Status == SubscriptionStatusCanceled {
		return fmt.Errorf("cannot change tier for canceled subscription")
	}

	if s.Status == SubscriptionStatusExpired {
		return fmt.Errorf("cannot change tier for expired subscription")
	}

	if s.IsInTrial() {
		return fmt.Errorf("cannot change tier during trial period")
	}

	if s.Tier == newTier {
		return fmt.Errorf("already on the selected tier")
	}

	return nil
}

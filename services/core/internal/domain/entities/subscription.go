package entities

import (
	"time"

	"github.com/google/uuid"

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
	BillingCycle BillingCycle                  `db:"billing_cycle" json:"billing_cycle"`

	// Status
	Status SubscriptionStatus `db:"status" json:"status"`

	// Trial
	TrialEndsAt *time.Time `db:"trial_ends_at" json:"trial_ends_at,omitempty"`
	IsTrialing  bool       `db:"is_trialing" json:"is_trialing"`

	// Billing
	CurrentPeriodStart time.Time  `db:"current_period_start" json:"current_period_start"`
	CurrentPeriodEnd   time.Time  `db:"current_period_end" json:"current_period_end"`
	NextBillingDate    *time.Time `db:"next_billing_date" json:"next_billing_date,omitempty"`

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
	BillingCycle       string
	SubscriptionStatus string
)

const (
	BillingCycleMonthly BillingCycle = "monthly"
	BillingCycleAnnual  BillingCycle = "annual"

	SubscriptionStatusActive   SubscriptionStatus = "active"
	SubscriptionStatusTrialing SubscriptionStatus = "trialing"
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

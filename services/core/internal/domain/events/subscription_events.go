package events

import (
	"time"

	"github.com/google/uuid"

	"github.com/paradoxe35/mytorra/services/core/internal/domain/valueobjects"
)

// SubscriptionCreated event is published when a subscription is created
type SubscriptionCreated struct {
	SubscriptionID uuid.UUID                     `json:"subscription_id"`
	UserID         uuid.UUID                     `json:"user_id"`
	EntityID       uuid.UUID                     `json:"entity_id"`
	EntityType     string                        `json:"entity_type"`
	Tier           valueobjects.SubscriptionTier `json:"tier"`
	IsTrialing     bool                          `json:"is_trialing"`
	CreatedAt      time.Time                     `json:"created_at"`
}

func (e SubscriptionCreated) EventType() string {
	return "subscription.created"
}

func (e SubscriptionCreated) EventData() interface{} {
	return e
}

// SubscriptionUpgraded event is published when a subscription is upgraded
type SubscriptionUpgraded struct {
	SubscriptionID uuid.UUID                     `json:"subscription_id"`
	UserID         uuid.UUID                     `json:"user_id"`
	OldTier        valueobjects.SubscriptionTier `json:"old_tier"`
	NewTier        valueobjects.SubscriptionTier `json:"new_tier"`
	UpgradedAt     time.Time                     `json:"upgraded_at"`
}

func (e SubscriptionUpgraded) EventType() string {
	return "subscription.upgraded"
}

func (e SubscriptionUpgraded) EventData() interface{} {
	return e
}

// SubscriptionDowngraded event is published when a subscription is downgraded
type SubscriptionDowngraded struct {
	SubscriptionID uuid.UUID                     `json:"subscription_id"`
	UserID         uuid.UUID                     `json:"user_id"`
	OldTier        valueobjects.SubscriptionTier `json:"old_tier"`
	NewTier        valueobjects.SubscriptionTier `json:"new_tier"`
	DowngradedAt   time.Time                     `json:"downgraded_at"`
}

func (e SubscriptionDowngraded) EventType() string {
	return "subscription.downgraded"
}

func (e SubscriptionDowngraded) EventData() interface{} {
	return e
}

// SubscriptionCanceled event is published when a subscription is canceled
type SubscriptionCanceled struct {
	SubscriptionID uuid.UUID `json:"subscription_id"`
	UserID         uuid.UUID `json:"user_id"`
	Reason         string    `json:"reason"`
	CanceledAt     time.Time `json:"canceled_at"`
}

func (e SubscriptionCanceled) EventType() string {
	return "subscription.canceled"
}

func (e SubscriptionCanceled) EventData() interface{} {
	return e
}

// SubscriptionPaused event is published when a subscription is paused
type SubscriptionPaused struct {
	SubscriptionID uuid.UUID  `json:"subscription_id"`
	UserID         uuid.UUID  `json:"user_id"`
	PausedAt       time.Time  `json:"paused_at"`
	ResumeAt       *time.Time `json:"resume_at,omitempty"`
}

func (e SubscriptionPaused) EventType() string {
	return "subscription.paused"
}

func (e SubscriptionPaused) EventData() interface{} {
	return e
}

// SubscriptionResumed event is published when a subscription is resumed
type SubscriptionResumed struct {
	SubscriptionID uuid.UUID `json:"subscription_id"`
	UserID         uuid.UUID `json:"user_id"`
	ResumedAt      time.Time `json:"resumed_at"`
}

func (e SubscriptionResumed) EventType() string {
	return "subscription.resumed"
}

func (e SubscriptionResumed) EventData() interface{} {
	return e
}

// SubscriptionTrialEnded event is published when a trial period ends
type SubscriptionTrialEnded struct {
	SubscriptionID uuid.UUID `json:"subscription_id"`
	UserID         uuid.UUID `json:"user_id"`
	EndedAt        time.Time `json:"ended_at"`
}

func (e SubscriptionTrialEnded) EventType() string {
	return "subscription.trial_ended"
}

func (e SubscriptionTrialEnded) EventData() interface{} {
	return e
}

// SubscriptionUsageWarning event is published when usage approaches limits
type SubscriptionUsageWarning struct {
	SubscriptionID uuid.UUID `json:"subscription_id"`
	UserID         uuid.UUID `json:"user_id"`
	Resource       string    `json:"resource"`
	CurrentUsage   int       `json:"usage"`
	Limit          int       `json:"limit"`
	Percentage     float64   `json:"percentage"`
}

func (e SubscriptionUsageWarning) EventType() string {
	return "subscription.usage_warning"
}

func (e SubscriptionUsageWarning) EventData() interface{} {
	return e
}

// SubscriptionTrialExtended event is published when trial is extended
type SubscriptionTrialExtended struct {
	SubscriptionID uuid.UUID  `json:"subscription_id"`
	UserID         uuid.UUID  `json:"user_id"`
	ExtensionDays  int        `json:"extension_days"`
	Reason         string     `json:"reason"`
	NewTrialEnd    *time.Time `json:"new_trial_end"`
}

func (e SubscriptionTrialExtended) EventType() string {
	return "subscription.trial_extended"
}

func (e SubscriptionTrialExtended) EventData() interface{} {
	return e
}

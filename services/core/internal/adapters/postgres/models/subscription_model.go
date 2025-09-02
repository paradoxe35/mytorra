package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
	"github.com/paradoxe35/torra/services/core/internal/domain/valueobjects"
)

// SubscriptionModel represents the database model for subscriptions
type SubscriptionModel struct {
	ID                     uuid.UUID       `db:"id"`
	UserID                 uuid.UUID       `db:"user_id"`
	EntityID               uuid.UUID       `db:"entity_id"`
	EntityType             string          `db:"entity_type"`
	TierID                 *uuid.UUID      `db:"tier_id"`
	Tier                   string          `db:"tier"`
	PlanName               string          `db:"plan_name"`
	Price                  decimal.Decimal `db:"price"`
	Currency               string          `db:"currency"`
	BillingCycle           string          `db:"billing_cycle"`
	Status                 string          `db:"status"`
	CurrentPeriodStart     time.Time       `db:"current_period_start"`
	CurrentPeriodEnd       time.Time       `db:"current_period_end"`
	TrialEndsAt            *time.Time      `db:"trial_ends_at"`
	IsTrialing             bool            `db:"is_trialing"`
	NextBillingDate        *time.Time      `db:"next_billing_date"`
	ProviderCustomerID     *string         `db:"stripe_customer_id"`
	ProviderSubscriptionID *string         `db:"stripe_subscription_id"`
	UsageLimits            JSONMap         `db:"usage_limits"`
	CurrentUsage           JSONMap         `db:"current_usage"`
	Metadata               JSONMap         `db:"metadata"`
	PausedAt               *time.Time      `db:"paused_at"`
	ResumeAt               *time.Time      `db:"resume_at"`
	CanceledAt             *time.Time      `db:"canceled_at"`
	CancelReason           *string         `db:"cancel_reason"`
	ActivatedAt            *time.Time      `db:"activated_at"`
	CreatedAt              time.Time       `db:"created_at"`
	UpdatedAt              time.Time       `db:"updated_at"`
}

// JSONMap is a custom type for JSON map fields
type JSONMap map[string]interface{}

// Value implements driver.Valuer interface
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements sql.Scanner interface
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = make(map[string]interface{})
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	var m map[string]interface{}
	if err := json.Unmarshal(bytes, &m); err != nil {
		return err
	}

	*j = m
	return nil
}

// ToEntity converts the database model to a domain entity
func (m *SubscriptionModel) ToEntity() *entities.Subscription {
	// Convert JSONMap to map[string]int for usage limits and current usage
	usageLimits := make(map[string]int)
	for k, v := range m.UsageLimits {
		if val, ok := v.(float64); ok {
			usageLimits[k] = int(val)
		}
	}

	currentUsage := make(map[string]int)
	for k, v := range m.CurrentUsage {
		if val, ok := v.(float64); ok {
			currentUsage[k] = int(val)
		}
	}

	return &entities.Subscription{
		ID:         m.ID,
		UserID:     m.UserID,
		EntityID:   m.EntityID,
		EntityType: m.EntityType,
		Tier:       valueobjects.SubscriptionTier(m.Tier),
		PlanName:   m.PlanName,
		Price: valueobjects.Money{
			Amount:   m.Price,
			Currency: m.Currency,
		},
		BillingCycle:           entities.SubscriptionBillingCycle(m.BillingCycle),
		Status:                 entities.SubscriptionStatus(m.Status),
		CurrentPeriodStart:     m.CurrentPeriodStart,
		CurrentPeriodEnd:       m.CurrentPeriodEnd,
		TrialEndsAt:            m.TrialEndsAt,
		IsTrialing:             m.IsTrialing,
		NextBillingDate:        m.NextBillingDate,
		ProviderCustomerID:     m.ProviderCustomerID,
		ProviderSubscriptionID: m.ProviderSubscriptionID,
		UsageLimits:            usageLimits,
		CurrentUsage:           currentUsage,
		Metadata:               m.Metadata,
		PausedAt:               m.PausedAt,
		ResumeAt:               m.ResumeAt,
		CanceledAt:             m.CanceledAt,
		CancelReason:           m.CancelReason,
		ActivatedAt:            m.ActivatedAt,
		CreatedAt:              m.CreatedAt,
		UpdatedAt:              m.UpdatedAt,
	}
}

// FromEntity converts a domain entity to a database model
func FromSubscriptionEntity(e *entities.Subscription) *SubscriptionModel {
	// Convert map[string]int to JSONMap for usage limits and current usage
	usageLimits := make(JSONMap)
	for k, v := range e.UsageLimits {
		usageLimits[k] = v
	}

	currentUsage := make(JSONMap)
	for k, v := range e.CurrentUsage {
		currentUsage[k] = v
	}

	metadata := make(JSONMap)
	if e.Metadata != nil {
		metadata = JSONMap(e.Metadata)
	}

	return &SubscriptionModel{
		ID:                     e.ID,
		UserID:                 e.UserID,
		EntityID:               e.EntityID,
		EntityType:             e.EntityType,
		Tier:                   string(e.Tier),
		PlanName:               e.PlanName,
		Price:                  e.Price.Amount,
		Currency:               e.Price.Currency,
		BillingCycle:           string(e.BillingCycle),
		Status:                 string(e.Status),
		CurrentPeriodStart:     e.CurrentPeriodStart,
		CurrentPeriodEnd:       e.CurrentPeriodEnd,
		TrialEndsAt:            e.TrialEndsAt,
		IsTrialing:             e.IsTrialing,
		NextBillingDate:        e.NextBillingDate,
		ProviderCustomerID:     e.ProviderCustomerID,
		ProviderSubscriptionID: e.ProviderSubscriptionID,
		UsageLimits:            usageLimits,
		CurrentUsage:           currentUsage,
		Metadata:               metadata,
		PausedAt:               e.PausedAt,
		ResumeAt:               e.ResumeAt,
		CanceledAt:             e.CanceledAt,
		CancelReason:           e.CancelReason,
		ActivatedAt:            e.ActivatedAt,
		CreatedAt:              e.CreatedAt,
		UpdatedAt:              e.UpdatedAt,
	}
}

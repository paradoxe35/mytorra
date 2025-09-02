package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
)

// SubscriptionTierModel represents the database model for subscription tiers
type SubscriptionTierModel struct {
	ID        uuid.UUID       `db:"id"`
	Name      string          `db:"name"`
	Code      string          `db:"code"`
	Price     decimal.Decimal `db:"price"`
	Currency  string          `db:"currency"`
	Features  JSONMap         `db:"features"`
	Limits    JSONMap         `db:"limits"`
	TrialDays int             `db:"trial_days"`
	SortOrder int             `db:"sort_order"`
	IsActive  bool            `db:"is_active"`
	CreatedAt time.Time       `db:"created_at"`
	UpdatedAt time.Time       `db:"updated_at"`
}

// ToEntity converts the database model to a domain entity
func (m *SubscriptionTierModel) ToEntity() *entities.SubscriptionTier {
	// Convert JSONMap to map[string]int for limits
	limits := make(map[string]int)
	for k, v := range m.Limits {
		if val, ok := v.(float64); ok {
			limits[k] = int(val)
		}
	}

	return &entities.SubscriptionTier{
		ID:        m.ID,
		Name:      m.Name,
		Code:      m.Code,
		Price:     m.Price,
		Currency:  m.Currency,
		Features:  m.Features,
		Limits:    limits,
		TrialDays: m.TrialDays,
		SortOrder: m.SortOrder,
		IsActive:  m.IsActive,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

// FromEntity converts a domain entity to a database model
func FromSubscriptionTierEntity(e *entities.SubscriptionTier) *SubscriptionTierModel {
	// Convert map[string]int to JSONMap for limits
	limits := make(JSONMap)
	for k, v := range e.Limits {
		limits[k] = v
	}

	features := make(JSONMap)
	if e.Features != nil {
		features = JSONMap(e.Features)
	}

	return &SubscriptionTierModel{
		ID:        e.ID,
		Name:      e.Name,
		Code:      e.Code,
		Price:     e.Price,
		Currency:  e.Currency,
		Features:  features,
		Limits:    limits,
		TrialDays: e.TrialDays,
		SortOrder: e.SortOrder,
		IsActive:  e.IsActive,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

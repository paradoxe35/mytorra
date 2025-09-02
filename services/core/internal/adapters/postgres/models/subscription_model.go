package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/paradoxe35/torra/packages/database"
	"github.com/paradoxe35/torra/services/core/internal/adapters/postgres/types"
	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
	"github.com/paradoxe35/torra/services/core/internal/domain/valueobjects"
)

// SubscriptionModel represents the GORM model for subscriptions table
type SubscriptionModel struct {
	database.BaseModel
	UserID     uuid.UUID  `gorm:"type:uuid;not null;index"`
	User       *UserModel `gorm:"foreignKey:UserID"`
	EntityID   uuid.UUID  `gorm:"type:uuid;not null;index"`
	EntityType string     `gorm:"type:varchar(50);not null"`

	// Plan details
	Tier         string          `gorm:"type:varchar(50);not null"` // Store tier as string
	PlanName     string          `gorm:"type:varchar(100);not null"`
	Price        decimal.Decimal `gorm:"type:decimal(10,2);not null"`
	Currency     string          `gorm:"type:varchar(3);not null"`
	BillingCycle string          `gorm:"type:varchar(20);not null"`

	// Status
	Status string `gorm:"type:varchar(20);not null;default:'pending'"`

	// Trial
	TrialEndsAt *time.Time `gorm:"type:timestamp"`
	IsTrialing  bool       `gorm:"type:boolean;default:false"`

	// Billing
	CurrentPeriodStart time.Time  `gorm:"type:timestamp;not null"`
	CurrentPeriodEnd   time.Time  `gorm:"type:timestamp;not null"`
	NextBillingDate    *time.Time `gorm:"type:timestamp"`

	// Payment provider integration
	ProviderCustomerID     *string `gorm:"type:varchar(255)"`
	ProviderSubscriptionID *string `gorm:"type:varchar(255);uniqueIndex"`

	// Metadata for storing additional information
	Metadata types.JSONBValue `gorm:"type:jsonb"`

	// Usage limits
	UsageLimits  types.JSONBValue `gorm:"type:jsonb"`
	CurrentUsage types.JSONBValue `gorm:"type:jsonb"`

	// Pause/Resume
	PausedAt *time.Time `gorm:"type:timestamp"`
	ResumeAt *time.Time `gorm:"type:timestamp"`

	// Cancellation
	CanceledAt   *time.Time `gorm:"type:timestamp"`
	CancelReason *string    `gorm:"type:text"`

	// Timestamps
	ActivatedAt *time.Time `gorm:"type:timestamp"`
}

// TableName specifies the table name for the SubscriptionModel
func (SubscriptionModel) TableName() string {
	return "subscriptions"
}

// ToEntity converts the GORM model to domain entity
func (s *SubscriptionModel) ToEntity() (*entities.Subscription, error) {
	entity := &entities.Subscription{
		ID:                     s.ID,
		UserID:                 s.UserID,
		EntityID:               s.EntityID,
		EntityType:             s.EntityType,
		Tier:                   valueobjects.SubscriptionTier(s.Tier),
		PlanName:               s.PlanName,
		Price:                  valueobjects.Money{Amount: s.Price, Currency: s.Currency},
		BillingCycle:           entities.SubscriptionBillingCycle(s.BillingCycle),
		Status:                 entities.SubscriptionStatus(s.Status),
		TrialEndsAt:            s.TrialEndsAt,
		IsTrialing:             s.IsTrialing,
		CurrentPeriodStart:     s.CurrentPeriodStart,
		CurrentPeriodEnd:       s.CurrentPeriodEnd,
		NextBillingDate:        s.NextBillingDate,
		ProviderCustomerID:     s.ProviderCustomerID,
		ProviderSubscriptionID: s.ProviderSubscriptionID,
		PausedAt:               s.PausedAt,
		ResumeAt:               s.ResumeAt,
		CanceledAt:             s.CanceledAt,
		CancelReason:           s.CancelReason,
		ActivatedAt:            s.ActivatedAt,
		CreatedAt:              s.CreatedAt,
		UpdatedAt:              s.UpdatedAt,
	}

	// Handle Metadata
	if s.Metadata != nil {
		entity.Metadata = make(map[string]interface{})
		s.Metadata.Unmarshal(&entity.Metadata)
	}

	// Handle UsageLimits
	if s.UsageLimits != nil {
		entity.UsageLimits = make(map[string]int)
		s.UsageLimits.Unmarshal(&entity.UsageLimits)
	}

	// Handle CurrentUsage
	if s.CurrentUsage != nil {
		entity.CurrentUsage = make(map[string]int)
		s.CurrentUsage.Unmarshal(&entity.CurrentUsage)
	}

	return entity, nil
}

// SubscriptionFromEntity creates a GORM model from domain entity
func SubscriptionFromEntity(entity *entities.Subscription) *SubscriptionModel {
	model := &SubscriptionModel{
		BaseModel: database.BaseModel{
			ID:        entity.ID,
			CreatedAt: entity.CreatedAt,
			UpdatedAt: entity.UpdatedAt,
		},
		UserID:                 entity.UserID,
		EntityID:               entity.EntityID,
		EntityType:             entity.EntityType,
		Tier:                   string(entity.Tier),
		PlanName:               entity.PlanName,
		Price:                  entity.Price.Amount,
		Currency:               entity.Price.Currency,
		BillingCycle:           string(entity.BillingCycle),
		Status:                 string(entity.Status),
		TrialEndsAt:            entity.TrialEndsAt,
		IsTrialing:             entity.IsTrialing,
		CurrentPeriodStart:     entity.CurrentPeriodStart,
		CurrentPeriodEnd:       entity.CurrentPeriodEnd,
		NextBillingDate:        entity.NextBillingDate,
		ProviderCustomerID:     entity.ProviderCustomerID,
		ProviderSubscriptionID: entity.ProviderSubscriptionID,
		PausedAt:               entity.PausedAt,
		ResumeAt:               entity.ResumeAt,
		CanceledAt:             entity.CanceledAt,
		CancelReason:           entity.CancelReason,
		ActivatedAt:            entity.ActivatedAt,
	}

	// Handle Metadata
	if entity.Metadata != nil {
		model.Metadata = types.NewJSONBValue(entity.Metadata)
	}

	// Handle UsageLimits
	if entity.UsageLimits != nil {
		model.UsageLimits = types.NewJSONBValue(entity.UsageLimits)
	}

	// Handle CurrentUsage
	if entity.CurrentUsage != nil {
		model.CurrentUsage = types.NewJSONBValue(entity.CurrentUsage)
	}

	return model
}

// SubscriptionTierModel represents the GORM model for subscription_tiers table
type SubscriptionTierModel struct {
	database.BaseModel
	Name      string           `gorm:"type:varchar(100);not null;uniqueIndex"`
	Code      string           `gorm:"type:varchar(50);not null;uniqueIndex"`
	Price     decimal.Decimal  `gorm:"type:decimal(10,2);not null"`
	Currency  string           `gorm:"type:varchar(3);not null;default:'USD'"`
	Features  types.JSONBValue `gorm:"type:jsonb"`
	Limits    types.JSONBValue `gorm:"type:jsonb"`
	TrialDays int              `gorm:"type:int;default:0"`
	SortOrder int              `gorm:"type:int;default:0"`
	IsActive  bool             `gorm:"type:boolean;default:true"`
}

// TableName specifies the table name for the SubscriptionTierModel
func (SubscriptionTierModel) TableName() string {
	return "subscription_tiers"
}

// ToEntity converts the GORM model to domain entity
func (st *SubscriptionTierModel) ToEntity() (*entities.SubscriptionTier, error) {
	entity := &entities.SubscriptionTier{
		ID:        st.ID,
		Name:      st.Name,
		Code:      st.Code,
		Price:     st.Price,
		Currency:  st.Currency,
		TrialDays: st.TrialDays,
		SortOrder: st.SortOrder,
		IsActive:  st.IsActive,
		CreatedAt: st.CreatedAt,
		UpdatedAt: st.UpdatedAt,
	}

	// Handle Features
	if st.Features != nil {
		entity.Features = make(map[string]interface{})
		st.Features.Unmarshal(&entity.Features)
	}

	// Handle Limits
	if st.Limits != nil {
		entity.Limits = make(map[string]int)
		st.Limits.Unmarshal(&entity.Limits)
	}

	return entity, nil
}

// SubscriptionTierFromEntity creates a GORM model from domain entity
func SubscriptionTierFromEntity(entity *entities.SubscriptionTier) *SubscriptionTierModel {
	model := &SubscriptionTierModel{
		BaseModel: database.BaseModel{
			ID:        entity.ID,
			CreatedAt: entity.CreatedAt,
			UpdatedAt: entity.UpdatedAt,
		},
		Name:      entity.Name,
		Code:      entity.Code,
		Price:     entity.Price,
		Currency:  entity.Currency,
		TrialDays: entity.TrialDays,
		SortOrder: entity.SortOrder,
		IsActive:  entity.IsActive,
	}

	// Handle Features
	if entity.Features != nil {
		model.Features = types.NewJSONBValue(entity.Features)
	}

	// Handle Limits
	if entity.Limits != nil {
		model.Limits = types.NewJSONBValue(entity.Limits)
	}

	return model
}

// BillingCycleModel represents the GORM model for billing_cycles table
type BillingCycleModel struct {
	database.BaseModel
	SubscriptionID uuid.UUID          `gorm:"type:uuid;not null;index"`
	Subscription   *SubscriptionModel `gorm:"foreignKey:SubscriptionID"`
	StartDate      time.Time          `gorm:"type:timestamp;not null"`
	EndDate        time.Time          `gorm:"type:timestamp;not null"`
	Amount         decimal.Decimal    `gorm:"type:decimal(10,2);not null"`
	Currency       string             `gorm:"type:varchar(3);not null"`
	Status         string             `gorm:"type:varchar(20);not null;default:'pending'"`
	InvoiceID      *string            `gorm:"type:varchar(255);uniqueIndex"`
	PaidAt         *time.Time         `gorm:"type:timestamp"`
}

// TableName specifies the table name for the BillingCycleModel
func (BillingCycleModel) TableName() string {
	return "billing_cycles"
}

// ToEntity converts the GORM model to domain entity
func (bc *BillingCycleModel) ToEntity() (*entities.BillingCycle, error) {
	entity := &entities.BillingCycle{
		ID:             bc.ID,
		SubscriptionID: bc.SubscriptionID,
		StartDate:      bc.StartDate,
		EndDate:        bc.EndDate,
		Amount:         bc.Amount,
		Currency:       bc.Currency,
		Status:         bc.Status,
		InvoiceID:      bc.InvoiceID,
		PaidAt:         bc.PaidAt,
		CreatedAt:      bc.CreatedAt,
		UpdatedAt:      bc.UpdatedAt,
	}

	return entity, nil
}

// BillingCycleFromEntity creates a GORM model from domain entity
func BillingCycleFromEntity(entity *entities.BillingCycle) *BillingCycleModel {
	model := &BillingCycleModel{
		BaseModel: database.BaseModel{
			ID:        entity.ID,
			CreatedAt: entity.CreatedAt,
			UpdatedAt: entity.UpdatedAt,
		},
		SubscriptionID: entity.SubscriptionID,
		StartDate:      entity.StartDate,
		EndDate:        entity.EndDate,
		Amount:         entity.Amount,
		Currency:       entity.Currency,
		Status:         entity.Status,
		InvoiceID:      entity.InvoiceID,
		PaidAt:         entity.PaidAt,
	}

	return model
}

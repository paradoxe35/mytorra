package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"

	"github.com/paradoxe35/torra/packages/database"
	"github.com/paradoxe35/torra/services/core/internal/adapters/postgres/types"
	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
	"github.com/paradoxe35/torra/services/core/internal/domain/valueobjects"
)

// OAuthProfileModel represents the GORM model for oauth_profiles table
type OAuthProfileModel struct {
	database.BaseModel
	UserID       uuid.UUID  `gorm:"type:uuid;not null;index"`
	User         *UserModel `gorm:"foreignKey:UserID"`
	Provider     string     `gorm:"type:varchar(50);not null;index"`
	ProviderID   string     `gorm:"type:varchar(255);not null;uniqueIndex:idx_provider_provider_id"`
	Email        string     `gorm:"type:varchar(255)"`
	Name         string     `gorm:"type:varchar(255)"`
	Picture      string     `gorm:"type:text"`
	AccessToken  string     `gorm:"type:text;not null"`
	RefreshToken string     `gorm:"type:text"`
	TokenExpiry  *time.Time
	Scopes       pq.StringArray   `gorm:"type:text[]"`
	RawProfile   types.JSONBValue `gorm:"type:jsonb"`
}

// TableName specifies the table name for the OAuthProfileModel
func (OAuthProfileModel) TableName() string {
	return "oauth_profiles"
}

// ToEntity converts the GORM model to domain entity
func (o *OAuthProfileModel) ToEntity() *entities.OAuthProfile {
	entity := &entities.OAuthProfile{
		ID:           o.ID,
		UserID:       o.UserID,
		Provider:     valueobjects.OAuthProvider(o.Provider),
		ProviderID:   o.ProviderID,
		Email:        o.Email,
		Name:         o.Name,
		Picture:      o.Picture,
		AccessToken:  o.AccessToken,
		RefreshToken: o.RefreshToken,
		TokenExpiry:  o.TokenExpiry,
		Scopes:       []string(o.Scopes),
		CreatedAt:    o.CreatedAt,
		UpdatedAt:    o.UpdatedAt,
	}

	// Handle RawProfile
	if o.RawProfile != nil {
		entity.RawProfile = json.RawMessage(o.RawProfile)
	}

	return entity
}

// OAuthProfileFromEntity creates a GORM model from domain entity
func OAuthProfileFromEntity(entity *entities.OAuthProfile) *OAuthProfileModel {
	model := &OAuthProfileModel{
		BaseModel: database.BaseModel{
			ID:        entity.ID,
			CreatedAt: entity.CreatedAt,
			UpdatedAt: entity.UpdatedAt,
		},
		UserID:       entity.UserID,
		Provider:     string(entity.Provider),
		ProviderID:   entity.ProviderID,
		Email:        entity.Email,
		Name:         entity.Name,
		Picture:      entity.Picture,
		AccessToken:  entity.AccessToken,
		RefreshToken: entity.RefreshToken,
		TokenExpiry:  entity.TokenExpiry,
		Scopes:       pq.StringArray(entity.Scopes),
	}

	// Handle RawProfile
	if entity.RawProfile != nil {
		model.RawProfile = types.JSONBValue(entity.RawProfile)
	}

	return model
}

// OAuthStateModel represents the GORM model for oauth_states table
type OAuthStateModel struct {
	State        string     `gorm:"type:varchar(255);primaryKey"`
	CodeVerifier *string    `gorm:"type:varchar(255)"`
	RedirectURI  string     `gorm:"type:text;not null"`
	Provider     string     `gorm:"type:varchar(50);not null"`
	IPAddress    *string    `gorm:"type:varchar(45)"`
	UserAgent    *string    `gorm:"type:text"`
	UserID       *uuid.UUID `gorm:"type:uuid"`
	User         *UserModel `gorm:"foreignKey:UserID"`
	CreatedAt    time.Time  `gorm:"not null;default:now()"`
	ExpiresAt    time.Time  `gorm:"not null"`
}

// TableName specifies the table name for the OAuthStateModel
func (OAuthStateModel) TableName() string {
	return "oauth_states"
}

// BeforeCreate hook to set default values
func (o *OAuthStateModel) BeforeCreate(tx *gorm.DB) error {
	if o.CreatedAt.IsZero() {
		o.CreatedAt = time.Now()
	}
	if o.ExpiresAt.IsZero() {
		o.ExpiresAt = time.Now().Add(10 * time.Minute)
	}
	return nil
}

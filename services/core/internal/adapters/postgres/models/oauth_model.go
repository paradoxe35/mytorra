package models

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/uptrace/bun"

	"github.com/paradoxe35/mytorra/packages/database"
	"github.com/paradoxe35/mytorra/services/core/internal/adapters/postgres/types"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/entities"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/valueobjects"
)

// OAuthProfileModel represents the Bun model for oauth_profiles table
type OAuthProfileModel struct {
	database.BaseModel `bun:"table:oauth_profiles,alias:op"`

	UserID       uuid.UUID        `bun:"user_id,type:uuid,notnull"`
	User         *UserModel       `bun:"rel:belongs-to,join:user_id=id"`
	Provider     string           `bun:"provider,notnull"`
	ProviderID   string           `bun:"provider_id,notnull,unique:idx_provider_provider_id"`
	Email        string           `bun:"email"`
	Name         string           `bun:"name"`
	Picture      string           `bun:"picture"`
	AccessToken  string           `bun:"access_token,notnull"`
	RefreshToken string           `bun:"refresh_token"`
	TokenExpiry  *time.Time       `bun:"token_expiry"`
	Scopes       pq.StringArray   `bun:"scopes,type:text[],array"`
	RawProfile   types.JSONBValue `bun:"raw_profile,type:jsonb"`
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

// OAuthStateModel represents the Bun model for oauth_states table
type OAuthStateModel struct {
	bun.BaseModel `bun:"table:oauth_states,alias:os"`

	State        string     `bun:"state,pk"`
	CodeVerifier *string    `bun:"code_verifier"`
	RedirectURI  string     `bun:"redirect_uri,notnull"`
	Provider     string     `bun:"provider,notnull"`
	IPAddress    *string    `bun:"ip_address"`
	UserAgent    *string    `bun:"user_agent"`
	UserID       *uuid.UUID `bun:"user_id,type:uuid"`
	User         *UserModel `bun:"rel:belongs-to,join:user_id=id"`
	CreatedAt    time.Time  `bun:"created_at,notnull,default:current_timestamp"`
	ExpiresAt    time.Time  `bun:"expires_at,notnull"`
}

// BeforeAppendModel is called before inserting or updating the model
func (o *OAuthStateModel) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		if o.CreatedAt.IsZero() {
			o.CreatedAt = time.Now()
		}
		if o.ExpiresAt.IsZero() {
			o.ExpiresAt = time.Now().Add(10 * time.Minute)
		}
	}
	return nil
}

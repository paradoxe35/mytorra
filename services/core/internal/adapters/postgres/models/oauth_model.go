package models

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// OAuthProfileModel represents the database model for oauth_profiles table
type OAuthProfileModel struct {
	ID           uuid.UUID       `db:"id"`
	UserID       uuid.UUID       `db:"user_id"`
	Provider     string          `db:"provider"`
	ProviderID   string          `db:"provider_id"`
	Email        sql.NullString  `db:"email"`
	Name         sql.NullString  `db:"name"`
	Picture      sql.NullString  `db:"picture"`
	AccessToken  string          `db:"access_token"`
	RefreshToken sql.NullString  `db:"refresh_token"`
	TokenExpiry  sql.NullTime    `db:"token_expiry"`
	Scopes       pq.StringArray  `db:"scopes"`
	RawProfile   json.RawMessage `db:"raw_profile"`
	CreatedAt    time.Time       `db:"created_at"`
	UpdatedAt    time.Time       `db:"updated_at"`
}

// OAuthStateModel represents the database model for oauth_states table
type OAuthStateModel struct {
	State        string         `db:"state"`
	CodeVerifier sql.NullString `db:"code_verifier"`
	RedirectURI  string         `db:"redirect_uri"`
	Provider     string         `db:"provider"`
	IPAddress    sql.NullString `db:"ip_address"`
	UserAgent    sql.NullString `db:"user_agent"`
	UserID       *uuid.UUID     `db:"user_id"`
	CreatedAt    time.Time      `db:"created_at"`
	ExpiresAt    time.Time      `db:"expires_at"`
}

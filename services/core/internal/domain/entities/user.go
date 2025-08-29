package entities

import (
	"time"

	"github.com/google/uuid"

	"github.com/paradoxe35/torra/services/core/internal/domain/valueobjects"
)

type User struct {
	ID              uuid.UUID          `db:"id" json:"id"`
	Email           valueobjects.Email `db:"email" json:"email"`
	Phone           valueobjects.Phone `db:"phone" json:"phone"`
	Name            string             `db:"name" json:"name"`
	PasswordHash    string             `db:"password_hash" json:"-"`
	ProfileImageURL string             `db:"profile_image_url" json:"profile_image_url"`

	// Verification
	EmailVerified     bool `db:"email_verified" json:"email_verified"`
	PhoneVerified     bool `db:"phone_verified" json:"phone_verified"`
	IDVerified        bool `db:"id_verified" json:"id_verified"`
	VerificationLevel int  `db:"verification_level" json:"verification_level"`

	// OAuth
	GoogleID   *string `db:"google_id" json:"google_id,omitempty"`
	FacebookID *string `db:"facebook_id" json:"facebook_id,omitempty"`

	// MFA
	MFAEnabled bool    `db:"mfa_enabled" json:"mfa_enabled"`
	MFASecret  *string `db:"mfa_secret" json:"-"`

	// Location
	Area         string   `db:"area" json:"area"` // "bukavu" or "goma"
	LastKnownLat *float64 `db:"last_known_lat" json:"last_known_lat,omitempty"`
	LastKnownLng *float64 `db:"last_known_lng" json:"last_known_lng,omitempty"`

	// Metadata
	Status      UserStatus `db:"status" json:"status"`
	LastLoginAt *time.Time `db:"last_login_at" json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusBanned    UserStatus = "banned"
	UserStatusDeleted   UserStatus = "deleted"
)

// Domain methods
func (u *User) CanLogin() bool {
	return u.Status == UserStatusActive && u.EmailVerified
}

func (u *User) RequiresMFA() bool {
	return u.MFAEnabled && u.MFASecret != nil
}

func (u *User) UpdateVerificationLevel() {
	level := 0
	if u.EmailVerified {
		level++
	}
	if u.PhoneVerified {
		level++
	}
	if u.IDVerified {
		level++
	}
	// Additional business logic for levels 4 and 5
	u.VerificationLevel = level
}

func (u *User) IsFromArea(area string) bool {
	return u.Area == area
}

func (u *User) SetLocation(lat, lng float64) {
	u.LastKnownLat = &lat
	u.LastKnownLng = &lng
	u.UpdatedAt = time.Now()
}

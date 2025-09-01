package models

import (
	"time"

	"github.com/google/uuid"

	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
	"github.com/paradoxe35/torra/services/core/internal/domain/valueobjects"
)

// UserModel represents the database model for users
type UserModel struct {
	ID                uuid.UUID  `db:"id"`
	Email             string     `db:"email"`
	Phone             string     `db:"phone"`
	Name              string     `db:"name"`
	PasswordHash      string     `db:"password_hash"`
	ProfileImageURL   string     `db:"profile_image_url"`
	EmailVerified     bool       `db:"email_verified"`
	PhoneVerified     bool       `db:"phone_verified"`
	IDVerified        bool       `db:"id_verified"`
	VerificationLevel int        `db:"verification_level"`
	GoogleID          *string    `db:"google_id"`
	FacebookID        *string    `db:"facebook_id"`
	MFAEnabled        bool       `db:"mfa_enabled"`
	MFASecret         *string    `db:"mfa_secret"`
	AreaID            uuid.UUID  `db:"area_id"`
	LastKnownLat      *float64   `db:"last_known_lat"`
	LastKnownLng      *float64   `db:"last_known_lng"`
	Status            string     `db:"status"`
	LastLoginAt       *time.Time `db:"last_login_at"`
	CreatedAt         time.Time  `db:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at"`
	DeletedAt         *time.Time `db:"deleted_at"`
}

// ToEntity converts the database model to a domain entity
func (m *UserModel) ToEntity() (*entities.User, error) {
	email, err := valueobjects.NewEmail(m.Email)
	if err != nil {
		return nil, err
	}

	phone, err := valueobjects.NewPhone(m.Phone)
	if err != nil {
		return nil, err
	}

	status := entities.UserStatus(m.Status)

	return &entities.User{
		ID:                m.ID,
		Email:             email,
		Phone:             phone,
		Name:              m.Name,
		PasswordHash:      m.PasswordHash,
		ProfileImageURL:   m.ProfileImageURL,
		EmailVerified:     m.EmailVerified,
		PhoneVerified:     m.PhoneVerified,
		IDVerified:        m.IDVerified,
		VerificationLevel: m.VerificationLevel,
		GoogleID:          m.GoogleID,
		FacebookID:        m.FacebookID,
		MFAEnabled:        m.MFAEnabled,
		MFASecret:         m.MFASecret,
		AreaID:            m.AreaID,
		LastKnownLat:      m.LastKnownLat,
		LastKnownLng:      m.LastKnownLng,
		Status:            status,
		LastLoginAt:       m.LastLoginAt,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
		DeletedAt:         m.DeletedAt,
	}, nil
}

// UserModelFromEntity creates a database model from a domain entity
func UserModelFromEntity(user *entities.User) *UserModel {
	return &UserModel{
		ID:                user.ID,
		Email:             user.Email.String(),
		Phone:             user.Phone.String(),
		Name:              user.Name,
		PasswordHash:      user.PasswordHash,
		ProfileImageURL:   user.ProfileImageURL,
		EmailVerified:     user.EmailVerified,
		PhoneVerified:     user.PhoneVerified,
		IDVerified:        user.IDVerified,
		VerificationLevel: user.VerificationLevel,
		GoogleID:          user.GoogleID,
		FacebookID:        user.FacebookID,
		MFAEnabled:        user.MFAEnabled,
		MFASecret:         user.MFASecret,
		AreaID:            user.AreaID,
		LastKnownLat:      user.LastKnownLat,
		LastKnownLng:      user.LastKnownLng,
		Status:            string(user.Status),
		LastLoginAt:       user.LastLoginAt,
		CreatedAt:         user.CreatedAt,
		UpdatedAt:         user.UpdatedAt,
		DeletedAt:         user.DeletedAt,
	}
}

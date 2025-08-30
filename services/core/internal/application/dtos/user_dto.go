package dtos

import (
	"time"

	"github.com/google/uuid"
)

// CreateUserRequest represents a user creation request
type CreateUserRequest struct {
	Email    string    `json:"email" validate:"required,email,max=255"`
	Phone    string    `json:"phone" validate:"required,e164"`
	Name     string    `json:"name" validate:"required,min=2,max=100"`
	Password string    `json:"password" validate:"required,min=12"`
	AreaID   uuid.UUID `json:"area_id" validate:"required"`
}

// UpdateUserRequest represents a user update request
type UpdateUserRequest struct {
	Name            *string    `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Phone           *string    `json:"phone,omitempty" validate:"omitempty,e164"`
	ProfileImageURL *string    `json:"profile_image_url,omitempty" validate:"omitempty,url"`
	AreaID          *uuid.UUID `json:"area_id,omitempty" validate:"omitempty"`
	LastKnownLat    *float64   `json:"last_known_lat,omitempty" validate:"omitempty,latitude"`
	LastKnownLng    *float64   `json:"last_known_lng,omitempty" validate:"omitempty,longitude"`
}

// UserResponse represents a user response
type UserResponse struct {
	ID                uuid.UUID  `json:"id"`
	Email             string     `json:"email"`
	Phone             string     `json:"phone"`
	Name              string     `json:"name"`
	ProfileImageURL   string     `json:"profile_image_url,omitempty"`
	EmailVerified     bool       `json:"email_verified"`
	PhoneVerified     bool       `json:"phone_verified"`
	IDVerified        bool       `json:"id_verified"`
	VerificationLevel int        `json:"verification_level"`
	AreaID            uuid.UUID  `json:"area_id"`
	AreaName          string     `json:"area_name,omitempty"`
	LastKnownLat      *float64   `json:"last_known_lat,omitempty"`
	LastKnownLng      *float64   `json:"last_known_lng,omitempty"`
	Status            string     `json:"status"`
	LastLoginAt       *time.Time `json:"last_login_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// VerifyEmailRequest represents an email verification request
type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

// VerifyPhoneRequest represents a phone verification request
type VerifyPhoneRequest struct {
	Code string `json:"code" validate:"required,len=6"`
}

// VerifyIDRequest represents a government ID verification request
type VerifyIDRequest struct {
	IDType      string `json:"id_type" validate:"required,oneof=passport national_id driver_license"`
	IDNumber    string `json:"id_number" validate:"required"`
	FrontImage  string `json:"front_image" validate:"required,base64"`
	BackImage   string `json:"back_image,omitempty" validate:"omitempty,base64"`
	SelfieImage string `json:"selfie_image" validate:"required,base64"`
}

// EnableMFARequest represents an MFA enablement request
type EnableMFARequest struct {
	Secret string `json:"secret" validate:"required"`
	Code   string `json:"code" validate:"required,len=6"`
}

// DisableMFARequest represents an MFA disable request
type DisableMFARequest struct {
	Code     string `json:"code" validate:"required,len=6"`
	Password string `json:"password" validate:"required"`
}

package dtos

import (
	"time"

	"github.com/google/uuid"
)

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	MFACode  string `json:"mfa_code,omitempty" validate:"omitempty,len=6"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Phone    string `json:"phone" validate:"required,e164"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Password string `json:"password" validate:"required,min=12"`
	Area     string `json:"area" validate:"required,oneof=bukavu goma"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	AccessToken  string        `json:"access_token,omitempty"`
	RefreshToken string        `json:"refresh_token,omitempty"`
	ExpiresIn    int           `json:"expires_in,omitempty"`
	TokenType    string        `json:"token_type,omitempty"`
	RequiresMFA  bool          `json:"requires_mfa"`
	User         *UserResponse `json:"user,omitempty"`
}

// RefreshTokenRequest represents a token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// ForgotPasswordRequest represents a forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ResetPasswordRequest represents a password reset request
type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=12"`
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=12"`
}

// OAuthCallbackRequest represents an OAuth callback request
type OAuthCallbackRequest struct {
	Code  string `json:"code" validate:"required"`
	State string `json:"state" validate:"required"`
}

// TokenClaims represents JWT token claims
type TokenClaims struct {
	UserID            uuid.UUID `json:"sub"`
	Email             string    `json:"email"`
	Name              string    `json:"name"`
	Area              string    `json:"area"`
	VerificationLevel int       `json:"verification_level"`
	IssuedAt          time.Time `json:"iat"`
	ExpiresAt         time.Time `json:"exp"`
	Issuer            string    `json:"iss"`
}

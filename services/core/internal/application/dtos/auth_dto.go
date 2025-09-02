package dtos

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/paradoxe35/mytorra/packages/errors"
	"github.com/paradoxe35/mytorra/packages/i18n"
	"github.com/paradoxe35/mytorra/packages/security"
)

// Custom validator instance
var validate *validator.Validate

func init() {
	validate = validator.New()

	// Register custom validation tags
	validate.RegisterValidation("password", validatePassword)
	validate.RegisterValidation("phone_e164", validatePhoneE164)
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	MFACode  string `json:"mfa_code,omitempty" validate:"omitempty,len=6,numeric"`
}

// Validate performs validation with i18n support
func (r *LoginRequest) Validate(ctx context.Context, translator *i18n.Translator) error {
	r.Email = strings.TrimSpace(strings.ToLower(r.Email))

	if err := validate.Struct(r); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldErr := range validationErrors {
				switch fieldErr.Field() {
				case "Email":
					return errors.ValidationError(translator.TWithContext(ctx, "validation.email"))
				case "Password":
					return errors.ValidationError(translator.TWithContext(ctx, "validation.required"))
				case "MFACode":
					return errors.ValidationError(translator.TWithContext(ctx, "auth.mfa_invalid"))
				}
			}
		}
	}

	return nil
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email           string     `json:"email" validate:"required,email,max=255"`
	Phone           string     `json:"phone" validate:"required,phone_e164"`
	Name            string     `json:"name" validate:"required,min=2,max=100"`
	Password        string     `json:"password" validate:"required,password"`
	ConfirmPassword string     `json:"confirm_password" validate:"required,eqfield=Password"`
	AreaID          *uuid.UUID `json:"area_id,omitempty"` // Optional - will be auto-detected if not provided
	AcceptTerms     bool       `json:"accept_terms" validate:"required,eq=true"`
	Language        string     `json:"language" validate:"omitempty,oneof=en fr sw"`

	// Location data for area detection (optional)
	Latitude  *float64 `json:"latitude,omitempty" validate:"omitempty,latitude"`
	Longitude *float64 `json:"longitude,omitempty" validate:"omitempty,longitude"`
}

// Validate performs validation with i18n support
func (r *RegisterRequest) Validate(ctx context.Context, translator *i18n.Translator) error {
	// Sanitize inputs
	r.Email = strings.TrimSpace(strings.ToLower(r.Email))
	r.Phone = strings.TrimSpace(r.Phone)
	r.Name = security.SanitizeInput(strings.TrimSpace(r.Name))

	// Run struct validation
	if err := validate.Struct(r); err != nil {
		return r.translateValidationError(ctx, err, translator)
	}

	// Additional security checks
	if security.ContainsSQLInjectionPattern(r.Name) {
		return errors.ValidationError(translator.TWithContext(ctx, "validation.invalid_input"))
	}

	return nil
}

// translateValidationError translates validation errors to user-friendly messages
func (r *RegisterRequest) translateValidationError(ctx context.Context, err error, translator *i18n.Translator) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var messages []string

		for _, fieldErr := range validationErrors {
			var msg string

			switch fieldErr.Field() {
			case "Email":
				if fieldErr.Tag() == "required" {
					msg = translator.TWithContext(ctx, "validation.required")
				} else {
					msg = translator.TWithContext(ctx, "validation.email")
				}

			case "Phone":
				if fieldErr.Tag() == "required" {
					msg = translator.TWithContext(ctx, "validation.required")
				} else {
					msg = translator.TWithContext(ctx, "validation.phone")
				}

			case "Name":
				switch fieldErr.Tag() {
				case "required":
					msg = translator.TWithContext(ctx, "validation.required")
				case "min":
					msg = translator.TWithContext(ctx, "validation.min_length", 2)
				case "max":
					msg = translator.TWithContext(ctx, "validation.max_length", 100)
				}

			case "Password":
				if fieldErr.Tag() == "required" {
					msg = translator.TWithContext(ctx, "validation.required")
				} else {
					msg = r.getPasswordErrorMessage(ctx, translator)
				}

			case "ConfirmPassword":
				if fieldErr.Tag() == "required" {
					msg = translator.TWithContext(ctx, "validation.required")
				} else {
					msg = translator.TWithContext(ctx, "validation.password_mismatch")
				}

			case "AreaID":
				msg = translator.TWithContext(ctx, "validation.area_invalid")

			case "AcceptTerms":
				msg = translator.TWithContext(ctx, "validation.terms_required")
			}

			if msg != "" {
				messages = append(messages, fmt.Sprintf("%s: %s", fieldErr.Field(), msg))
			}
		}

		if len(messages) > 0 {
			return errors.ValidationError(strings.Join(messages, "; "))
		}
	}

	return errors.ValidationError(translator.TWithContext(ctx, "validation.invalid_input"))
}

// getPasswordErrorMessage returns specific password validation error
func (r *RegisterRequest) getPasswordErrorMessage(ctx context.Context, translator *i18n.Translator) string {
	if err := security.ValidatePassword(r.Password); err != nil {
		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "12 characters"):
			return translator.TWithContext(ctx, "validation.password.min", 12)
		case strings.Contains(errMsg, "uppercase"):
			return translator.TWithContext(ctx, "validation.password.uppercase")
		case strings.Contains(errMsg, "lowercase"):
			return translator.TWithContext(ctx, "validation.password.lowercase")
		case strings.Contains(errMsg, "number"):
			return translator.TWithContext(ctx, "validation.password.number")
		case strings.Contains(errMsg, "special"):
			return translator.TWithContext(ctx, "validation.password.special")
		}
	}
	return translator.TWithContext(ctx, "validation.password.min", 12)
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

// Validate performs validation
func (r *ForgotPasswordRequest) Validate(ctx context.Context, translator *i18n.Translator) error {
	r.Email = strings.TrimSpace(strings.ToLower(r.Email))

	if err := validate.Struct(r); err != nil {
		return errors.ValidationError(translator.TWithContext(ctx, "validation.email"))
	}

	return nil
}

// ResetPasswordRequest represents a password reset request
type ResetPasswordRequest struct {
	Token           string `json:"token" validate:"required"`
	Password        string `json:"password" validate:"required,password"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}

// Validate performs validation
func (r *ResetPasswordRequest) Validate(ctx context.Context, translator *i18n.Translator) error {
	if err := validate.Struct(r); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldErr := range validationErrors {
				switch fieldErr.Field() {
				case "Token":
					return errors.ValidationError(translator.TWithContext(ctx, "validation.required"))
				case "Password":
					if fieldErr.Tag() == "required" {
						return errors.ValidationError(translator.TWithContext(ctx, "validation.required"))
					}
					// Get specific password error
					if err := security.ValidatePassword(r.Password); err != nil {
						return errors.ValidationError(err.Error())
					}
				case "ConfirmPassword":
					return errors.ValidationError(translator.TWithContext(ctx, "validation.password_mismatch"))
				}
			}
		}
	}

	return nil
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,password"`
}

// VerifyEmailRequest represents email verification request
type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

// VerifyPhoneRequest represents phone verification request
type VerifyPhoneRequest struct {
	Code string `json:"code" validate:"required,len=6,numeric"`
}

// Validate performs validation
func (r *VerifyPhoneRequest) Validate(ctx context.Context, translator *i18n.Translator) error {
	if err := validate.Struct(r); err != nil {
		return errors.ValidationError(translator.TWithContext(ctx, "validation.invalid_code"))
	}
	return nil
}

// EnableMFARequest represents MFA enablement request
type EnableMFARequest struct {
	Password string `json:"password" validate:"required"`
}

// DisableMFARequest represents MFA disable request
type DisableMFARequest struct {
	Password string `json:"password" validate:"required"`
	MFACode  string `json:"mfa_code" validate:"required,len=6,numeric"`
}

// UpdateAreaRequest represents area update request
type UpdateAreaRequest struct {
	AreaID uuid.UUID `json:"area_id" validate:"required"`
}

// OAuthCallbackRequest represents an OAuth callback request
type OAuthCallbackRequest struct {
	Code     string `json:"code" validate:"required"`
	State    string `json:"state" validate:"required"`
	Provider string `json:"provider" validate:"required,oneof=google facebook"`
}

// TokenClaims represents JWT token claims
type TokenClaims struct {
	UserID            uuid.UUID `json:"sub"`
	Email             string    `json:"email"`
	Name              string    `json:"name"`
	AreaID            uuid.UUID `json:"area_id"`
	VerificationLevel int       `json:"verification_level"`
	IssuedAt          time.Time `json:"iat"`
	ExpiresAt         time.Time `json:"exp"`
	Issuer            string    `json:"iss"`
}

// Custom validators

func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	return security.ValidatePassword(password) == nil
}

func validatePhoneE164(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	return security.ValidatePhone(phone) == nil
}

// MFASetupResponse represents MFA setup response
type MFASetupResponse struct {
	Secret      string   `json:"secret"`
	QRCode      string   `json:"qr_code"`
	BackupCodes []string `json:"backup_codes"`
}

// EnableMFAResponse represents the response after enabling MFA
type EnableMFAResponse struct {
	Secret string `json:"secret"`
	QRCode string `json:"qr_code"`
}

// VerificationStatusResponse represents the user's verification status
type VerificationStatusResponse struct {
	EmailVerified     bool `json:"email_verified"`
	PhoneVerified     bool `json:"phone_verified"`
	IDVerified        bool `json:"id_verified"`
	VerificationLevel int  `json:"verification_level"`
}

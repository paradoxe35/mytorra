package dtos_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/paradoxe35/torra/packages/i18n"
	"github.com/paradoxe35/torra/services/core/internal/application/dtos"
)

func setupTranslator(t *testing.T) *i18n.Translator {
	translator, err := i18n.New("en", "fr", "sw")
	require.NoError(t, err)
	return translator
}

func TestRegisterRequest_Validate(t *testing.T) {
	translator := setupTranslator(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		request dtos.RegisterRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid registration request",
			request: dtos.RegisterRequest{
				Email:           "test@example.com",
				Phone:           "+250780123456",
				Name:            "John Doe",
				Password:        "SecurePass123!@#",
				ConfirmPassword: "SecurePass123!@#",
				AreaID:          func() *uuid.UUID { id := uuid.New(); return &id }(),
				AcceptTerms:     true,
				Language:        "en",
			},
			wantErr: false,
		},
		{
			name: "invalid email format",
			request: dtos.RegisterRequest{
				Email:           "invalid-email",
				Phone:           "+250780123456",
				Name:            "John Doe",
				Password:        "SecurePass123!@#",
				ConfirmPassword: "SecurePass123!@#",
				AreaID:          func() *uuid.UUID { id := uuid.New(); return &id }(),
				AcceptTerms:     true,
			},
			wantErr: true,
			errMsg:  "email",
		},
		{
			name: "invalid phone format",
			request: dtos.RegisterRequest{
				Email:           "test@example.com",
				Phone:           "0780123456", // Missing country code
				Name:            "John Doe",
				Password:        "SecurePass123!@#",
				ConfirmPassword: "SecurePass123!@#",
				AreaID:          func() *uuid.UUID { id := uuid.New(); return &id }(),
				AcceptTerms:     true,
			},
			wantErr: true,
			errMsg:  "phone",
		},
		{
			name: "password too short",
			request: dtos.RegisterRequest{
				Email:           "test@example.com",
				Phone:           "+250780123456",
				Name:            "John Doe",
				Password:        "Short1!",
				ConfirmPassword: "Short1!",
				AreaID:          func() *uuid.UUID { id := uuid.New(); return &id }(),
				AcceptTerms:     true,
			},
			wantErr: true,
			errMsg:  "12 characters",
		},
		{
			name: "password mismatch",
			request: dtos.RegisterRequest{
				Email:           "test@example.com",
				Phone:           "+250780123456",
				Name:            "John Doe",
				Password:        "SecurePass123!@#",
				ConfirmPassword: "DifferentPass123!@#",
				AreaID:          func() *uuid.UUID { id := uuid.New(); return &id }(),
				AcceptTerms:     true,
			},
			wantErr: true,
			errMsg:  "ConfirmPassword",
		},
		{
			name: "terms not accepted",
			request: dtos.RegisterRequest{
				Email:           "test@example.com",
				Phone:           "+250780123456",
				Name:            "John Doe",
				Password:        "SecurePass123!@#",
				ConfirmPassword: "SecurePass123!@#",
				AreaID:          func() *uuid.UUID { id := uuid.New(); return &id }(),
				AcceptTerms:     false,
			},
			wantErr: true,
			errMsg:  "terms",
		},
		{
			name: "name too short",
			request: dtos.RegisterRequest{
				Email:           "test@example.com",
				Phone:           "+250780123456",
				Name:            "J",
				Password:        "SecurePass123!@#",
				ConfirmPassword: "SecurePass123!@#",
				AreaID:          func() *uuid.UUID { id := uuid.New(); return &id }(),
				AcceptTerms:     true,
			},
			wantErr: true,
			errMsg:  "Name",
		},
		{
			name: "invalid language",
			request: dtos.RegisterRequest{
				Email:           "test@example.com",
				Phone:           "+250780123456",
				Name:            "John Doe",
				Password:        "SecurePass123!@#",
				ConfirmPassword: "SecurePass123!@#",
				AreaID:          func() *uuid.UUID { id := uuid.New(); return &id }(),
				AcceptTerms:     true,
				Language:        "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate(ctx, translator)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoginRequest_Validate(t *testing.T) {
	translator := setupTranslator(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		request dtos.LoginRequest
		wantErr bool
	}{
		{
			name: "valid login request",
			request: dtos.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "valid login with MFA",
			request: dtos.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
				MFACode:  "123456",
			},
			wantErr: false,
		},
		{
			name: "invalid email",
			request: dtos.LoginRequest{
				Email:    "invalid-email",
				Password: "password123",
			},
			wantErr: true,
		},
		{
			name: "missing password",
			request: dtos.LoginRequest{
				Email:    "test@example.com",
				Password: "",
			},
			wantErr: true,
		},
		{
			name: "invalid MFA code",
			request: dtos.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
				MFACode:  "12345", // Too short
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate(ctx, translator)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestForgotPasswordRequest_Validate(t *testing.T) {
	translator := setupTranslator(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		request dtos.ForgotPasswordRequest
		wantErr bool
	}{
		{
			name: "valid email",
			request: dtos.ForgotPasswordRequest{
				Email: "test@example.com",
			},
			wantErr: false,
		},
		{
			name: "invalid email",
			request: dtos.ForgotPasswordRequest{
				Email: "not-an-email",
			},
			wantErr: true,
		},
		{
			name: "empty email",
			request: dtos.ForgotPasswordRequest{
				Email: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate(ctx, translator)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestResetPasswordRequest_Validate(t *testing.T) {
	translator := setupTranslator(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		request dtos.ResetPasswordRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid reset request",
			request: dtos.ResetPasswordRequest{
				Token:           "valid-token",
				Password:        "NewSecurePass123!@#",
				ConfirmPassword: "NewSecurePass123!@#",
			},
			wantErr: false,
		},
		{
			name: "missing token",
			request: dtos.ResetPasswordRequest{
				Token:           "",
				Password:        "NewSecurePass123!@#",
				ConfirmPassword: "NewSecurePass123!@#",
			},
			wantErr: true,
		},
		{
			name: "weak password",
			request: dtos.ResetPasswordRequest{
				Token:           "valid-token",
				Password:        "weak",
				ConfirmPassword: "weak",
			},
			wantErr: true,
		},
		{
			name: "password mismatch",
			request: dtos.ResetPasswordRequest{
				Token:           "valid-token",
				Password:        "NewSecurePass123!@#",
				ConfirmPassword: "DifferentPass123!@#",
			},
			wantErr: true,
			errMsg:  "Passwords do not match",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate(ctx, translator)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegisterRequest_SanitizeInput(t *testing.T) {
	translator := setupTranslator(t)
	ctx := context.Background()

	request := dtos.RegisterRequest{
		Email:           "  TEST@EXAMPLE.COM  ",
		Phone:           "  +250780123456  ",
		Name:            "  John Doe  ",
		Password:        "SecurePass123!@#",
		ConfirmPassword: "SecurePass123!@#",
		AreaID:          func() *uuid.UUID { id := uuid.New(); return &id }(),
		AcceptTerms:     true,
	}

	err := request.Validate(ctx, translator)
	assert.NoError(t, err)

	// Check that inputs were sanitized
	assert.Equal(t, "test@example.com", request.Email)
	assert.Equal(t, "+250780123456", request.Phone)
	assert.Equal(t, "John Doe", request.Name)
}

func TestPasswordValidation_SpecificErrors(t *testing.T) {
	translator := setupTranslator(t)
	ctx := i18n.WithLanguage(context.Background(), "en")

	tests := []struct {
		name     string
		password string
		errMsg   string
	}{
		{
			name:     "no uppercase",
			password: "lowercase123!@#",
			errMsg:   "uppercase",
		},
		{
			name:     "no lowercase",
			password: "UPPERCASE123!@#",
			errMsg:   "lowercase",
		},
		{
			name:     "no numbers",
			password: "NoNumbers!@#ABC",
			errMsg:   "number",
		},
		{
			name:     "no special characters",
			password: "NoSpecial123ABC",
			errMsg:   "special",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := dtos.RegisterRequest{
				Email:           "test@example.com",
				Phone:           "+250780123456",
				Name:            "John Doe",
				Password:        tt.password,
				ConfirmPassword: tt.password,
				AreaID:          func() *uuid.UUID { id := uuid.New(); return &id }(),
				AcceptTerms:     true,
			}

			err := request.Validate(ctx, translator)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestI18nTranslation(t *testing.T) {
	translator := setupTranslator(t)

	// Test French translation
	ctxFr := i18n.WithLanguage(context.Background(), "fr")
	request := dtos.RegisterRequest{
		Email:           "invalid-email",
		Phone:           "+250780123456",
		Name:            "John Doe",
		Password:        "SecurePass123!@#",
		ConfirmPassword: "SecurePass123!@#",
		AreaID:          func() *uuid.UUID { id := uuid.New(); return &id }(),
		AcceptTerms:     true,
	}

	err := request.Validate(ctxFr, translator)
	assert.Error(t, err)
	// Should contain French error message
	assert.Contains(t, err.Error(), "email")

	// Test with unsupported language (should fallback to default)
	ctxUnsupported := i18n.WithLanguage(context.Background(), "zh")
	err = request.Validate(ctxUnsupported, translator)
	assert.Error(t, err)
}

func BenchmarkRegisterRequest_Validate(b *testing.B) {
	translator, _ := i18n.New("en")
	ctx := context.Background()

	request := dtos.RegisterRequest{
		Email:           "test@example.com",
		Phone:           "+250780123456",
		Name:            "John Doe",
		Password:        "SecurePass123!@#",
		ConfirmPassword: "SecurePass123!@#",
		AreaID:          func() *uuid.UUID { id := uuid.New(); return &id }(),
		AcceptTerms:     true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = request.Validate(ctx, translator)
	}
}

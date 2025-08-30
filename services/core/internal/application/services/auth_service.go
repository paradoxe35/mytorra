package services

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/paradoxe35/torra/packages/errors"
	"github.com/paradoxe35/torra/packages/i18n"
	"github.com/paradoxe35/torra/packages/logger"
	"github.com/paradoxe35/torra/packages/security"
	"github.com/paradoxe35/torra/services/core/internal/application/dtos"
	"github.com/paradoxe35/torra/services/core/internal/application/ports"
	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
	"github.com/paradoxe35/torra/services/core/internal/domain/valueobjects"
)

type AuthService struct {
	userRepo        ports.UserRepository
	cache           ports.Cache
	logger          *logger.Logger
	translator      *i18n.Translator
	jwtSecret       []byte
	jwtIssuer       string
	accessDuration  time.Duration
	refreshDuration time.Duration
}

func NewAuthService(
	userRepo ports.UserRepository,
	cache ports.Cache,
	logger *logger.Logger,
	translator *i18n.Translator,
	jwtSecret string,
	jwtIssuer string,
	accessDuration time.Duration,
	refreshDuration time.Duration,
) *AuthService {
	return &AuthService{
		userRepo:        userRepo,
		cache:           cache,
		logger:          logger,
		translator:      translator,
		jwtSecret:       []byte(jwtSecret),
		jwtIssuer:       jwtIssuer,
		accessDuration:  accessDuration,
		refreshDuration: refreshDuration,
	}
}

func (s *AuthService) Login(ctx context.Context, req dtos.LoginRequest) (*dtos.AuthResponse, error) {
	// Validate request
	if err := req.Validate(ctx, s.translator); err != nil {
		return nil, err
	}

	// Find user by email
	email, err := valueobjects.NewEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		s.logger.WithError(err).Debug("User not found")
		return nil, errors.ErrInvalidCredentials
	}

	// Check if user can login
	if !user.CanLogin() {
		return nil, errors.ErrAccountSuspended
	}

	// Verify password
	if err := security.CheckPassword(req.Password, user.PasswordHash); err != nil {
		return nil, errors.ErrInvalidCredentials
	}

	// Check MFA if enabled
	if user.RequiresMFA() && req.MFACode == "" {
		return &dtos.AuthResponse{
			RequiresMFA: true,
		}, nil
	}

	if user.RequiresMFA() {
		// TODO: Implement MFA verification
		// if !s.verifyMFACode(user.MFASecret, req.MFACode) {
		//     return nil, errors.ErrInvalidMFACode
		// }
	}

	// Generate tokens
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Update last login
	now := time.Now()
	user.LastLoginAt = &now
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.WithError(err).Error("Failed to update last login")
	}

	// Store refresh token in cache
	refreshKey := fmt.Sprintf("refresh:%s", user.ID)
	if err := s.cache.Set(ctx, refreshKey, []byte(refreshToken), s.refreshDuration); err != nil {
		s.logger.WithError(err).Error("Failed to store refresh token")
	}

	return &dtos.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(s.accessDuration.Seconds()),
		User:         s.toUserResponse(user),
		RequiresMFA:  false,
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, req dtos.RefreshTokenRequest) (*dtos.AuthResponse, error) {
	// Parse refresh token
	token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, errors.ErrInvalidToken
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.ErrInvalidToken
	}

	// Get user ID from claims
	userIDStr, ok := claims["sub"].(string)
	if !ok {
		return nil, errors.ErrInvalidToken
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.ErrInvalidToken
	}

	// Check if refresh token exists in cache
	refreshKey := fmt.Sprintf("refresh:%s", userID)
	storedToken, err := s.cache.Get(ctx, refreshKey)
	if err != nil || string(storedToken) != req.RefreshToken {
		return nil, errors.ErrInvalidToken
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	// Generate new tokens
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Update refresh token in cache
	if err := s.cache.Set(ctx, refreshKey, []byte(newRefreshToken), s.refreshDuration); err != nil {
		s.logger.WithError(err).Error("Failed to update refresh token")
	}

	return &dtos.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int(s.accessDuration.Seconds()),
		User:         s.toUserResponse(user),
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, userID string) error {
	// Remove refresh token from cache
	refreshKey := fmt.Sprintf("refresh:%s", userID)
	if err := s.cache.Delete(ctx, refreshKey); err != nil {
		return fmt.Errorf("failed to remove refresh token: %w", err)
	}

	// TODO: Add token to blacklist if needed

	return nil
}

func (s *AuthService) ForgotPassword(ctx context.Context, req dtos.ForgotPasswordRequest) error {
	// Validate request
	if err := req.Validate(ctx, s.translator); err != nil {
		return err
	}

	// Find user by email
	email, err := valueobjects.NewEmail(req.Email)
	if err != nil {
		return fmt.Errorf("invalid email: %w", err)
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		// Don't reveal if user exists or not
		return nil
	}

	// Generate reset token
	resetToken := uuid.New().String()
	resetKey := fmt.Sprintf("password_reset:%s", resetToken)

	// Store reset token with user ID (expires in 1 hour)
	if err := s.cache.Set(ctx, resetKey, []byte(user.ID.String()), time.Hour); err != nil {
		return fmt.Errorf("failed to store reset token: %w", err)
	}

	// TODO: Send password reset email
	// s.emailService.SendPasswordResetEmail(user.Email, resetToken)

	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, req dtos.ResetPasswordRequest) error {
	// Validate request
	if err := req.Validate(ctx, s.translator); err != nil {
		return err
	}

	// Get user ID from reset token
	resetKey := fmt.Sprintf("password_reset:%s", req.Token)
	userIDBytes, err := s.cache.Get(ctx, resetKey)
	if err != nil {
		return errors.ErrInvalidToken
	}

	userID, err := uuid.Parse(string(userIDBytes))
	if err != nil {
		return errors.ErrInvalidToken
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return errors.ErrUserNotFound
	}

	// Hash new password
	hashedPassword, err := security.HashPassword(req.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	user.PasswordHash = hashedPassword
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Delete reset token
	if err := s.cache.Delete(ctx, resetKey); err != nil {
		s.logger.WithError(err).Error("Failed to delete reset token")
	}

	// TODO: Send password changed notification email

	return nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	// Get user ID from verification token
	verifyKey := fmt.Sprintf("email_verify:%s", token)
	userIDBytes, err := s.cache.Get(ctx, verifyKey)
	if err != nil {
		return errors.ErrInvalidToken
	}

	userID, err := uuid.Parse(string(userIDBytes))
	if err != nil {
		return errors.ErrInvalidToken
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return errors.ErrUserNotFound
	}

	// Update email verification status
	user.EmailVerified = true
	user.UpdateVerificationLevel()
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Delete verification token
	if err := s.cache.Delete(ctx, verifyKey); err != nil {
		s.logger.WithError(err).Error("Failed to delete verification token")
	}

	return nil
}

func (s *AuthService) VerifyPhone(ctx context.Context, req dtos.VerifyPhoneRequest) error {
	// Validate request
	if err := req.Validate(ctx, s.translator); err != nil {
		return err
	}

	// Get user ID from context or token
	// TODO: Implement phone verification logic

	return nil
}

func (s *AuthService) generateAccessToken(user *entities.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":     user.ID.String(),
		"email":   user.Email.String(),
		"name":    user.Name,
		"area_id": user.AreaID.String(),
		"iat":     time.Now().Unix(),
		"exp":     time.Now().Add(s.accessDuration).Unix(),
		"iss":     s.jwtIssuer,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *AuthService) generateRefreshToken(user *entities.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":  user.ID.String(),
		"type": "refresh",
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(s.refreshDuration).Unix(),
		"iss":  s.jwtIssuer,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// toUserResponse converts a User entity to UserResponse DTO
func (s *AuthService) toUserResponse(user *entities.User) *dtos.UserResponse {
	return &dtos.UserResponse{
		ID:                user.ID,
		Email:             user.Email.String(),
		Phone:             user.Phone.String(),
		Name:              user.Name,
		ProfileImageURL:   user.ProfileImageURL,
		EmailVerified:     user.EmailVerified,
		PhoneVerified:     user.PhoneVerified,
		IDVerified:        user.IDVerified,
		VerificationLevel: user.VerificationLevel,
		AreaID:            user.AreaID,
		LastKnownLat:      user.LastKnownLat,
		LastKnownLng:      user.LastKnownLng,
		Status:            string(user.Status),
		LastLoginAt:       user.LastLoginAt,
		CreatedAt:         user.CreatedAt,
		UpdatedAt:         user.UpdatedAt,
	}
}

package services

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/paradoxe35/torra/packages/config"
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
	userRepo         ports.UserRepository
	refreshTokenRepo ports.RefreshTokenRepository
	jwtKeyRepo       ports.JWTKeyRepository
	cache            ports.Cache
	logger           logger.Interface
	translator       *i18n.Translator
	jwtIssuer        string
	accessDuration   time.Duration
	refreshDuration  time.Duration
}

func NewAuthService(
	userRepo ports.UserRepository,
	refreshTokenRepo ports.RefreshTokenRepository,
	jwtKeyRepo ports.JWTKeyRepository,
	cache ports.Cache,
	logger logger.Interface,
	translator *i18n.Translator,
	jwtConfig config.JWTConfig,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtKeyRepo:       jwtKeyRepo,
		cache:            cache,
		logger:           logger,
		translator:       translator,
		jwtIssuer:        jwtConfig.Issuer,
		accessDuration:   jwtConfig.AccessDuration,
		refreshDuration:  jwtConfig.RefreshDuration,
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
	accessToken, err := s.generateAccessToken(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Update last login
	now := time.Now()
	user.LastLoginAt = &now
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.WithError(err).Error("Failed to update last login")
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
	// Hash the provided token to find it in database
	tokenHash := security.HashRefreshToken(req.RefreshToken)

	// Find refresh token in database
	refreshTokenEntity, err := s.refreshTokenRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, errors.ErrInvalidToken
	}

	// Check if token is valid
	if !refreshTokenEntity.IsValid() {
		return nil, errors.ErrTokenExpired
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, refreshTokenEntity.UserID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	// Check if user can login
	if !user.CanLogin() {
		return nil, errors.ErrAccountSuspended
	}

	// Generate new access token
	accessToken, err := s.generateAccessToken(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Optionally rotate refresh token (for extra security)
	// For now, we'll keep the same refresh token but update last used time
	refreshTokenEntity.UpdateLastUsed()
	if err := s.refreshTokenRepo.Update(ctx, refreshTokenEntity); err != nil {
		s.logger.WithError(err).Error("Failed to update refresh token last used time")
	}

	return &dtos.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: req.RefreshToken, // Return same refresh token
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

func (s *AuthService) generateAccessToken(ctx context.Context, user *entities.User) (string, error) {
	// Get active JWT key for signing
	activeKey, err := s.jwtKeyRepo.FindActiveKey(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get active signing key: %w", err)
	}

	// Parse private key
	privateKey, err := security.DecodePrivateKeyPEM(activeKey.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode private key: %w", err)
	}

	// Create claims
	claims := jwt.MapClaims{
		"sub":     user.ID.String(),
		"email":   user.Email.String(),
		"name":    user.Name,
		"area_id": user.AreaID.String(),
		"iat":     time.Now().Unix(),
		"exp":     time.Now().Add(s.accessDuration).Unix(),
		"iss":     s.jwtIssuer,
	}

	// Create token with RS256
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// Add key ID to header
	token.Header["kid"] = activeKey.KeyID

	// Sign with private key
	return token.SignedString(privateKey)
}

func (s *AuthService) generateRefreshToken(ctx context.Context, user *entities.User) (string, error) {
	// Generate secure opaque token
	refreshToken, err := security.GenerateRefreshToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Hash token for storage
	tokenHash := security.HashRefreshToken(refreshToken)

	// Create refresh token entity
	refreshTokenEntity := &entities.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(s.refreshDuration),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// TODO: Get device info from request context
	// refreshTokenEntity.DeviceID = ...
	// refreshTokenEntity.IPAddress = ...
	// refreshTokenEntity.UserAgent = ...

	// Store in database
	if err := s.refreshTokenRepo.Create(ctx, refreshTokenEntity); err != nil {
		return "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Return the raw token (not the hash)
	return refreshToken, nil
}

// toUserResponse converts a User entity to UserResponse DTO
// GenerateTokensForUser generates access and refresh tokens for a user (used by OAuth service)
func (s *AuthService) GenerateTokensForUser(ctx context.Context, user *entities.User) (*dtos.AuthResponse, error) {
	// Generate access token
	accessToken, err := s.generateAccessToken(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := s.generateRefreshToken(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &dtos.AuthResponse{
		User:         s.toUserResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(s.accessDuration.Seconds()),
	}, nil
}

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

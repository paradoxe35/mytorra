package services

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/google/uuid"

	// Use shared packages
	appErrors "github.com/paradoxe35/torra/packages/errors"
	"github.com/paradoxe35/torra/packages/i18n"
	"github.com/paradoxe35/torra/packages/logger"
	"github.com/paradoxe35/torra/packages/security"

	"github.com/paradoxe35/torra/services/core/internal/application/dtos"
	"github.com/paradoxe35/torra/services/core/internal/application/ports"
	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
	"github.com/paradoxe35/torra/services/core/internal/domain/events"
	"github.com/paradoxe35/torra/services/core/internal/domain/valueobjects"
)

type UserService struct {
	userRepo       ports.UserRepository
	areaService    *AreaService
	eventPublisher ports.EventPublisher
	cache          ports.Cache
	logger         logger.Interface
	translator     *i18n.Translator
}

func NewUserService(
	userRepo ports.UserRepository,
	areaService *AreaService,
	eventPublisher ports.EventPublisher,
	cache ports.Cache,
	logger logger.Interface,
	translator *i18n.Translator,
) *UserService {
	return &UserService{
		userRepo:       userRepo,
		areaService:    areaService,
		eventPublisher: eventPublisher,
		cache:          cache,
		logger:         logger,
		translator:     translator,
	}
}

func (s *UserService) CreateUser(ctx context.Context, req dtos.RegisterRequest) (*entities.User, error) {
	// Validate request with i18n support
	if err := req.Validate(ctx, s.translator); err != nil {
		return nil, err
	}

	// Create value objects
	email, err := valueobjects.NewEmail(req.Email)
	if err != nil {
		return nil, appErrors.ValidationError(s.translator.TWithContext(ctx, "validation.email"))
	}

	phone, err := valueobjects.NewPhone(req.Phone)
	if err != nil {
		return nil, appErrors.ValidationError(s.translator.TWithContext(ctx, "validation.phone"))
	}

	// Check if user exists
	existing, _ := s.userRepo.FindByEmail(ctx, email)
	if existing != nil {
		return nil, appErrors.ConflictError(s.translator.TWithContext(ctx, "user.already_exists"))
	}

	// Validate area exists and is active
	area, err := s.areaService.GetAreaByID(ctx, req.AreaID)
	if err != nil {
		return nil, appErrors.ValidationError(s.translator.TWithContext(ctx, "area.not_found"))
	}
	if !area.IsActive() {
		return nil, appErrors.ValidationError(s.translator.TWithContext(ctx, "area.not_active"))
	}

	// Hash password using security package
	hashedPassword, err := security.HashPassword(req.Password)
	if err != nil {
		s.logger.WithContext(ctx).WithError(err).Error("Failed to hash password")
		return nil, appErrors.InternalError(s.translator.TWithContext(ctx, "error.internal"))
	}

	// Create user entity
	user := &entities.User{
		ID:                uuid.New(),
		Email:             email,
		Phone:             phone,
		Name:              req.Name,
		PasswordHash:      string(hashedPassword),
		AreaID:            req.AreaID,
		Status:            entities.UserStatusActive,
		EmailVerified:     false,
		PhoneVerified:     false,
		IDVerified:        false,
		VerificationLevel: 0,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Save to repository
	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	// Publish event
	event := events.NewUserCreatedEvent(user.ID, user.Email.String(), user.AreaID.String())
	if err := s.eventPublisher.PublishAsync(ctx, event); err != nil {
		s.logger.Error("failed to publish user created event", "error", err, "user_id", user.ID)
	}

	// Invalidate cache
	s.invalidateUserCache(ctx, user.ID)

	return user, nil
}

func (s *UserService) GetUser(ctx context.Context, userID uuid.UUID) (*entities.User, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("user:%s", userID)
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
		var user entities.User
		if err := json.Unmarshal(cached, &user); err == nil {
			return &user, nil
		}
	}

	// Get from repository
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Update cache
	if data, err := json.Marshal(user); err == nil {
		s.cache.Set(ctx, cacheKey, data, 5*time.Minute)
	}

	return user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, userID uuid.UUID, req dtos.UpdateUserRequest) (*entities.User, error) {
	// Get existing user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Phone != nil {
		phone, err := valueobjects.NewPhone(*req.Phone)
		if err != nil {
			return nil, fmt.Errorf("invalid phone: %w", err)
		}
		user.Phone = phone
	}
	if req.ProfileImageURL != nil {
		user.ProfileImageURL = *req.ProfileImageURL
	}
	if req.AreaID != nil {
		// Validate area exists and is active
		area, err := s.areaService.GetAreaByID(ctx, *req.AreaID)
		if err != nil {
			return nil, fmt.Errorf("invalid area: %w", err)
		}
		if !area.IsActive() {
			return nil, fmt.Errorf("area is not active")
		}
		user.AreaID = *req.AreaID
	}
	if req.LastKnownLat != nil && req.LastKnownLng != nil {
		user.SetLocation(*req.LastKnownLat, *req.LastKnownLng)
	}

	user.UpdatedAt = time.Now()

	// Save changes
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Invalidate cache
	s.invalidateUserCache(ctx, userID)

	return user, nil
}

func (s *UserService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	// Soft delete
	now := time.Now()
	user.DeletedAt = &now
	user.Status = entities.UserStatusDeleted
	user.UpdatedAt = now

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// Publish event
	event := events.NewUserDeletedEvent(userID)
	if err := s.eventPublisher.PublishAsync(ctx, event); err != nil {
		s.logger.Error("failed to publish user deleted event", "error", err, "user_id", userID)
	}

	// Invalidate cache
	s.invalidateUserCache(ctx, userID)

	return nil
}

func (s *UserService) UpdateVerificationStatus(ctx context.Context, userID uuid.UUID, verificationType string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	switch verificationType {
	case "email":
		user.EmailVerified = true
	case "phone":
		user.PhoneVerified = true
	case "id":
		user.IDVerified = true
	default:
		return fmt.Errorf("invalid verification type: %s", verificationType)
	}

	user.UpdateVerificationLevel()
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update verification status: %w", err)
	}

	// Publish event
	event := events.NewUserVerifiedEvent(userID, verificationType, user.VerificationLevel)
	if err := s.eventPublisher.PublishAsync(ctx, event); err != nil {
		s.logger.Error("failed to publish user verified event", "error", err, "user_id", userID)
	}

	// Invalidate cache
	s.invalidateUserCache(ctx, userID)

	return nil
}

func (s *UserService) ListUsersByArea(ctx context.Context, area string, limit, offset int) ([]*entities.User, error) {
	return s.userRepo.ListByArea(ctx, area, limit, offset)
}

func (s *UserService) GetUserCount(ctx context.Context) (int64, error) {
	// Try cache first
	cacheKey := "users:count"
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
		var count int64
		if err := json.Unmarshal(cached, &count); err == nil {
			return count, nil
		}
	}

	// Get from repository
	count, err := s.userRepo.Count(ctx)
	if err != nil {
		return 0, err
	}

	// Update cache
	if data, err := json.Marshal(count); err == nil {
		s.cache.Set(ctx, cacheKey, data, 1*time.Minute)
	}

	return count, nil
}

// EnableMFA enables multi-factor authentication for a user
func (s *UserService) EnableMFA(ctx context.Context, userID uuid.UUID) (*dtos.EnableMFAResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user.MFAEnabled {
		return nil, appErrors.ConflictError(s.translator.TWithContext(ctx, "mfa.already_enabled"))
	}

	// Generate MFA secret using security package
	secret, err := security.GenerateMFASecret()
	if err != nil {
		s.logger.WithContext(ctx).WithError(err).Error("Failed to generate MFA secret")
		return nil, appErrors.InternalError(s.translator.TWithContext(ctx, "error.internal"))
	}

	secretStr := string(secret)
	user.MFASecret = &secretStr
	user.MFAEnabled = true
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to enable MFA: %w", err)
	}

	// Invalidate cache
	s.invalidateUserCache(ctx, userID)

	return &dtos.EnableMFAResponse{
		Secret: string(secret),
		QRCode: fmt.Sprintf("otpauth://totp/Torra:%s?secret=%s&issuer=Torra", user.Email.String(), secret),
	}, nil
}

// DisableMFA disables multi-factor authentication for a user
func (s *UserService) DisableMFA(ctx context.Context, userID uuid.UUID, req dtos.DisableMFARequest) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	if !user.MFAEnabled {
		return appErrors.ConflictError(s.translator.TWithContext(ctx, "mfa.not_enabled"))
	}

	// Verify current password or MFA code
	if req.Password != "" {
		if err := security.CheckPassword(req.Password, user.PasswordHash); err != nil {
			return appErrors.ValidationError(s.translator.TWithContext(ctx, "auth.invalid_password"))
		}
	} else if req.MFACode != "" {
		valid, err := security.VerifyMFACode(req.MFACode, *user.MFASecret)
		if err != nil || !valid {
			return appErrors.ValidationError(s.translator.TWithContext(ctx, "mfa.invalid_code"))
		}
	} else {
		return appErrors.ValidationError(s.translator.TWithContext(ctx, "validation.password_or_mfa_required"))
	}

	user.MFAEnabled = false
	user.MFASecret = nil
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to disable MFA: %w", err)
	}

	// Invalidate cache
	s.invalidateUserCache(ctx, userID)

	return nil
}

// UpdateUserArea updates the user's area
func (s *UserService) UpdateUserArea(ctx context.Context, userID, areaID uuid.UUID) (*entities.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Validate area exists and is active
	area, err := s.areaService.GetAreaByID(ctx, areaID)
	if err != nil {
		return nil, appErrors.ValidationError(s.translator.TWithContext(ctx, "area.not_found"))
	}
	if !area.IsActive() {
		return nil, appErrors.ValidationError(s.translator.TWithContext(ctx, "area.not_active"))
	}

	user.AreaID = areaID
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user area: %w", err)
	}

	// Invalidate cache
	s.invalidateUserCache(ctx, userID)

	return user, nil
}

// GetVerificationStatus returns the user's verification status
func (s *UserService) GetVerificationStatus(ctx context.Context, userID uuid.UUID) (*dtos.VerificationStatusResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &dtos.VerificationStatusResponse{
		EmailVerified:     user.EmailVerified,
		PhoneVerified:     user.PhoneVerified,
		IDVerified:        user.IDVerified,
		VerificationLevel: user.VerificationLevel,
	}, nil
}

// VerifyGovernmentID verifies a user's government ID
func (s *UserService) VerifyGovernmentID(ctx context.Context, userID uuid.UUID, file multipart.File, header *multipart.FileHeader) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.IDVerified {
		return appErrors.ConflictError(s.translator.TWithContext(ctx, "verification.already_verified"))
	}

	// TODO: Implement actual ID verification logic
	// This would typically involve:
	// 1. Upload file to secure storage (S3, etc.)
	// 2. Queue verification job for manual or automated review
	// 3. Store verification status

	// For now, we'll mark as pending verification
	user.IDVerified = false // Will be set to true after manual verification
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to submit ID for verification: %w", err)
	}

	// Publish event for ID verification submission
	event := events.NewIDVerificationSubmittedEvent(userID, header.Filename)
	if err := s.eventPublisher.PublishAsync(ctx, event); err != nil {
		s.logger.Error("failed to publish ID verification submitted event", "error", err, "user_id", userID)
	}

	// Invalidate cache
	s.invalidateUserCache(ctx, userID)

	s.logger.Info("ID verification submitted", "user_id", userID, "filename", header.Filename)
	return nil
}

// Helper methods
func (s *UserService) invalidateUserCache(ctx context.Context, userID uuid.UUID) {
	cacheKey := fmt.Sprintf("user:%s", userID)
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		s.logger.Warn("failed to invalidate user cache", "error", err, "user_id", userID)
	}
}

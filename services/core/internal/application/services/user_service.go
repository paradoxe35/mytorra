package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/paradoxe35/torra/services/core/internal/application/dtos"
	"github.com/paradoxe35/torra/services/core/internal/application/ports"
	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
	"github.com/paradoxe35/torra/services/core/internal/domain/errors"
	"github.com/paradoxe35/torra/services/core/internal/domain/events"
	"github.com/paradoxe35/torra/services/core/internal/domain/valueobjects"
)

type UserService struct {
	userRepo       ports.UserRepository
	areaService    *AreaService
	eventPublisher ports.EventPublisher
	cache          ports.Cache
	logger         ports.Logger
}

func NewUserService(
	userRepo ports.UserRepository,
	areaService *AreaService,
	eventPublisher ports.EventPublisher,
	cache ports.Cache,
	logger ports.Logger,
) *UserService {
	return &UserService{
		userRepo:       userRepo,
		areaService:    areaService,
		eventPublisher: eventPublisher,
		cache:          cache,
		logger:         logger,
	}
}

func (s *UserService) CreateUser(ctx context.Context, req dtos.CreateUserRequest) (*entities.User, error) {
	// Create value objects
	email, err := valueobjects.NewEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	phone, err := valueobjects.NewPhone(req.Phone)
	if err != nil {
		return nil, fmt.Errorf("invalid phone: %w", err)
	}

	// Check if user exists
	existing, _ := s.userRepo.FindByEmail(ctx, email)
	if existing != nil {
		return nil, errors.ErrUserAlreadyExists
	}

	// Validate area exists and is active
	area, err := s.areaService.GetAreaByID(ctx, req.AreaID)
	if err != nil {
		return nil, fmt.Errorf("invalid area: %w", err)
	}
	if !area.IsActive() {
		return nil, fmt.Errorf("area is not active")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
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

// Helper methods
func (s *UserService) invalidateUserCache(ctx context.Context, userID uuid.UUID) {
	cacheKey := fmt.Sprintf("user:%s", userID)
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		s.logger.Warn("failed to invalidate user cache", "error", err, "user_id", userID)
	}
}

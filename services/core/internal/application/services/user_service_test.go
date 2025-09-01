package services_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/paradoxe35/torra/packages/cache"
	"github.com/paradoxe35/torra/packages/i18n"
	"github.com/paradoxe35/torra/packages/logger"
	"github.com/paradoxe35/torra/services/core/internal/adapters/postgres/repositories"
	"github.com/paradoxe35/torra/services/core/internal/application/dtos"
	"github.com/paradoxe35/torra/services/core/internal/application/services"
	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
	"github.com/paradoxe35/torra/services/core/internal/domain/valueobjects"
)

func TestUserService_CreateUser(t *testing.T) {
	// Setup
	mockRepo := repositories.NewMockUserRepository()
	mockAreaRepo := repositories.NewMockAreaRepository()
	mockEventPublisher := &repositories.MockEventPublisher{}
	mockCache := cache.NewMockCache()
	mockLogger := logger.NewMockLogger()

	// Create real logger and translator
	testLogger := logger.New("test", "test")
	translator, err := i18n.New("en")
	require.NoError(t, err)

	// Create AreaService with mocks
	areaService := services.NewAreaService(
		mockAreaRepo,
		mockEventPublisher,
		mockCache,
		mockLogger,
	)

	// Create LocationService
	locationService := services.NewLocationService(
		areaService,
		testLogger,
	)

	service := services.NewUserService(
		mockRepo,
		areaService,
		locationService,
		mockEventPublisher,
		mockCache,
		testLogger,
		translator,
	)

	ctx := context.Background()
	areaID := uuid.New()

	t.Run("successful user creation", func(t *testing.T) {
		req := dtos.RegisterRequest{
			Email:           "test@example.com",
			Phone:           "+250780123456",
			Name:            "Test User",
			Password:        "SecurePass123!@#",
			ConfirmPassword: "SecurePass123!@#",
			AreaID:          &areaID,
			AcceptTerms:     true,
		}

		// Setup expectations
		email, _ := valueobjects.NewEmail(req.Email)

		// Area is active
		activeArea := &entities.Area{
			ID:     areaID,
			Status: valueobjects.AreaStatusActive,
		}

		// Mock area service calls
		mockAreaRepo.On("FindByID", ctx, areaID).Return(activeArea, nil).Once()
		mockCache.On("Get", ctx, fmt.Sprintf("area:%s", areaID)).Return(nil, fmt.Errorf("not found")).Once()
		mockCache.On("Set", ctx, fmt.Sprintf("area:%s", areaID), mock.Anything, mock.Anything).Return(nil).Once()
		mockRepo.On("FindByEmail", ctx, email).Return(nil, nil).Once()
		mockRepo.On("Save", ctx, mock.AnythingOfType("*entities.User")).Return(nil).Once()
		mockEventPublisher.On("PublishAsync", ctx, mock.Anything).Return(nil).Once()
		mockCache.On("Delete", ctx, mock.AnythingOfType("string")).Return(nil).Once()

		// Execute
		user, err := service.CreateUser(ctx, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, req.Email, user.Email.String())
		assert.Equal(t, req.Name, user.Name)
		assert.Equal(t, areaID, user.AreaID)
		assert.Equal(t, entities.UserStatusActive, user.Status)
		assert.False(t, user.EmailVerified)

		mockRepo.AssertExpectations(t)
		mockAreaRepo.AssertExpectations(t)
		mockEventPublisher.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("user already exists", func(t *testing.T) {
		req := dtos.RegisterRequest{
			Email:           "existing@example.com",
			Phone:           "+250780123456",
			Name:            "Existing User",
			Password:        "SecurePass123!@#",
			ConfirmPassword: "SecurePass123!@#",
			AreaID:          &areaID,
			AcceptTerms:     true,
		}

		email, _ := valueobjects.NewEmail(req.Email)
		existingUser := &entities.User{
			ID:    uuid.New(),
			Email: email,
		}

		mockRepo.On("FindByEmail", ctx, email).Return(existingUser, nil).Once()

		// Execute
		user, err := service.CreateUser(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "already exists")

		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid area", func(t *testing.T) {
		req := dtos.RegisterRequest{
			Email:           "test2@example.com",
			Phone:           "+250780123456",
			Name:            "Test User",
			Password:        "SecurePass123!@#",
			ConfirmPassword: "SecurePass123!@#",
			AreaID:          &areaID,
			AcceptTerms:     true,
		}

		email, _ := valueobjects.NewEmail(req.Email)

		mockRepo.On("FindByEmail", ctx, email).Return(nil, nil).Once()
		mockCache.On("Get", ctx, fmt.Sprintf("area:%s", areaID)).Return(nil, fmt.Errorf("not found")).Once()
		mockAreaRepo.On("FindByID", ctx, areaID).Return(nil, assert.AnError).Once()

		// Allow cache set attempts without strict count
		mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(nil)

		// Execute
		user, err := service.CreateUser(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, strings.ToLower(err.Error()), "area")

		mockRepo.AssertExpectations(t)
		mockAreaRepo.AssertExpectations(t)
	})

	t.Run("inactive area", func(t *testing.T) {
		req := dtos.RegisterRequest{
			Email:           "test3@example.com",
			Phone:           "+250780123456",
			Name:            "Test User",
			Password:        "SecurePass123!@#",
			ConfirmPassword: "SecurePass123!@#",
			AreaID:          &areaID,
			AcceptTerms:     true,
		}

		email, _ := valueobjects.NewEmail(req.Email)

		// Area is inactive
		inactiveArea := &entities.Area{
			ID:     areaID,
			Status: valueobjects.AreaStatusInactive,
		}

		mockRepo.On("FindByEmail", ctx, email).Return(nil, nil).Once()
		mockAreaRepo.On("FindByID", ctx, areaID).Return(inactiveArea, nil).Once()
		mockCache.On("Get", ctx, fmt.Sprintf("area:%s", areaID)).Return(nil, fmt.Errorf("not found")).Once()

		// Allow cache set attempts without strict count
		mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(nil)

		// Execute
		user, err := service.CreateUser(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, strings.ToLower(err.Error()), "not active")

		mockRepo.AssertExpectations(t)
		mockAreaRepo.AssertExpectations(t)
	})

	t.Run("validation errors", func(t *testing.T) {
		testCases := []struct {
			name   string
			req    dtos.RegisterRequest
			errMsg string
		}{
			{
				name: "invalid email",
				req: dtos.RegisterRequest{
					Email:           "invalid-email",
					Phone:           "+250780123456",
					Name:            "Test User",
					Password:        "SecurePass123!@#",
					ConfirmPassword: "SecurePass123!@#",
					AreaID:          &areaID,
					AcceptTerms:     true,
				},
				errMsg: "email",
			},
			{
				name: "weak password",
				req: dtos.RegisterRequest{
					Email:           "test@example.com",
					Phone:           "+250780123456",
					Name:            "Test User",
					Password:        "weak",
					ConfirmPassword: "weak",
					AreaID:          &areaID,
					AcceptTerms:     true,
				},
				errMsg: "password",
			},
			{
				name: "password mismatch",
				req: dtos.RegisterRequest{
					Email:           "test@example.com",
					Phone:           "+250780123456",
					Name:            "Test User",
					Password:        "SecurePass123!@#",
					ConfirmPassword: "DifferentPass123!@#",
					AreaID:          &areaID,
					AcceptTerms:     true,
				},
				errMsg: "ConfirmPassword",
			},
			{
				name: "terms not accepted",
				req: dtos.RegisterRequest{
					Email:           "test@example.com",
					Phone:           "+250780123456",
					Name:            "Test User",
					Password:        "SecurePass123!@#",
					ConfirmPassword: "SecurePass123!@#",
					AreaID:          &areaID,
					AcceptTerms:     false,
				},
				errMsg: "terms",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				user, err := service.CreateUser(ctx, tc.req)

				assert.Error(t, err)
				assert.Nil(t, user)
				assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tc.errMsg))
			})
		}
	})
}

func TestUserService_GetUser(t *testing.T) {
	// Setup
	mockRepo := repositories.NewMockUserRepository()
	mockAreaRepo := repositories.NewMockAreaRepository()
	mockEventPublisher := &repositories.MockEventPublisher{}
	mockCache := cache.NewMockCache()
	mockLogger := logger.NewMockLogger()

	testLogger := logger.New("test", "test")
	translator, err := i18n.New("en")
	require.NoError(t, err)

	// Create AreaService with mocks
	areaService := services.NewAreaService(
		mockAreaRepo,
		mockEventPublisher,
		mockCache,
		mockLogger,
	)

	locationService := services.NewLocationService(
		areaService,
		testLogger,
	)

	service := services.NewUserService(
		mockRepo,
		areaService,
		locationService,
		mockEventPublisher,
		mockCache,
		testLogger,
		translator,
	)

	ctx := context.Background()
	userID := uuid.New()

	t.Run("get user from cache", func(t *testing.T) {
		email, _ := valueobjects.NewEmail("test@example.com")
		phone, _ := valueobjects.NewPhone("+250780123456")

		cachedUser := &entities.User{
			ID:    userID,
			Email: email,
			Phone: phone,
			Name:  "Test User",
		}

		cachedData, _ := json.Marshal(cachedUser)
		cacheKey := fmt.Sprintf("user:%s", userID)

		mockCache.On("Get", ctx, cacheKey).Return(cachedData, nil).Once()

		// Execute
		user, err := service.GetUser(ctx, userID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userID, user.ID)

		mockCache.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "FindByID")
	})

	t.Run("get user from repository when cache miss", func(t *testing.T) {
		email, _ := valueobjects.NewEmail("test@example.com")
		phone, _ := valueobjects.NewPhone("+250780123456")

		repoUser := &entities.User{
			ID:    userID,
			Email: email,
			Phone: phone,
			Name:  "Test User",
		}

		cacheKey := fmt.Sprintf("user:%s", userID)

		mockCache.On("Get", ctx, cacheKey).Return(nil, assert.AnError).Once()
		mockRepo.On("FindByID", ctx, userID).Return(repoUser, nil).Once()
		mockCache.On("Set", ctx, cacheKey, mock.Anything, 5*time.Minute).Return(nil).Once()

		// Execute
		user, err := service.GetUser(ctx, userID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userID, user.ID)

		mockCache.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})
}

func BenchmarkUserService_CreateUser(b *testing.B) {
	// Setup
	mockRepo := repositories.NewMockUserRepository()
	mockAreaRepo := repositories.NewMockAreaRepository()
	mockEventPublisher := &repositories.MockEventPublisher{}
	mockCache := cache.NewMockCache()
	mockLogger := logger.NewMockLogger()

	testLogger := logger.New("test", "test")
	translator, _ := i18n.New("en")

	// Create AreaService with mocks
	areaService := services.NewAreaService(
		mockAreaRepo,
		mockEventPublisher,
		mockCache,
		mockLogger,
	)

	locationService := services.NewLocationService(
		areaService,
		testLogger,
	)

	service := services.NewUserService(
		mockRepo,
		areaService,
		locationService,
		mockEventPublisher,
		mockCache,
		testLogger,
		translator,
	)

	ctx := context.Background()
	areaID := uuid.New()

	activeArea := &entities.Area{
		ID:     areaID,
		Status: valueobjects.AreaStatusActive,
	}

	// Setup mock expectations for all iterations
	mockAreaRepo.On("FindByID", ctx, areaID).Return(activeArea, nil)
	mockCache.On("Get", ctx, fmt.Sprintf("area:%s", areaID)).Return(nil, fmt.Errorf("not found"))
	mockCache.On("Set", ctx, fmt.Sprintf("area:%s", areaID), mock.Anything, mock.Anything).Return(nil)
	mockRepo.On("FindByEmail", ctx, mock.Anything).Return(nil, nil)
	mockRepo.On("Save", ctx, mock.Anything).Return(nil)
	mockEventPublisher.On("PublishAsync", ctx, mock.Anything).Return(nil)
	mockCache.On("Delete", ctx, mock.Anything).Return(nil)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := dtos.RegisterRequest{
			Email:           fmt.Sprintf("test%d@example.com", i),
			Phone:           "+250780123456",
			Name:            "Test User",
			Password:        "SecurePass123!@#",
			ConfirmPassword: "SecurePass123!@#",
			AreaID:          &areaID,
			AcceptTerms:     true,
		}

		_, _ = service.CreateUser(ctx, req)
	}
}

func TestUserService_CreateUser_LocationDetection(t *testing.T) {
	// Setup
	mockRepo := repositories.NewMockUserRepository()
	mockAreaRepo := repositories.NewMockAreaRepository()
	mockEventPublisher := &repositories.MockEventPublisher{}
	mockCache := cache.NewMockCache()
	mockLogger := logger.NewMockLogger()

	// Create real logger and translator
	testLogger := logger.New("test", "test")
	translator, err := i18n.New("en")
	require.NoError(t, err)

	// Create AreaService with mocks
	areaService := services.NewAreaService(
		mockAreaRepo,
		mockEventPublisher,
		mockCache,
		mockLogger,
	)

	locationService := services.NewLocationService(
		areaService,
		testLogger,
	)

	service := services.NewUserService(
		mockRepo,
		areaService,
		locationService,
		mockEventPublisher,
		mockCache,
		testLogger,
		translator,
	)

	ctx := context.WithValue(context.Background(), services.ClientIPKey, "127.0.0.1") // Use localhost to avoid external API calls
	areaID := uuid.New()

	// Create a proper area entity with ContainsCoordinate method working
	activeArea := &entities.Area{
		ID:                areaID,
		Status:            valueobjects.AreaStatusActive,
		Type:              valueobjects.AreaTypeTown,
		Name:              "Test Area",
		CenterLat:         -2.5085,
		CenterLng:         28.8473,
		OperationalRadius: 50, // 50km radius
	}

	t.Run("successful user creation with explicit area ID", func(t *testing.T) {
		req := dtos.RegisterRequest{
			Email:           "test@example.com",
			Phone:           "+250780123456",
			Name:            "Test User",
			Password:        "SecurePass123!@#",
			ConfirmPassword: "SecurePass123!@#",
			AcceptTerms:     true,
			AreaID:          &areaID, // Explicitly provide area ID
		}

		email, _ := valueobjects.NewEmail(req.Email)

		// Setup expectations
		mockRepo.On("FindByEmail", ctx, email).Return(nil, fmt.Errorf("not found")).Once()
		// Mock area lookup by ID
		mockAreaRepo.On("FindByID", ctx, areaID).Return(activeArea, nil).Once()
		mockCache.On("Get", ctx, fmt.Sprintf("area:%s", areaID)).Return(nil, fmt.Errorf("not found")).Once()
		mockCache.On("Set", ctx, fmt.Sprintf("area:%s", areaID), mock.Anything, mock.Anything).Return(nil).Once()
		mockRepo.On("Save", ctx, mock.AnythingOfType("*entities.User")).Return(nil).Once()
		mockEventPublisher.On("PublishAsync", ctx, mock.Anything).Return(nil).Once()
		mockCache.On("Delete", ctx, mock.AnythingOfType("string")).Return(nil).Once()

		// Execute
		user, err := service.CreateUser(ctx, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, req.Email, user.Email.String())
		assert.Equal(t, req.Name, user.Name)
		assert.Equal(t, areaID, user.AreaID)

		mockRepo.AssertExpectations(t)
		mockAreaRepo.AssertExpectations(t)
		mockEventPublisher.AssertExpectations(t)
	})

	t.Run("area detection fails - no location data", func(t *testing.T) {
		ctxNoIP := context.Background() // No client IP in context

		req := dtos.RegisterRequest{
			Email:           "test3@example.com",
			Phone:           "+250780123458",
			Name:            "Test User 3",
			Password:        "SecurePass123!@#",
			ConfirmPassword: "SecurePass123!@#",
			AcceptTerms:     true,
			// No GPS coordinates, no area ID, and no IP in context
		}

		email, _ := valueobjects.NewEmail(req.Email)

		// Setup expectations
		mockRepo.On("FindByEmail", ctxNoIP, email).Return(nil, fmt.Errorf("not found")).Once()

		// Execute
		user, err := service.CreateUser(ctxNoIP, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "Unable to detect area from provided location data")

		mockRepo.AssertExpectations(t)
	})
}

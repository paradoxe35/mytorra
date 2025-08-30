package services_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/paradoxe35/torra/packages/i18n"
	"github.com/paradoxe35/torra/packages/logger"
	"github.com/paradoxe35/torra/services/core/internal/application/dtos"
	"github.com/paradoxe35/torra/services/core/internal/application/services"
	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
	"github.com/paradoxe35/torra/services/core/internal/domain/valueobjects"
)

// Mock implementations
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email valueobjects.Email) (*entities.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) Save(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Update(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) FindByPhone(ctx context.Context, phone valueobjects.Phone) (*entities.User, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) ListByArea(ctx context.Context, area string, limit, offset int) ([]*entities.User, error) {
	args := m.Called(ctx, area, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.User), args.Error(1)
}

func (m *MockUserRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

type MockEventPublisher struct {
	mock.Mock
}

func (m *MockEventPublisher) Publish(ctx context.Context, event interface{}) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishAsync(ctx context.Context, event interface{}) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishBatch(ctx context.Context, events []interface{}) error {
	args := m.Called(ctx, events)
	return args.Error(0)
}

type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(ctx context.Context, key string) ([]byte, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockCache) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCache) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	args := m.Called(ctx, key, ttl)
	return args.Error(0)
}

func (m *MockCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(time.Duration), args.Error(1)
}

func (m *MockCache) HGet(ctx context.Context, key, field string) ([]byte, error) {
	args := m.Called(ctx, key, field)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockCache) HSet(ctx context.Context, key, field string, value []byte) error {
	args := m.Called(ctx, key, field, value)
	return args.Error(0)
}

func (m *MockCache) HGetAll(ctx context.Context, key string) (map[string][]byte, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string][]byte), args.Error(1)
}

func (m *MockCache) HDel(ctx context.Context, key string, fields ...string) error {
	args := m.Called(ctx, key, fields)
	return args.Error(0)
}

func (m *MockCache) LPush(ctx context.Context, key string, values ...[]byte) error {
	args := m.Called(ctx, key, values)
	return args.Error(0)
}

func (m *MockCache) RPop(ctx context.Context, key string) ([]byte, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockCache) LRange(ctx context.Context, key string, start, stop int64) ([][]byte, error) {
	args := m.Called(ctx, key, start, stop)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([][]byte), args.Error(1)
}

func (m *MockCache) SAdd(ctx context.Context, key string, members ...[]byte) error {
	args := m.Called(ctx, key, members)
	return args.Error(0)
}

func (m *MockCache) SMembers(ctx context.Context, key string) ([][]byte, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([][]byte), args.Error(1)
}

func (m *MockCache) SRem(ctx context.Context, key string, members ...[]byte) error {
	args := m.Called(ctx, key, members)
	return args.Error(0)
}

func (m *MockCache) SIsMember(ctx context.Context, key string, member []byte) (bool, error) {
	args := m.Called(ctx, key, member)
	return args.Get(0).(bool), args.Error(1)
}

type MockAreaRepository struct {
	mock.Mock
}

func (m *MockAreaRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Area, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Area), args.Error(1)
}

func (m *MockAreaRepository) FindByCode(ctx context.Context, code string) (*entities.Area, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Area), args.Error(1)
}

func (m *MockAreaRepository) FindByParentID(ctx context.Context, parentID *uuid.UUID) ([]*entities.Area, error) {
	args := m.Called(ctx, parentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Area), args.Error(1)
}

func (m *MockAreaRepository) FindByType(ctx context.Context, areaType valueobjects.AreaType) ([]*entities.Area, error) {
	args := m.Called(ctx, areaType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Area), args.Error(1)
}

func (m *MockAreaRepository) FindActive(ctx context.Context) ([]*entities.Area, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Area), args.Error(1)
}

func (m *MockAreaRepository) FindByCoordinate(ctx context.Context, lat, lng float64) ([]*entities.Area, error) {
	args := m.Called(ctx, lat, lng)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Area), args.Error(1)
}

func (m *MockAreaRepository) Save(ctx context.Context, area *entities.Area) error {
	args := m.Called(ctx, area)
	return args.Error(0)
}

func (m *MockAreaRepository) Update(ctx context.Context, area *entities.Area) error {
	args := m.Called(ctx, area)
	return args.Error(0)
}

func (m *MockAreaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAreaRepository) GetHierarchy(ctx context.Context, areaID uuid.UUID) ([]*entities.Area, error) {
	args := m.Called(ctx, areaID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Area), args.Error(1)
}

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string, fields ...logger.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Info(msg string, fields ...logger.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Warn(msg string, fields ...logger.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Error(msg string, fields ...logger.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Fatal(msg string, fields ...logger.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) With(fields ...logger.Field) logger.Logger {
	args := m.Called(fields)
	return args.Get(0).(logger.Logger)
}

func TestUserService_CreateUser(t *testing.T) {
	// Setup
	mockRepo := new(MockUserRepository)
	mockAreaRepo := new(MockAreaRepository)
	mockEventPublisher := new(MockEventPublisher)
	mockCache := new(MockCache)
	mockLogger := new(MockLogger)

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

	service := services.NewUserService(
		mockRepo,
		areaService,
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
			AreaID:          areaID,
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
		mockEventPublisher.On("Publish", ctx, mock.Anything).Return(nil).Once()
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
		mockAreaService.AssertExpectations(t)
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
			AreaID:          areaID,
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
			AreaID:          areaID,
			AcceptTerms:     true,
		}

		email, _ := valueobjects.NewEmail(req.Email)

		mockRepo.On("FindByEmail", ctx, email).Return(nil, nil).Once()
		mockAreaService.On("GetAreaByID", ctx, areaID).Return(nil, assert.AnError).Once()

		// Execute
		user, err := service.CreateUser(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "area")

		mockRepo.AssertExpectations(t)
		mockAreaService.AssertExpectations(t)
	})

	t.Run("inactive area", func(t *testing.T) {
		req := dtos.RegisterRequest{
			Email:           "test3@example.com",
			Phone:           "+250780123456",
			Name:            "Test User",
			Password:        "SecurePass123!@#",
			ConfirmPassword: "SecurePass123!@#",
			AreaID:          areaID,
			AcceptTerms:     true,
		}

		email, _ := valueobjects.NewEmail(req.Email)

		// Area is inactive
		inactiveArea := &entities.Area{
			ID:     areaID,
			Status: valueobjects.AreaStatusInactive,
		}

		mockRepo.On("FindByEmail", ctx, email).Return(nil, nil).Once()
		mockAreaService.On("GetAreaByID", ctx, areaID).Return(inactiveArea, nil).Once()

		// Execute
		user, err := service.CreateUser(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "not active")

		mockRepo.AssertExpectations(t)
		mockAreaService.AssertExpectations(t)
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
					AreaID:          areaID,
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
					AreaID:          areaID,
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
					AreaID:          areaID,
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
					AreaID:          areaID,
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
				assert.Contains(t, err.Error(), tc.errMsg)
			})
		}
	})
}

func TestUserService_GetUser(t *testing.T) {
	// Setup
	mockRepo := new(MockUserRepository)
	mockAreaService := new(MockAreaService)
	mockEventPublisher := new(MockEventPublisher)
	mockCache := new(MockCache)

	testLogger := logger.New("test", "test")
	translator, err := i18n.New("en")
	require.NoError(t, err)

	service := services.NewUserService(
		mockRepo,
		mockAreaService,
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
	mockRepo := new(MockUserRepository)
	mockAreaService := new(MockAreaService)
	mockEventPublisher := new(MockEventPublisher)
	mockCache := new(MockCache)

	testLogger := logger.New("test", "test")
	translator, _ := i18n.New("en")

	service := services.NewUserService(
		mockRepo,
		mockAreaService,
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
	mockAreaService.On("GetAreaByID", ctx, areaID).Return(activeArea, nil)
	mockRepo.On("FindByEmail", ctx, mock.Anything).Return(nil, nil)
	mockRepo.On("Save", ctx, mock.Anything).Return(nil)
	mockEventPublisher.On("Publish", ctx, mock.Anything).Return(nil)
	mockCache.On("Delete", ctx, mock.Anything).Return(nil)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := dtos.RegisterRequest{
			Email:           fmt.Sprintf("test%d@example.com", i),
			Phone:           "+250780123456",
			Name:            "Test User",
			Password:        "SecurePass123!@#",
			ConfirmPassword: "SecurePass123!@#",
			AreaID:          areaID,
			AcceptTerms:     true,
		}

		_, _ = service.CreateUser(ctx, req)
	}
}

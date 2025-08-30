package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/paradoxe35/torra/packages/logger"
	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
	"github.com/paradoxe35/torra/services/core/internal/domain/valueobjects"
)

// MockUserRepository is a mock implementation for testing
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

// MockAreaRepository is a mock implementation for testing
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

// MockEventPublisher is a mock implementation for testing
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

// MockCache is a mock implementation for testing
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

// MockLogger is a mock implementation for testing
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) Info(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) Warn(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) Error(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) Fatal(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) With(args ...interface{}) logger.Logger {
	mockArgs := m.Called(args)
	return mockArgs.Get(0).(logger.Logger)
}

func (m *MockLogger) WithContext(ctx context.Context) logger.Logger {
	args := m.Called(ctx)
	return args.Get(0).(logger.Logger)
}

func (m *MockLogger) WithError(err error) logger.Logger {
	args := m.Called(err)
	return args.Get(0).(logger.Logger)
}

func (m *MockLogger) WithField(key string, value interface{}) logger.Logger {
	args := m.Called(key, value)
	return args.Get(0).(logger.Logger)
}

func (m *MockLogger) WithFields(fields map[string]interface{}) logger.Logger {
	args := m.Called(fields)
	return args.Get(0).(logger.Logger)
}

package cache

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockCache is a mock implementation of the Cache interface for testing
type MockCache struct {
	mock.Mock
}

// NewMockCache creates a new MockCache instance
func NewMockCache() *MockCache {
	return &MockCache{}
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

// MockBasicCache is a mock implementation of the BasicCache interface
type MockBasicCache struct {
	mock.Mock
}

func (m *MockBasicCache) Get(ctx context.Context, key string) ([]byte, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockBasicCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockBasicCache) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockBasicCache) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockBasicCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(time.Duration), args.Error(1)
}

func (m *MockBasicCache) Flush(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

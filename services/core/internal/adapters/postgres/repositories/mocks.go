package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/paradoxe35/mytorra/packages/cache"
	"github.com/paradoxe35/mytorra/packages/events"
	"github.com/paradoxe35/mytorra/services/core/internal/application/ports"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/entities"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/valueobjects"
)

// MockUserRepository implements ports.UserRepository for tests
type MockUserRepository struct{ mock.Mock }

func NewMockUserRepository() *MockUserRepository { return &MockUserRepository{} }

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

func (m *MockUserRepository) FindByPhone(ctx context.Context, phone valueobjects.Phone) (*entities.User, error) {
	args := m.Called(ctx, phone)
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

// MockAreaRepository implements ports.AreaRepository for tests
type MockAreaRepository struct{ mock.Mock }

func NewMockAreaRepository() *MockAreaRepository { return &MockAreaRepository{} }

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

// NewMockCache returns the shared cache.MockCache for tests to keep legacy imports working
func NewMockCache() *cache.MockCache { return &cache.MockCache{} }

// MockEventPublisher is an alias to the shared events.MockEventPublisher to keep tests compatible
type MockEventPublisher = events.MockEventPublisher

// Ensure our mocks satisfy the interfaces
var (
	_ ports.UserRepository = (*MockUserRepository)(nil)
	_ ports.AreaRepository = (*MockAreaRepository)(nil)
)

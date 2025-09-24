package mocks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/paradoxe35/mytorra/services/core/internal/domain/entities"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/valueobjects"
)

// JWTKeyRepository mock
type JWTKeyRepository struct {
	mock.Mock
}

func (m *JWTKeyRepository) Create(ctx context.Context, key *entities.JWTKey) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *JWTKeyRepository) FindByKeyID(ctx context.Context, keyID string) (*entities.JWTKey, error) {
	args := m.Called(ctx, keyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.JWTKey), args.Error(1)
}

func (m *JWTKeyRepository) FindActiveKey(ctx context.Context) (*entities.JWTKey, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.JWTKey), args.Error(1)
}

func (m *JWTKeyRepository) FindAllActive(ctx context.Context) ([]*entities.JWTKey, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.JWTKey), args.Error(1)
}

func (m *JWTKeyRepository) FindAll(ctx context.Context) ([]*entities.JWTKey, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.JWTKey), args.Error(1)
}

func (m *JWTKeyRepository) Update(ctx context.Context, key *entities.JWTKey) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *JWTKeyRepository) RotateKey(ctx context.Context, newKey *entities.JWTKey) error {
	args := m.Called(ctx, newKey)
	return args.Error(0)
}

func (m *JWTKeyRepository) DeactivateExpired(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *JWTKeyRepository) DeleteOldKeys(ctx context.Context, olderThan time.Duration) (int64, error) {
	args := m.Called(ctx, olderThan)
	return args.Get(0).(int64), args.Error(1)
}

// RefreshTokenRepository mock
type RefreshTokenRepository struct {
	mock.Mock
}

func (m *RefreshTokenRepository) Create(ctx context.Context, token *entities.RefreshToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *RefreshTokenRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*entities.RefreshToken, error) {
	args := m.Called(ctx, tokenHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.RefreshToken), args.Error(1)
}

func (m *RefreshTokenRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.RefreshToken, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.RefreshToken), args.Error(1)
}

func (m *RefreshTokenRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.RefreshToken, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.RefreshToken), args.Error(1)
}

func (m *RefreshTokenRepository) FindActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.RefreshToken, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.RefreshToken), args.Error(1)
}

func (m *RefreshTokenRepository) Update(ctx context.Context, token *entities.RefreshToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *RefreshTokenRepository) RevokeByID(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *RefreshTokenRepository) RevokeByDeviceID(ctx context.Context, userID uuid.UUID, deviceID string) error {
	args := m.Called(ctx, userID, deviceID)
	return args.Error(0)
}

func (m *RefreshTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *RefreshTokenRepository) CountActiveByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

// AreaRepository mock
type AreaRepository struct {
	mock.Mock
}

func (m *AreaRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Area, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Area), args.Error(1)
}

func (m *AreaRepository) FindByCode(ctx context.Context, code string) (*entities.Area, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Area), args.Error(1)
}

func (m *AreaRepository) FindByParentID(ctx context.Context, parentID *uuid.UUID) ([]*entities.Area, error) {
	args := m.Called(ctx, parentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Area), args.Error(1)
}

func (m *AreaRepository) FindByType(ctx context.Context, areaType valueobjects.AreaType) ([]*entities.Area, error) {
	args := m.Called(ctx, areaType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Area), args.Error(1)
}

func (m *AreaRepository) FindActive(ctx context.Context) ([]*entities.Area, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Area), args.Error(1)
}

func (m *AreaRepository) FindByCoordinate(ctx context.Context, lat, lng float64) ([]*entities.Area, error) {
	args := m.Called(ctx, lat, lng)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Area), args.Error(1)
}

func (m *AreaRepository) Save(ctx context.Context, area *entities.Area) error {
	args := m.Called(ctx, area)
	return args.Error(0)
}

func (m *AreaRepository) Update(ctx context.Context, area *entities.Area) error {
	args := m.Called(ctx, area)
	return args.Error(0)
}

func (m *AreaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *AreaRepository) GetHierarchy(ctx context.Context, areaID uuid.UUID) ([]*entities.Area, error) {
	args := m.Called(ctx, areaID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Area), args.Error(1)
}

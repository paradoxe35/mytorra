package mocks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/paradoxe35/torra/services/core/internal/application/ports"
	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
	"github.com/paradoxe35/torra/services/core/internal/domain/valueobjects"
)

// OAuthRepository mock
type OAuthRepository struct {
	mock.Mock
}

func (m *OAuthRepository) CreateOAuthProfile(ctx context.Context, profile *entities.OAuthProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *OAuthRepository) GetOAuthProfile(ctx context.Context, provider valueobjects.OAuthProvider, providerID string) (*entities.OAuthProfile, error) {
	args := m.Called(ctx, provider, providerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.OAuthProfile), args.Error(1)
}

func (m *OAuthRepository) GetOAuthProfileByUserAndProvider(ctx context.Context, userID uuid.UUID, provider valueobjects.OAuthProvider) (*entities.OAuthProfile, error) {
	args := m.Called(ctx, userID, provider)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.OAuthProfile), args.Error(1)
}

func (m *OAuthRepository) GetUserOAuthProfiles(ctx context.Context, userID uuid.UUID) ([]*entities.OAuthProfile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.OAuthProfile), args.Error(1)
}

func (m *OAuthRepository) UpdateOAuthTokens(ctx context.Context, profileID uuid.UUID, accessToken, refreshToken string, expiry *time.Time) error {
	args := m.Called(ctx, profileID, accessToken, refreshToken, expiry)
	return args.Error(0)
}

func (m *OAuthRepository) UpdateOAuthProfile(ctx context.Context, profile *entities.OAuthProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *OAuthRepository) DeleteOAuthProfile(ctx context.Context, userID uuid.UUID, provider valueobjects.OAuthProvider) error {
	args := m.Called(ctx, userID, provider)
	return args.Error(0)
}

func (m *OAuthRepository) SaveOAuthState(ctx context.Context, state *valueobjects.OAuthState) error {
	args := m.Called(ctx, state)
	return args.Error(0)
}

func (m *OAuthRepository) GetOAuthState(ctx context.Context, state string) (*valueobjects.OAuthState, error) {
	args := m.Called(ctx, state)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*valueobjects.OAuthState), args.Error(1)
}

func (m *OAuthRepository) DeleteOAuthState(ctx context.Context, state string) error {
	args := m.Called(ctx, state)
	return args.Error(0)
}

func (m *OAuthRepository) CleanupExpiredStates(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// OAuthProvider mock
type OAuthProvider struct {
	mock.Mock
}

func (m *OAuthProvider) GetAuthorizationURL(state, codeVerifier, redirectURI string) string {
	args := m.Called(state, codeVerifier, redirectURI)
	return args.String(0)
}

func (m *OAuthProvider) ExchangeCode(ctx context.Context, code, codeVerifier, redirectURI string) (*ports.OAuthTokens, error) {
	args := m.Called(ctx, code, codeVerifier, redirectURI)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ports.OAuthTokens), args.Error(1)
}

func (m *OAuthProvider) GetUserProfile(ctx context.Context, accessToken string) (*ports.OAuthUserProfile, error) {
	args := m.Called(ctx, accessToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ports.OAuthUserProfile), args.Error(1)
}

func (m *OAuthProvider) RefreshToken(ctx context.Context, refreshToken string) (*ports.OAuthTokens, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ports.OAuthTokens), args.Error(1)
}

func (m *OAuthProvider) RevokeToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

// UserRepository mock (partial, add methods as needed)
type UserRepository struct {
	mock.Mock
}

func (m *UserRepository) FindByEmail(ctx context.Context, email valueobjects.Email) (*entities.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *UserRepository) FindByPhone(ctx context.Context, phone valueobjects.Phone) (*entities.User, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *UserRepository) Create(ctx context.Context, user *entities.User) (*entities.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *UserRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *UserRepository) ListByArea(ctx context.Context, area string, limit, offset int) ([]*entities.User, error) {
	args := m.Called(ctx, area, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.User), args.Error(1)
}

func (m *UserRepository) Save(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *UserRepository) Update(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// EventBus mock
type EventBus struct {
	mock.Mock
}

func (m *EventBus) Publish(ctx context.Context, event interface{}) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *EventBus) PublishAsync(ctx context.Context, event interface{}) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *EventBus) PublishBatch(ctx context.Context, events []interface{}) error {
	args := m.Called(ctx, events)
	return args.Error(0)
}

func (m *EventBus) Subscribe(ctx context.Context, eventType string, handler func(event interface{})) error {
	args := m.Called(ctx, eventType, handler)
	return args.Error(0)
}

func (m *EventBus) Close() error {
	args := m.Called()
	return args.Error(0)
}

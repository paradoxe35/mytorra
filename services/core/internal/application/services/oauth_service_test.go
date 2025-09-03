package services

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"encoding/pem"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/paradoxe35/mytorra/packages/cache"
	"github.com/paradoxe35/mytorra/packages/config"
	"github.com/paradoxe35/mytorra/packages/i18n"
	"github.com/paradoxe35/mytorra/packages/logger"
	"github.com/paradoxe35/mytorra/packages/security"
	"github.com/paradoxe35/mytorra/services/core/internal/application/dtos"
	"github.com/paradoxe35/mytorra/services/core/internal/application/ports"
	"github.com/paradoxe35/mytorra/services/core/internal/application/ports/mocks"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/entities"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/valueobjects"
)

// Helper function to generate test RSA keys
func generateTestRSAKeys() (*rsa.PrivateKey, string, string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, "", "", err
	}

	// Convert private key to PEM format
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	privateKeyStr := string(pem.EncodeToMemory(privateKeyPEM))

	// Convert public key to PEM format
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, "", "", err
	}
	publicKeyPEM := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	publicKeyStr := string(pem.EncodeToMemory(publicKeyPEM))

	return privateKey, privateKeyStr, publicKeyStr, nil
}

// Helper function to generate a test JWT key entity
func generateTestJWTKey() *entities.JWTKey {
	_, privateKeyPEM, publicKeyPEM, _ := generateTestRSAKeys()
	keyID := base64.RawURLEncoding.EncodeToString([]byte(uuid.New().String()))[:22]

	return &entities.JWTKey{
		ID:         uuid.New(),
		KeyID:      keyID,
		PrivateKey: privateKeyPEM,
		PublicKey:  publicKeyPEM,
		Algorithm:  "RS256",
		IsActive:   true,
		ExpiresAt:  time.Now().Add(30 * 24 * time.Hour),
		CreatedAt:  time.Now(),
	}
}

// OAuthServiceTestSetup contains all the dependencies and mocks for OAuth service testing
type OAuthServiceTestSetup struct {
	Service          *OAuthService
	OAuthRepo        *mocks.OAuthRepository
	UserRepo         *mocks.UserRepository
	AreaRepo         *mocks.AreaRepository
	EventBus         *mocks.EventBus
	OAuthProvider    *mocks.OAuthProvider
	JWTKeyRepo       *mocks.JWTKeyRepository
	RefreshTokenRepo *mocks.RefreshTokenRepository
	Cache            *cache.MockCache
	AuthService      *AuthService
	UserService      *UserService
	AreaService      *AreaService
	LocationService  *LocationService
}

func setupOAuthServiceTest(t *testing.T) *OAuthServiceTestSetup {
	oauthRepo := new(mocks.OAuthRepository)
	userRepo := new(mocks.UserRepository)
	areaRepo := new(mocks.AreaRepository)
	eventBus := new(mocks.EventBus)
	oauthProvider := new(mocks.OAuthProvider)
	jwtKeyRepo := new(mocks.JWTKeyRepository)
	refreshTokenRepo := new(mocks.RefreshTokenRepository)
	cacheMock := cache.NewMockCache()

	translator, _ := i18n.New("en", "en", "fr")
	logger := logger.New("test", "test")
	encryptor := security.NewAESEncryptor("test-secret-key-32-chars-long!!!!!")

	authService := &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtKeyRepo:       jwtKeyRepo,
		cache:            cacheMock,
		logger:           logger,
		translator:       translator,
		accessDuration:   15 * time.Minute,
		refreshDuration:  30 * 24 * time.Hour,
		jwtIssuer:        "test-issuer",
	}

	// Create AreaService
	areaService := &AreaService{
		areaRepo:       areaRepo,
		eventPublisher: eventBus,
		cache:          cacheMock,
		logger:         logger,
	}

	// Create LocationService
	locationService := &LocationService{
		areaService: areaService,
		logger:      logger,
	}

	userService := &UserService{
		userRepo:        userRepo,
		areaService:     areaService,
		locationService: locationService,
		eventPublisher:  eventBus,
		cache:           cacheMock,
		logger:          logger,
		translator:      translator,
	}

	oauthConfig := config.OAuthConfig{
		StateExpiry: 10 * time.Minute,
		PKCEEnabled: true,
		Google: config.OAuthProviderConfig{
			ClientID:     "test-google-client",
			ClientSecret: "test-google-secret",
			RedirectURI:  "http://localhost/callback",
		},
	}

	service := &OAuthService{
		oauthRepo:   oauthRepo,
		userRepo:    userRepo,
		authService: authService,
		userService: userService,
		providers: map[valueobjects.OAuthProvider]ports.OAuthProvider{
			valueobjects.OAuthProviderGoogle: oauthProvider,
		},
		eventBus:    eventBus,
		logger:      logger,
		translator:  translator,
		encryptor:   encryptor,
		stateExpiry: oauthConfig.StateExpiry,
		pkceEnabled: oauthConfig.PKCEEnabled,
	}

	return &OAuthServiceTestSetup{
		Service:          service,
		OAuthRepo:        oauthRepo,
		UserRepo:         userRepo,
		AreaRepo:         areaRepo,
		EventBus:         eventBus,
		OAuthProvider:    oauthProvider,
		JWTKeyRepo:       jwtKeyRepo,
		RefreshTokenRepo: refreshTokenRepo,
		Cache:            cacheMock,
		AuthService:      authService,
		UserService:      userService,
		AreaService:      areaService,
		LocationService:  locationService,
	}
}

func TestOAuthService_InitiateOAuth_Success(t *testing.T) {
	setup := setupOAuthServiceTest(t)

	ctx := context.Background()
	req := dtos.OAuthInitRequest{
		Provider:    "google",
		RedirectURI: "http://localhost/callback",
	}

	// Mock save state
	setup.OAuthRepo.On("SaveOAuthState", mock.Anything, mock.Anything).Return(nil)

	// Mock provider authorization URL
	setup.OAuthProvider.On("GetAuthorizationURL", mock.Anything, mock.Anything, req.RedirectURI).
		Return("https://accounts.google.com/oauth/authorize?state=test")

	response, err := setup.Service.InitiateOAuth(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.AuthorizationURL)
	assert.NotEmpty(t, response.State)
	assert.Contains(t, response.AuthorizationURL, "google.com")

	setup.OAuthRepo.AssertExpectations(t)
}

func TestOAuthService_InitiateOAuth_InvalidProvider(t *testing.T) {
	setup := setupOAuthServiceTest(t)

	ctx := context.Background()
	req := dtos.OAuthInitRequest{
		Provider:    "invalid",
		RedirectURI: "http://localhost/callback",
	}

	response, err := setup.Service.InitiateOAuth(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "OAuth provider not found")
}

func TestOAuthService_HandleCallback_NewUser(t *testing.T) {
	setup := setupOAuthServiceTest(t)

	ctx := context.Background()
	req := dtos.OAuthCallbackRequest{
		Code:     "test-code",
		State:    "test-state",
		Provider: "google",
	}

	// Create test OAuth state
	oauthState := &valueobjects.OAuthState{
		State:        req.State,
		CodeVerifier: "test-verifier",
		RedirectURI:  "http://localhost/callback",
		Provider:     valueobjects.OAuthProviderGoogle,
		ExpiresAt:    time.Now().Add(10 * time.Minute),
	}

	// Mock get state
	setup.OAuthRepo.On("GetOAuthState", ctx, req.State).Return(oauthState, nil)
	setup.OAuthRepo.On("DeleteOAuthState", ctx, req.State).Return(nil)

	// Mock token exchange
	tokens := &ports.OAuthTokens{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		ExpiresIn:    3600,
		TokenType:    "Bearer",
		Scopes:       []string{"email", "profile"},
	}
	setup.OAuthProvider.On("ExchangeCode", ctx, req.Code, oauthState.CodeVerifier, oauthState.RedirectURI).
		Return(tokens, nil)

	// Mock user profile
	profile := &ports.OAuthUserProfile{
		ID:       "google-user-123",
		Email:    "test@example.com",
		Name:     "Test User",
		Picture:  "https://example.com/picture.jpg",
		Verified: true,
		Raw:      map[string]interface{}{"id": "google-user-123"},
	}
	setup.OAuthProvider.On("GetUserProfile", ctx, tokens.AccessToken).Return(profile, nil)

	// Mock no existing OAuth profile
	setup.OAuthRepo.On("GetOAuthProfile", ctx, valueobjects.OAuthProviderGoogle, profile.ID).
		Return(nil, sql.ErrNoRows)

	// Mock no existing user with email
	email, _ := valueobjects.NewEmail(profile.Email)
	setup.UserRepo.On("FindByEmail", ctx, email).Return(nil, sql.ErrNoRows).Times(2) // Called by OAuth service and UserService

	// Mock area for user creation (needed by UserService)
	testArea := &entities.Area{
		ID:       uuid.New(),
		Name:     "Default Area",
		Code:     "DEFAULT",
		Type:     valueobjects.AreaTypeTown,
		Status:   valueobjects.AreaStatusActive,
		ParentID: nil,
	}

	// Mock area detection for OAuth users with default coordinates (0, 0)
	setup.AreaRepo.On("FindByCoordinate", ctx, 0.0, 0.0).Return([]*entities.Area{testArea}, nil)

	// Mock create user via UserRepo (UserService uses Save internally)
	setup.UserRepo.On("Save", ctx, mock.MatchedBy(func(u *entities.User) bool {
		return u.Email == email && u.Name == profile.Name
	})).Return(nil) // Save returns error only, not the user

	// Mock create OAuth profile - UserID will be generated, just check provider and email
	setup.OAuthRepo.On("CreateOAuthProfile", ctx, mock.MatchedBy(func(p *entities.OAuthProfile) bool {
		return p.Provider == valueobjects.OAuthProviderGoogle && p.Email == profile.Email
	})).Return(nil)

	// Mock JWT key for token generation
	testJWTKey := generateTestJWTKey()
	setup.JWTKeyRepo.On("FindActiveKey", ctx).Return(testJWTKey, nil)

	// Mock refresh token creation - UserID will be generated, just check it's not empty
	setup.RefreshTokenRepo.On("Create", ctx, mock.MatchedBy(func(t *entities.RefreshToken) bool {
		return t.UserID != uuid.Nil
	})).Return(nil)

	// Mock cache operations for user data
	setup.Cache.On("Set", ctx, mock.AnythingOfType("string"), mock.Anything, mock.Anything).Return(nil).Maybe()
	setup.Cache.On("Get", ctx, mock.AnythingOfType("string")).Return(nil, nil).Maybe()
	setup.Cache.On("Delete", ctx, mock.AnythingOfType("string")).Return(nil).Maybe() // User cache invalidation

	// Mock event publishing - UserService uses PublishAsync, OAuthService uses Publish
	setup.EventBus.On("PublishAsync", ctx, mock.Anything).Return(nil).Once() // UserService publishes UserCreatedEvent
	setup.EventBus.On("Publish", ctx, mock.Anything).Return(nil).Times(2)    // OAuthService publishes linked and login events

	response, err := setup.Service.HandleCallback(ctx, req)

	// The service should successfully handle the OAuth callback
	assert.NoError(t, err)
	assert.NotNil(t, response)
	if response != nil {
		assert.NotEmpty(t, response.AccessToken)
		assert.NotEmpty(t, response.RefreshToken)
		assert.NotEqual(t, uuid.Nil, response.User.ID) // User ID should be generated
		assert.Equal(t, profile.Email, response.User.Email)
		assert.Equal(t, profile.Name, response.User.Name)
	}

	// Verify all expected mocks were called
	setup.OAuthRepo.AssertExpectations(t)
	setup.UserRepo.AssertExpectations(t)
	setup.OAuthProvider.AssertExpectations(t)
	setup.JWTKeyRepo.AssertExpectations(t)
	setup.RefreshTokenRepo.AssertExpectations(t)
	setup.EventBus.AssertExpectations(t)
}

func TestOAuthService_HandleCallback_ExistingUser(t *testing.T) {
	setup := setupOAuthServiceTest(t)

	ctx := context.Background()
	req := dtos.OAuthCallbackRequest{
		Code:     "test-code",
		State:    "test-state",
		Provider: "google",
	}

	// Create test OAuth state
	oauthState := &valueobjects.OAuthState{
		State:        req.State,
		CodeVerifier: "test-verifier",
		RedirectURI:  "http://localhost/callback",
		Provider:     valueobjects.OAuthProviderGoogle,
		ExpiresAt:    time.Now().Add(10 * time.Minute),
	}

	// Create existing user
	existingUser := &entities.User{
		ID:            uuid.New(),
		Email:         func() valueobjects.Email { e, _ := valueobjects.NewEmail("test@example.com"); return e }(),
		Name:          "Test User",
		Status:        entities.UserStatusActive,
		EmailVerified: true,
	}

	// Create existing OAuth profile
	existingProfile := &entities.OAuthProfile{
		ID:         uuid.New(),
		UserID:     existingUser.ID,
		Provider:   valueobjects.OAuthProviderGoogle,
		ProviderID: "google-user-123",
		Email:      "test@example.com",
		Name:       "Test User",
	}

	// Mock get state
	setup.OAuthRepo.On("GetOAuthState", ctx, req.State).Return(oauthState, nil)
	setup.OAuthRepo.On("DeleteOAuthState", ctx, req.State).Return(nil)

	// Mock token exchange
	tokens := &ports.OAuthTokens{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		ExpiresIn:    3600,
		TokenType:    "Bearer",
	}
	setup.OAuthProvider.On("ExchangeCode", ctx, req.Code, oauthState.CodeVerifier, oauthState.RedirectURI).
		Return(tokens, nil)

	// Mock user profile
	profile := &ports.OAuthUserProfile{
		ID:       "google-user-123",
		Email:    "test@example.com",
		Name:     "Test User",
		Verified: true,
	}
	setup.OAuthProvider.On("GetUserProfile", ctx, tokens.AccessToken).Return(profile, nil)

	// Mock existing OAuth profile found
	setup.OAuthRepo.On("GetOAuthProfile", ctx, valueobjects.OAuthProviderGoogle, profile.ID).
		Return(existingProfile, nil)

	// Mock FindByID for the user
	setup.UserRepo.On("FindByID", ctx, existingUser.ID).Return(existingUser, nil)

	// Mock update OAuth profile
	setup.OAuthRepo.On("UpdateOAuthProfile", ctx, mock.Anything).Return(nil)

	// Mock JWT key for token generation
	testJWTKey := generateTestJWTKey()
	setup.JWTKeyRepo.On("FindActiveKey", ctx).Return(testJWTKey, nil)

	// Mock refresh token creation
	setup.RefreshTokenRepo.On("Create", ctx, mock.MatchedBy(func(t *entities.RefreshToken) bool {
		return t.UserID == existingUser.ID
	})).Return(nil)

	// Mock cache operations
	setup.Cache.On("Set", ctx, mock.AnythingOfType("string"), mock.Anything, mock.Anything).Return(nil).Maybe()
	setup.Cache.On("Get", ctx, mock.AnythingOfType("string")).Return(nil, nil).Maybe()

	// Mock event publishing
	setup.EventBus.On("Publish", ctx, mock.Anything).Return(nil)

	response, err := setup.Service.HandleCallback(ctx, req)

	// The service should successfully handle the OAuth callback for existing user
	assert.NoError(t, err)
	if response != nil {
		assert.NotEmpty(t, response.AccessToken)
		assert.NotEmpty(t, response.RefreshToken)
		assert.Equal(t, existingUser.ID, response.User.ID)
	}

	// Verify all expected mocks were called
	setup.OAuthRepo.AssertExpectations(t)
	setup.UserRepo.AssertExpectations(t)
	setup.OAuthProvider.AssertExpectations(t)
	setup.JWTKeyRepo.AssertExpectations(t)
	setup.RefreshTokenRepo.AssertExpectations(t)
}

func TestOAuthService_HandleCallback_InvalidState(t *testing.T) {
	setup := setupOAuthServiceTest(t)

	ctx := context.Background()
	req := dtos.OAuthCallbackRequest{
		Code:     "test-code",
		State:    "invalid-state",
		Provider: "google",
	}

	// Mock state not found
	setup.OAuthRepo.On("GetOAuthState", ctx, req.State).Return(nil, sql.ErrNoRows)

	response, err := setup.Service.HandleCallback(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "Invalid OAuth state")

	setup.OAuthRepo.AssertExpectations(t)
}

func TestOAuthService_HandleCallback_ExpiredState(t *testing.T) {
	setup := setupOAuthServiceTest(t)

	ctx := context.Background()
	req := dtos.OAuthCallbackRequest{
		Code:     "test-code",
		State:    "test-state",
		Provider: "google",
	}

	// Create expired OAuth state
	expiredState := &valueobjects.OAuthState{
		State:        req.State,
		CodeVerifier: "test-verifier",
		RedirectURI:  "http://localhost/callback",
		Provider:     valueobjects.OAuthProviderGoogle,
		ExpiresAt:    time.Now().Add(-10 * time.Minute), // Expired
	}

	// Mock get expired state
	setup.OAuthRepo.On("GetOAuthState", ctx, req.State).Return(expiredState, nil)
	setup.OAuthRepo.On("DeleteOAuthState", ctx, req.State).Return(nil)

	response, err := setup.Service.HandleCallback(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "OAuth state has expired")

	setup.OAuthRepo.AssertExpectations(t)
}

func TestOAuthService_UnlinkOAuthAccount_Success(t *testing.T) {
	setup := setupOAuthServiceTest(t)

	ctx := context.Background()
	userID := uuid.New()
	req := dtos.OAuthUnlinkRequest{
		Provider: "google",
	}

	// Create user with password (has alternative auth method)
	user := &entities.User{
		ID:           userID,
		Email:        func() valueobjects.Email { e, _ := valueobjects.NewEmail("test@example.com"); return e }(),
		PasswordHash: "hashed-password",
	}

	// Create OAuth profile to unlink
	profile := &entities.OAuthProfile{
		ID:           uuid.New(),
		UserID:       userID,
		Provider:     valueobjects.OAuthProviderGoogle,
		ProviderID:   "google-user-123",
		AccessToken:  "test-token",
		RefreshToken: "test-refresh",
	}

	// Mock get OAuth profile
	setup.OAuthRepo.On("GetOAuthProfileByUserAndProvider", ctx, userID, valueobjects.OAuthProviderGoogle).
		Return(profile, nil)

	// Mock get user
	setup.UserRepo.On("FindByID", ctx, userID).Return(user, nil)

	// Mock get all OAuth profiles (user has only 1)
	setup.OAuthRepo.On("GetUserOAuthProfiles", ctx, userID).Return([]*entities.OAuthProfile{profile}, nil)

	// Mock revoke token
	setup.OAuthProvider.On("RevokeToken", ctx, profile.AccessToken).Return(nil)

	// Mock delete OAuth profile
	setup.OAuthRepo.On("DeleteOAuthProfile", ctx, userID, valueobjects.OAuthProviderGoogle).Return(nil)

	// Mock event publishing
	setup.EventBus.On("Publish", ctx, mock.Anything).Return(nil)

	err := setup.Service.UnlinkOAuthAccount(ctx, userID, req)

	assert.NoError(t, err)

	setup.OAuthRepo.AssertExpectations(t)
	setup.UserRepo.AssertExpectations(t)
	setup.OAuthProvider.AssertExpectations(t)
	setup.EventBus.AssertExpectations(t)
}

func TestOAuthService_UnlinkOAuthAccount_LastAuthMethod(t *testing.T) {
	setup := setupOAuthServiceTest(t)

	ctx := context.Background()
	userID := uuid.New()
	req := dtos.OAuthUnlinkRequest{
		Provider: "google",
	}

	// Create user WITHOUT password (OAuth is only auth method)
	user := &entities.User{
		ID:           userID,
		Email:        func() valueobjects.Email { e, _ := valueobjects.NewEmail("test@example.com"); return e }(),
		PasswordHash: "", // No password
	}

	// Create OAuth profile (only one)
	profile := &entities.OAuthProfile{
		ID:         uuid.New(),
		UserID:     userID,
		Provider:   valueobjects.OAuthProviderGoogle,
		ProviderID: "google-user-123",
	}

	// Mock get OAuth profile
	setup.OAuthRepo.On("GetOAuthProfileByUserAndProvider", ctx, userID, valueobjects.OAuthProviderGoogle).
		Return(profile, nil)

	// Mock get user
	setup.UserRepo.On("FindByID", ctx, userID).Return(user, nil)

	// Mock get all OAuth profiles (only 1)
	setup.OAuthRepo.On("GetUserOAuthProfiles", ctx, userID).Return([]*entities.OAuthProfile{profile}, nil)

	err := setup.Service.UnlinkOAuthAccount(ctx, userID, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Cannot unlink the last authentication method")

	setup.OAuthRepo.AssertExpectations(t)
	setup.UserRepo.AssertExpectations(t)
}

func TestOAuthService_GetUserOAuthProfiles(t *testing.T) {
	setup := setupOAuthServiceTest(t)

	ctx := context.Background()
	userID := uuid.New()

	// Create test OAuth profiles
	profiles := []*entities.OAuthProfile{
		{
			ID:         uuid.New(),
			UserID:     userID,
			Provider:   valueobjects.OAuthProviderGoogle,
			ProviderID: "google-123",
			Email:      "test@gmail.com",
			Name:       "Test User",
			Picture:    "https://example.com/pic.jpg",
			CreatedAt:  time.Now(),
		},
		{
			ID:         uuid.New(),
			UserID:     userID,
			Provider:   valueobjects.OAuthProviderFacebook,
			ProviderID: "fb-456",
			Email:      "test@facebook.com",
			Name:       "Test FB User",
			CreatedAt:  time.Now(),
		},
	}

	// Mock get profiles
	setup.OAuthRepo.On("GetUserOAuthProfiles", ctx, userID).Return(profiles, nil)

	response, err := setup.Service.GetUserOAuthProfiles(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Profiles, 2)
	assert.Equal(t, "google", response.Profiles[0].Provider)
	assert.Equal(t, "facebook", response.Profiles[1].Provider)

	setup.OAuthRepo.AssertExpectations(t)
}

func TestOAuthService_RefreshOAuthToken(t *testing.T) {
	setup := setupOAuthServiceTest(t)

	ctx := context.Background()
	userID := uuid.New()
	provider := valueobjects.OAuthProviderGoogle

	// Create OAuth profile with refresh token
	profile := &entities.OAuthProfile{
		ID:           uuid.New(),
		UserID:       userID,
		Provider:     provider,
		RefreshToken: "test-refresh-token",
	}

	// Mock get profile
	setup.OAuthRepo.On("GetOAuthProfileByUserAndProvider", ctx, userID, provider).Return(profile, nil)

	// Mock refresh token
	newTokens := &ports.OAuthTokens{
		AccessToken:  "new-access-token",
		RefreshToken: "test-refresh-token", // Same refresh token
		ExpiresIn:    3600,
		TokenType:    "Bearer",
	}
	setup.OAuthProvider.On("RefreshToken", ctx, profile.RefreshToken).Return(newTokens, nil)

	// Mock update tokens
	setup.OAuthRepo.On("UpdateOAuthTokens", ctx, profile.ID, newTokens.AccessToken, newTokens.RefreshToken, mock.MatchedBy(func(t *time.Time) bool {
		return t != nil && t.After(time.Now())
	})).Return(nil)

	err := setup.Service.RefreshOAuthToken(ctx, userID, provider)

	assert.NoError(t, err)

	setup.OAuthRepo.AssertExpectations(t)
	setup.OAuthProvider.AssertExpectations(t)
}

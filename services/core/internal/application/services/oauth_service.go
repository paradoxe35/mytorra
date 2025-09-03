package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/paradoxe35/mytorra/packages/config"
	"github.com/paradoxe35/mytorra/packages/errors"
	"github.com/paradoxe35/mytorra/packages/i18n"
	"github.com/paradoxe35/mytorra/packages/logger"
	"github.com/paradoxe35/mytorra/packages/security"
	"github.com/paradoxe35/mytorra/services/core/internal/adapters/oauth"
	"github.com/paradoxe35/mytorra/services/core/internal/application/dtos"
	"github.com/paradoxe35/mytorra/services/core/internal/application/ports"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/entities"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/events"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/valueobjects"
)

// OAuthService handles OAuth-related business logic
type OAuthService struct {
	oauthRepo   ports.OAuthRepository
	userRepo    ports.UserRepository
	authService *AuthService
	userService *UserService
	providers   map[valueobjects.OAuthProvider]ports.OAuthProvider
	eventBus    ports.EventBus
	logger      logger.Interface
	translator  *i18n.Translator
	encryptor   security.Encryptor
	stateExpiry time.Duration
	pkceEnabled bool
}

// NewOAuthService creates a new OAuth service instance
func NewOAuthService(
	oauthRepo ports.OAuthRepository,
	userRepo ports.UserRepository,
	authService *AuthService,
	userService *UserService,
	eventBus ports.EventBus,
	logger logger.Interface,
	translator *i18n.Translator,
	encryptor security.Encryptor,
	oauthConfig config.OAuthConfig,
) *OAuthService {
	// Initialize OAuth providers
	providers := make(map[valueobjects.OAuthProvider]ports.OAuthProvider)

	// Initialize Google provider if configured
	if oauthConfig.Google.ClientID != "" && oauthConfig.Google.ClientSecret != "" {
		providers[valueobjects.OAuthProviderGoogle] = oauth.NewGoogleOAuthProvider(
			oauthConfig.Google.ClientID,
			oauthConfig.Google.ClientSecret,
			oauthConfig.Google.RedirectURI,
			logger,
		)
	}

	// Initialize Facebook provider if configured
	if oauthConfig.Facebook.AppID != "" && oauthConfig.Facebook.AppSecret != "" {
		providers[valueobjects.OAuthProviderFacebook] = oauth.NewFacebookOAuthProvider(
			oauthConfig.Facebook.AppID,
			oauthConfig.Facebook.AppSecret,
			oauthConfig.Facebook.RedirectURI,
			logger,
		)
	}

	return &OAuthService{
		oauthRepo:   oauthRepo,
		userRepo:    userRepo,
		authService: authService,
		userService: userService,
		providers:   providers,
		eventBus:    eventBus,
		logger:      logger,
		translator:  translator,
		encryptor:   encryptor,
		stateExpiry: oauthConfig.StateExpiry,
		pkceEnabled: oauthConfig.PKCEEnabled,
	}
}

// InitiateOAuth initiates the OAuth flow
func (s *OAuthService) InitiateOAuth(ctx context.Context, req dtos.OAuthInitRequest) (*dtos.OAuthInitResponse, error) {
	// Parse and validate provider
	provider, err := valueobjects.ParseOAuthProvider(req.Provider)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errors.ErrBadRequest, s.translator.TWithContext(ctx, "oauth.provider_not_found"))
	}

	// Check if provider is configured
	oauthProvider, exists := s.providers[provider]
	if !exists {
		return nil, fmt.Errorf("%w: %s", errors.ErrBadRequest, s.translator.TWithContext(ctx, "oauth.provider_not_configured"))
	}

	// Generate state for CSRF protection
	state, err := s.generateSecureToken(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}

	// Generate PKCE code verifier if enabled
	var codeVerifier string
	if s.pkceEnabled {
		codeVerifier, err = s.generateSecureToken(43) // 43 chars for base64url encoded 32 bytes
		if err != nil {
			return nil, fmt.Errorf("failed to generate code verifier: %w", err)
		}
	}

	// Get client IP and user agent from context
	ipAddress := s.getClientIPFromContext(ctx)
	userAgent := s.getUserAgentFromContext(ctx)

	// Check if user is authenticated (for linking accounts)
	var userID *string
	if userIDValue := ctx.Value("user_id"); userIDValue != nil {
		if uid, ok := userIDValue.(string); ok {
			userID = &uid
		}
	}

	// Save state for validation
	oauthState := &valueobjects.OAuthState{
		State:        state,
		CodeVerifier: codeVerifier,
		RedirectURI:  req.RedirectURI,
		Provider:     provider,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		UserID:       userID,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(s.stateExpiry),
	}

	if err := s.oauthRepo.SaveOAuthState(ctx, oauthState); err != nil {
		return nil, fmt.Errorf("failed to save oauth state: %w", err)
	}

	// Generate authorization URL
	authURL := oauthProvider.GetAuthorizationURL(state, codeVerifier, req.RedirectURI)

	return &dtos.OAuthInitResponse{
		AuthorizationURL: authURL,
		State:            state,
	}, nil
}

// HandleCallback processes the OAuth callback
func (s *OAuthService) HandleCallback(ctx context.Context, req dtos.OAuthCallbackRequest) (*dtos.OAuthCallbackResponse, error) {
	// Parse and validate provider
	provider, err := valueobjects.ParseOAuthProvider(req.Provider)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errors.ErrBadRequest, s.translator.TWithContext(ctx, "oauth.provider_not_found"))
	}

	// Check if provider is configured
	oauthProvider, exists := s.providers[provider]
	if !exists {
		return nil, fmt.Errorf("%w: %s", errors.ErrBadRequest, s.translator.TWithContext(ctx, "oauth.provider_not_configured"))
	}

	// Validate state
	oauthState, err := s.oauthRepo.GetOAuthState(ctx, req.State)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errors.ErrUnauthorized, s.translator.TWithContext(ctx, "oauth.state_invalid"))
	}

	// Check if state is expired
	if oauthState.IsExpired() {
		// Clean up expired state
		_ = s.oauthRepo.DeleteOAuthState(ctx, req.State)
		return nil, fmt.Errorf("%w: %s", errors.ErrUnauthorized, s.translator.TWithContext(ctx, "oauth.state_expired"))
	}

	// Verify provider matches
	if oauthState.Provider != provider {
		return nil, fmt.Errorf("%w: %s", errors.ErrUnauthorized, s.translator.TWithContext(ctx, "oauth.state_invalid"))
	}

	// Delete state after validation (one-time use)
	defer func() {
		_ = s.oauthRepo.DeleteOAuthState(ctx, req.State)
	}()

	// Exchange authorization code for tokens
	tokens, err := oauthProvider.ExchangeCode(ctx, req.Code, oauthState.CodeVerifier, oauthState.RedirectURI)
	if err != nil {
		s.logger.WithError(err).Error("Failed to exchange OAuth code")
		return nil, fmt.Errorf("%w: %s", errors.ErrUnauthorized, s.translator.TWithContext(ctx, "oauth.code_invalid"))
	}

	// Get user profile from provider
	profile, err := oauthProvider.GetUserProfile(ctx, tokens.AccessToken)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get OAuth user profile")
		return nil, fmt.Errorf("%w: %s", errors.ErrInternalServer, s.translator.TWithContext(ctx, "oauth.profile_fetch_failed"))
	}

	// Check if this is a linking request (user already authenticated)
	if oauthState.UserID != nil {
		userID, err := uuid.Parse(*oauthState.UserID)
		if err == nil {
			return s.linkOAuthAccount(ctx, userID, provider, profile, tokens)
		}
	}

	// Try to find existing OAuth profile
	existingProfile, err := s.oauthRepo.GetOAuthProfile(ctx, provider, profile.ID)
	if err == nil && existingProfile != nil {
		// User exists, log them in
		return s.loginWithOAuth(ctx, existingProfile, tokens)
	}

	// Check if user exists with same email
	var user *entities.User
	if profile.Email != "" {
		email, err := valueobjects.NewEmail(profile.Email)
		if err == nil {
			user, _ = s.userRepo.FindByEmail(ctx, email)
		}
	}

	if user != nil {
		// Link OAuth to existing user with same email
		return s.linkOAuthAccount(ctx, user.ID, provider, profile, tokens)
	}

	// Create new user
	return s.createUserWithOAuth(ctx, provider, profile, tokens)
}

// LinkOAuthAccount links an OAuth account to an existing user
func (s *OAuthService) LinkOAuthAccount(ctx context.Context, userID uuid.UUID, req dtos.OAuthLinkRequest) (*dtos.OAuthInitResponse, error) {
	// Check if user exists
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errors.ErrNotFound, s.translator.TWithContext(ctx, "user.not_found"))
	}

	// Parse provider
	provider, err := valueobjects.ParseOAuthProvider(req.Provider)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errors.ErrBadRequest, s.translator.TWithContext(ctx, "oauth.provider_not_found"))
	}

	// Check if already linked
	existingProfile, _ := s.oauthRepo.GetOAuthProfileByUserAndProvider(ctx, user.ID, provider)
	if existingProfile != nil {
		return nil, fmt.Errorf("%w: %s", errors.ErrConflict, s.translator.TWithContext(ctx, "oauth.account_already_linked", map[string]interface{}{"Provider": provider.String()}))
	}

	// Initiate OAuth flow with user context
	ctx = context.WithValue(ctx, "user_id", userID.String())
	return s.InitiateOAuth(ctx, dtos.OAuthInitRequest{
		Provider:    req.Provider,
		RedirectURI: req.RedirectURI,
	})
}

// UnlinkOAuthAccount unlinks an OAuth account from a user
func (s *OAuthService) UnlinkOAuthAccount(ctx context.Context, userID uuid.UUID, req dtos.OAuthUnlinkRequest) error {
	// Parse provider
	provider, err := valueobjects.ParseOAuthProvider(req.Provider)
	if err != nil {
		return fmt.Errorf("%w: %s", errors.ErrBadRequest, s.translator.TWithContext(ctx, "oauth.provider_not_found"))
	}

	// Check if OAuth profile exists
	profile, err := s.oauthRepo.GetOAuthProfileByUserAndProvider(ctx, userID, provider)
	if err != nil {
		return fmt.Errorf("%w: %s", errors.ErrNotFound, s.translator.TWithContext(ctx, "oauth.account_not_linked", map[string]interface{}{"Provider": provider.String()}))
	}

	// Check if this is the last authentication method
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("%w: %s", errors.ErrNotFound, s.translator.TWithContext(ctx, "user.not_found"))
	}

	// Count OAuth profiles
	profiles, err := s.oauthRepo.GetUserOAuthProfiles(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user oauth profiles: %w", err)
	}

	// Check if user has password or other OAuth accounts
	if user.PasswordHash == "" && len(profiles) <= 1 {
		return fmt.Errorf("%w: %s", errors.ErrBadRequest, s.translator.TWithContext(ctx, "oauth.last_auth_method"))
	}

	// Revoke tokens if provider supports it
	if oauthProvider, exists := s.providers[provider]; exists && profile.AccessToken != "" {
		if err := oauthProvider.RevokeToken(ctx, profile.AccessToken); err != nil {
			s.logger.WithError(err).Warn("Failed to revoke OAuth token")
		}
	}

	// Delete OAuth profile
	if err := s.oauthRepo.DeleteOAuthProfile(ctx, userID, provider); err != nil {
		return fmt.Errorf("failed to delete oauth profile: %w", err)
	}

	// Emit event
	if s.eventBus != nil {
		event := events.NewUserUnlinkedOAuthEvent(userID, provider)
		if err := s.eventBus.Publish(ctx, event); err != nil {
			s.logger.WithError(err).Error("Failed to publish user unlinked oauth event")
		}
	}

	return nil
}

// GetUserOAuthProfiles returns all OAuth profiles for a user
func (s *OAuthService) GetUserOAuthProfiles(ctx context.Context, userID uuid.UUID) (*dtos.OAuthProfilesResponse, error) {
	profiles, err := s.oauthRepo.GetUserOAuthProfiles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user oauth profiles: %w", err)
	}

	response := &dtos.OAuthProfilesResponse{
		Profiles: make([]*dtos.OAuthProfileResponse, len(profiles)),
	}

	for i, profile := range profiles {
		response.Profiles[i] = &dtos.OAuthProfileResponse{
			ID:          profile.ID,
			Provider:    profile.Provider.String(),
			Email:       profile.Email,
			Name:        profile.Name,
			Picture:     profile.Picture,
			ConnectedAt: profile.CreatedAt,
		}
	}

	return response, nil
}

// RefreshOAuthToken refreshes an expired OAuth token
func (s *OAuthService) RefreshOAuthToken(ctx context.Context, userID uuid.UUID, provider valueobjects.OAuthProvider) error {
	// Get OAuth profile
	profile, err := s.oauthRepo.GetOAuthProfileByUserAndProvider(ctx, userID, provider)
	if err != nil {
		return fmt.Errorf("oauth profile not found: %w", err)
	}

	// Check if provider supports refresh
	oauthProvider, exists := s.providers[provider]
	if !exists {
		return fmt.Errorf("oauth provider not configured")
	}

	// Check if we have a refresh token
	if profile.RefreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	// Refresh the token
	tokens, err := oauthProvider.RefreshToken(ctx, profile.RefreshToken)
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	// Update tokens in database
	var expiry *time.Time
	if tokens.ExpiresIn > 0 {
		e := time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)
		expiry = &e
	}

	if err := s.oauthRepo.UpdateOAuthTokens(ctx, profile.ID, tokens.AccessToken, tokens.RefreshToken, expiry); err != nil {
		return fmt.Errorf("failed to update oauth tokens: %w", err)
	}

	return nil
}

// Helper methods

func (s *OAuthService) linkOAuthAccount(ctx context.Context, userID uuid.UUID, provider valueobjects.OAuthProvider, profile *ports.OAuthUserProfile, tokens *ports.OAuthTokens) (*dtos.OAuthCallbackResponse, error) {
	// Create OAuth profile
	rawProfile, _ := json.Marshal(profile.Raw)
	oauthProfile := entities.NewOAuthProfile(userID, provider, profile.ID, profile.Email, profile.Name)
	oauthProfile.Picture = profile.Picture
	oauthProfile.RawProfile = rawProfile
	oauthProfile.UpdateTokens(tokens.AccessToken, tokens.RefreshToken, tokens.ExpiresIn)
	oauthProfile.Scopes = tokens.Scopes

	if err := s.oauthRepo.CreateOAuthProfile(ctx, oauthProfile); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil, fmt.Errorf("%w: %s", errors.ErrConflict, s.translator.TWithContext(ctx, "oauth.account_already_linked", map[string]interface{}{"Provider": provider.String()}))
		}
		return nil, fmt.Errorf("failed to create oauth profile: %w", err)
	}

	// Get user for response
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Generate auth tokens
	authResponse, err := s.authService.GenerateTokensForUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Emit event
	if s.eventBus != nil {
		event := events.NewUserLinkedOAuthEvent(userID, provider, profile.ID, profile.Email)
		if err := s.eventBus.Publish(ctx, event); err != nil {
			s.logger.WithError(err).Error("Failed to publish user linked oauth event")
		}
	}

	return &dtos.OAuthCallbackResponse{
		User:         authResponse.User,
		AccessToken:  authResponse.AccessToken,
		RefreshToken: authResponse.RefreshToken,
		ExpiresIn:    int64(authResponse.ExpiresIn),
		IsNewUser:    false,
	}, nil
}

func (s *OAuthService) loginWithOAuth(ctx context.Context, profile *entities.OAuthProfile, tokens *ports.OAuthTokens) (*dtos.OAuthCallbackResponse, error) {
	// Update tokens
	profile.UpdateTokens(tokens.AccessToken, tokens.RefreshToken, tokens.ExpiresIn)
	if err := s.oauthRepo.UpdateOAuthProfile(ctx, profile); err != nil {
		s.logger.WithError(err).Warn("Failed to update OAuth tokens")
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, profile.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Check if user can login
	if !user.CanLogin() {
		return nil, errors.ErrAccountSuspended
	}

	// Generate auth tokens
	authResponse, err := s.authService.GenerateTokensForUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Emit event
	if s.eventBus != nil {
		ipAddress := s.getClientIPFromContext(ctx)
		userAgent := s.getUserAgentFromContext(ctx)
		event := events.NewOAuthLoginEvent(user.ID, profile.Provider, ipAddress, userAgent)
		if err := s.eventBus.Publish(ctx, event); err != nil {
			s.logger.WithError(err).Error("Failed to publish oauth login event")
		}
	}

	return &dtos.OAuthCallbackResponse{
		User:         authResponse.User,
		AccessToken:  authResponse.AccessToken,
		RefreshToken: authResponse.RefreshToken,
		ExpiresIn:    int64(authResponse.ExpiresIn),
		IsNewUser:    false,
	}, nil
}

func (s *OAuthService) createUserWithOAuth(ctx context.Context, provider valueobjects.OAuthProvider, profile *ports.OAuthUserProfile, tokens *ports.OAuthTokens) (*dtos.OAuthCallbackResponse, error) {
	// Create user request with placeholder location coordinates
	// This allows the location service to detect an appropriate area
	lat := 0.0
	lng := 0.0
	createUserReq := dtos.RegisterRequest{
		Email:           profile.Email,
		Name:            profile.Name,
		Phone:           "+1234567890",     // Placeholder phone number for OAuth users
		Password:        "OAuth2User@2024", // Placeholder password for OAuth users (meets 12 char requirement)
		ConfirmPassword: "OAuth2User@2024",
		AcceptTerms:     true,
		// Provide default coordinates for OAuth users (e.g., default city center)
		Latitude:  &lat,
		Longitude: &lng,
	}

	// Create user
	user, err := s.userService.CreateUser(ctx, createUserReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create OAuth profile
	rawProfile, _ := json.Marshal(profile.Raw)
	oauthProfile := entities.NewOAuthProfile(user.ID, provider, profile.ID, profile.Email, profile.Name)
	oauthProfile.Picture = profile.Picture
	oauthProfile.RawProfile = rawProfile
	oauthProfile.UpdateTokens(tokens.AccessToken, tokens.RefreshToken, tokens.ExpiresIn)
	oauthProfile.Scopes = tokens.Scopes

	if err := s.oauthRepo.CreateOAuthProfile(ctx, oauthProfile); err != nil {
		return nil, fmt.Errorf("failed to create oauth profile: %w", err)
	}

	// Generate auth tokens
	authResponse, err := s.authService.GenerateTokensForUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Emit events
	if s.eventBus != nil {
		// OAuth login event
		ipAddress := s.getClientIPFromContext(ctx)
		userAgent := s.getUserAgentFromContext(ctx)
		loginEvent := events.NewOAuthLoginEvent(user.ID, provider, ipAddress, userAgent)
		if err := s.eventBus.Publish(ctx, loginEvent); err != nil {
			s.logger.WithError(err).Error("Failed to publish oauth login event")
		}

		// User linked OAuth event
		linkedEvent := events.NewUserLinkedOAuthEvent(user.ID, provider, profile.ID, profile.Email)
		if err := s.eventBus.Publish(ctx, linkedEvent); err != nil {
			s.logger.WithError(err).Error("Failed to publish user linked oauth event")
		}
	}

	return &dtos.OAuthCallbackResponse{
		User:         authResponse.User,
		AccessToken:  authResponse.AccessToken,
		RefreshToken: authResponse.RefreshToken,
		ExpiresIn:    int64(authResponse.ExpiresIn),
		IsNewUser:    true,
	}, nil
}

func (s *OAuthService) generateSecureToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b)[:length], nil
}

func (s *OAuthService) getClientIPFromContext(ctx context.Context) string {
	if ip := ctx.Value("client_ip"); ip != nil {
		if ipStr, ok := ip.(string); ok {
			return ipStr
		}
	}
	return ""
}

func (s *OAuthService) getUserAgentFromContext(ctx context.Context) string {
	if ua := ctx.Value("user_agent"); ua != nil {
		if uaStr, ok := ua.(string); ok {
			return uaStr
		}
	}
	return ""
}

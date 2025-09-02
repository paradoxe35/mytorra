package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	httpPkg "github.com/paradoxe35/mytorra/packages/http"
	"github.com/paradoxe35/mytorra/packages/i18n"
	"github.com/paradoxe35/mytorra/packages/logger"
	"github.com/paradoxe35/mytorra/services/core/internal/adapters/http/middleware"
	"github.com/paradoxe35/mytorra/services/core/internal/application/dtos"
	"github.com/paradoxe35/mytorra/services/core/internal/application/services"
)

// OAuthHandler handles OAuth-related HTTP requests
type OAuthHandler struct {
	oauthService *services.OAuthService
	logger       logger.Interface
	translator   *i18n.Translator
}

// NewOAuthHandler creates a new OAuth handler instance
func NewOAuthHandler(oauthService *services.OAuthService, logger logger.Interface, translator *i18n.Translator) *OAuthHandler {
	return &OAuthHandler{
		oauthService: oauthService,
		logger:       logger,
		translator:   translator,
	}
}

// InitiateOAuth initiates the OAuth flow
// @Summary Initiate OAuth flow
// @Description Generate OAuth authorization URL
// @Tags OAuth
// @Accept json
// @Produce json
// @Param request body dtos.OAuthInitRequest true "OAuth init request"
// @Success 200 {object} dtos.OAuthInitResponse
// @Failure 400 {object} httpPkg.ErrorResponse
// @Router /api/v1/auth/oauth/init [post]
func (h *OAuthHandler) InitiateOAuth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Add client IP and user agent to context for state tracking
	ctx = context.WithValue(ctx, "client_ip", httpPkg.GetClientIP(r))
	ctx = context.WithValue(ctx, "user_agent", r.Header.Get("User-Agent"))

	var req dtos.OAuthInitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST",
			h.translator.TWithContext(ctx, "error.invalid_request_body"))
		return
	}

	response, err := h.oauthService.InitiateOAuth(ctx, req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to initiate OAuth")
		httpPkg.RespondError(w, http.StatusBadRequest, "OAUTH_INIT_FAILED",
			h.translator.TWithContext(ctx, "oauth.init_failed"))
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, response)
}

// HandleCallback handles OAuth callback
// @Summary Handle OAuth callback
// @Description Process OAuth provider callback
// @Tags OAuth
// @Accept json
// @Produce json
// @Param request body dtos.OAuthCallbackRequest true "OAuth callback request"
// @Success 200 {object} dtos.OAuthCallbackResponse
// @Failure 400 {object} httpPkg.ErrorResponse
// @Router /api/v1/auth/oauth/callback [post]
func (h *OAuthHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Add client IP and user agent to context
	ctx = context.WithValue(ctx, "client_ip", httpPkg.GetClientIP(r))
	ctx = context.WithValue(ctx, "user_agent", r.Header.Get("User-Agent"))

	var req dtos.OAuthCallbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST",
			h.translator.TWithContext(ctx, "error.invalid_request_body"))
		return
	}

	response, err := h.oauthService.HandleCallback(ctx, req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to handle OAuth callback")
		httpPkg.RespondError(w, http.StatusBadRequest, "OAUTH_CALLBACK_FAILED",
			h.translator.TWithContext(ctx, "oauth.callback_failed"))
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, response)
}

// GetOAuthProfiles returns user's connected OAuth profiles
// @Summary Get OAuth profiles
// @Description Get user's connected OAuth accounts
// @Tags OAuth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dtos.OAuthProfilesResponse
// @Failure 401 {object} httpPkg.ErrorResponse
// @Router /api/v1/auth/oauth/profiles [get]
func (h *OAuthHandler) GetOAuthProfiles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context (set by auth middleware)
	userIDStr := ctx.Value(middleware.UserIDKey).(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_USER_ID",
			h.translator.TWithContext(ctx, "error.invalid_user_id"))
		return
	}

	response, err := h.oauthService.GetUserOAuthProfiles(ctx, userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get OAuth profiles")
		httpPkg.RespondError(w, http.StatusInternalServerError, "GET_PROFILES_FAILED",
			h.translator.TWithContext(ctx, "oauth.get_profiles_failed"))
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, response)
}

// LinkOAuthAccount links an OAuth account to existing user
// @Summary Link OAuth account
// @Description Link an OAuth account to the authenticated user
// @Tags OAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dtos.OAuthLinkRequest true "OAuth link request"
// @Success 200 {object} dtos.OAuthInitResponse
// @Failure 400 {object} httpPkg.ErrorResponse
// @Router /api/v1/auth/oauth/link [post]
func (h *OAuthHandler) LinkOAuthAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context (set by auth middleware)
	userIDStr := ctx.Value(middleware.UserIDKey).(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_USER_ID",
			h.translator.TWithContext(ctx, "error.invalid_user_id"))
		return
	}

	var req dtos.OAuthLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST",
			h.translator.TWithContext(ctx, "error.invalid_request_body"))
		return
	}

	response, err := h.oauthService.LinkOAuthAccount(ctx, userID, req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to link OAuth account")
		httpPkg.RespondError(w, http.StatusBadRequest, "LINK_FAILED",
			h.translator.TWithContext(ctx, "oauth.link_failed"))
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, response)
}

// UnlinkOAuthAccount unlinks an OAuth account from user
// @Summary Unlink OAuth account
// @Description Unlink an OAuth account from the authenticated user
// @Tags OAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dtos.OAuthUnlinkRequest true "OAuth unlink request"
// @Success 200 {object} httpPkg.SuccessResponse
// @Failure 400 {object} httpPkg.ErrorResponse
// @Router /api/v1/auth/oauth/unlink [delete]
func (h *OAuthHandler) UnlinkOAuthAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context (set by auth middleware)
	userIDStr := ctx.Value(middleware.UserIDKey).(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_USER_ID",
			h.translator.TWithContext(ctx, "error.invalid_user_id"))
		return
	}

	var req dtos.OAuthUnlinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST",
			h.translator.TWithContext(ctx, "error.invalid_request_body"))
		return
	}

	if err := h.oauthService.UnlinkOAuthAccount(ctx, userID, req); err != nil {
		h.logger.WithError(err).Error("Failed to unlink OAuth account")
		httpPkg.RespondError(w, http.StatusBadRequest, "UNLINK_FAILED",
			h.translator.TWithContext(ctx, "oauth.unlink_failed"))
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, map[string]string{
		"message": h.translator.TWithContext(ctx, "oauth.unlink_success"),
	})
}

// RegisterRoutes registers OAuth routes
func (h *OAuthHandler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	// Public routes
	r.Post("/oauth/init", h.InitiateOAuth)
	r.Post("/oauth/callback", h.HandleCallback)

	// Protected routes - require authentication
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)
		r.Get("/oauth/profiles", h.GetOAuthProfiles)
		r.Post("/oauth/link", h.LinkOAuthAccount)
		r.Delete("/oauth/unlink", h.UnlinkOAuthAccount)
	})
}

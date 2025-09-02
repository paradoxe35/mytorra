package handlers

import (
	"encoding/json"
	"net/http"

	httpPkg "github.com/paradoxe35/mytorra/packages/http"
	"github.com/paradoxe35/mytorra/packages/i18n"
	"github.com/paradoxe35/mytorra/packages/logger"
	"github.com/paradoxe35/mytorra/services/core/internal/adapters/http/middleware"
	"github.com/paradoxe35/mytorra/services/core/internal/application/dtos"
	"github.com/paradoxe35/mytorra/services/core/internal/application/services"
)

type AuthHandler struct {
	authService *services.AuthService
	userService *services.UserService
	logger      logger.Interface
	translator  *i18n.Translator
}

func NewAuthHandler(authService *services.AuthService, userService *services.UserService, logger logger.Interface, translator *i18n.Translator) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
		logger:      logger,
		translator:  translator,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req dtos.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", h.translator.TWithContext(ctx, "error.invalid_request_body"))
		return
	}

	user, err := h.userService.CreateUser(ctx, req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create user")
		httpPkg.RespondError(w, http.StatusBadRequest, "REGISTRATION_FAILED", err.Error())
		return
	}

	httpPkg.RespondJSON(w, http.StatusCreated, user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req dtos.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", h.translator.TWithContext(ctx, "error.invalid_request_body"))
		return
	}

	response, err := h.authService.Login(ctx, req)
	if err != nil {
		h.logger.WithError(err).Error("Login failed")
		httpPkg.RespondError(w, http.StatusUnauthorized, "LOGIN_FAILED", h.translator.TWithContext(ctx, "auth.login_failed"))
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, response)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req dtos.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", h.translator.TWithContext(ctx, "error.invalid_request_body"))
		return
	}

	response, err := h.authService.RefreshToken(ctx, req)
	if err != nil {
		h.logger.WithError(err).Error("Token refresh failed")
		httpPkg.RespondError(w, http.StatusUnauthorized, "REFRESH_FAILED", h.translator.TWithContext(ctx, "auth.refresh_failed"))
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, response)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Get user ID from context (set by auth middleware)
	userID := ctx.Value(middleware.UserIDKey).(string)

	if err := h.authService.Logout(ctx, userID); err != nil {
		h.logger.WithError(err).Error("Logout failed")
		httpPkg.RespondError(w, http.StatusInternalServerError, "LOGOUT_FAILED", h.translator.TWithContext(ctx, "auth.logout_failed"))
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, map[string]string{
		"message": h.translator.TWithContext(ctx, "auth.logout_success_message"),
	})
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req dtos.ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", h.translator.TWithContext(ctx, "error.invalid_request_body"))
		return
	}

	if err := h.authService.ForgotPassword(ctx, req); err != nil {
		h.logger.WithError(err).Error("Forgot password failed")
		httpPkg.RespondError(w, http.StatusBadRequest, "REQUEST_FAILED", h.translator.TWithContext(ctx, "auth.request_failed"))
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, map[string]string{
		"message": h.translator.TWithContext(ctx, "auth.password_reset_email_sent"),
	})
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req dtos.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", h.translator.TWithContext(ctx, "error.invalid_request_body"))
		return
	}

	if err := h.authService.ResetPassword(ctx, req); err != nil {
		h.logger.WithError(err).Error("Reset password failed")
		httpPkg.RespondError(w, http.StatusBadRequest, "RESET_FAILED", err.Error())
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, map[string]string{
		"message": h.translator.TWithContext(ctx, "auth.password_reset_success_message"),
	})
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	token := r.URL.Query().Get("token")
	if token == "" {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_TOKEN", h.translator.TWithContext(ctx, "error.invalid_token"))
		return
	}

	if err := h.authService.VerifyEmail(ctx, token); err != nil {
		h.logger.WithError(err).Error("Email verification failed")
		httpPkg.RespondError(w, http.StatusBadRequest, "VERIFICATION_FAILED", h.translator.TWithContext(ctx, "auth.verification_failed"))
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, map[string]string{
		"message": h.translator.TWithContext(ctx, "auth.email_verified_message"),
	})
}

func (h *AuthHandler) VerifyPhone(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req dtos.VerifyPhoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", h.translator.TWithContext(ctx, "error.invalid_request_body"))
		return
	}

	if err := h.authService.VerifyPhone(ctx, req); err != nil {
		h.logger.WithError(err).Error("Phone verification failed")
		httpPkg.RespondError(w, http.StatusBadRequest, "VERIFICATION_FAILED", h.translator.TWithContext(ctx, "auth.verification_failed"))
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, map[string]string{
		"message": h.translator.TWithContext(ctx, "auth.phone_verified_message"),
	})
}

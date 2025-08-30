package handlers

import (
	"encoding/json"
	"net/http"

	httpPkg "github.com/paradoxe35/torra/packages/http"
	"github.com/paradoxe35/torra/packages/logger"
	"github.com/paradoxe35/torra/services/core/internal/application/dtos"
	"github.com/paradoxe35/torra/services/core/internal/application/services"
)

type AuthHandler struct {
	authService *services.AuthService
	userService *services.UserService
	logger      *logger.Logger
}

func NewAuthHandler(authService *services.AuthService, userService *services.UserService, logger *logger.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
		logger:      logger,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dtos.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	user, err := h.userService.CreateUser(r.Context(), req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create user")
		httpPkg.RespondError(w, http.StatusBadRequest, "REGISTRATION_FAILED", err.Error())
		return
	}

	httpPkg.RespondJSON(w, http.StatusCreated, user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dtos.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	response, err := h.authService.Login(r.Context(), req)
	if err != nil {
		h.logger.WithError(err).Error("Login failed")
		httpPkg.RespondError(w, http.StatusUnauthorized, "LOGIN_FAILED", "Invalid credentials")
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, response)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req dtos.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	response, err := h.authService.RefreshToken(r.Context(), req)
	if err != nil {
		h.logger.WithError(err).Error("Token refresh failed")
		httpPkg.RespondError(w, http.StatusUnauthorized, "REFRESH_FAILED", "Invalid refresh token")
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, response)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID := r.Context().Value("user_id").(string)

	if err := h.authService.Logout(r.Context(), userID); err != nil {
		h.logger.WithError(err).Error("Logout failed")
		httpPkg.RespondError(w, http.StatusInternalServerError, "LOGOUT_FAILED", "Failed to logout")
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "Successfully logged out",
	})
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req dtos.ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if err := h.authService.ForgotPassword(r.Context(), req); err != nil {
		h.logger.WithError(err).Error("Forgot password failed")
		httpPkg.RespondError(w, http.StatusBadRequest, "REQUEST_FAILED", err.Error())
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "Password reset email sent",
	})
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req dtos.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if err := h.authService.ResetPassword(r.Context(), req); err != nil {
		h.logger.WithError(err).Error("Reset password failed")
		httpPkg.RespondError(w, http.StatusBadRequest, "RESET_FAILED", err.Error())
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "Password reset successful",
	})
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_TOKEN", "Verification token is required")
		return
	}

	if err := h.authService.VerifyEmail(r.Context(), token); err != nil {
		h.logger.WithError(err).Error("Email verification failed")
		httpPkg.RespondError(w, http.StatusBadRequest, "VERIFICATION_FAILED", err.Error())
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "Email verified successfully",
	})
}

func (h *AuthHandler) VerifyPhone(w http.ResponseWriter, r *http.Request) {
	var req dtos.VerifyPhoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if err := h.authService.VerifyPhone(r.Context(), req); err != nil {
		h.logger.WithError(err).Error("Phone verification failed")
		httpPkg.RespondError(w, http.StatusBadRequest, "VERIFICATION_FAILED", err.Error())
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "Phone verified successfully",
	})
}

package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	httpPkg "github.com/paradoxe35/mytorra/packages/http"
	"github.com/paradoxe35/mytorra/packages/i18n"
	"github.com/paradoxe35/mytorra/packages/logger"
	"github.com/paradoxe35/mytorra/services/core/internal/adapters/http/middleware"
	"github.com/paradoxe35/mytorra/services/core/internal/application/dtos"
	"github.com/paradoxe35/mytorra/services/core/internal/application/services"
)

type UserHandler struct {
	userService *services.UserService
	logger      logger.Interface
	translator  *i18n.Translator
}

func NewUserHandler(userService *services.UserService, logger logger.Interface, translator *i18n.Translator) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
		translator:  translator,
	}
}

func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Get user ID from context (set by auth middleware)
	userIDStr := ctx.Value(middleware.UserIDKey).(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_USER_ID", h.translator.TWithContext(ctx, "error.invalid_user_id"))
		return
	}

	user, err := h.userService.GetUser(ctx, userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get user")
		httpPkg.RespondError(w, http.StatusNotFound, "USER_NOT_FOUND", h.translator.TWithContext(ctx, "user.not_found"))
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, user)
}

func (h *UserHandler) UpdateCurrentUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDStr := ctx.Value(middleware.UserIDKey).(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_USER_ID", h.translator.TWithContext(ctx, "error.invalid_user_id"))
		return
	}

	var req dtos.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", h.translator.TWithContext(ctx, "error.invalid_request_body"))
		return
	}

	user, err := h.userService.UpdateUser(ctx, userID, req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update user")
		httpPkg.RespondError(w, http.StatusBadRequest, "UPDATE_FAILED", h.translator.TWithContext(ctx, "user.update_failed"))
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, user)
}

func (h *UserHandler) DeleteCurrentUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDStr := ctx.Value(middleware.UserIDKey).(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_USER_ID", h.translator.TWithContext(ctx, "error.invalid_user_id"))
		return
	}

	if err := h.userService.DeleteUser(ctx, userID); err != nil {
		h.logger.WithError(err).Error("Failed to delete user")
		httpPkg.RespondError(w, http.StatusBadRequest, "DELETE_FAILED", h.translator.TWithContext(ctx, "user.delete_failed"))
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, map[string]string{
		"message": h.translator.TWithContext(ctx, "user.deleted_successfully"),
	})
}

func (h *UserHandler) EnableMFA(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDStr := ctx.Value(middleware.UserIDKey).(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_USER_ID", h.translator.TWithContext(ctx, "error.invalid_user_id"))
		return
	}

	response, err := h.userService.EnableMFA(ctx, userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to enable MFA")
		httpPkg.RespondError(w, http.StatusBadRequest, "MFA_ENABLE_FAILED", h.translator.TWithContext(ctx, "user.mfa_enable_failed"))
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, response)
}

func (h *UserHandler) DisableMFA(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDStr := ctx.Value(middleware.UserIDKey).(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_USER_ID", h.translator.TWithContext(ctx, "error.invalid_user_id"))
		return
	}

	var req dtos.DisableMFARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", h.translator.TWithContext(ctx, "error.invalid_request_body"))
		return
	}

	if err := h.userService.DisableMFA(ctx, userID, req); err != nil {
		h.logger.WithError(err).Error("Failed to disable MFA")
		httpPkg.RespondError(w, http.StatusBadRequest, "MFA_DISABLE_FAILED", h.translator.TWithContext(ctx, "user.mfa_disable_failed"))
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, map[string]string{
		"message": h.translator.TWithContext(ctx, "user.mfa_disabled_successfully"),
	})
}

func (h *UserHandler) UpdateArea(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDStr := ctx.Value(middleware.UserIDKey).(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_USER_ID", h.translator.TWithContext(ctx, "error.invalid_user_id"))
		return
	}

	var req dtos.UpdateAreaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", h.translator.TWithContext(ctx, "error.invalid_request_body"))
		return
	}

	user, err := h.userService.UpdateUserArea(ctx, userID, req.AreaID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update user area")
		httpPkg.RespondError(w, http.StatusBadRequest, "UPDATE_FAILED", h.translator.TWithContext(ctx, "user.update_failed"))
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, user)
}

func (h *UserHandler) GetVerificationStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDStr := ctx.Value(middleware.UserIDKey).(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_USER_ID", h.translator.TWithContext(ctx, "error.invalid_user_id"))
		return
	}

	status, err := h.userService.GetVerificationStatus(ctx, userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get verification status")
		httpPkg.RespondError(w, http.StatusBadRequest, "REQUEST_FAILED", h.translator.TWithContext(ctx, "auth.request_failed"))
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, status)
}

func (h *UserHandler) VerifyGovernmentID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDStr := ctx.Value(middleware.UserIDKey).(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_USER_ID", h.translator.TWithContext(ctx, "error.invalid_user_id"))
		return
	}

	// Handle multipart form for file upload
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB max
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", h.translator.TWithContext(ctx, "error.invalid_form_data"))
		return
	}

	file, header, err := r.FormFile("government_id")
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "FILE_REQUIRED", h.translator.TWithContext(ctx, "error.file_required"))
		return
	}
	defer file.Close()

	if err := h.userService.VerifyGovernmentID(ctx, userID, file, header); err != nil {
		h.logger.WithError(err).Error("Failed to verify government ID")
		httpPkg.RespondError(w, http.StatusBadRequest, "VERIFICATION_FAILED", err.Error())
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, map[string]string{
		"message": h.translator.TWithContext(ctx, "user.government_id_verification_initiated"),
	})
}

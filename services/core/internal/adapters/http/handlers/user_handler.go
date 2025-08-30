package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	httpPkg "github.com/paradoxe35/torra/packages/http"
	"github.com/paradoxe35/torra/packages/logger"
	"github.com/paradoxe35/torra/services/core/internal/application/dtos"
	"github.com/paradoxe35/torra/services/core/internal/application/services"
)

type UserHandler struct {
	userService *services.UserService
	logger      *logger.Logger
}

func NewUserHandler(userService *services.UserService, logger *logger.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
	}
}

func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userIDStr := r.Context().Value("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID")
		return
	}

	user, err := h.userService.GetUser(r.Context(), userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get user")
		httpPkg.RespondError(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, user)
}

func (h *UserHandler) UpdateCurrentUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID")
		return
	}

	var req dtos.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	user, err := h.userService.UpdateUser(r.Context(), userID, req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update user")
		httpPkg.RespondError(w, http.StatusBadRequest, "UPDATE_FAILED", err.Error())
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, user)
}

func (h *UserHandler) DeleteCurrentUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID")
		return
	}

	if err := h.userService.DeleteUser(r.Context(), userID); err != nil {
		h.logger.WithError(err).Error("Failed to delete user")
		httpPkg.RespondError(w, http.StatusBadRequest, "DELETE_FAILED", err.Error())
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "User deleted successfully",
	})
}

func (h *UserHandler) EnableMFA(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID")
		return
	}

	response, err := h.userService.EnableMFA(r.Context(), userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to enable MFA")
		httpPkg.RespondError(w, http.StatusBadRequest, "MFA_ENABLE_FAILED", err.Error())
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, response)
}

func (h *UserHandler) DisableMFA(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID")
		return
	}

	var req dtos.DisableMFARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if err := h.userService.DisableMFA(r.Context(), userID, req); err != nil {
		h.logger.WithError(err).Error("Failed to disable MFA")
		httpPkg.RespondError(w, http.StatusBadRequest, "MFA_DISABLE_FAILED", err.Error())
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "MFA disabled successfully",
	})
}

func (h *UserHandler) UpdateArea(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID")
		return
	}

	var req dtos.UpdateAreaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	user, err := h.userService.UpdateUserArea(r.Context(), userID, req.AreaID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update user area")
		httpPkg.RespondError(w, http.StatusBadRequest, "UPDATE_FAILED", err.Error())
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, user)
}

func (h *UserHandler) GetVerificationStatus(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID")
		return
	}

	status, err := h.userService.GetVerificationStatus(r.Context(), userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get verification status")
		httpPkg.RespondError(w, http.StatusBadRequest, "REQUEST_FAILED", err.Error())
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, status)
}

func (h *UserHandler) VerifyGovernmentID(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID")
		return
	}

	// Handle multipart form for file upload
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB max
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Failed to parse form data")
		return
	}

	file, header, err := r.FormFile("government_id")
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "FILE_REQUIRED", "Government ID image is required")
		return
	}
	defer file.Close()

	if err := h.userService.VerifyGovernmentID(r.Context(), userID, file, header); err != nil {
		h.logger.WithError(err).Error("Failed to verify government ID")
		httpPkg.RespondError(w, http.StatusBadRequest, "VERIFICATION_FAILED", err.Error())
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "Government ID verification initiated",
	})
}

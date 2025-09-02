package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/shopspring/decimal"

	httpPkg "github.com/paradoxe35/torra/packages/http"
	"github.com/paradoxe35/torra/packages/i18n"
	"github.com/paradoxe35/torra/packages/logger"
	"github.com/paradoxe35/torra/services/core/internal/adapters/http/middleware"
	"github.com/paradoxe35/torra/services/core/internal/application/dtos"
	"github.com/paradoxe35/torra/services/core/internal/application/services"
	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
)

// SubscriptionHandler handles subscription-related HTTP requests
type SubscriptionHandler struct {
	subscriptionService *services.SubscriptionService
	logger              logger.Interface
	translator          *i18n.Translator
}

// NewSubscriptionHandler creates a new subscription handler
func NewSubscriptionHandler(
	subscriptionService *services.SubscriptionService,
	logger logger.Interface,
	translator *i18n.Translator,
) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionService: subscriptionService,
		logger:              logger,
		translator:          translator,
	}
}

// CreateSubscription handles POST /api/v1/subscriptions
func (h *SubscriptionHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userIDStr := ctx.Value(middleware.UserIDKey).(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_USER_ID",
			h.translator.TWithContext(ctx, "error.invalid_user_id"))
		return
	}

	// Parse request body
	var req dtos.CreateSubscriptionDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST",
			h.translator.TWithContext(ctx, "error.invalid_request_body"))
		return
	}

	// Set user ID from context
	req.UserID = userID

	// Create subscription
	subscription, err := h.subscriptionService.CreateSubscription(ctx, req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create subscription")
		httpPkg.RespondError(w, http.StatusInternalServerError, "SUBSCRIPTION_CREATION_FAILED",
			err.Error())
		return
	}

	// Build response
	response := h.buildSubscriptionResponse(subscription)
	httpPkg.RespondJSON(w, http.StatusCreated, response)
}

// GetCurrentSubscription handles GET /api/v1/subscriptions/current
func (h *SubscriptionHandler) GetCurrentSubscription(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userIDStr := ctx.Value(middleware.UserIDKey).(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_USER_ID",
			h.translator.TWithContext(ctx, "error.invalid_user_id"))
		return
	}

	// Get user's subscriptions
	subscriptions, err := h.subscriptionService.GetUserSubscriptions(ctx, userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get user subscriptions")
		httpPkg.RespondError(w, http.StatusInternalServerError, "FETCH_FAILED",
			h.translator.TWithContext(ctx, "error.internal"))
		return
	}

	// Find active subscription
	var activeSubscription *dtos.SubscriptionResponseDTO
	for _, sub := range subscriptions {
		if sub.IsActive() {
			activeSubscription = h.buildSubscriptionResponse(sub)
			break
		}
	}

	if activeSubscription == nil {
		httpPkg.RespondError(w, http.StatusNotFound, "SUBSCRIPTION_NOT_FOUND",
			h.translator.TWithContext(ctx, "subscription.not_found"))
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, activeSubscription)
}

// GetSubscription handles GET /api/v1/subscriptions/:id
func (h *SubscriptionHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get subscription ID from URL
	subscriptionIDStr := chi.URLParam(r, "id")
	subscriptionID, err := uuid.Parse(subscriptionIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_ID",
			h.translator.TWithContext(ctx, "error.bad_request"))
		return
	}

	// Get subscription
	subscription, err := h.subscriptionService.GetSubscription(ctx, subscriptionID)
	if err != nil {
		httpPkg.RespondError(w, http.StatusNotFound, "SUBSCRIPTION_NOT_FOUND",
			err.Error())
		return
	}

	// Verify ownership
	userIDStr := ctx.Value(middleware.UserIDKey).(string)
	userID, _ := uuid.Parse(userIDStr)
	if subscription.UserID != userID {
		httpPkg.RespondError(w, http.StatusForbidden, "FORBIDDEN",
			h.translator.TWithContext(ctx, "error.forbidden"))
		return
	}

	response := h.buildSubscriptionResponse(subscription)
	httpPkg.RespondJSON(w, http.StatusOK, response)
}

// UpgradeSubscription handles PUT /api/v1/subscriptions/:id/upgrade
func (h *SubscriptionHandler) UpgradeSubscription(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get subscription ID from URL
	subscriptionIDStr := chi.URLParam(r, "id")
	subscriptionID, err := uuid.Parse(subscriptionIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_ID",
			h.translator.TWithContext(ctx, "error.bad_request"))
		return
	}

	// Parse request body
	var req dtos.UpgradeSubscriptionDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST",
			h.translator.TWithContext(ctx, "error.invalid_request_body"))
		return
	}

	// Verify ownership
	subscription, err := h.subscriptionService.GetSubscription(ctx, subscriptionID)
	if err != nil {
		httpPkg.RespondError(w, http.StatusNotFound, "SUBSCRIPTION_NOT_FOUND",
			err.Error())
		return
	}

	userIDStr := ctx.Value(middleware.UserIDKey).(string)
	userID, _ := uuid.Parse(userIDStr)
	if subscription.UserID != userID {
		httpPkg.RespondError(w, http.StatusForbidden, "FORBIDDEN",
			h.translator.TWithContext(ctx, "error.forbidden"))
		return
	}

	// Upgrade subscription
	updatedSubscription, err := h.subscriptionService.UpgradeSubscription(ctx, subscriptionID, req.NewTierID, req.Immediate)
	if err != nil {
		h.logger.WithError(err).Error("Failed to upgrade subscription")
		httpPkg.RespondError(w, http.StatusBadRequest, "UPGRADE_FAILED",
			err.Error())
		return
	}

	response := h.buildSubscriptionResponse(updatedSubscription)
	httpPkg.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"message":      h.translator.TWithContext(ctx, "subscription.upgraded"),
		"subscription": response,
	})
}

// DowngradeSubscription handles PUT /api/v1/subscriptions/:id/downgrade
func (h *SubscriptionHandler) DowngradeSubscription(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get subscription ID from URL
	subscriptionIDStr := chi.URLParam(r, "id")
	subscriptionID, err := uuid.Parse(subscriptionIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_ID",
			h.translator.TWithContext(ctx, "error.bad_request"))
		return
	}

	// Parse request body
	var req dtos.DowngradeSubscriptionDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST",
			h.translator.TWithContext(ctx, "error.invalid_request_body"))
		return
	}

	// Verify ownership
	subscription, err := h.subscriptionService.GetSubscription(ctx, subscriptionID)
	if err != nil {
		httpPkg.RespondError(w, http.StatusNotFound, "SUBSCRIPTION_NOT_FOUND",
			err.Error())
		return
	}

	userIDStr := ctx.Value(middleware.UserIDKey).(string)
	userID, _ := uuid.Parse(userIDStr)
	if subscription.UserID != userID {
		httpPkg.RespondError(w, http.StatusForbidden, "FORBIDDEN",
			h.translator.TWithContext(ctx, "error.forbidden"))
		return
	}

	// Downgrade subscription
	updatedSubscription, err := h.subscriptionService.DowngradeSubscription(ctx, subscriptionID, req.NewTierID, req.EffectiveNextCycle)
	if err != nil {
		h.logger.WithError(err).Error("Failed to downgrade subscription")
		httpPkg.RespondError(w, http.StatusBadRequest, "DOWNGRADE_FAILED",
			err.Error())
		return
	}

	response := h.buildSubscriptionResponse(updatedSubscription)
	httpPkg.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"message":      h.translator.TWithContext(ctx, "subscription.downgraded"),
		"subscription": response,
	})
}

// PauseSubscription handles PUT /api/v1/subscriptions/:id/pause
func (h *SubscriptionHandler) PauseSubscription(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get subscription ID from URL
	subscriptionIDStr := chi.URLParam(r, "id")
	subscriptionID, err := uuid.Parse(subscriptionIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_ID",
			h.translator.TWithContext(ctx, "error.bad_request"))
		return
	}

	// Parse request body
	var req dtos.PauseSubscriptionDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Body is optional for pause
	}

	// Verify ownership
	subscription, err := h.subscriptionService.GetSubscription(ctx, subscriptionID)
	if err != nil {
		httpPkg.RespondError(w, http.StatusNotFound, "SUBSCRIPTION_NOT_FOUND",
			err.Error())
		return
	}

	userIDStr := ctx.Value(middleware.UserIDKey).(string)
	userID, _ := uuid.Parse(userIDStr)
	if subscription.UserID != userID {
		httpPkg.RespondError(w, http.StatusForbidden, "FORBIDDEN",
			h.translator.TWithContext(ctx, "error.forbidden"))
		return
	}

	// Pause subscription
	if err := h.subscriptionService.PauseSubscription(ctx, subscriptionID, req.PausedUntil); err != nil {
		h.logger.WithError(err).Error("Failed to pause subscription")
		httpPkg.RespondError(w, http.StatusBadRequest, "PAUSE_FAILED",
			err.Error())
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, map[string]string{
		"message": h.translator.TWithContext(ctx, "subscription.paused"),
	})
}

// ResumeSubscription handles PUT /api/v1/subscriptions/:id/resume
func (h *SubscriptionHandler) ResumeSubscription(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get subscription ID from URL
	subscriptionIDStr := chi.URLParam(r, "id")
	subscriptionID, err := uuid.Parse(subscriptionIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_ID",
			h.translator.TWithContext(ctx, "error.bad_request"))
		return
	}

	// Verify ownership
	subscription, err := h.subscriptionService.GetSubscription(ctx, subscriptionID)
	if err != nil {
		httpPkg.RespondError(w, http.StatusNotFound, "SUBSCRIPTION_NOT_FOUND",
			err.Error())
		return
	}

	userIDStr := ctx.Value(middleware.UserIDKey).(string)
	userID, _ := uuid.Parse(userIDStr)
	if subscription.UserID != userID {
		httpPkg.RespondError(w, http.StatusForbidden, "FORBIDDEN",
			h.translator.TWithContext(ctx, "error.forbidden"))
		return
	}

	// Resume subscription
	if err := h.subscriptionService.ResumeSubscription(ctx, subscriptionID); err != nil {
		h.logger.WithError(err).Error("Failed to resume subscription")
		httpPkg.RespondError(w, http.StatusBadRequest, "RESUME_FAILED",
			err.Error())
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, map[string]string{
		"message": h.translator.TWithContext(ctx, "subscription.resumed"),
	})
}

// CancelSubscription handles DELETE /api/v1/subscriptions/:id
func (h *SubscriptionHandler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get subscription ID from URL
	subscriptionIDStr := chi.URLParam(r, "id")
	subscriptionID, err := uuid.Parse(subscriptionIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_ID",
			h.translator.TWithContext(ctx, "error.bad_request"))
		return
	}

	// Parse request body
	var req dtos.CancelSubscriptionDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST",
			h.translator.TWithContext(ctx, "error.invalid_request_body"))
		return
	}

	// Verify ownership
	subscription, err := h.subscriptionService.GetSubscription(ctx, subscriptionID)
	if err != nil {
		httpPkg.RespondError(w, http.StatusNotFound, "SUBSCRIPTION_NOT_FOUND",
			err.Error())
		return
	}

	userIDStr := ctx.Value(middleware.UserIDKey).(string)
	userID, _ := uuid.Parse(userIDStr)
	if subscription.UserID != userID {
		httpPkg.RespondError(w, http.StatusForbidden, "FORBIDDEN",
			h.translator.TWithContext(ctx, "error.forbidden"))
		return
	}

	// Cancel subscription
	if err := h.subscriptionService.CancelSubscription(ctx, subscriptionID, req.Reason, req.CancelImmediately); err != nil {
		h.logger.WithError(err).Error("Failed to cancel subscription")
		httpPkg.RespondError(w, http.StatusBadRequest, "CANCELLATION_FAILED",
			err.Error())
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, map[string]string{
		"message": h.translator.TWithContext(ctx, "subscription.cancelled"),
	})
}

// GetSubscriptionTiers handles GET /api/v1/subscription-tiers
func (h *SubscriptionHandler) GetSubscriptionTiers(w http.ResponseWriter, r *http.Request) {
	_ = r.Context()

	// This would fetch from tier repository - for now return mock data
	tiers := []dtos.TierDTO{
		{
			ID:       uuid.New(),
			Name:     "Basic",
			Code:     "basic",
			Price:    decimal.NewFromFloat(9.99),
			Currency: "USD",
			Features: map[string]interface{}{
				"support":      "email",
				"api_access":   true,
				"max_projects": 3,
			},
			Limits: map[string]int{
				"api_calls":  1000,
				"storage_gb": 10,
			},
			TrialDays: 7,
		},
		{
			ID:       uuid.New(),
			Name:     "Professional",
			Code:     "professional",
			Price:    decimal.NewFromFloat(29.99),
			Currency: "USD",
			Features: map[string]interface{}{
				"support":      "priority",
				"api_access":   true,
				"max_projects": 10,
				"analytics":    true,
			},
			Limits: map[string]int{
				"api_calls":  10000,
				"storage_gb": 100,
			},
			TrialDays: 14,
		},
		{
			ID:       uuid.New(),
			Name:     "Premium",
			Code:     "premium",
			Price:    decimal.NewFromFloat(99.99),
			Currency: "USD",
			Features: map[string]interface{}{
				"support":       "24/7",
				"api_access":    true,
				"max_projects":  -1,
				"analytics":     true,
				"custom_domain": true,
			},
			Limits: map[string]int{
				"api_calls":  -1,
				"storage_gb": 1000,
			},
			TrialDays: 30,
		},
	}

	httpPkg.RespondJSON(w, http.StatusOK, tiers)
}

// GetUserSubscriptions handles GET /api/v1/subscriptions
func (h *SubscriptionHandler) GetUserSubscriptions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userIDStr := ctx.Value(middleware.UserIDKey).(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_USER_ID",
			h.translator.TWithContext(ctx, "error.invalid_user_id"))
		return
	}

	subscriptions, err := h.subscriptionService.GetUserSubscriptions(ctx, userID)
	if err != nil {
		httpPkg.RespondError(w, http.StatusInternalServerError, "FETCH_ERROR",
			h.translator.TWithContext(ctx, "error.internal"))
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, subscriptions)
}

// GetUsage handles GET /api/v1/subscriptions/{id}/usage
func (h *SubscriptionHandler) GetUsage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	subscriptionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_ID",
			h.translator.TWithContext(ctx, "error.invalid_id"))
		return
	}

	// Get subscription to check ownership
	subscription, err := h.subscriptionService.GetSubscription(ctx, subscriptionID)
	if err != nil {
		httpPkg.RespondError(w, http.StatusNotFound, "NOT_FOUND",
			h.translator.TWithContext(ctx, "subscription.not_found"))
		return
	}

	// Check if user owns the subscription
	userIDStr := ctx.Value(middleware.UserIDKey).(string)
	userID, _ := uuid.Parse(userIDStr)
	if subscription.UserID != userID {
		httpPkg.RespondError(w, http.StatusForbidden, "FORBIDDEN",
			h.translator.TWithContext(ctx, "error.forbidden"))
		return
	}

	// Return usage data
	usage := map[string]interface{}{
		"limits":  subscription.UsageLimits,
		"current": subscription.CurrentUsage,
	}

	httpPkg.RespondJSON(w, http.StatusOK, usage)
}

// UpdateUsage handles POST /api/v1/subscriptions/{id}/usage
func (h *SubscriptionHandler) UpdateUsage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	subscriptionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_ID",
			h.translator.TWithContext(ctx, "error.invalid_id"))
		return
	}

	// Parse request body
	var req struct {
		Resource string `json:"resource"`
		Usage    int    `json:"usage"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST",
			h.translator.TWithContext(ctx, "error.invalid_request_body"))
		return
	}

	// Enforce usage limits
	if err := h.subscriptionService.EnforceUsageLimits(ctx, subscriptionID, req.Resource, req.Usage); err != nil {
		httpPkg.RespondError(w, http.StatusForbidden, "USAGE_LIMIT_EXCEEDED",
			err.Error())
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, map[string]string{
		"message": h.translator.TWithContext(ctx, "subscription.usage_updated"),
	})
}

// GetAvailableTiers handles GET /api/v1/subscription-tiers
func (h *SubscriptionHandler) GetAvailableTiers(w http.ResponseWriter, r *http.Request) {
	// This is the same as GetSubscriptionTiers
	h.GetSubscriptionTiers(w, r)
}

// GetTierDetails handles GET /api/v1/subscription-tiers/{code}
func (h *SubscriptionHandler) GetTierDetails(w http.ResponseWriter, r *http.Request) {
	_ = r.Context()
	code := chi.URLParam(r, "code")

	// Mock tier details
	tier := dtos.TierDTO{
		ID:       uuid.New(),
		Name:     "Professional",
		Code:     code,
		Price:    decimal.NewFromFloat(25.00),
		Currency: "USD",
		Features: map[string]interface{}{
			"analytics":  true,
			"support":    "priority",
			"api_access": true,
		},
		Limits: map[string]int{
			"users":    50,
			"projects": 20,
			"storage":  100,
		},
	}

	httpPkg.RespondJSON(w, http.StatusOK, tier)
}

// ExtendTrial handles POST /api/v1/subscriptions/:id/extend-trial
func (h *SubscriptionHandler) ExtendTrial(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get subscription ID from URL
	subscriptionIDStr := chi.URLParam(r, "id")
	subscriptionID, err := uuid.Parse(subscriptionIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_ID",
			h.translator.TWithContext(ctx, "error.bad_request"))
		return
	}

	// Parse request body
	var req dtos.ExtendTrialDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST",
			h.translator.TWithContext(ctx, "error.invalid_request_body"))
		return
	}

	// Only admin can extend trials - check for admin role
	// For now, we'll just check ownership
	subscription, err := h.subscriptionService.GetSubscription(ctx, subscriptionID)
	if err != nil {
		httpPkg.RespondError(w, http.StatusNotFound, "SUBSCRIPTION_NOT_FOUND",
			err.Error())
		return
	}

	userIDStr := ctx.Value(middleware.UserIDKey).(string)
	userID, _ := uuid.Parse(userIDStr)
	if subscription.UserID != userID {
		httpPkg.RespondError(w, http.StatusForbidden, "FORBIDDEN",
			h.translator.TWithContext(ctx, "error.forbidden"))
		return
	}

	// Extend trial
	if err := h.subscriptionService.ExtendTrial(ctx, subscriptionID, req.ExtensionDays, req.Reason); err != nil {
		h.logger.WithError(err).Error("Failed to extend trial")
		httpPkg.RespondError(w, http.StatusBadRequest, "TRIAL_EXTENSION_FAILED",
			err.Error())
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, map[string]string{
		"message": h.translator.TWithContext(ctx, "subscription.trial_extended", req.ExtensionDays),
	})
}

// buildSubscriptionResponse builds a subscription response DTO
func (h *SubscriptionHandler) buildSubscriptionResponse(subscription *entities.Subscription) *dtos.SubscriptionResponseDTO {
	return &dtos.SubscriptionResponseDTO{
		ID: subscription.ID,
		User: dtos.UserSummaryDTO{
			ID: subscription.UserID,
		},
		Tier: dtos.TierDTO{
			Name:     subscription.PlanName,
			Code:     string(subscription.Tier),
			Price:    subscription.Price.Amount,
			Currency: subscription.Price.Currency,
			Limits:   subscription.UsageLimits,
		},
		Status: string(subscription.Status),
		CurrentPeriod: dtos.PeriodDTO{
			StartDate: subscription.CurrentPeriodStart,
			EndDate:   subscription.CurrentPeriodEnd,
			DaysLeft:  subscription.DaysUntilRenewal(),
		},
		NextBilling:  subscription.NextBillingDate,
		TrialEndsAt:  subscription.TrialEndsAt,
		PausedAt:     subscription.PausedAt,
		CanceledAt:   subscription.CanceledAt,
		UsageLimits:  subscription.UsageLimits,
		CurrentUsage: subscription.CurrentUsage,
		CreatedAt:    subscription.CreatedAt,
		UpdatedAt:    subscription.UpdatedAt,
	}
}

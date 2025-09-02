package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	httpPkg "github.com/paradoxe35/torra/packages/http"
	"github.com/paradoxe35/torra/packages/i18n"
	"github.com/paradoxe35/torra/packages/logger"
	"github.com/paradoxe35/torra/services/core/internal/adapters/http/middleware"
	"github.com/paradoxe35/torra/services/core/internal/application/dtos"
	"github.com/paradoxe35/torra/services/core/internal/application/services"
	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
)

// BillingHandler handles billing-related HTTP requests
type BillingHandler struct {
	billingService      *services.BillingService
	subscriptionService *services.SubscriptionService
	logger              logger.Interface
	translator          *i18n.Translator
}

// NewBillingHandler creates a new billing handler
func NewBillingHandler(
	billingService *services.BillingService,
	subscriptionService *services.SubscriptionService,
	logger logger.Interface,
	translator *i18n.Translator,
) *BillingHandler {
	return &BillingHandler{
		billingService:      billingService,
		subscriptionService: subscriptionService,
		logger:              logger,
		translator:          translator,
	}
}

// GetPaymentHistory handles GET /api/v1/billing/history
func (h *BillingHandler) GetPaymentHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userIDStr := ctx.Value(middleware.UserIDKey).(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_USER_ID",
			h.translator.TWithContext(ctx, "error.invalid_user_id"))
		return
	}

	// Get pagination params
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Get user's active subscription
	subscriptions, err := h.subscriptionService.GetUserSubscriptions(ctx, userID)
	if err != nil || len(subscriptions) == 0 {
		httpPkg.RespondJSON(w, http.StatusOK, dtos.PaymentHistoryDTO{
			Payments: []dtos.PaymentTransactionDTO{},
			Total:    0,
			Page:     page,
			PageSize: pageSize,
		})
		return
	}

	// Get payment history for the first subscription
	// In a real implementation, you might want to aggregate across all subscriptions
	transactions, err := h.billingService.GetPaymentHistory(ctx, subscriptions[0].ID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get payment history")
		httpPkg.RespondError(w, http.StatusInternalServerError, "FETCH_FAILED",
			h.translator.TWithContext(ctx, "error.internal"))
		return
	}

	// Convert to DTOs
	paymentDTOs := make([]dtos.PaymentTransactionDTO, len(transactions))
	for i, tx := range transactions {
		paymentDTOs[i] = dtos.PaymentTransactionDTO{
			ID:                   tx.ID,
			SubscriptionID:       tx.SubscriptionID,
			Amount:               tx.Amount,
			Currency:             tx.Currency,
			Type:                 tx.Type,
			Status:               tx.Status,
			GatewayTransactionID: tx.GatewayTransactionID,
			ProcessedAt:          tx.ProcessedAt,
			FailureReason:        tx.FailureReason,
			CreatedAt:            tx.CreatedAt,
		}
	}

	response := dtos.PaymentHistoryDTO{
		Payments: paymentDTOs,
		Total:    int64(len(paymentDTOs)),
		Page:     page,
		PageSize: pageSize,
	}

	httpPkg.RespondJSON(w, http.StatusOK, response)
}

// GetInvoice handles GET /api/v1/billing/invoices/:id
func (h *BillingHandler) GetInvoice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get invoice ID from URL
	invoiceIDStr := chi.URLParam(r, "id")
	invoiceID, err := uuid.Parse(invoiceIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_ID",
			h.translator.TWithContext(ctx, "error.bad_request"))
		return
	}

	// Generate invoice (this would typically fetch from storage)
	invoice, err := h.billingService.GenerateInvoice(ctx, invoiceID)
	if err != nil {
		httpPkg.RespondError(w, http.StatusNotFound, "INVOICE_NOT_FOUND",
			h.translator.TWithContext(ctx, "billing.invoice_not_found"))
		return
	}

	// Verify ownership
	userIDStr := ctx.Value(middleware.UserIDKey).(string)
	userID, _ := uuid.Parse(userIDStr)

	// Get subscription to verify ownership
	subscription, err := h.subscriptionService.GetSubscription(ctx, invoice.SubscriptionID)
	if err != nil || subscription.UserID != userID {
		httpPkg.RespondError(w, http.StatusForbidden, "FORBIDDEN",
			h.translator.TWithContext(ctx, "error.forbidden"))
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, invoice)
}

// AddPaymentMethod handles POST /api/v1/billing/payment-methods
func (h *BillingHandler) AddPaymentMethod(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userIDStr := ctx.Value(middleware.UserIDKey).(string)
	_, err := uuid.Parse(userIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_USER_ID",
			h.translator.TWithContext(ctx, "error.invalid_user_id"))
		return
	}

	// Parse request body
	var req dtos.AddPaymentMethodDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST",
			h.translator.TWithContext(ctx, "error.invalid_request_body"))
		return
	}

	// In a real implementation, this would interact with the payment gateway
	// to add the payment method to the customer's account

	// For now, just return success
	httpPkg.RespondJSON(w, http.StatusCreated, map[string]interface{}{
		"message": h.translator.TWithContext(ctx, "payment.method_added"),
		"payment_method": dtos.PaymentMethodDTO{
			ID:        uuid.New().String(),
			Type:      req.Type,
			IsDefault: req.SetDefault,
		},
	})
}

// RemovePaymentMethod handles DELETE /api/v1/billing/payment-methods/:id
func (h *BillingHandler) RemovePaymentMethod(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get payment method ID from URL
	_ = chi.URLParam(r, "id")

	// Get user ID from context
	userIDStr := ctx.Value(middleware.UserIDKey).(string)
	_, err := uuid.Parse(userIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_USER_ID",
			h.translator.TWithContext(ctx, "error.invalid_user_id"))
		return
	}

	// In a real implementation, this would:
	// 1. Verify the payment method belongs to the user
	// 2. Check if it's not the last payment method
	// 3. Remove it from the payment gateway

	httpPkg.RespondJSON(w, http.StatusOK, map[string]string{
		"message": h.translator.TWithContext(ctx, "payment.method_removed"),
	})
}

// HandlePaymentWebhook handles POST /webhooks/payment
func (h *BillingHandler) HandlePaymentWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Read webhook payload
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.WithError(err).Error("Failed to read webhook payload")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get signature from header
	signature := r.Header.Get("Stripe-Signature") // Or whatever header your gateway uses

	// Process webhook
	if err := h.billingService.HandlePaymentWebhook(ctx, payload, signature); err != nil {
		h.logger.WithError(err).Error("Failed to process webhook")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Return success
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// GetUpcomingInvoice handles GET /api/v1/billing/upcoming-invoice
func (h *BillingHandler) GetUpcomingInvoice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userIDStr := ctx.Value(middleware.UserIDKey).(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_USER_ID",
			h.translator.TWithContext(ctx, "error.invalid_user_id"))
		return
	}

	// Get user's active subscription
	subscriptions, err := h.subscriptionService.GetUserSubscriptions(ctx, userID)
	if err != nil || len(subscriptions) == 0 {
		httpPkg.RespondError(w, http.StatusNotFound, "SUBSCRIPTION_NOT_FOUND",
			h.translator.TWithContext(ctx, "subscription.not_found"))
		return
	}

	var activeSubscription *entities.Subscription
	for _, sub := range subscriptions {
		if sub.IsActive() {
			activeSubscription = sub
			break
		}
	}

	if activeSubscription == nil {
		httpPkg.RespondError(w, http.StatusNotFound, "SUBSCRIPTION_NOT_FOUND",
			h.translator.TWithContext(ctx, "subscription.not_found"))
		return
	}

	// Get upcoming invoice preview
	upcomingInvoice, err := h.billingService.GetUpcomingInvoice(ctx, activeSubscription.ID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get upcoming invoice")
		httpPkg.RespondError(w, http.StatusInternalServerError, "FETCH_FAILED",
			h.translator.TWithContext(ctx, "error.internal"))
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, upcomingInvoice)
}

// ProcessPayment handles POST /api/v1/billing/process-payment
func (h *BillingHandler) ProcessPayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req dtos.ProcessPaymentDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST",
			h.translator.TWithContext(ctx, "error.invalid_request_body"))
		return
	}

	// Verify ownership of subscription
	subscription, err := h.subscriptionService.GetSubscription(ctx, req.SubscriptionID)
	if err != nil {
		httpPkg.RespondError(w, http.StatusNotFound, "SUBSCRIPTION_NOT_FOUND",
			h.translator.TWithContext(ctx, "subscription.not_found"))
		return
	}

	userIDStr := ctx.Value(middleware.UserIDKey).(string)
	userID, _ := uuid.Parse(userIDStr)
	if subscription.UserID != userID {
		httpPkg.RespondError(w, http.StatusForbidden, "FORBIDDEN",
			h.translator.TWithContext(ctx, "error.forbidden"))
		return
	}

	// In a real implementation, this would process the payment
	// For now, return success
	httpPkg.RespondJSON(w, http.StatusOK, map[string]string{
		"message": h.translator.TWithContext(ctx, "payment.success"),
	})
}

// RefundPayment handles POST /api/v1/billing/refund
func (h *BillingHandler) RefundPayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req dtos.RefundDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST",
			h.translator.TWithContext(ctx, "error.invalid_request_body"))
		return
	}

	// Process refund
	if err := h.billingService.RefundPayment(ctx, req.TransactionID, req.Amount, req.Reason); err != nil {
		h.logger.WithError(err).Error("Failed to process refund")
		httpPkg.RespondError(w, http.StatusBadRequest, "REFUND_FAILED",
			err.Error())
		return
	}

	httpPkg.RespondJSON(w, http.StatusOK, map[string]string{
		"message": h.translator.TWithContext(ctx, "payment.refund_success"),
	})
}

// GetBillingSettings handles GET /api/v1/billing/settings
func (h *BillingHandler) GetBillingSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userIDStr := ctx.Value(middleware.UserIDKey).(string)
	_, err := uuid.Parse(userIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_USER_ID",
			h.translator.TWithContext(ctx, "error.invalid_user_id"))
		return
	}

	// In a real implementation, this would fetch the user's billing settings
	// For now, return mock data
	settings := dtos.BillingSettingsDTO{
		AutoRenew:            true,
		PaymentMethodID:      "pm_default",
		BillingEmail:         "user@example.com",
		SendInvoiceReminders: true,
		ReminderDaysBefore:   []int{7, 3, 1},
	}

	httpPkg.RespondJSON(w, http.StatusOK, settings)
}

// UpdateBillingSettings handles PUT /api/v1/billing/settings
func (h *BillingHandler) UpdateBillingSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userIDStr := ctx.Value(middleware.UserIDKey).(string)
	_, err := uuid.Parse(userIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_USER_ID",
			h.translator.TWithContext(ctx, "error.invalid_user_id"))
		return
	}

	// Parse request body
	var req dtos.UpdateBillingSettingsDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST",
			h.translator.TWithContext(ctx, "error.invalid_request_body"))
		return
	}

	// In a real implementation, this would update the user's billing settings
	// For now, just return success
	httpPkg.RespondJSON(w, http.StatusOK, map[string]string{
		"message": h.translator.TWithContext(ctx, "billing.settings_updated"),
	})
}

// HandleStripeWebhook handles Stripe webhook events
func (h *BillingHandler) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	_ = r.Context()

	// In a real implementation, this would:
	// 1. Verify the webhook signature
	// 2. Parse the webhook payload
	// 3. Process the event (payment success, failure, etc.)
	// 4. Update the subscription status accordingly

	httpPkg.RespondJSON(w, http.StatusOK, map[string]string{
		"status": "received",
	})
}

// GetBillingCycles gets billing cycles for the user
func (h *BillingHandler) GetBillingCycles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userIDStr := ctx.Value(middleware.UserIDKey).(string)
	_, err := uuid.Parse(userIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_USER_ID",
			h.translator.TWithContext(ctx, "error.invalid_user_id"))
		return
	}

	// Return empty list for now
	httpPkg.RespondJSON(w, http.StatusOK, []interface{}{})
}

// GetBillingCycle gets a specific billing cycle
func (h *BillingHandler) GetBillingCycle(w http.ResponseWriter, r *http.Request) {
	_ = r.Context()
	_ = chi.URLParam(r, "id")

	httpPkg.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"id":     chi.URLParam(r, "id"),
		"status": "paid",
	})
}

// GetInvoices gets invoices for the user
func (h *BillingHandler) GetInvoices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userIDStr := ctx.Value(middleware.UserIDKey).(string)
	_, err := uuid.Parse(userIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_USER_ID",
			h.translator.TWithContext(ctx, "error.invalid_user_id"))
		return
	}

	// Return empty list for now
	httpPkg.RespondJSON(w, http.StatusOK, []interface{}{})
}

// DownloadInvoice downloads an invoice PDF
func (h *BillingHandler) DownloadInvoice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = chi.URLParam(r, "id")

	// In a real implementation, this would generate and return a PDF
	httpPkg.RespondError(w, http.StatusNotImplemented, "NOT_IMPLEMENTED",
		h.translator.TWithContext(ctx, "error.not_implemented"))
}

// GetPaymentMethods gets payment methods for the user
func (h *BillingHandler) GetPaymentMethods(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userIDStr := ctx.Value(middleware.UserIDKey).(string)
	_, err := uuid.Parse(userIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_USER_ID",
			h.translator.TWithContext(ctx, "error.invalid_user_id"))
		return
	}

	// Return empty list for now
	httpPkg.RespondJSON(w, http.StatusOK, []interface{}{})
}

// DeletePaymentMethod deletes a payment method
func (h *BillingHandler) DeletePaymentMethod(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = chi.URLParam(r, "id")

	// Get user ID from context
	userIDStr := ctx.Value(middleware.UserIDKey).(string)
	_, err := uuid.Parse(userIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_USER_ID",
			h.translator.TWithContext(ctx, "error.invalid_user_id"))
		return
	}

	// In a real implementation, this would delete the payment method
	httpPkg.RespondJSON(w, http.StatusOK, map[string]string{
		"message": h.translator.TWithContext(ctx, "payment.method_deleted"),
	})
}

// SetDefaultPaymentMethod sets a payment method as default
func (h *BillingHandler) SetDefaultPaymentMethod(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = chi.URLParam(r, "id")

	// Get user ID from context
	userIDStr := ctx.Value(middleware.UserIDKey).(string)
	_, err := uuid.Parse(userIDStr)
	if err != nil {
		httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_USER_ID",
			h.translator.TWithContext(ctx, "error.invalid_user_id"))
		return
	}

	// In a real implementation, this would set the payment method as default
	httpPkg.RespondJSON(w, http.StatusOK, map[string]string{
		"message": h.translator.TWithContext(ctx, "payment.method_set_default"),
	})
}

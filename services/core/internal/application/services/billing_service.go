package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/paradoxe35/mytorra/packages/events"
	"github.com/paradoxe35/mytorra/packages/i18n"
	"github.com/paradoxe35/mytorra/packages/logger"
	"github.com/paradoxe35/mytorra/services/core/internal/application/dtos"
	"github.com/paradoxe35/mytorra/services/core/internal/application/ports"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/entities"
	domainEvents "github.com/paradoxe35/mytorra/services/core/internal/domain/events"
)

// BillingService handles billing and payment processing
type BillingService struct {
	subscriptionRepo    ports.SubscriptionRepository
	billingRepo         ports.BillingCycleRepository
	paymentRepo         ports.PaymentTransactionRepository
	paymentGateway      ports.PaymentGateway
	notificationService ports.NotificationService
	eventBus            events.EventPublisher
	logger              logger.Interface
	translator          *i18n.Translator
}

// NewBillingService creates a new billing service
func NewBillingService(
	subscriptionRepo ports.SubscriptionRepository,
	billingRepo ports.BillingCycleRepository,
	paymentRepo ports.PaymentTransactionRepository,
	paymentGateway ports.PaymentGateway,
	notificationService ports.NotificationService,
	eventBus events.EventPublisher,
	logger logger.Interface,
	translator *i18n.Translator,
) *BillingService {
	return &BillingService{
		subscriptionRepo:    subscriptionRepo,
		billingRepo:         billingRepo,
		paymentRepo:         paymentRepo,
		paymentGateway:      paymentGateway,
		notificationService: notificationService,
		eventBus:            eventBus,
		logger:              logger,
		translator:          translator,
	}
}

// ProcessRecurringBilling processes recurring billing for all active subscriptions
func (b *BillingService) ProcessRecurringBilling(ctx context.Context) error {
	// Find all pending billing cycles
	pendingCycles, err := b.billingRepo.FindPending(ctx)
	if err != nil {
		b.logger.WithError(err).Error("Failed to find pending billing cycles")
		return err
	}

	successCount := 0
	failureCount := 0

	for _, cycle := range pendingCycles {
		if err := b.processBillingCycle(ctx, cycle); err != nil {
			b.logger.WithError(err).WithField("billing_cycle_id", cycle.ID).Error("Failed to process billing cycle")
			failureCount++
		} else {
			successCount++
		}
	}

	b.logger.WithFields(map[string]any{
		"success_count": successCount,
		"failure_count": failureCount,
		"total":         len(pendingCycles),
	}).Info("Recurring billing processing completed")

	return nil
}

// processBillingCycle processes a single billing cycle
func (b *BillingService) processBillingCycle(ctx context.Context, cycle *entities.BillingCycle) error {
	// Get subscription
	subscription, err := b.subscriptionRepo.FindByID(ctx, cycle.SubscriptionID)
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	// Skip if subscription is not active
	if !subscription.IsActive() {
		b.logger.WithField("subscription_id", subscription.ID).Info("Skipping billing for inactive subscription")
		return nil
	}

	// Skip if in trial
	if subscription.IsInTrial() {
		b.logger.WithField("subscription_id", subscription.ID).Info("Skipping billing for trial subscription")
		return nil
	}

	// Create payment transaction
	transaction := &entities.PaymentTransaction{
		ID:             uuid.New(),
		SubscriptionID: subscription.ID,
		BillingCycleID: &cycle.ID,
		Amount:         cycle.Amount,
		Currency:       cycle.Currency,
		Type:           string(entities.PaymentTransactionTypeSubscription),
		Status:         string(entities.PaymentTransactionStatusPending),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := b.paymentRepo.Create(ctx, transaction); err != nil {
		return fmt.Errorf("failed to create payment transaction: %w", err)
	}

	// Process payment through gateway
	if subscription.ProviderCustomerID != nil && *subscription.ProviderCustomerID != "" {
		paymentReq := ports.ChargePaymentRequest{
			CustomerID:  *subscription.ProviderCustomerID,
			Amount:      cycle.Amount,
			Currency:    cycle.Currency,
			Description: fmt.Sprintf("Subscription payment for %s", subscription.PlanName),
			ReferenceID: cycle.ID.String(),
		}

		paymentResp, err := b.paymentGateway.ChargePayment(ctx, paymentReq)
		if err != nil {
			b.handlePaymentFailure(ctx, transaction, cycle, subscription, err.Error())
			return fmt.Errorf("payment failed: %w", err)
		}

		// Check payment status
		if paymentResp.Status != "succeeded" {
			b.handlePaymentFailure(ctx, transaction, cycle, subscription, "Payment not successful")
			return fmt.Errorf("payment not successful: %s", paymentResp.Status)
		}

		// Mark transaction as successful
		if err := transaction.MarkAsSuccessful(paymentResp.TransactionID); err != nil {
			return err
		}

		// Mark billing cycle as paid
		if err := cycle.MarkAsPaid(paymentResp.TransactionID); err != nil {
			return err
		}

		// Update entities
		if err := b.paymentRepo.Update(ctx, transaction); err != nil {
			return err
		}

		if err := b.billingRepo.Update(ctx, cycle); err != nil {
			return err
		}

		// Create next billing cycle
		b.createNextBillingCycle(ctx, subscription)

		// Publish success event
		b.eventBus.Publish(ctx, domainEvents.BillingPaymentSuccessful{
			SubscriptionID: subscription.ID,
			BillingCycleID: cycle.ID,
			TransactionID:  transaction.ID,
			Amount:         cycle.Amount,
			Currency:       cycle.Currency,
			ProcessedAt:    time.Now(),
		})

		b.logger.WithFields(map[string]any{
			"subscription_id":  subscription.ID,
			"billing_cycle_id": cycle.ID,
			"amount":           cycle.Amount,
		}).Info("Billing cycle processed successfully")
	}

	return nil
}

// handlePaymentFailure handles failed payment attempts
func (b *BillingService) handlePaymentFailure(ctx context.Context, transaction *entities.PaymentTransaction, cycle *entities.BillingCycle, subscription *entities.Subscription, reason string) {
	// Mark transaction as failed
	transaction.MarkAsFailed(reason)
	transaction.IncrementRetryCount()
	b.paymentRepo.Update(ctx, transaction)

	// Mark billing cycle as failed
	cycle.MarkAsFailed()
	b.billingRepo.Update(ctx, cycle)

	// Check retry count
	if transaction.CanRetry() {
		// Schedule retry
		b.schedulePaymentRetry(ctx, transaction, cycle, subscription)
	} else {
		// Max retries reached, mark subscription as past due
		subscription.Status = entities.SubscriptionStatusPastDue
		b.subscriptionRepo.Update(ctx, subscription)

		// Send notification
		b.sendPaymentFailureNotification(ctx, subscription, cycle)

		// Publish event
		b.eventBus.Publish(ctx, domainEvents.BillingPaymentFailed{
			SubscriptionID: subscription.ID,
			BillingCycleID: cycle.ID,
			TransactionID:  transaction.ID,
			Amount:         cycle.Amount,
			Currency:       cycle.Currency,
			FailureReason:  reason,
			FailedAt:       time.Now(),
		})
	}
}

// schedulePaymentRetry schedules a payment retry
func (b *BillingService) schedulePaymentRetry(ctx context.Context, transaction *entities.PaymentTransaction, cycle *entities.BillingCycle, subscription *entities.Subscription) {
	// Calculate retry time (exponential backoff)
	retryHours := 6 * (1 << transaction.RetryCount) // 6, 12, 24 hours
	retryTime := time.Now().Add(time.Duration(retryHours) * time.Hour)

	// Update cycle with next retry time
	cycle.UpdatedAt = time.Now()
	b.billingRepo.Update(ctx, cycle)

	// Publish retry event
	b.eventBus.Publish(ctx, domainEvents.BillingPaymentRetryScheduled{
		SubscriptionID: subscription.ID,
		BillingCycleID: cycle.ID,
		TransactionID:  transaction.ID,
		NextRetryAt:    retryTime,
		RetryCount:     transaction.RetryCount,
	})

	b.logger.WithFields(logger.Fields{
		"subscription_id": subscription.ID,
		"retry_at":        retryTime,
		"retry_count":     transaction.RetryCount,
	}).Info("Payment retry scheduled")
}

// RetryFailedPayments retries failed payment attempts
func (b *BillingService) RetryFailedPayments(ctx context.Context) error {
	// Find transactions eligible for retry
	failedTransactions, err := b.paymentRepo.FindFailedForRetry(ctx, 3)
	if err != nil {
		return err
	}

	for _, transaction := range failedTransactions {
		// Get billing cycle
		if transaction.BillingCycleID != nil {
			cycle, err := b.billingRepo.FindByID(ctx, *transaction.BillingCycleID)
			if err != nil {
				continue
			}

			// Retry payment
			if err := b.processBillingCycle(ctx, cycle); err != nil {
				b.logger.WithError(err).WithField("transaction_id", transaction.ID).Error("Payment retry failed")
			}
		}
	}

	return nil
}

// HandlePaymentWebhook handles payment webhooks from the gateway
func (b *BillingService) HandlePaymentWebhook(ctx context.Context, payload []byte, signature string) error {
	// Process webhook
	webhookEvent, err := b.paymentGateway.ProcessWebhook(ctx, payload, signature)
	if err != nil {
		return fmt.Errorf("failed to process webhook: %w", err)
	}

	// Use the webhook event type
	eventType := webhookEvent.Type

	// Parse webhook payload
	var webhookData map[string]interface{}
	if err := json.Unmarshal(webhookEvent.Data, &webhookData); err != nil {
		return fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	// Handle different webhook events
	switch eventType {
	case "payment_intent.succeeded":
		return b.handlePaymentSucceeded(ctx, webhookData)
	case "payment_intent.payment_failed":
		return b.handlePaymentFailed(ctx, webhookData)
	case "subscription.updated":
		return b.handleSubscriptionUpdated(ctx, webhookData)
	case "subscription.deleted":
		return b.handleSubscriptionDeleted(ctx, webhookData)
	default:
		b.logger.WithField("event_type", eventType).Info("Unhandled webhook event")
	}

	return nil
}

// handlePaymentSucceeded handles successful payment webhook
func (b *BillingService) handlePaymentSucceeded(ctx context.Context, data map[string]interface{}) error {
	// Extract metadata
	metadata, ok := data["data"].(map[string]interface{})["metadata"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("metadata not found in webhook")
	}

	transactionIDStr, ok := metadata["transaction_id"].(string)
	if !ok {
		return fmt.Errorf("transaction_id not found in metadata")
	}

	transactionID, err := uuid.Parse(transactionIDStr)
	if err != nil {
		return err
	}

	// Get transaction
	transaction, err := b.paymentRepo.FindByID(ctx, transactionID)
	if err != nil {
		return err
	}

	// Update transaction status
	if !transaction.IsSuccessful() {
		chargeID, _ := data["data"].(map[string]interface{})["charge"].(string)
		transaction.MarkAsSuccessful(chargeID)
		b.paymentRepo.Update(ctx, transaction)

		// Update billing cycle if applicable
		if transaction.BillingCycleID != nil {
			cycle, err := b.billingRepo.FindByID(ctx, *transaction.BillingCycleID)
			if err == nil {
				cycle.MarkAsPaid(chargeID)
				b.billingRepo.Update(ctx, cycle)
			}
		}

		// Log webhook success
		b.logger.WithFields(logger.Fields{
			"transaction_id": transaction.ID,
			"charge_id":      chargeID,
		}).Info("Webhook payment success processed")
	}

	return nil
}

// handlePaymentFailed handles failed payment webhook
func (b *BillingService) handlePaymentFailed(ctx context.Context, data map[string]interface{}) error {
	// Extract metadata
	metadata, ok := data["data"].(map[string]interface{})["metadata"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("metadata not found in webhook")
	}

	transactionIDStr, ok := metadata["transaction_id"].(string)
	if !ok {
		return fmt.Errorf("transaction_id not found in metadata")
	}

	transactionID, err := uuid.Parse(transactionIDStr)
	if err != nil {
		return err
	}

	// Get transaction
	transaction, err := b.paymentRepo.FindByID(ctx, transactionID)
	if err != nil {
		return err
	}

	// Update transaction status
	if !transaction.IsFailed() {
		failureMessage, _ := data["data"].(map[string]interface{})["failure_message"].(string)
		transaction.MarkAsFailed(failureMessage)
		b.paymentRepo.Update(ctx, transaction)

		// Log webhook failure
		b.logger.WithFields(logger.Fields{
			"transaction_id":  transaction.ID,
			"failure_message": failureMessage,
		}).Error("Webhook payment failure processed")
	}

	return nil
}

// handleSubscriptionUpdated handles subscription update webhook
func (b *BillingService) handleSubscriptionUpdated(ctx context.Context, data map[string]interface{}) error {
	// Implementation for subscription update webhook
	b.logger.WithField("data", data).Info("Subscription updated webhook received")
	return nil
}

// handleSubscriptionDeleted handles subscription deletion webhook
func (b *BillingService) handleSubscriptionDeleted(ctx context.Context, data map[string]interface{}) error {
	// Implementation for subscription deletion webhook
	b.logger.WithField("data", data).Info("Subscription deleted webhook received")
	return nil
}

// GenerateInvoice generates an invoice for a billing cycle
func (b *BillingService) GenerateInvoice(ctx context.Context, billingCycleID uuid.UUID) (*dtos.InvoiceDTO, error) {
	// Get billing cycle
	cycle, err := b.billingRepo.FindByID(ctx, billingCycleID)
	if err != nil {
		return nil, fmt.Errorf("%s", b.translator.TWithContext(ctx, "billing.cycle_not_found"))
	}

	// Get subscription
	subscription, err := b.subscriptionRepo.FindByID(ctx, cycle.SubscriptionID)
	if err != nil {
		return nil, fmt.Errorf("%s", b.translator.TWithContext(ctx, "subscription.not_found"))
	}

	// Generate invoice number
	invoiceNumber := fmt.Sprintf("INV-%s-%d", subscription.ID.String()[:8], time.Now().Unix())

	// Create invoice DTO
	invoice := &dtos.InvoiceDTO{
		ID:             uuid.New(),
		Number:         invoiceNumber,
		SubscriptionID: subscription.ID,
		BillingCycleID: cycle.ID,
		CustomerName:   "Customer Name",        // Should get from user service
		CustomerEmail:  "customer@example.com", // Should get from user service
		LineItems: []dtos.LineItemDTO{
			{
				Description: fmt.Sprintf("%s Subscription - %s to %s",
					subscription.PlanName,
					cycle.StartDate.Format("Jan 02, 2006"),
					cycle.EndDate.Format("Jan 02, 2006")),
				Quantity:  1,
				UnitPrice: cycle.Amount,
				Amount:    cycle.Amount,
			},
		},
		Subtotal:  cycle.Amount,
		Tax:       decimal.Zero, // Calculate tax if applicable
		Total:     cycle.Amount,
		Currency:  cycle.Currency,
		Status:    cycle.Status,
		DueDate:   cycle.EndDate,
		PaidAt:    cycle.PaidAt,
		CreatedAt: time.Now(),
	}

	// Store invoice ID in billing cycle
	cycle.InvoiceID = &invoiceNumber
	b.billingRepo.Update(ctx, cycle)

	// Publish event
	b.eventBus.Publish(ctx, domainEvents.BillingInvoiceGenerated{
		SubscriptionID: subscription.ID,
		BillingCycleID: cycle.ID,
		InvoiceID:      invoice.ID.String(),
		Amount:         invoice.Total,
		Currency:       invoice.Currency,
		DueDate:        invoice.DueDate,
		GeneratedAt:    time.Now(),
	})

	return invoice, nil
}

// RefundPayment processes a refund for a payment
func (b *BillingService) RefundPayment(ctx context.Context, transactionID uuid.UUID, amount decimal.Decimal, reason string) error {
	// Get transaction
	transaction, err := b.paymentRepo.FindByID(ctx, transactionID)
	if err != nil {
		return fmt.Errorf("%s", b.translator.TWithContext(ctx, "payment.not_found"))
	}

	if !transaction.IsSuccessful() {
		return fmt.Errorf("%s", b.translator.TWithContext(ctx, "payment.cannot_refund"))
	}

	// Process refund through gateway
	refundReq := ports.RefundPaymentRequest{
		PaymentID: transaction.GatewayTransactionID,
		Amount:    amount,
		Reason:    reason,
	}

	refundResp, err := b.paymentGateway.RefundPayment(ctx, refundReq)
	if err != nil {
		return fmt.Errorf("%s", b.translator.TWithContext(ctx, "payment.refund_failed"))
	}

	// Create refund transaction
	now := time.Now()
	refundTransaction := &entities.PaymentTransaction{
		ID:                   uuid.New(),
		SubscriptionID:       transaction.SubscriptionID,
		Amount:               amount,
		Currency:             transaction.Currency,
		Type:                 string(entities.PaymentTransactionTypeRefund),
		Status:               string(entities.PaymentTransactionStatusSucceeded),
		GatewayTransactionID: refundResp.RefundID,
		ProcessedAt:          &now,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	// Store gateway response
	refundTransaction.SetGatewayResponse(refundResp)

	if err := b.paymentRepo.Create(ctx, refundTransaction); err != nil {
		return err
	}

	// Update original transaction status
	transaction.Status = string(entities.PaymentTransactionStatusRefunded)
	b.paymentRepo.Update(ctx, transaction)

	// Publish event
	b.eventBus.Publish(ctx, domainEvents.BillingRefundProcessed{
		SubscriptionID: transaction.SubscriptionID,
		BillingCycleID: uuid.UUID{}, // TODO: Add billing cycle ID to transaction
		TransactionID:  transactionID,
		RefundID:       refundTransaction.ID.String(),
		Amount:         amount,
		Currency:       transaction.Currency,
		Reason:         reason,
		ProcessedAt:    time.Now(),
	})

	return nil
}

// createNextBillingCycle creates the next billing cycle for a subscription
func (b *BillingService) createNextBillingCycle(ctx context.Context, subscription *entities.Subscription) error {
	// Calculate next billing period
	startDate := subscription.CurrentPeriodEnd.AddDate(0, 0, 1)
	endDate := startDate.AddDate(0, 1, 0) // Monthly billing

	// Create new billing cycle
	cycle := &entities.BillingCycle{
		ID:             uuid.New(),
		SubscriptionID: subscription.ID,
		StartDate:      startDate,
		EndDate:        endDate,
		Amount:         subscription.Price.Amount,
		Currency:       subscription.Price.Currency,
		Status:         string(entities.BillingCycleStatusPending),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := b.billingRepo.Create(ctx, cycle); err != nil {
		return err
	}

	// Update subscription dates
	subscription.CurrentPeriodStart = startDate
	subscription.CurrentPeriodEnd = endDate
	nextBilling := endDate.AddDate(0, 0, 1)
	subscription.NextBillingDate = &nextBilling
	subscription.UpdatedAt = time.Now()

	return b.subscriptionRepo.Update(ctx, subscription)
}

// sendPaymentFailureNotification sends notification for payment failure
func (b *BillingService) sendPaymentFailureNotification(ctx context.Context, subscription *entities.Subscription, cycle *entities.BillingCycle) {
	// Create notification
	notification := &entities.Notification{
		ID:      uuid.New(),
		UserID:  subscription.UserID,
		Type:    "payment_failed",
		Title:   b.translator.TWithContext(ctx, "subscription.payment_failed"),
		Message: fmt.Sprintf(b.translator.TWithContext(ctx, "billing.amount_due"), cycle.Amount.String()),
		Data: map[string]interface{}{
			"subscription_id":  subscription.ID,
			"billing_cycle_id": cycle.ID,
			"amount":           cycle.Amount.String(),
			"currency":         cycle.Currency,
		},
		CreatedAt: time.Now(),
	}

	if err := b.notificationService.Send(ctx, notification); err != nil {
		b.logger.WithError(err).Error("Failed to send payment failure notification")
	}
}

// GetPaymentHistory gets payment history for a subscription
func (b *BillingService) GetPaymentHistory(ctx context.Context, subscriptionID uuid.UUID) ([]*entities.PaymentTransaction, error) {
	return b.paymentRepo.FindBySubscriptionID(ctx, subscriptionID)
}

// GetUpcomingInvoice generates a preview of the upcoming invoice
func (b *BillingService) GetUpcomingInvoice(ctx context.Context, subscriptionID uuid.UUID) (*dtos.UpcomingInvoiceDTO, error) {
	subscription, err := b.subscriptionRepo.FindByID(ctx, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("%s", b.translator.TWithContext(ctx, "subscription.not_found"))
	}

	// Get current billing cycle
	currentCycle, err := b.billingRepo.FindCurrentBySubscriptionID(ctx, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("%s", b.translator.TWithContext(ctx, "billing.cycle_not_found"))
	}

	upcoming := &dtos.UpcomingInvoiceDTO{
		SubscriptionID: subscriptionID,
		PeriodStart:    currentCycle.EndDate.AddDate(0, 0, 1),
		PeriodEnd:      currentCycle.EndDate.AddDate(0, 1, 1),
		Amount:         subscription.Price.Amount,
		Currency:       subscription.Price.Currency,
		DueDate:        currentCycle.EndDate,
		LineItems: []dtos.LineItemDTO{
			{
				Description: fmt.Sprintf("%s Subscription", subscription.PlanName),
				Quantity:    1,
				UnitPrice:   subscription.Price.Amount,
				Amount:      subscription.Price.Amount,
			},
		},
	}

	return upcoming, nil
}

package services

import (
	"context"
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
	"github.com/paradoxe35/mytorra/services/core/internal/domain/valueobjects"
)

// SubscriptionService handles subscription business logic
type SubscriptionService struct {
	subscriptionRepo ports.SubscriptionRepository
	tierRepo         ports.SubscriptionTierRepository
	billingRepo      ports.BillingCycleRepository
	paymentGateway   ports.PaymentGateway
	userRepo         ports.UserRepository
	eventBus         events.EventPublisher
	logger           logger.Interface
	translator       *i18n.Translator
}

// NewSubscriptionService creates a new subscription service
func NewSubscriptionService(
	subscriptionRepo ports.SubscriptionRepository,
	tierRepo ports.SubscriptionTierRepository,
	billingRepo ports.BillingCycleRepository,
	paymentGateway ports.PaymentGateway,
	userRepo ports.UserRepository,
	eventBus events.EventPublisher,
	logger logger.Interface,
	translator *i18n.Translator,
) *SubscriptionService {
	return &SubscriptionService{
		subscriptionRepo: subscriptionRepo,
		tierRepo:         tierRepo,
		billingRepo:      billingRepo,
		paymentGateway:   paymentGateway,
		userRepo:         userRepo,
		eventBus:         eventBus,
		logger:           logger,
		translator:       translator,
	}
}

// CreateSubscription creates a new subscription
func (s *SubscriptionService) CreateSubscription(ctx context.Context, dto dtos.CreateSubscriptionDTO) (*entities.Subscription, error) {
	// Check for existing active subscription
	existing, err := s.subscriptionRepo.FindActiveByUserID(ctx, dto.UserID)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("%s", s.translator.TWithContext(ctx, "subscription.already_exists"))
	}

	// Get the tier
	tier, err := s.tierRepo.FindByID(ctx, dto.TierID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to find tier")
		return nil, fmt.Errorf("%s", s.translator.TWithContext(ctx, "subscription.tier_not_found"))
	}

	if !tier.IsActive {
		return nil, fmt.Errorf("%s", s.translator.TWithContext(ctx, "subscription.tier_inactive"))
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, dto.UserID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to find user")
		return nil, fmt.Errorf("%s", s.translator.TWithContext(ctx, "user.not_found"))
	}

	// Create subscription entity
	subscription := &entities.Subscription{
		ID:         uuid.New(),
		UserID:     dto.UserID,
		EntityID:   dto.EntityID,
		EntityType: dto.EntityType,
		Tier:       valueobjects.SubscriptionTier(tier.Code),
		PlanName:   tier.Name,
		Price: valueobjects.Money{
			Amount:   tier.Price,
			Currency: tier.Currency,
		},
		BillingCycle: entities.BillingCycleMonthly,
		Status:       entities.SubscriptionStatusActive,
		UsageLimits:  tier.Limits,
		CurrentUsage: make(map[string]int),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Handle trial period
	if dto.StartTrial && tier.TrialDays > 0 {
		trialEnd := time.Now().AddDate(0, 0, tier.TrialDays)
		subscription.TrialEndsAt = &trialEnd
		subscription.IsTrialing = true
		subscription.Status = entities.SubscriptionStatusTrialing
	}

	// Set billing periods
	subscription.CurrentPeriodStart = time.Now()
	if subscription.IsTrialing {
		subscription.CurrentPeriodEnd = *subscription.TrialEndsAt
	} else {
		subscription.CurrentPeriodEnd = time.Now().AddDate(0, 1, 0)
	}
	nextBilling := subscription.CurrentPeriodEnd.AddDate(0, 0, 1)
	subscription.NextBillingDate = &nextBilling

	// Create customer in payment gateway if not exists
	customerID, err := s.paymentGateway.CreateCustomer(ctx, user)
	if err != nil {
		s.logger.WithError(err).Error("Failed to create customer in payment gateway")
		return nil, fmt.Errorf("%s", s.translator.TWithContext(ctx, "payment.failed"))
	}

	// Store customer ID
	subscription.ProviderCustomerID = &customerID

	// Create subscription in payment gateway if not in trial
	if !subscription.IsTrialing {
		subReq := ports.CreateSubscriptionRequest{
			CustomerID:     customerID,
			SubscriptionID: subscription.ID.String(),
			TierCode:       tier.Code,
			TierName:       tier.Name,
			Amount:         tier.Price,
			Currency:       tier.Currency,
			BillingCycle:   subscription.BillingCycle,
			TrialDays:      0,
			Metadata: map[string]string{
				"subscription_id": subscription.ID.String(),
				"user_id":         user.ID.String(),
			},
		}

		gatewaySubscription, err := s.paymentGateway.CreateSubscription(ctx, subReq)
		if err != nil {
			s.logger.WithError(err).Error("Failed to create subscription in payment gateway")
			return nil, fmt.Errorf("%s", s.translator.TWithContext(ctx, "subscription.creation_failed"))
		}

		// Store gateway subscription ID
		providerSubID := gatewaySubscription.ProviderSubscriptionID
		subscription.ProviderSubscriptionID = &providerSubID
	}

	// Save subscription
	if err := s.subscriptionRepo.Create(ctx, subscription); err != nil {
		s.logger.WithError(err).Error("Failed to save subscription")
		return nil, fmt.Errorf("%s", s.translator.TWithContext(ctx, "error.internal"))
	}

	// Create initial billing cycle
	billingCycle := &entities.BillingCycle{
		ID:             uuid.New(),
		SubscriptionID: subscription.ID,
		StartDate:      subscription.CurrentPeriodStart,
		EndDate:        subscription.CurrentPeriodEnd,
		Amount:         s.getInitialAmount(subscription.IsTrialing, tier.Price),
		Currency:       tier.Currency,
		Status:         string(entities.BillingCycleStatusPending),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.billingRepo.Create(ctx, billingCycle); err != nil {
		s.logger.WithError(err).Error("Failed to create billing cycle")
	}

	// Publish event
	s.eventBus.Publish(ctx, domainEvents.SubscriptionCreated{
		SubscriptionID: subscription.ID,
		UserID:         subscription.UserID,
		EntityID:       subscription.EntityID,
		EntityType:     subscription.EntityType,
		Tier:           subscription.Tier,
		IsTrialing:     subscription.IsTrialing,
		CreatedAt:      time.Now(),
	})

	s.logger.WithFields(logger.Fields{
		"subscription_id": subscription.ID,
		"user_id":         subscription.UserID,
		"tier":            subscription.Tier,
	}).Info("Subscription created successfully")

	return subscription, nil
}

// UpgradeSubscription upgrades a subscription to a higher tier
func (s *SubscriptionService) UpgradeSubscription(ctx context.Context, subscriptionID, newTierID uuid.UUID, immediate bool) (*entities.Subscription, error) {
	// Get subscription
	subscription, err := s.subscriptionRepo.FindByID(ctx, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("%s", s.translator.TWithContext(ctx, "subscription.not_found"))
	}

	// Check if upgrade is allowed
	if !subscription.CanUpgrade() {
		return nil, fmt.Errorf("%s", s.translator.TWithContext(ctx, "subscription.cannot_upgrade"))
	}

	// Get new tier
	newTier, err := s.tierRepo.FindByID(ctx, newTierID)
	if err != nil {
		return nil, fmt.Errorf("%s", s.translator.TWithContext(ctx, "subscription.tier_not_found"))
	}

	// Check if it's actually an upgrade
	currentTier, err := s.tierRepo.FindByCode(ctx, string(subscription.Tier))
	if err != nil {
		return nil, fmt.Errorf("%s", s.translator.TWithContext(ctx, "subscription.tier_not_found"))
	}

	if !newTier.IsUpgradeFrom(currentTier) {
		return nil, fmt.Errorf("%s", s.translator.TWithContext(ctx, "subscription.not_an_upgrade"))
	}

	// Calculate proration if immediate
	if immediate {
		_ = subscription.CalculateProration(newTier.Price)
	}

	// Update subscription
	subscription.Tier = valueobjects.SubscriptionTier(newTier.Code)
	subscription.PlanName = newTier.Name
	subscription.Price = valueobjects.Money{
		Amount:   newTier.Price,
		Currency: newTier.Currency,
	}
	subscription.UsageLimits = newTier.Limits
	subscription.UpdatedAt = time.Now()

	// Update in payment gateway
	if gatewaySubID, ok := subscription.Metadata["gateway_subscription_id"].(string); ok {
		updateReq := ports.UpdateSubscriptionRequest{
			ProviderSubscriptionID: gatewaySubID,
			ProrationBehavior:      "create_prorations",
		}

		_, err = s.paymentGateway.UpdateSubscription(ctx, updateReq)
		if err != nil {
			s.logger.WithError(err).Error("Failed to update subscription in payment gateway")
			return nil, fmt.Errorf("%s", s.translator.TWithContext(ctx, "subscription.upgrade_failed"))
		}
	}

	// Save updated subscription
	if err := s.subscriptionRepo.Update(ctx, subscription); err != nil {
		return nil, fmt.Errorf("%s", s.translator.TWithContext(ctx, "error.internal"))
	}

	// Publish event
	s.eventBus.Publish(ctx, domainEvents.SubscriptionUpgraded{
		SubscriptionID: subscription.ID,
		UserID:         subscription.UserID,
		OldTier:        valueobjects.SubscriptionTier(currentTier.Code),
		NewTier:        valueobjects.SubscriptionTier(newTier.Code),
		UpgradedAt:     time.Now(),
	})

	return subscription, nil
}

// DowngradeSubscription downgrades a subscription to a lower tier
func (s *SubscriptionService) DowngradeSubscription(ctx context.Context, subscriptionID, newTierID uuid.UUID, effectiveNextCycle bool) (*entities.Subscription, error) {
	// Get subscription
	subscription, err := s.subscriptionRepo.FindByID(ctx, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("%s", s.translator.TWithContext(ctx, "subscription.not_found"))
	}

	// Check if downgrade is allowed
	if !subscription.CanDowngrade() {
		return nil, fmt.Errorf("%s", s.translator.TWithContext(ctx, "subscription.cannot_downgrade"))
	}

	// Get new tier
	newTier, err := s.tierRepo.FindByID(ctx, newTierID)
	if err != nil {
		return nil, fmt.Errorf("%s", s.translator.TWithContext(ctx, "subscription.tier_not_found"))
	}

	// Check if it's actually a downgrade
	currentTier, err := s.tierRepo.FindByCode(ctx, string(subscription.Tier))
	if err != nil {
		return nil, fmt.Errorf("%s", s.translator.TWithContext(ctx, "subscription.tier_not_found"))
	}

	if !newTier.IsDowngradeFrom(currentTier) {
		return nil, fmt.Errorf("%s", s.translator.TWithContext(ctx, "subscription.not_a_downgrade"))
	}

	if effectiveNextCycle {
		// Schedule downgrade for next billing cycle
		subscription.Metadata["scheduled_tier_change"] = map[string]interface{}{
			"new_tier_id":    newTier.ID,
			"new_tier_code":  newTier.Code,
			"effective_date": subscription.CurrentPeriodEnd,
		}
	} else {
		// Apply downgrade immediately
		subscription.Tier = valueobjects.SubscriptionTier(newTier.Code)
		subscription.PlanName = newTier.Name
		subscription.Price = valueobjects.Money{
			Amount:   newTier.Price,
			Currency: newTier.Currency,
		}
		subscription.UsageLimits = newTier.Limits
	}

	subscription.UpdatedAt = time.Now()

	// Update in payment gateway
	if gatewaySubID, ok := subscription.Metadata["gateway_subscription_id"].(string); ok {
		updateReq := ports.UpdateSubscriptionRequest{
			ProviderSubscriptionID: gatewaySubID,
			ProrationBehavior:      "none", // No proration for downgrades
		}

		_, err = s.paymentGateway.UpdateSubscription(ctx, updateReq)
		if err != nil {
			s.logger.WithError(err).Error("Failed to update subscription in payment gateway")
			return nil, fmt.Errorf("%s", s.translator.TWithContext(ctx, "subscription.downgrade_failed"))
		}
	}

	// Save updated subscription
	if err := s.subscriptionRepo.Update(ctx, subscription); err != nil {
		return nil, fmt.Errorf("%s", s.translator.TWithContext(ctx, "error.internal"))
	}

	// Publish event
	s.eventBus.Publish(ctx, domainEvents.SubscriptionDowngraded{
		SubscriptionID: subscription.ID,
		UserID:         subscription.UserID,
		OldTier:        valueobjects.SubscriptionTier(currentTier.Code),
		NewTier:        valueobjects.SubscriptionTier(newTier.Code),
		DowngradedAt:   time.Now(),
	})

	return subscription, nil
}

// PauseSubscription pauses a subscription
func (s *SubscriptionService) PauseSubscription(ctx context.Context, subscriptionID uuid.UUID, pausedUntil *time.Time) error {
	subscription, err := s.subscriptionRepo.FindByID(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("%s", s.translator.TWithContext(ctx, "subscription.not_found"))
	}

	if subscription.Status != entities.SubscriptionStatusActive {
		return fmt.Errorf("%s", s.translator.TWithContext(ctx, "subscription.cannot_pause"))
	}

	subscription.Pause()
	if pausedUntil != nil {
		subscription.ResumeAt = pausedUntil
	}

	// Pause in payment gateway
	if gatewaySubID, ok := subscription.Metadata["gateway_subscription_id"].(string); ok {
		if err := s.paymentGateway.PauseSubscription(ctx, gatewaySubID); err != nil {
			s.logger.WithError(err).Error("Failed to pause subscription in payment gateway")
			return fmt.Errorf("%s", s.translator.TWithContext(ctx, "subscription.pause_failed"))
		}
	}

	if err := s.subscriptionRepo.Update(ctx, subscription); err != nil {
		return fmt.Errorf("%s", s.translator.TWithContext(ctx, "error.internal"))
	}

	// Publish event
	event := domainEvents.SubscriptionPaused{
		SubscriptionID: subscription.ID,
		UserID:         subscription.UserID,
		PausedAt:       time.Now(),
		ResumeAt:       pausedUntil,
	}
	s.eventBus.Publish(ctx, event)

	return nil
}

// ResumeSubscription resumes a paused subscription
func (s *SubscriptionService) ResumeSubscription(ctx context.Context, subscriptionID uuid.UUID) error {
	subscription, err := s.subscriptionRepo.FindByID(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("%s", s.translator.TWithContext(ctx, "subscription.not_found"))
	}

	if subscription.Status != entities.SubscriptionStatusPaused {
		return fmt.Errorf("%s", s.translator.TWithContext(ctx, "subscription.not_paused"))
	}

	subscription.Resume()

	// Resume in payment gateway
	if gatewaySubID, ok := subscription.Metadata["gateway_subscription_id"].(string); ok {
		if err := s.paymentGateway.ResumeSubscription(ctx, gatewaySubID); err != nil {
			s.logger.WithError(err).Error("Failed to resume subscription in payment gateway")
			return fmt.Errorf("%s", s.translator.TWithContext(ctx, "subscription.resume_failed"))
		}
	}

	if err := s.subscriptionRepo.Update(ctx, subscription); err != nil {
		return fmt.Errorf("%s", s.translator.TWithContext(ctx, "error.internal"))
	}

	// Publish event
	event := domainEvents.SubscriptionResumed{
		SubscriptionID: subscription.ID,
		UserID:         subscription.UserID,
		ResumedAt:      time.Now(),
	}
	s.eventBus.Publish(ctx, event)

	return nil
}

// CancelSubscription cancels a subscription
func (s *SubscriptionService) CancelSubscription(ctx context.Context, subscriptionID uuid.UUID, reason string, immediate bool) error {
	subscription, err := s.subscriptionRepo.FindByID(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("%s", s.translator.TWithContext(ctx, "subscription.not_found"))
	}

	if err := subscription.Cancel(reason); err != nil {
		return err
	}

	// Cancel in payment gateway
	if gatewaySubID, ok := subscription.Metadata["gateway_subscription_id"].(string); ok {
		if immediate {
			if err := s.paymentGateway.CancelSubscription(ctx, gatewaySubID, immediate); err != nil {
				s.logger.WithError(err).Error("Failed to cancel subscription in payment gateway")
				return fmt.Errorf("%s", s.translator.TWithContext(ctx, "subscription.cancel_failed"))
			}
		} else {
			// Cancel at period end
			updateReq := ports.UpdateSubscriptionRequest{
				ProviderSubscriptionID: gatewaySubID,
				ProrationBehavior:      "none",
			}
			if _, err := s.paymentGateway.UpdateSubscription(ctx, updateReq); err != nil {
				s.logger.WithError(err).Error("Failed to schedule subscription cancellation")
				return fmt.Errorf("%s", s.translator.TWithContext(ctx, "subscription.cancel_failed"))
			}
		}
	}

	if err := s.subscriptionRepo.Update(ctx, subscription); err != nil {
		return fmt.Errorf("%s", s.translator.TWithContext(ctx, "error.internal"))
	}

	// Publish event
	event := domainEvents.SubscriptionCanceled{
		SubscriptionID: subscription.ID,
		UserID:         subscription.UserID,
		Reason:         reason,
		CanceledAt:     time.Now(),
	}
	s.eventBus.Publish(ctx, event)

	return nil
}

// ProcessTrialExpiration processes expired trials
func (s *SubscriptionService) ProcessTrialExpiration(ctx context.Context) error {
	// Find subscriptions with expired trials
	expiring, err := s.subscriptionRepo.FindExpiring(ctx, time.Now())
	if err != nil {
		return err
	}

	for _, subscription := range expiring {
		if subscription.IsTrialing && subscription.TrialEndsAt != nil && time.Now().After(*subscription.TrialEndsAt) {
			// Trial has ended
			subscription.IsTrialing = false
			subscription.Status = entities.SubscriptionStatusActive

			// Start billing
			if err := s.startBilling(ctx, subscription); err != nil {
				s.logger.WithError(err).Error("Failed to start billing after trial")
				subscription.Status = entities.SubscriptionStatusExpired
			}

			if err := s.subscriptionRepo.Update(ctx, subscription); err != nil {
				s.logger.WithError(err).Error("Failed to update subscription after trial expiration")
				continue
			}

			// Publish event
			event := domainEvents.SubscriptionTrialEnded{
				SubscriptionID: subscription.ID,
				UserID:         subscription.UserID,
				EndedAt:        time.Now(),
			}
			s.eventBus.Publish(ctx, event)
		}
	}

	return nil
}

// EnforceUsageLimits checks and enforces usage limits
func (s *SubscriptionService) EnforceUsageLimits(ctx context.Context, subscriptionID uuid.UUID, resource string, requestedUsage int) error {
	subscription, err := s.subscriptionRepo.FindByID(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("%s", s.translator.TWithContext(ctx, "subscription.not_found"))
	}

	limit, hasLimit := subscription.UsageLimits[resource]
	if !hasLimit || limit == -1 {
		// No limit or unlimited
		return nil
	}

	currentUsage := subscription.CurrentUsage[resource]
	if currentUsage+requestedUsage > limit {
		return fmt.Errorf("%s", s.translator.TWithContext(ctx, "subscription.limit_reached", resource))
	}

	// Update usage
	subscription.CurrentUsage[resource] = currentUsage + requestedUsage
	subscription.UpdatedAt = time.Now()

	if err := s.subscriptionRepo.Update(ctx, subscription); err != nil {
		return fmt.Errorf("%s", s.translator.TWithContext(ctx, "error.internal"))
	}

	// Check if approaching limit (80%)
	percentage := float64(subscription.CurrentUsage[resource]) / float64(limit)
	if percentage > 0.8 {
		event := domainEvents.SubscriptionUsageWarning{
			SubscriptionID: subscription.ID,
			UserID:         subscription.UserID,
			Resource:       resource,
			CurrentUsage:   subscription.CurrentUsage[resource],
			Limit:          limit,
			Percentage:     percentage * 100,
		}
		s.eventBus.Publish(ctx, event)
	}

	return nil
}

// GetSubscription gets a subscription by ID
func (s *SubscriptionService) GetSubscription(ctx context.Context, subscriptionID uuid.UUID) (*entities.Subscription, error) {
	subscription, err := s.subscriptionRepo.FindByID(ctx, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("%s", s.translator.TWithContext(ctx, "subscription.not_found"))
	}
	return subscription, nil
}

// GetUserSubscriptions gets all subscriptions for a user
func (s *SubscriptionService) GetUserSubscriptions(ctx context.Context, userID uuid.UUID) ([]*entities.Subscription, error) {
	return s.subscriptionRepo.FindByUserID(ctx, userID)
}

// ExtendTrial extends the trial period
func (s *SubscriptionService) ExtendTrial(ctx context.Context, subscriptionID uuid.UUID, days int, reason string) error {
	subscription, err := s.subscriptionRepo.FindByID(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("%s", s.translator.TWithContext(ctx, "subscription.not_found"))
	}

	if err := subscription.ExtendTrial(days); err != nil {
		return fmt.Errorf("%s", s.translator.TWithContext(ctx, "subscription.cannot_extend_trial"))
	}

	if err := s.subscriptionRepo.Update(ctx, subscription); err != nil {
		return fmt.Errorf("%s", s.translator.TWithContext(ctx, "error.internal"))
	}

	// Publish event
	event := domainEvents.SubscriptionTrialExtended{
		SubscriptionID: subscription.ID,
		UserID:         subscription.UserID,
		ExtensionDays:  days,
		Reason:         reason,
		NewTrialEnd:    subscription.TrialEndsAt,
	}
	s.eventBus.Publish(ctx, event)

	return nil
}

// startBilling starts billing for a subscription after trial
func (s *SubscriptionService) startBilling(ctx context.Context, subscription *entities.Subscription) error {
	// Get payment method
	if gatewayCustomerID, ok := subscription.Metadata["gateway_customer_id"].(string); ok {
		// Create subscription in payment gateway
		tier, err := s.tierRepo.FindByCode(ctx, string(subscription.Tier))
		if err != nil {
			return err
		}

		subReq := ports.CreateSubscriptionRequest{
			CustomerID: gatewayCustomerID,
			TrialDays:  0,
			Metadata: map[string]string{
				"subscription_id": subscription.ID.String(),
			},
		}

		gatewaySubscription, err := s.paymentGateway.CreateSubscription(ctx, subReq)
		if err != nil {
			return err
		}

		subscription.Metadata["gateway_subscription_id"] = gatewaySubscription.ProviderSubscriptionID

		// Create new billing cycle
		billingCycle := &entities.BillingCycle{
			ID:             uuid.New(),
			SubscriptionID: subscription.ID,
			StartDate:      time.Now(),
			EndDate:        time.Now().AddDate(0, 1, 0),
			Amount:         tier.Price,
			Currency:       tier.Currency,
			Status:         string(entities.BillingCycleStatusPending),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		return s.billingRepo.Create(ctx, billingCycle)
	}

	return fmt.Errorf("no payment method found")
}

// Helper method to get initial amount for billing cycle
func (s *SubscriptionService) getInitialAmount(isTrialing bool, tierPrice decimal.Decimal) decimal.Decimal {
	if isTrialing {
		return decimal.Zero
	}
	return tierPrice
}

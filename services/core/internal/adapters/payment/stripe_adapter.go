package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/invoice"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/paymentmethod"
	"github.com/stripe/stripe-go/v76/price"
	"github.com/stripe/stripe-go/v76/product"
	"github.com/stripe/stripe-go/v76/refund"
	"github.com/stripe/stripe-go/v76/subscription"
	"github.com/stripe/stripe-go/v76/webhook"

	"github.com/paradoxe35/mytorra/packages/logger"
	"github.com/paradoxe35/mytorra/services/core/internal/application/ports"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/entities"
)

type StripeAdapter struct {
	apiKey          string
	webhookSecret   string
	logger          logger.Interface
	defaultCurrency string
}

func NewStripeAdapter(apiKey, webhookSecret string, logger logger.Interface) *StripeAdapter {
	stripe.Key = apiKey
	return &StripeAdapter{
		apiKey:          apiKey,
		webhookSecret:   webhookSecret,
		logger:          logger,
		defaultCurrency: "USD",
	}
}

func (s *StripeAdapter) CreateCustomer(ctx context.Context, user *entities.User) (string, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(user.Email.String()),
		Name:  stripe.String(user.Name),
		Metadata: map[string]string{
			"user_id": user.ID.String(),
		},
	}

	if user.Phone.String() != "" {
		params.Phone = stripe.String(user.Phone.String())
	}

	c, err := customer.New(params)
	if err != nil {
		s.logger.WithError(err).Error("Failed to create Stripe customer")
		return "", fmt.Errorf("failed to create customer: %w", err)
	}

	return c.ID, nil
}

func (s *StripeAdapter) CreateSubscription(ctx context.Context, req ports.CreateSubscriptionRequest) (*ports.SubscriptionResponse, error) {
	// Create or get product
	productID, err := s.getOrCreateProduct(req.TierCode, req.TierName)
	if err != nil {
		return nil, err
	}

	// Create price
	priceID, err := s.createPrice(productID, req.Amount, req.Currency, req.BillingCycle)
	if err != nil {
		return nil, err
	}

	// Create subscription
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(req.CustomerID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(priceID),
			},
		},
		Metadata: map[string]string{
			"subscription_id": req.SubscriptionID,
			"tier_code":       req.TierCode,
		},
	}

	// Add trial period if specified
	if req.TrialDays > 0 {
		params.TrialPeriodDays = stripe.Int64(int64(req.TrialDays))
	}

	sub, err := subscription.New(params)
	if err != nil {
		s.logger.WithError(err).Error("Failed to create Stripe subscription")
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	return s.mapSubscriptionResponse(sub), nil
}

func (s *StripeAdapter) UpdateSubscription(ctx context.Context, req ports.UpdateSubscriptionRequest) (*ports.SubscriptionResponse, error) {
	// Get the subscription
	sub, err := subscription.Get(req.ProviderSubscriptionID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	// Create new price if tier changed
	if req.NewTierCode != "" {
		productID, err := s.getOrCreateProduct(req.NewTierCode, req.NewTierName)
		if err != nil {
			return nil, err
		}

		priceID, err := s.createPrice(productID, req.NewAmount, req.Currency, req.BillingCycle)
		if err != nil {
			return nil, err
		}

		// Update subscription item
		params := &stripe.SubscriptionParams{
			Items: []*stripe.SubscriptionItemsParams{
				{
					ID:    stripe.String(sub.Items.Data[0].ID),
					Price: stripe.String(priceID),
				},
			},
			ProrationBehavior: stripe.String("create_prorations"),
		}

		sub, err = subscription.Update(req.ProviderSubscriptionID, params)
		if err != nil {
			return nil, fmt.Errorf("failed to update subscription: %w", err)
		}
	}

	return s.mapSubscriptionResponse(sub), nil
}

func (s *StripeAdapter) CancelSubscription(ctx context.Context, providerSubscriptionID string, immediate bool) error {
	params := &stripe.SubscriptionCancelParams{}

	if immediate {
		params.InvoiceNow = stripe.Bool(true)
		params.Prorate = stripe.Bool(true)
	} else {
		// Cancel at end of period
		updateParams := &stripe.SubscriptionParams{
			CancelAtPeriodEnd: stripe.Bool(true),
		}
		_, err := subscription.Update(providerSubscriptionID, updateParams)
		return err
	}

	_, err := subscription.Cancel(providerSubscriptionID, params)
	return err
}

func (s *StripeAdapter) PauseSubscription(ctx context.Context, providerSubscriptionID string) error {
	params := &stripe.SubscriptionParams{
		PauseCollection: &stripe.SubscriptionPauseCollectionParams{
			Behavior: stripe.String("mark_uncollectible"),
		},
	}

	_, err := subscription.Update(providerSubscriptionID, params)
	return err
}

func (s *StripeAdapter) ResumeSubscription(ctx context.Context, providerSubscriptionID string) error {
	params := &stripe.SubscriptionParams{}
	params.AddMetadata("pause_collection", "")

	_, err := subscription.Update(providerSubscriptionID, params)
	return err
}

func (s *StripeAdapter) ChargePayment(ctx context.Context, req ports.ChargePaymentRequest) (*ports.PaymentResponse, error) {
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(req.Amount.Mul(decimal.NewFromInt(100)).IntPart()), // Convert to cents
		Currency: stripe.String(req.Currency),
		Customer: stripe.String(req.CustomerID),
		Metadata: map[string]string{
			"reference_id": req.ReferenceID,
			"description":  req.Description,
		},
	}

	if req.PaymentMethodID != "" {
		params.PaymentMethod = stripe.String(req.PaymentMethodID)
		params.Confirm = stripe.Bool(true)
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		s.logger.WithError(err).Error("Failed to create payment intent")
		return nil, fmt.Errorf("failed to charge payment: %w", err)
	}

	return &ports.PaymentResponse{
		PaymentID:     pi.ID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		Status:        mapPaymentStatus(pi.Status),
		TransactionID: pi.ID,
	}, nil
}

func (s *StripeAdapter) RefundPayment(ctx context.Context, req ports.RefundPaymentRequest) (*ports.RefundResponse, error) {
	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(req.PaymentID),
		Amount:        stripe.Int64(req.Amount.Mul(decimal.NewFromInt(100)).IntPart()),
		Reason:        stripe.String(req.Reason),
	}

	r, err := refund.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to refund payment: %w", err)
	}

	return &ports.RefundResponse{
		RefundID:      r.ID,
		Amount:        req.Amount,
		Currency:      string(r.Currency),
		Status:        mapRefundStatus(r.Status),
		TransactionID: r.ID,
	}, nil
}

func (s *StripeAdapter) CreatePaymentMethod(ctx context.Context, req ports.CreatePaymentMethodRequest) (*ports.PaymentMethodResponse, error) {
	params := &stripe.PaymentMethodParams{
		Type: stripe.String(string(stripe.PaymentMethodTypeCard)),
		Card: &stripe.PaymentMethodCardParams{
			Token: stripe.String(req.Token),
		},
	}

	pm, err := paymentmethod.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment method: %w", err)
	}

	// Attach to customer
	attachParams := &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(req.CustomerID),
	}
	pm, err = paymentmethod.Attach(pm.ID, attachParams)
	if err != nil {
		return nil, fmt.Errorf("failed to attach payment method: %w", err)
	}

	// Set as default if requested
	if req.SetDefault {
		customerParams := &stripe.CustomerParams{
			InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
				DefaultPaymentMethod: stripe.String(pm.ID),
			},
		}
		_, err = customer.Update(req.CustomerID, customerParams)
		if err != nil {
			s.logger.WithError(err).Warn("Failed to set default payment method")
		}
	}

	return &ports.PaymentMethodResponse{
		MethodID:  pm.ID,
		Type:      string(pm.Type),
		Last4:     pm.Card.Last4,
		Brand:     string(pm.Card.Brand),
		ExpMonth:  int(pm.Card.ExpMonth),
		ExpYear:   int(pm.Card.ExpYear),
		IsDefault: req.SetDefault,
	}, nil
}

func (s *StripeAdapter) DeletePaymentMethod(ctx context.Context, methodID string) error {
	_, err := paymentmethod.Detach(methodID, nil)
	return err
}

func (s *StripeAdapter) GetPaymentMethods(ctx context.Context, customerID string) ([]*ports.PaymentMethodResponse, error) {
	params := &stripe.PaymentMethodListParams{
		Customer: stripe.String(customerID),
		Type:     stripe.String(string(stripe.PaymentMethodTypeCard)),
	}

	iter := paymentmethod.List(params)
	var methods []*ports.PaymentMethodResponse

	for iter.Next() {
		pm := iter.PaymentMethod()
		methods = append(methods, &ports.PaymentMethodResponse{
			MethodID: pm.ID,
			Type:     string(pm.Type),
			Last4:    pm.Card.Last4,
			Brand:    string(pm.Card.Brand),
			ExpMonth: int(pm.Card.ExpMonth),
			ExpYear:  int(pm.Card.ExpYear),
		})
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to list payment methods: %w", err)
	}

	return methods, nil
}

func (s *StripeAdapter) GetInvoice(ctx context.Context, invoiceID string) (*ports.InvoiceResponse, error) {
	inv, err := invoice.Get(invoiceID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	return s.mapInvoiceResponse(inv), nil
}

func (s *StripeAdapter) ProcessWebhook(ctx context.Context, payload []byte, signature string) (*ports.WebhookEvent, error) {
	event, err := webhook.ConstructEvent(payload, signature, s.webhookSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to verify webhook: %w", err)
	}

	return &ports.WebhookEvent{
		ID:        event.ID,
		Type:      string(event.Type),
		Data:      event.Data.Raw,
		Timestamp: time.Unix(event.Created, 0),
	}, nil
}

// Helper methods
func (s *StripeAdapter) getOrCreateProduct(tierCode, tierName string) (string, error) {
	// Search for existing product
	params := &stripe.ProductSearchParams{}
	params.Query = fmt.Sprintf("metadata['tier_code']:'%s'", tierCode)

	iter := product.Search(params)
	if iter.Next() {
		return iter.Product().ID, nil
	}

	// Create new product
	productParams := &stripe.ProductParams{
		Name: stripe.String(tierName),
		Metadata: map[string]string{
			"tier_code": tierCode,
		},
	}

	p, err := product.New(productParams)
	if err != nil {
		return "", fmt.Errorf("failed to create product: %w", err)
	}

	return p.ID, nil
}

func (s *StripeAdapter) createPrice(productID string, amount decimal.Decimal, currency string, billingCycle entities.SubscriptionBillingCycle) (string, error) {
	var recurring *stripe.PriceRecurringParams

	switch billingCycle {
	case entities.BillingCycleMonthly:
		recurring = &stripe.PriceRecurringParams{
			Interval: stripe.String("month"),
		}
	case entities.BillingCycleAnnual:
		recurring = &stripe.PriceRecurringParams{
			Interval: stripe.String("year"),
		}
	default:
		// One-time price
		recurring = nil
	}

	params := &stripe.PriceParams{
		Product:    stripe.String(productID),
		UnitAmount: stripe.Int64(amount.Mul(decimal.NewFromInt(100)).IntPart()),
		Currency:   stripe.String(currency),
		Recurring:  recurring,
	}

	p, err := price.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create price: %w", err)
	}

	return p.ID, nil
}

func (s *StripeAdapter) mapSubscriptionResponse(sub *stripe.Subscription) *ports.SubscriptionResponse {
	var trialEnd *time.Time
	if sub.TrialEnd > 0 {
		t := time.Unix(sub.TrialEnd, 0)
		trialEnd = &t
	}

	return &ports.SubscriptionResponse{
		ProviderSubscriptionID: sub.ID,
		Status:                 mapSubscriptionStatus(sub.Status),
		CurrentPeriodStart:     time.Unix(sub.CurrentPeriodStart, 0),
		CurrentPeriodEnd:       time.Unix(sub.CurrentPeriodEnd, 0),
		TrialEnd:               trialEnd,
	}
}

func (s *StripeAdapter) mapInvoiceResponse(inv *stripe.Invoice) *ports.InvoiceResponse {
	amount := decimal.NewFromInt(inv.Total).Div(decimal.NewFromInt(100))

	var items []ports.InvoiceItem
	for _, line := range inv.Lines.Data {
		items = append(items, ports.InvoiceItem{
			Description: line.Description,
			Amount:      decimal.NewFromInt(line.Amount).Div(decimal.NewFromInt(100)),
			Quantity:    int(line.Quantity),
		})
	}

	return &ports.InvoiceResponse{
		InvoiceID:   inv.ID,
		Number:      inv.Number,
		Amount:      amount,
		Currency:    string(inv.Currency),
		Status:      mapInvoiceStatus(inv.Status),
		DueDate:     time.Unix(inv.DueDate, 0),
		Items:       items,
		DownloadURL: inv.InvoicePDF,
	}
}

func mapSubscriptionStatus(status stripe.SubscriptionStatus) string {
	switch status {
	case stripe.SubscriptionStatusActive:
		return "active"
	case stripe.SubscriptionStatusTrialing:
		return "trialing"
	case stripe.SubscriptionStatusPastDue:
		return "past_due"
	case stripe.SubscriptionStatusCanceled:
		return "canceled"
	case stripe.SubscriptionStatusUnpaid:
		return "unpaid"
	default:
		return string(status)
	}
}

func mapPaymentStatus(status stripe.PaymentIntentStatus) string {
	switch status {
	case stripe.PaymentIntentStatusSucceeded:
		return "succeeded"
	case stripe.PaymentIntentStatusProcessing:
		return "processing"
	case stripe.PaymentIntentStatusRequiresPaymentMethod:
		return "requires_payment_method"
	case stripe.PaymentIntentStatusCanceled:
		return "canceled"
	default:
		return string(status)
	}
}

func mapRefundStatus(status stripe.RefundStatus) string {
	switch status {
	case stripe.RefundStatusSucceeded:
		return "succeeded"
	case stripe.RefundStatusPending:
		return "pending"
	case stripe.RefundStatusFailed:
		return "failed"
	case stripe.RefundStatusCanceled:
		return "canceled"
	default:
		return string(status)
	}
}

func mapInvoiceStatus(status stripe.InvoiceStatus) string {
	switch status {
	case stripe.InvoiceStatusDraft:
		return "draft"
	case stripe.InvoiceStatusOpen:
		return "open"
	case stripe.InvoiceStatusPaid:
		return "paid"
	case stripe.InvoiceStatusVoid:
		return "void"
	case stripe.InvoiceStatusUncollectible:
		return "uncollectible"
	default:
		return string(status)
	}
}

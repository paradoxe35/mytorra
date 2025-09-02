package dtos

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// BillingCycleDTO represents a billing cycle
type BillingCycleDTO struct {
	ID             uuid.UUID       `json:"id"`
	SubscriptionID uuid.UUID       `json:"subscription_id"`
	StartDate      time.Time       `json:"start_date"`
	EndDate        time.Time       `json:"end_date"`
	Amount         decimal.Decimal `json:"amount"`
	Currency       string          `json:"currency"`
	Status         string          `json:"status"`
	InvoiceID      *string         `json:"invoice_id,omitempty"`
	PaidAt         *time.Time      `json:"paid_at,omitempty"`
	IsOverdue      bool            `json:"is_overdue"`
	DaysRemaining  int             `json:"days_remaining"`
}

// PaymentHistoryDTO represents payment history
type PaymentHistoryDTO struct {
	Payments []PaymentTransactionDTO `json:"payments"`
	Total    int64                   `json:"total"`
	Page     int                     `json:"page"`
	PageSize int                     `json:"page_size"`
}

// PaymentTransactionDTO represents a payment transaction
type PaymentTransactionDTO struct {
	ID                   uuid.UUID       `json:"id"`
	SubscriptionID       uuid.UUID       `json:"subscription_id"`
	Amount               decimal.Decimal `json:"amount"`
	Currency             string          `json:"currency"`
	Type                 string          `json:"type"`
	Status               string          `json:"status"`
	GatewayTransactionID string          `json:"gateway_transaction_id,omitempty"`
	ProcessedAt          *time.Time      `json:"processed_at,omitempty"`
	FailureReason        *string         `json:"failure_reason,omitempty"`
	CreatedAt            time.Time       `json:"created_at"`
}

// InvoiceDTO represents an invoice
type InvoiceDTO struct {
	ID             uuid.UUID       `json:"id"`
	Number         string          `json:"number"`
	SubscriptionID uuid.UUID       `json:"subscription_id"`
	BillingCycleID uuid.UUID       `json:"billing_cycle_id"`
	CustomerName   string          `json:"customer_name"`
	CustomerEmail  string          `json:"customer_email"`
	BillingAddress AddressDTO      `json:"billing_address"`
	LineItems      []LineItemDTO   `json:"line_items"`
	Subtotal       decimal.Decimal `json:"subtotal"`
	Tax            decimal.Decimal `json:"tax"`
	Total          decimal.Decimal `json:"total"`
	Currency       string          `json:"currency"`
	Status         string          `json:"status"`
	DueDate        time.Time       `json:"due_date"`
	PaidAt         *time.Time      `json:"paid_at,omitempty"`
	DownloadURL    string          `json:"download_url,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

// LineItemDTO represents an invoice line item
type LineItemDTO struct {
	Description string          `json:"description"`
	Quantity    int             `json:"quantity"`
	UnitPrice   decimal.Decimal `json:"unit_price"`
	Amount      decimal.Decimal `json:"amount"`
}

// AddressDTO represents a billing address
type AddressDTO struct {
	Line1      string `json:"line1"`
	Line2      string `json:"line2,omitempty"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// ProcessPaymentDTO represents a payment processing request
type ProcessPaymentDTO struct {
	SubscriptionID  uuid.UUID       `json:"subscription_id" validate:"required"`
	PaymentMethodID string          `json:"payment_method_id" validate:"required"`
	Amount          decimal.Decimal `json:"amount" validate:"required,gt=0"`
	Currency        string          `json:"currency" validate:"required,len=3"`
	Description     string          `json:"description,omitempty"`
}

// RefundDTO represents a refund request
type RefundDTO struct {
	TransactionID uuid.UUID       `json:"transaction_id" validate:"required"`
	Amount        decimal.Decimal `json:"amount" validate:"required,gt=0"`
	Reason        string          `json:"reason" validate:"required,min=3,max=500"`
}

// RefundResponseDTO represents a refund response
type RefundResponseDTO struct {
	ID            uuid.UUID       `json:"id"`
	TransactionID uuid.UUID       `json:"transaction_id"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
	Status        string          `json:"status"`
	Reason        string          `json:"reason"`
	ProcessedAt   time.Time       `json:"processed_at"`
}

// WebhookPayloadDTO represents a webhook payload from payment gateway
type WebhookPayloadDTO struct {
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Signature string                 `json:"signature"`
	Timestamp time.Time              `json:"timestamp"`
}

// BillingSettingsDTO represents billing settings for a subscription
type BillingSettingsDTO struct {
	AutoRenew            bool   `json:"auto_renew"`
	PaymentMethodID      string `json:"payment_method_id"`
	BillingEmail         string `json:"billing_email"`
	SendInvoiceReminders bool   `json:"send_invoice_reminders"`
	ReminderDaysBefore   []int  `json:"reminder_days_before"`
}

// UpdateBillingSettingsDTO represents request to update billing settings
type UpdateBillingSettingsDTO struct {
	AutoRenew            *bool   `json:"auto_renew,omitempty"`
	PaymentMethodID      *string `json:"payment_method_id,omitempty"`
	BillingEmail         *string `json:"billing_email,omitempty" validate:"omitempty,email"`
	SendInvoiceReminders *bool   `json:"send_invoice_reminders,omitempty"`
	ReminderDaysBefore   []int   `json:"reminder_days_before,omitempty"`
}

// UpcomingInvoiceDTO represents an upcoming invoice preview
type UpcomingInvoiceDTO struct {
	SubscriptionID uuid.UUID       `json:"subscription_id"`
	PeriodStart    time.Time       `json:"period_start"`
	PeriodEnd      time.Time       `json:"period_end"`
	Amount         decimal.Decimal `json:"amount"`
	Currency       string          `json:"currency"`
	DueDate        time.Time       `json:"due_date"`
	LineItems      []LineItemDTO   `json:"line_items"`
}

// PaymentRetryDTO represents a payment retry attempt
type PaymentRetryDTO struct {
	TransactionID      uuid.UUID `json:"transaction_id" validate:"required"`
	UseAlternateMethod bool      `json:"use_alternate_method"`
	PaymentMethodID    string    `json:"payment_method_id,omitempty"`
}

// BillingReportDTO represents a billing report
type BillingReportDTO struct {
	Period           string          `json:"period"`
	TotalRevenue     decimal.Decimal `json:"total_revenue"`
	TotalRefunds     decimal.Decimal `json:"total_refunds"`
	NetRevenue       decimal.Decimal `json:"net_revenue"`
	TransactionCount int64           `json:"transaction_count"`
	FailedPayments   int64           `json:"failed_payments"`
	SuccessRate      float64         `json:"success_rate"`
	AverageValue     decimal.Decimal `json:"average_value"`
	Currency         string          `json:"currency"`
}

package entities

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID      uuid.UUID              `db:"id" json:"id"`
	UserID  uuid.UUID              `db:"user_id" json:"user_id"`
	Type    NotificationType       `db:"type" json:"type"`
	Channel NotificationChannel    `db:"channel" json:"channel"`
	Title   string                 `db:"title" json:"title"`
	Message string                 `db:"message" json:"message"`
	Data    map[string]interface{} `db:"data" json:"data"`

	// Status tracking
	Status      NotificationStatus `db:"status" json:"status"`
	Read        bool               `db:"read" json:"read"`
	ReadAt      *time.Time         `db:"read_at" json:"read_at,omitempty"`
	DeliveredAt *time.Time         `db:"delivered_at" json:"delivered_at,omitempty"`
	FailedAt    *time.Time         `db:"failed_at" json:"failed_at,omitempty"`
	RetryCount  int                `db:"retry_count" json:"retry_count"`

	// Timestamps
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type (
	NotificationType    string
	NotificationChannel string
	NotificationStatus  string
)

const (
	// Notification types
	NotificationTypeWelcome             NotificationType = "welcome"
	NotificationTypeVerification        NotificationType = "verification"
	NotificationTypePasswordReset       NotificationType = "password_reset"
	NotificationTypePaymentSuccess      NotificationType = "payment_success"
	NotificationTypePaymentFailed       NotificationType = "payment_failed"
	NotificationTypeSubscriptionRenewal NotificationType = "subscription_renewal"
	NotificationTypeSubscriptionExpired NotificationType = "subscription_expired"
	NotificationTypeNewMessage          NotificationType = "new_message"
	NotificationTypeNewOrder            NotificationType = "new_order"
	NotificationTypeOrderUpdate         NotificationType = "order_update"

	// Notification channels
	NotificationChannelEmail NotificationChannel = "email"
	NotificationChannelSMS   NotificationChannel = "sms"
	NotificationChannelPush  NotificationChannel = "push"
	NotificationChannelInApp NotificationChannel = "in_app"

	// Notification status
	NotificationStatusPending   NotificationStatus = "pending"
	NotificationStatusSent      NotificationStatus = "sent"
	NotificationStatusDelivered NotificationStatus = "delivered"
	NotificationStatusFailed    NotificationStatus = "failed"
)

// Domain methods
func (n *Notification) MarkAsRead() {
	now := time.Now()
	n.Read = true
	n.ReadAt = &now
	n.UpdatedAt = now
}

func (n *Notification) MarkAsDelivered() {
	now := time.Now()
	n.Status = NotificationStatusDelivered
	n.DeliveredAt = &now
	n.UpdatedAt = now
}

func (n *Notification) MarkAsFailed() {
	now := time.Now()
	n.Status = NotificationStatusFailed
	n.FailedAt = &now
	n.RetryCount++
	n.UpdatedAt = now
}

func (n *Notification) CanRetry() bool {
	return n.Status == NotificationStatusFailed && n.RetryCount < 3
}

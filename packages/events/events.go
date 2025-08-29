package events

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type EventType string

const (
	UserCreated         EventType = "user.created"
	UserUpdated         EventType = "user.updated"
	UserDeleted         EventType = "user.deleted"
	
	ProductListed       EventType = "product.listed"
	ProductUpdated      EventType = "product.updated"
	ProductSold         EventType = "product.sold"
	ProductDelisted     EventType = "product.delisted"
	
	OrderCreated        EventType = "order.created"
	OrderPaid           EventType = "order.paid"
	OrderShipped        EventType = "order.shipped"
	OrderDelivered      EventType = "order.delivered"
	OrderCancelled      EventType = "order.cancelled"
	
	PaymentProcessed    EventType = "payment.processed"
	PaymentFailed       EventType = "payment.failed"
	PaymentRefunded     EventType = "payment.refunded"
	
	SubscriptionCreated EventType = "subscription.created"
	SubscriptionUpdated EventType = "subscription.updated"
	SubscriptionExpired EventType = "subscription.expired"
	
	NotificationSent    EventType = "notification.sent"
	NotificationFailed  EventType = "notification.failed"
)

type BaseEvent struct {
	ID        string    `json:"id"`
	Type      EventType `json:"type"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
}

func NewBaseEvent(eventType EventType, source string) BaseEvent {
	return BaseEvent{
		ID:        uuid.New().String(),
		Type:      eventType,
		Version:   "1.0",
		Timestamp: time.Now().UTC(),
		Source:    source,
	}
}

type UserCreatedEvent struct {
	BaseEvent
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Phone     string `json:"phone,omitempty"`
}

type ProductListedEvent struct {
	BaseEvent
	ProductID   string  `json:"product_id"`
	SellerID    string  `json:"seller_id"`
	Title       string  `json:"title"`
	Category    string  `json:"category"`
	Price       float64 `json:"price"`
	Currency    string  `json:"currency"`
	Condition   string  `json:"condition"`
	Location    string  `json:"location"`
}

type OrderCreatedEvent struct {
	BaseEvent
	OrderID    string    `json:"order_id"`
	BuyerID    string    `json:"buyer_id"`
	SellerID   string    `json:"seller_id"`
	ProductIDs []string  `json:"product_ids"`
	TotalAmount float64  `json:"total_amount"`
	Currency   string    `json:"currency"`
}

type PaymentProcessedEvent struct {
	BaseEvent
	PaymentID   string    `json:"payment_id"`
	OrderID     string    `json:"order_id"`
	UserID      string    `json:"user_id"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Method      string    `json:"method"`
	Status      string    `json:"status"`
}

type SubscriptionCreatedEvent struct {
	BaseEvent
	SubscriptionID string    `json:"subscription_id"`
	UserID         string    `json:"user_id"`
	PlanID         string    `json:"plan_id"`
	PlanName       string    `json:"plan_name"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	Amount         float64   `json:"amount"`
	Currency       string    `json:"currency"`
}

type NotificationSentEvent struct {
	BaseEvent
	NotificationID string `json:"notification_id"`
	UserID         string `json:"user_id"`
	Type           string `json:"notification_type"`
	Channel        string `json:"channel"`
	Subject        string `json:"subject,omitempty"`
	Success        bool   `json:"success"`
}

type EventEnvelope struct {
	Event    interface{} `json:"event"`
	Metadata Metadata    `json:"metadata"`
}

type Metadata struct {
	CorrelationID string            `json:"correlation_id"`
	CausationID   string            `json:"causation_id,omitempty"`
	UserID        string            `json:"user_id,omitempty"`
	IPAddress     string            `json:"ip_address,omitempty"`
	UserAgent     string            `json:"user_agent,omitempty"`
	Headers       map[string]string `json:"headers,omitempty"`
}

func Marshal(event interface{}) ([]byte, error) {
	return json.Marshal(event)
}

func Unmarshal(data []byte, event interface{}) error {
	return json.Unmarshal(data, event)
}
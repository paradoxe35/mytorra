package ports

import (
	"context"
)

// NotificationSender defines the interface for sending notifications
type NotificationSender interface {
	EmailSender
	SMSSender
	PushNotificationSender
}

// EmailSender defines the interface for sending emails
type EmailSender interface {
	SendEmail(ctx context.Context, req EmailRequest) error
	SendBulkEmails(ctx context.Context, requests []EmailRequest) error
	ValidateEmail(ctx context.Context, email string) (bool, error)
}

// SMSSender defines the interface for sending SMS messages
type SMSSender interface {
	SendSMS(ctx context.Context, req SMSRequest) error
	SendBulkSMS(ctx context.Context, requests []SMSRequest) error
	ValidatePhoneNumber(ctx context.Context, phone string) (bool, error)
}

// PushNotificationSender defines the interface for sending push notifications
type PushNotificationSender interface {
	SendPushNotification(ctx context.Context, req PushNotificationRequest) error
	SendBulkPushNotifications(ctx context.Context, requests []PushNotificationRequest) error
	RegisterDevice(ctx context.Context, userID, deviceToken string) error
	UnregisterDevice(ctx context.Context, deviceToken string) error
}

// EmailRequest represents an email sending request
type EmailRequest struct {
	To           []string
	CC           []string
	BCC          []string
	From         string
	FromName     string
	Subject      string
	HTMLBody     string
	TextBody     string
	Attachments  []EmailAttachment
	TemplateID   string
	TemplateData map[string]interface{}
}

// EmailAttachment represents an email attachment
type EmailAttachment struct {
	Filename string
	Content  []byte
	MimeType string
}

// SMSRequest represents an SMS sending request
type SMSRequest struct {
	To      string
	From    string
	Message string
}

// PushNotificationRequest represents a push notification request
type PushNotificationRequest struct {
	DeviceTokens []string
	Title        string
	Body         string
	Data         map[string]string
	Badge        *int
	Sound        string
	Icon         string
	ClickAction  string
}

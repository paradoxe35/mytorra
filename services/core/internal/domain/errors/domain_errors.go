package errors

import "errors"

var (
	// User errors
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrInvalidPhone       = errors.New("invalid phone format")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrAccountSuspended   = errors.New("account is suspended")
	ErrAccountBanned      = errors.New("account is banned")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidMFACode     = errors.New("invalid MFA code")

	// Payment errors
	ErrPaymentNotFound         = errors.New("payment not found")
	ErrInsufficientFunds       = errors.New("insufficient funds")
	ErrPaymentFailed           = errors.New("payment failed")
	ErrPaymentAlreadyProcessed = errors.New("payment already processed")
	ErrRefundNotAllowed        = errors.New("refund not allowed")
	ErrInvalidAmount           = errors.New("invalid amount")
	ErrInvalidCurrency         = errors.New("invalid currency")

	// Subscription errors
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrSubscriptionExpired  = errors.New("subscription expired")
	ErrSubscriptionPaused   = errors.New("subscription paused")
	ErrUsageLimitExceeded   = errors.New("usage limit exceeded")
	ErrInvalidTier          = errors.New("invalid subscription tier")
	ErrUpgradeNotAllowed    = errors.New("upgrade not allowed")
	ErrDowngradeNotAllowed  = errors.New("downgrade not allowed")

	// Notification errors
	ErrNotificationNotFound = errors.New("notification not found")
	ErrDeliveryFailed       = errors.New("notification delivery failed")

	// General errors
	ErrInvalidInput        = errors.New("invalid input")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrInternalServerError = errors.New("internal server error")
	ErrDatabaseError       = errors.New("database error")
	ErrCacheError          = errors.New("cache error")
	ErrEventPublishError   = errors.New("event publish error")
)

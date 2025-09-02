package valueobjects

import "fmt"

// SubscriptionStatus represents the status of a subscription
type SubscriptionStatus string

const (
	SubscriptionStatusActive   SubscriptionStatus = "active"
	SubscriptionStatusTrialing SubscriptionStatus = "trialing"
	SubscriptionStatusPastDue  SubscriptionStatus = "past_due"
	SubscriptionStatusPaused   SubscriptionStatus = "paused"
	SubscriptionStatusCanceled SubscriptionStatus = "canceled"
	SubscriptionStatusExpired  SubscriptionStatus = "expired"
)

// String returns the string representation of the subscription status
func (s SubscriptionStatus) String() string {
	return string(s)
}

// IsValid checks if the subscription status is valid
func (s SubscriptionStatus) IsValid() bool {
	switch s {
	case SubscriptionStatusActive, SubscriptionStatusTrialing,
		SubscriptionStatusPastDue, SubscriptionStatusPaused,
		SubscriptionStatusCanceled, SubscriptionStatusExpired:
		return true
	default:
		return false
	}
}

// IsActive checks if the subscription status represents an active state
func (s SubscriptionStatus) IsActive() bool {
	return s == SubscriptionStatusActive || s == SubscriptionStatusTrialing
}

// IsCancellable checks if the subscription can be canceled
func (s SubscriptionStatus) IsCancellable() bool {
	return s == SubscriptionStatusActive || s == SubscriptionStatusTrialing ||
		s == SubscriptionStatusPastDue || s == SubscriptionStatusPaused
}

// IsPausable checks if the subscription can be paused
func (s SubscriptionStatus) IsPausable() bool {
	return s == SubscriptionStatusActive || s == SubscriptionStatusTrialing
}

// IsResumable checks if the subscription can be resumed
func (s SubscriptionStatus) IsResumable() bool {
	return s == SubscriptionStatusPaused
}

// CanRenew checks if the subscription can be renewed
func (s SubscriptionStatus) CanRenew() bool {
	return s == SubscriptionStatusActive || s == SubscriptionStatusPastDue
}

// RequiresPayment checks if the subscription requires payment action
func (s SubscriptionStatus) RequiresPayment() bool {
	return s == SubscriptionStatusPastDue
}

// GetDisplayName returns the display name for the subscription status
func (s SubscriptionStatus) GetDisplayName() string {
	switch s {
	case SubscriptionStatusActive:
		return "Active"
	case SubscriptionStatusTrialing:
		return "Trial"
	case SubscriptionStatusPastDue:
		return "Past Due"
	case SubscriptionStatusPaused:
		return "Paused"
	case SubscriptionStatusCanceled:
		return "Canceled"
	case SubscriptionStatusExpired:
		return "Expired"
	default:
		return string(s)
	}
}

// GetBadgeColor returns a suggested badge color for UI display
func (s SubscriptionStatus) GetBadgeColor() string {
	switch s {
	case SubscriptionStatusActive:
		return "green"
	case SubscriptionStatusTrialing:
		return "blue"
	case SubscriptionStatusPastDue:
		return "orange"
	case SubscriptionStatusPaused:
		return "yellow"
	case SubscriptionStatusCanceled:
		return "red"
	case SubscriptionStatusExpired:
		return "gray"
	default:
		return "gray"
	}
}

// ParseSubscriptionStatus parses a string into a SubscriptionStatus
func ParseSubscriptionStatus(s string) (SubscriptionStatus, error) {
	status := SubscriptionStatus(s)
	if !status.IsValid() {
		return "", fmt.Errorf("invalid subscription status: %s", s)
	}
	return status, nil
}

// AllowedTransitions returns the valid status transitions from the current status
func (s SubscriptionStatus) AllowedTransitions() []SubscriptionStatus {
	switch s {
	case SubscriptionStatusActive:
		return []SubscriptionStatus{
			SubscriptionStatusPastDue,
			SubscriptionStatusPaused,
			SubscriptionStatusCanceled,
			SubscriptionStatusExpired,
		}
	case SubscriptionStatusTrialing:
		return []SubscriptionStatus{
			SubscriptionStatusActive,
			SubscriptionStatusCanceled,
			SubscriptionStatusExpired,
		}
	case SubscriptionStatusPastDue:
		return []SubscriptionStatus{
			SubscriptionStatusActive,
			SubscriptionStatusCanceled,
			SubscriptionStatusExpired,
		}
	case SubscriptionStatusPaused:
		return []SubscriptionStatus{
			SubscriptionStatusActive,
			SubscriptionStatusCanceled,
			SubscriptionStatusExpired,
		}
	case SubscriptionStatusCanceled:
		return []SubscriptionStatus{} // Terminal state
	case SubscriptionStatusExpired:
		return []SubscriptionStatus{} // Terminal state
	default:
		return []SubscriptionStatus{}
	}
}

// CanTransitionTo checks if the status can transition to another status
func (s SubscriptionStatus) CanTransitionTo(newStatus SubscriptionStatus) bool {
	allowed := s.AllowedTransitions()
	for _, status := range allowed {
		if status == newStatus {
			return true
		}
	}
	return false
}

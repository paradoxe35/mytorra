package events

import (
	"time"

	"github.com/google/uuid"
)

const (
	EventTypeUserCreated             = "user.created"
	EventTypeUserVerified            = "user.verified"
	EventTypeUserDeleted             = "user.deleted"
	EventTypeUserUpdated             = "user.updated"
	EventTypeIDVerificationSubmitted = "user.id_verification_submitted"
)

type UserCreatedEvent struct {
	EventID   string    `json:"event_id"`
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Area      string    `json:"area"`
	CreatedAt time.Time `json:"created_at"`
	EventType string    `json:"event_type"`
	Version   string    `json:"version"`
}

func NewUserCreatedEvent(userID uuid.UUID, email, area string) UserCreatedEvent {
	return UserCreatedEvent{
		EventID:   uuid.New().String(),
		UserID:    userID,
		Email:     email,
		Area:      area,
		CreatedAt: time.Now(),
		EventType: EventTypeUserCreated,
		Version:   "1.0",
	}
}

type UserVerifiedEvent struct {
	EventID           string    `json:"event_id"`
	UserID            uuid.UUID `json:"user_id"`
	VerificationType  string    `json:"verification_type"` // email, phone, id
	VerificationLevel int       `json:"verification_level"`
	VerifiedAt        time.Time `json:"verified_at"`
	EventType         string    `json:"event_type"`
	Version           string    `json:"version"`
}

func NewUserVerifiedEvent(userID uuid.UUID, verificationType string, level int) UserVerifiedEvent {
	return UserVerifiedEvent{
		EventID:           uuid.New().String(),
		UserID:            userID,
		VerificationType:  verificationType,
		VerificationLevel: level,
		VerifiedAt:        time.Now(),
		EventType:         EventTypeUserVerified,
		Version:           "1.0",
	}
}

type UserDeletedEvent struct {
	EventID   string    `json:"event_id"`
	UserID    uuid.UUID `json:"user_id"`
	DeletedAt time.Time `json:"deleted_at"`
	EventType string    `json:"event_type"`
	Version   string    `json:"version"`
}

func NewUserDeletedEvent(userID uuid.UUID) UserDeletedEvent {
	return UserDeletedEvent{
		EventID:   uuid.New().String(),
		UserID:    userID,
		DeletedAt: time.Now(),
		EventType: EventTypeUserDeleted,
		Version:   "1.0",
	}
}

type IDVerificationSubmittedEvent struct {
	EventID     string    `json:"event_id"`
	UserID      uuid.UUID `json:"user_id"`
	Filename    string    `json:"filename"`
	SubmittedAt time.Time `json:"submitted_at"`
	EventType   string    `json:"event_type"`
	Version     string    `json:"version"`
}

func NewIDVerificationSubmittedEvent(userID uuid.UUID, filename string) IDVerificationSubmittedEvent {
	return IDVerificationSubmittedEvent{
		EventID:     uuid.New().String(),
		UserID:      userID,
		Filename:    filename,
		SubmittedAt: time.Now(),
		EventType:   EventTypeIDVerificationSubmitted,
		Version:     "1.0",
	}
}

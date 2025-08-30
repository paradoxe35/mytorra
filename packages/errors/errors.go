package errors

import (
	"errors"
	"fmt"
	"net/http"
)

type ErrorType string

const (
	ErrorTypeValidation    ErrorType = "VALIDATION"
	ErrorTypeAuthorization ErrorType = "AUTHORIZATION"
	ErrorTypeNotFound      ErrorType = "NOT_FOUND"
	ErrorTypeConflict      ErrorType = "CONFLICT"
	ErrorTypeInternal      ErrorType = "INTERNAL"
	ErrorTypeExternal      ErrorType = "EXTERNAL"
	ErrorTypeRateLimit     ErrorType = "RATE_LIMIT"
	ErrorTypeBadRequest    ErrorType = "BAD_REQUEST"
)

type AppError struct {
	Type    ErrorType `json:"type"`
	Message string    `json:"message"`
	Code    string    `json:"code,omitempty"`
	Details string    `json:"details,omitempty"`
	Err     error     `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func (e *AppError) HTTPStatus() int {
	switch e.Type {
	case ErrorTypeValidation, ErrorTypeBadRequest:
		return http.StatusBadRequest
	case ErrorTypeAuthorization:
		return http.StatusUnauthorized
	case ErrorTypeNotFound:
		return http.StatusNotFound
	case ErrorTypeConflict:
		return http.StatusConflict
	case ErrorTypeRateLimit:
		return http.StatusTooManyRequests
	case ErrorTypeInternal, ErrorTypeExternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func New(errType ErrorType, message string) *AppError {
	return &AppError{
		Type:    errType,
		Message: message,
	}
}

func NewWithCode(errType ErrorType, message, code string) *AppError {
	return &AppError{
		Type:    errType,
		Message: message,
		Code:    code,
	}
}

func Wrap(err error, errType ErrorType, message string) *AppError {
	return &AppError{
		Type:    errType,
		Message: message,
		Err:     err,
	}
}

func ValidationError(message string) *AppError {
	return New(ErrorTypeValidation, message)
}

func NotFoundError(resource string) *AppError {
	return New(ErrorTypeNotFound, fmt.Sprintf("%s not found", resource))
}

func UnauthorizedError(message string) *AppError {
	return New(ErrorTypeAuthorization, message)
}

func ConflictError(message string) *AppError {
	return New(ErrorTypeConflict, message)
}

func InternalError(message string) *AppError {
	return New(ErrorTypeInternal, message)
}

func BadRequestError(message string) *AppError {
	return New(ErrorTypeBadRequest, message)
}

func RateLimitError() *AppError {
	return New(ErrorTypeRateLimit, "Rate limit exceeded")
}

func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

func GetAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return nil
}

// Common error variables
var (
	ErrInvalidToken       = UnauthorizedError("Invalid or expired token")
	ErrInvalidCredentials = UnauthorizedError("Invalid credentials")
	ErrAccountSuspended   = UnauthorizedError("Account is suspended")
	ErrUserNotFound       = NotFoundError("User")
	ErrEmailExists        = ConflictError("Email already exists")
	ErrPhoneExists        = ConflictError("Phone already exists")
)

package api

import (
	"fmt"
	"net/http"
)

// ErrorType represents different categories of API errors
type ErrorType string

const (
	ErrorTypeAuth       ErrorType = "authentication"
	ErrorTypeNetwork    ErrorType = "network"
	ErrorTypeValidation ErrorType = "validation"
	ErrorTypeServer     ErrorType = "server"
	ErrorTypeTimeout    ErrorType = "timeout"
	ErrorTypeUnknown    ErrorType = "unknown"
)

// APIError represents a structured error from the qBittorrent API
type APIError struct {
	Type       ErrorType
	StatusCode int
	Message    string
	Cause      error
}

func (e *APIError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s error: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s error: %s", e.Type, e.Message)
}

func (e *APIError) Unwrap() error {
	return e.Cause
}

// IsAuthError returns true if the error is an authentication error
func IsAuthError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Type == ErrorTypeAuth
	}
	return false
}

// IsNetworkError returns true if the error is a network error
func IsNetworkError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Type == ErrorTypeNetwork
	}
	return false
}

// IsServerError returns true if the error is a server error
func IsServerError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Type == ErrorTypeServer
	}
	return false
}

// NewAuthError creates a new authentication error
func NewAuthError(message string, cause error) *APIError {
	return &APIError{
		Type:    ErrorTypeAuth,
		Message: message,
		Cause:   cause,
	}
}

// NewNetworkError creates a new network error
func NewNetworkError(message string, cause error) *APIError {
	return &APIError{
		Type:    ErrorTypeNetwork,
		Message: message,
		Cause:   cause,
	}
}

// NewServerError creates a new server error
func NewServerError(statusCode int, message string, cause error) *APIError {
	return &APIError{
		Type:       ErrorTypeServer,
		StatusCode: statusCode,
		Message:    message,
		Cause:      cause,
	}
}

// NewValidationError creates a new validation error
func NewValidationError(message string, cause error) *APIError {
	return &APIError{
		Type:    ErrorTypeValidation,
		Message: message,
		Cause:   cause,
	}
}

// NewTimeoutError creates a new timeout error
func NewTimeoutError(message string, cause error) *APIError {
	return &APIError{
		Type:    ErrorTypeTimeout,
		Message: message,
		Cause:   cause,
	}
}

// WrapHTTPError wraps an HTTP response into an appropriate APIError
func WrapHTTPError(resp *http.Response, cause error) *APIError {
	switch resp.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return NewAuthError("authentication failed", cause)
	case http.StatusBadRequest:
		return NewValidationError("invalid request", cause)
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		return NewServerError(resp.StatusCode, "server error", cause)
	default:
		return &APIError{
			Type:       ErrorTypeUnknown,
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
			Cause:      cause,
		}
	}
}

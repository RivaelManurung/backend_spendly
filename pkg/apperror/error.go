package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError is a structured error that can be mapped to HTTP status codes.
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// New creates a new AppError.
func New(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Common errors
func NotFound(message string, err error) *AppError {
	if message == "" {
		message = "Resource not found"
	}
	return New(http.StatusNotFound, message, err)
}

func BadRequest(message string, err error) *AppError {
	if message == "" {
		message = "Bad request"
	}
	return New(http.StatusBadRequest, message, err)
}

func Unauthorized(message string, err error) *AppError {
	if message == "" {
		message = "Unauthorized"
	}
	return New(http.StatusUnauthorized, message, err)
}

func Forbidden(message string, err error) *AppError {
	if message == "" {
		message = "Forbidden"
	}
	return New(http.StatusForbidden, message, err)
}

func Internal(message string, err error) *AppError {
	if message == "" {
		message = "Internal server error"
	}
	return New(http.StatusInternalServerError, message, err)
}

func ValidationError(message string, err error) *AppError {
	if message == "" {
		message = "Validation error"
	}
	return New(http.StatusBadRequest, message, err)
}

// Error definitions for easy comparison
var (
	ErrNotFound      = errors.New("not_found")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
	ErrBadRequest    = errors.New("bad_request")
	ErrInternal      = errors.New("internal_error")
)

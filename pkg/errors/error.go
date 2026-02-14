package errors

import (
	"errors"
	"fmt"
	"strings"
)

const (
	StatusOK                  = 200
	StatusCreated             = 201
	StatusNoContent           = 204
	StatusBadRequest          = 400
	StatusUnauthorized        = 401
	StatusForbidden           = 403
	StatusNotFound            = 404
	StatusRequestTimeout      = 408
	StatusMethodNotAllowed    = 405
	StatusConflict            = 409
	StatusTooManyRequests     = 429
	StatusInternalServerError = 500
)

const (
	ErrorTypeDatabaseError       = "DATABASE_ERROR"
	ErrorTypeNotFound            = "NOT_FOUND"
	ErrorTypeInvalidRequest      = "INVALID_REQUEST"
	ErrorTypeUnauthorized        = "UNAUTHORIZED"
	ErrorTypeForbidden           = "FORBIDDEN"
	ErrorTypeConflict            = "CONFLICT"
	ErrorTypeInternalServerError = "INTERNAL_SERVER_ERROR"
	ErrorTypeUnknown             = "UNKNOWN_ERROR"
	ErrorTypeNoContent           = "NO_CONTENT"
	ErrorTypeTooManyRequests     = "TOO_MANY_REQUESTS"
	ErrorTypeRateLimitExceeded   = "RATE_LIMIT_EXCEEDED"
	ErrorTypeRequestTimeout      = "REQUEST_TIMEOUT"
	ErrorTypeMethodNotAllowed    = "METHOD_NOT_ALLOWED"
)

type Error struct {
	Type    string
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func New(errType, message string, err error) *Error {
	return &Error{
		Type:    errType,
		Message: message,
		Err:     err,
	}
}

func NewNotFoundError(message string, err error) *Error {
	return New(ErrorTypeNotFound, message, err)
}

func NewInvalidRequestError(message string, err error) *Error {
	return New(ErrorTypeInvalidRequest, message, err)
}

func NewDatabaseError(message string, err error) *Error {
	return New(ErrorTypeDatabaseError, message, err)
}

func NewConflictError(message string, err error) *Error {
	return New(ErrorTypeConflict, message, err)
}

func NewUnauthorizedError(message string, err error) *Error {
	return New(ErrorTypeUnauthorized, message, err)
}

func NewForbiddenError(message string, err error) *Error {
	return New(ErrorTypeForbidden, message, err)
}

func NewInternalServerError(message string, err error) *Error {
	return New(ErrorTypeInternalServerError, message, err)
}

func NewNoContentError(message string, err error) *Error {
	return New(ErrorTypeNoContent, message, err)
}

func Type(err error) string {
	if err == nil {
		return ""
	}

	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr.Type
	}

	return ErrorTypeUnknown
}

func DeduceErrorTypeFromErrorString(err error) string {
	if err == nil {
		return ""
	}

	errMsg := err.Error()
	switch {
	case errMsg == "":
		return ""
	case strings.Contains(strings.ToLower(errMsg), strings.ToLower("not found")):
		return ErrorTypeNotFound
	case strings.Contains(strings.ToLower(errMsg), strings.ToLower("unauthorized")):
		return ErrorTypeUnauthorized
	case strings.Contains(strings.ToLower(errMsg), strings.ToLower("forbidden")):
		return ErrorTypeForbidden
	case strings.Contains(strings.ToLower(errMsg), strings.ToLower("conflict")):
		return ErrorTypeConflict
	case strings.Contains(strings.ToLower(errMsg), strings.ToLower("database")):
		return ErrorTypeDatabaseError
	case strings.Contains(strings.ToLower(errMsg), strings.ToLower("invalid request")):
		return ErrorTypeInvalidRequest
	case strings.Contains(strings.ToLower(errMsg), strings.ToLower("no content")):
		return ErrorTypeNoContent
	}

	return ErrorTypeUnknown
}

func IsDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return DeduceErrorTypeFromErrorString(err) == ErrorTypeConflict ||
		strings.Contains(strings.ToLower(errMsg), strings.ToLower("duplicate")) ||
		strings.Contains(strings.ToLower(errMsg), strings.ToLower("unique constraint")) ||
		strings.Contains(strings.ToLower(errMsg), strings.ToLower("UNIQUE constraint failed"))
}

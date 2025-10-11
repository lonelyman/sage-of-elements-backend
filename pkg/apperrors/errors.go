package apperrors

import "github.com/gofiber/fiber/v2"

// ====================================================================================
// Standard Error Codes
// ====================================================================================
const (
	// Authentication
	ErrUnauthorized     = "UNAUTHORIZED"
	ErrInvalidToken     = "INVALID_TOKEN"
	ErrTokenExpired     = "TOKEN_EXPIRED"
	ErrPermissionDenied = "PERMISSION_DENIED"

	// Validation
	ErrValidation    = "VALIDATION_ERROR"
	ErrMissingParam  = "MISSING_PARAMETER"
	ErrInvalidFormat = "INVALID_FORMAT"

	// Resource
	ErrNotFound      = "NOT_FOUND"
	ErrAlreadyExists = "ALREADY_EXISTS"

	// System
	ErrSystem      = "SYSTEM_ERROR"
	ErrExternalAPI = "EXTERNAL_API_ERROR"
)

// ====================================================================================
// AppError Struct
// ====================================================================================
type AppError struct {
	HTTPStatus int         `json:"-"`
	Code       string      `json:"code"`
	Message    string      `json:"message"`
	Details    interface{} `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	return e.Message
}

// ====================================================================================
// Base Constructors
// ====================================================================================

// New creates a new AppError without details.
func New(status int, code, message string) *AppError {
	return &AppError{HTTPStatus: status, Code: code, Message: message}
}

// NewWithDetails creates a new AppError with details.
func NewWithDetails(status int, code, message string, details interface{}) *AppError {
	return &AppError{HTTPStatus: status, Code: code, Message: message, Details: details}
}

// ====================================================================================
// Helper Functions (for easy use in Service layer)
// ====================================================================================

// --- Authentication Errors ---

func UnauthorizedError(message string) *AppError {
	return New(fiber.StatusUnauthorized, ErrUnauthorized, message) // 401
}

func InvalidTokenError(message string, details interface{}) *AppError {
	return NewWithDetails(fiber.StatusUnauthorized, ErrInvalidToken, message, details) // 401
}

func PermissionDeniedError(message string) *AppError {
	return New(fiber.StatusForbidden, ErrPermissionDenied, message) // 403
}

// --- Validation Errors ---

func ValidationError(message string, details interface{}) *AppError {
	return NewWithDetails(fiber.StatusBadRequest, ErrValidation, message, details) // 400
}

func InvalidFormatError(message string, details interface{}) *AppError {
	return NewWithDetails(fiber.StatusBadRequest, ErrInvalidFormat, message, details) // 400
}

// --- Resource Errors ---

func NotFoundError(message string) *AppError {
	return New(fiber.StatusNotFound, ErrNotFound, message) // 404
}

func AlreadyExistsError(message string, details interface{}) *AppError {
	return NewWithDetails(fiber.StatusConflict, ErrAlreadyExists, message, details) // 409
}

// --- System Errors ---

// SystemError is for generic internal errors with a user-friendly message.
func SystemError(message string) *AppError {
	return New(fiber.StatusInternalServerError, ErrSystem, message) // 500
}

// SystemErrorWithDetails is for internal errors where we want to log the original error.
func SystemErrorWithDetails(message string, details interface{}) *AppError {
	return NewWithDetails(fiber.StatusInternalServerError, ErrSystem, message, details) // 500
}

// ExternalAPIError is for errors when calling third-party services.
func ExternalAPIError(message string, details interface{}) *AppError {
	return NewWithDetails(fiber.StatusBadGateway, ErrExternalAPI, message, details) // 502
}

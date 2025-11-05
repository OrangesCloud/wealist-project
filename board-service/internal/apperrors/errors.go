package apperrors

import "fmt"

// AppError represents an application-specific error
type AppError struct {
	Code       string // Error code (e.g., "UNAUTHORIZED", "NOT_FOUND")
	Message    string // User-friendly message
	HTTPStatus int    // HTTP status code
	Err        error  // Original error (for logging)
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError
func New(code, message string, status int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: status,
	}
}

// Wrap creates a new AppError wrapping an existing error
func Wrap(err error, code, message string, status int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: status,
		Err:        err,
	}
}

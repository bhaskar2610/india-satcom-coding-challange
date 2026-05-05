package errors

import "errors"

// Sentinel errors used across layers so callers can type-switch or compare
// without importing concrete error types.
var (
	ErrTaskNotFound      = errors.New("task not found")
	ErrValidation        = errors.New("validation error")
	ErrInternalServer    = errors.New("internal server error")
)

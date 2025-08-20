// Package errors provides custom error types for handling retry logic and permanent failures.
// It defines two main error types:
// - PermanentError: indicates that an operation should not be retried
// - RetryAfterError: indicates that an operation should be retried after a specified duration

package errors

import (
	"fmt"
	"time"
)

// WithErrorCode is an interface that allows errors to be associated with an ErrorCode.
// It provides a method to retrieve the error code associated with the error.
type WithErrorCode interface {
	GetCode() ErrorCode
}

var _ WithErrorCode = &ErrorCodeWrapper{}
var _ WithErrorCode = &ErrorWithCode{}

// ErrorCodeWrapper wraps an error with an associated ErrorCode.
type ErrorCodeWrapper struct {
	code ErrorCode // The error code associated with the error
	err  error     // The underlying error
}

func (w *ErrorCodeWrapper) Error() string {
	return fmt.Sprintf("[ERR,%s]: %s", w.code.String(), w.err.Error())
}

func (w *ErrorCodeWrapper) GetCode() ErrorCode {
	return w.code
}

// ErrorWithCode is a custom error type that includes an error code.
type ErrorWithCode struct {
	code    ErrorCode
	message string
}

func (e *ErrorWithCode) Error() string {
	return fmt.Sprintf("[ERR,%s]: %s", e.code.String(), e.message)
}

func (w *ErrorWithCode) GetCode() ErrorCode {
	return w.code
}

// PermanentError signals that the operation should not be retried.
// This error type is used when an operation fails due to a permanent condition
// (e.g., invalid input, authentication failure) that cannot be resolved by retrying.
type PermanentError struct {
	err error // The underlying error that caused the permanent failure
}

// Error returns a string representation of the Permanent error.
// It delegates to the underlying error's Error method.
func (e *PermanentError) Error() string {
	return fmt.Sprintf("[ERR|Permanent]: %s", e.err.Error())
}

// RetryAfterError signals that the operation should be retried after the given duration.
// This error type is used when an operation fails but should be retried after a specific
// time period (e.g., rate limiting, temporary server unavailability).
type RetryAfterError struct {
	err      error         // The underlying error that caused the retry
	duration time.Duration // The duration to wait before retrying
}

// Error returns a string representation of the RetryAfter error.
// It formats the error message to include the retry duration.
func (e *RetryAfterError) Error() string {
	return fmt.Sprintf("[ERR|Retry After %s]: %s", e.duration, e.err.Error())
}

// Duration returns the duration to wait before retrying the operation.
func (e *RetryAfterError) Duration() time.Duration {
	return e.duration
}

// APIError represents an error returned by the provider's API.
type APIError struct {
	StatusCode int    // HTTP status code
	Message    string // Error message from the API
	Err        error  // Original error if any
}

// Error returns a formatted error message.
func (e *APIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("API Error: Status=%d, Message='%s', OriginalErr=%v", e.StatusCode, e.Message, e.Err)
	}
	return fmt.Sprintf("API Error: Status=%d, Message='%s'", e.StatusCode, e.Message)
}

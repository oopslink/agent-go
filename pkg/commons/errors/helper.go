package errors

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

// New create a new ErrorWithCode with the specified code
func New(code ErrorCode) error {
	return &ErrorWithCode{
		code:    code,
		message: code.DefaultMessage,
	}
}

// Errorf creates a new ErrorWithCode with the specified code and formatted message.
// It uses fmt.Sprintf to format the message with the provided arguments.
func Errorf(code ErrorCode, format string, args ...interface{}) error {
	return &ErrorWithCode{
		code:    code,
		message: fmt.Sprintf(format, args...),
	}
}

// Wrap creates a new error that wraps the given error with an ErrorCode.
func Wrap(code ErrorCode, err error) error {
	return &ErrorCodeWrapper{
		code: code,
		err:  err,
	}
}

// Unwrap retrieves the underlying error from a wrapped error.
// It checks for various error types that may contain an underlying error.
func Unwrap(err error) error {
	if err == nil {
		return nil
	}

	var withErrorCodeErr *ErrorCodeWrapper
	if errors.As(err, &withErrorCodeErr) {
		return withErrorCodeErr.err
	}

	var permanentError *PermanentError
	if errors.As(err, &permanentError) {
		return permanentError.err
	}

	var retryAfterError *RetryAfterError
	if errors.As(err, &retryAfterError) {
		return retryAfterError.err
	}

	var apiError *APIError
	if errors.As(err, &apiError) {
		return apiError.Err
	}

	return err
}

// GetErrorCode extracts the ErrorCode from an error.
// If the error does not have an associated code, it returns NoErrorCode.
func GetErrorCode(err error) ErrorCode {
	if err == nil {
		return NoErrorCode
	}

	withErrorCode, ok := err.(WithErrorCode)
	if ok {
		return withErrorCode.GetCode()
	}

	return NoErrorCode
}

func Is(err1, err2 error) bool {
	return errors.Is(err1, err2)
}

// IsCode checks if the error matches the given error code.
func IsCode(err error, code ErrorCode) bool {
	if err == nil {
		return false
	}

	errorCode := GetErrorCode(err)
	if errorCode.Equal(code) {
		return true
	}
	if code.Name == "" && errorCode.Code == code.Code {
		return true
	}
	if code.Name != "" && errorCode.Name == code.Name {
		return true
	}
	if code.IsZero() && errorCode.IsZero() {
		return true
	}
	return false
}

// Permanent wraps the given err in a *PermanentError.
// If err is nil, it returns nil. This function is used to mark errors as permanent
// so that retry mechanisms will not attempt to retry the operation.
func Permanent(err error) error {
	if err == nil {
		return nil
	}
	return &PermanentError{
		err: err,
	}
}

// IsPermanent checks if the given error is a PermanentError.
func IsPermanent(err error) bool {
	_, ok := err.(*PermanentError)
	return ok
}

// RetryAfter returns a RetryAfter error that specifies how long to wait before retrying.
// The duration is specified in seconds and converted to time.Duration.
func RetryAfter(err error, seconds int) error {
	return &RetryAfterError{
		err:      err,
		duration: time.Duration(seconds) * time.Second,
	}
}

// IsRetryAfter checks if the given error is a RetryAfterError.
func IsRetryAfter(err error) bool {
	_, ok := err.(*RetryAfterError)
	return ok
}

// IsRetryableError determines if an error should trigger a retry.
// It checks for common retryable conditions like network timeouts,
// rate limiting, and server errors.
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	var apiErr *APIError
	if errors.As(err, &apiErr) {
		switch apiErr.StatusCode {
		case http.StatusConflict, http.StatusTooManyRequests,
			http.StatusInternalServerError, http.StatusBadGateway,
			http.StatusServiceUnavailable, http.StatusGatewayTimeout:
			return true
		default:
			return false
		}
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	var retryAfterErr *RetryAfterError
	if errors.As(err, &retryAfterErr) {
		return true
	}

	// TODO: check other errors

	return false
}

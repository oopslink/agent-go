package errors

import "fmt"

// ErrorCode represents an error with both numeric and string codes
type ErrorCode struct {
	Code           int64  // Numeric error code
	Name           string // String error code
	DefaultMessage string // String default message
}

// Predefined error codes
var (
	// NoErrorCode is a special value indicating no error code is set.
	NoErrorCode = ErrorCode{0, "NO_ERROR", "No error code"}

	// Generic error categories (1000-1999)
	UnknownError   = ErrorCode{1000, "UNKNOWN_ERROR", "Unknown error"}
	InternalError  = ErrorCode{1001, "INTERNAL_ERROR", "internal error"}
	NotImplemented = ErrorCode{1002, "NOT_IMPLEMENTED", "Not implemented"}
	Unavailable    = ErrorCode{1003, "UNAVAILABLE", "Unavailable"}

	// Authentication & Authorization (2000-2999)
	Unauthorized = ErrorCode{2001, "UNAUTHORIZED", "Unauthorized"}
	Forbidden    = ErrorCode{2002, "FORBIDDEN", "Forbidden"}
	AccessDenied = ErrorCode{2003, "ACCESS_DENIED", "Access denied"}

	// Input & Validation (3000-3999)
	InvalidInput    = ErrorCode{3001, "INVALID_INPUT", "Invalid input"}
	InvalidFormat   = ErrorCode{3002, "INVALID_FORMAT", "Invalid format"}
	MissingRequired = ErrorCode{3003, "MISSING_REQUIRED", "Missing required"}
	OutOfRange      = ErrorCode{3004, "OUT_OF_RANGE", "Out of range"}

	// Resource Management (4000-4999)
	NotFound      = ErrorCode{4001, "NOT_FOUND", "Not found"}
	AlreadyExists = ErrorCode{4002, "ALREADY_EXISTS", "Already exists"}
	Conflict      = ErrorCode{4003, "CONFLICT", "Conflict"}
	ResourceBusy  = ErrorCode{4004, "RESOURCE_BUSY", "Resource busy"}

	// Operation State (5000-5999)
	Timeout   = ErrorCode{5001, "TIMEOUT", "Timeout"}
	Cancelled = ErrorCode{5002, "CANCELLED", "Cancelled"}
	Failed    = ErrorCode{5003, "FAILED", "Failed"}
	Retry     = ErrorCode{5004, "RETRY", "Retry"}

	// Rate & Quota (6000-6999)
	RateLimited   = ErrorCode{6001, "RATE_LIMITED", "Rate limited"}
	QuotaExceeded = ErrorCode{6002, "QUOTA_EXCEEDED", "Quota exceeded"}
)

// String returns the string representation of the error code
func (ec *ErrorCode) String() string {
	if ec.Name != "" {
		return ec.Name
	}
	return fmt.Sprintf("ERROR_%d", ec.Code)
}

// Equal checks if two error codes are equal
func (ec *ErrorCode) Equal(other ErrorCode) bool {
	return ec.Code == other.Code
}

// IsZero checks if the error code is unset
func (ec *ErrorCode) IsZero() bool {
	return ec.Code == 0 && ec.Name == ""
}

// NewErrorCode creates a new error code with the given numeric and string codes
func NewErrorCode(code int64, name, message string) ErrorCode {
	return ErrorCode{
		Code:           code,
		Name:           name,
		DefaultMessage: message,
	}
}

// GetCode returns the numeric code
func (ec *ErrorCode) GetCode() int64 {
	return ec.Code
}

// GetName returns the string code
func (ec *ErrorCode) GetName() string {
	return ec.Name
}

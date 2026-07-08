// Package web: error code constants and structured error types.
package web

// Error code constants
const (
	CodeOK                 = 200
	CodeBadRequest         = 400
	CodeUnauthorized       = 401
	CodeForbidden          = 403
	CodeNotFound           = 404
	CodeMethodNotAllowed   = 405
	CodeTooManyRequests    = 429
	CodeInternalError      = 500
	CodeServiceUnavailable = 503
	// Business error codes (1000+)
	CodeValidationFailed = 1001
	CodeDuplicateEntry   = 1002
	CodeTokenExpired     = 1003
	CodeTokenInvalid     = 1004
)

// Pre-defined ErrorCode instances. Use with errors.Is/As or pass to Error().
var (
	ErrBadRequest         = NewErrorCode(CodeBadRequest, "bad request")
	ErrUnauthorized       = NewErrorCode(CodeUnauthorized, "unauthorized")
	ErrForbidden          = NewErrorCode(CodeForbidden, "forbidden")
	ErrNotFound           = NewErrorCode(CodeNotFound, "not found")
	ErrMethodNotAllowed   = NewErrorCode(CodeMethodNotAllowed, "method not allowed")
	ErrTooManyRequests    = NewErrorCode(CodeTooManyRequests, "too many requests")
	ErrInternalError      = NewErrorCode(CodeInternalError, "internal server error")
	ErrServiceUnavailable = NewErrorCode(CodeServiceUnavailable, "service unavailable")
	ErrValidationFailed   = NewErrorCode(CodeValidationFailed, "validation failed")
	ErrDuplicateEntry     = NewErrorCode(CodeDuplicateEntry, "duplicate entry")
	ErrTokenExpired       = NewErrorCode(CodeTokenExpired, "token expired")
	ErrTokenInvalid       = NewErrorCode(CodeTokenInvalid, "token invalid")
)

// ErrorCode is a structured error that can be used as Message.Data.
type ErrorCode struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
	Err     error  `json:"-"`
}

// Error implements the error interface.
func (e *ErrorCode) Error() string {
	if e.Detail != "" {
		return e.Detail
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	if e.Message != "" {
		return e.Message
	}
	return ""
}

// NewErrorCode creates a new ErrorCode with the given code and message.
func NewErrorCode(code int, msg string) *ErrorCode {
	return &ErrorCode{Code: code, Message: msg}
}

// NewErrorCodeWithError creates a new ErrorCode with the given code and wraps the original error.
func NewErrorCodeWithError(code int, err error) *ErrorCode {
	return &ErrorCode{Code: code, Err: err}
}

// Unwrap returns the wrapped error, supporting errors.Is/As chains.
func (e *ErrorCode) Unwrap() error {
	return e.Err
}

// WithDetail attaches additional detail information to the ErrorCode.
func (e *ErrorCode) WithDetail(detail string) *ErrorCode {
	e.Detail = detail
	return e
}

// WithError wraps the original error into the ErrorCode.
func (e *ErrorCode) WithError(err error) *ErrorCode {
	e.Err = err
	return e
}

// NewBadRequest returns a new 400 Bad Request error (copy of ErrBadRequest).
func NewBadRequest() *ErrorCode { return ErrBadRequest.clone() }

// NewUnauthorized returns a new 401 Unauthorized error (copy of ErrUnauthorized).
func NewUnauthorized() *ErrorCode { return ErrUnauthorized.clone() }

// NewForbidden returns a new 403 Forbidden error (copy of ErrForbidden).
func NewForbidden() *ErrorCode { return ErrForbidden.clone() }

// NewNotFound returns a new 404 Not Found error (copy of ErrNotFound).
func NewNotFound() *ErrorCode { return ErrNotFound.clone() }

// NewMethodNotAllowed returns a new 405 Method Not Allowed error (copy of ErrMethodNotAllowed).
func NewMethodNotAllowed() *ErrorCode { return ErrMethodNotAllowed.clone() }

// NewTooManyRequests returns a new 429 Too Many Requests error (copy of ErrTooManyRequests).
func NewTooManyRequests() *ErrorCode { return ErrTooManyRequests.clone() }

// NewInternalError returns a new 500 Internal Server Error error (copy of ErrInternalError).
func NewInternalError() *ErrorCode { return ErrInternalError.clone() }

// NewServiceUnavailable returns a new 503 Service Unavailable error (copy of ErrServiceUnavailable).
func NewServiceUnavailable() *ErrorCode { return ErrServiceUnavailable.clone() }

// NewValidationError returns a new 1001 Validation Failed error (copy of ErrValidationFailed).
func NewValidationError() *ErrorCode { return ErrValidationFailed.clone() }

// NewDuplicateEntry returns a new 1002 Duplicate Entry error (copy of ErrDuplicateEntry).
func NewDuplicateEntry() *ErrorCode { return ErrDuplicateEntry.clone() }

// NewTokenExpired returns a new 1003 Token Expired error (copy of ErrTokenExpired).
func NewTokenExpired() *ErrorCode { return ErrTokenExpired.clone() }

// NewTokenInvalid returns a new 1004 Token Invalid error (copy of ErrTokenInvalid).
func NewTokenInvalid() *ErrorCode { return ErrTokenInvalid.clone() }

// clone returns a shallow copy of the ErrorCode.
func (e *ErrorCode) clone() *ErrorCode {
	cp := *e
	return &cp
}

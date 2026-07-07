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

// ErrorCode is a structured error that can be used as Message.Data.
type ErrorCode struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

// Error implements the error interface.
func (e *ErrorCode) Error() string {
	return e.Message
}

// NewErrorCode creates a new ErrorCode with the given code and message.
func NewErrorCode(code int, msg string) *ErrorCode {
	return &ErrorCode{Code: code, Message: msg}
}

// WithDetail attaches additional detail information to the ErrorCode.
func (e *ErrorCode) WithDetail(detail string) *ErrorCode {
	e.Detail = detail
	return e
}

// ToMessage converts ErrorCode to *Message.
func (e *ErrorCode) ToMessage() *Message {
	return &Message{
		Code: e.Code,
		Data: e,
	}
}

// NewBadRequest creates a 400 Bad Request response message.
func NewBadRequest(msg string) *Message {
	return &Message{Code: CodeBadRequest, Data: msg}
}

// NewBadRequestWithDetail creates a 400 Bad Request response with additional detail.
func NewBadRequestWithDetail(msg, detail string) *Message {
	return &Message{Code: CodeBadRequest, Data: NewErrorCode(CodeBadRequest, msg).WithDetail(detail)}
}

// NewUnauthorized creates a 401 Unauthorized response message.
func NewUnauthorized(msg string) *Message {
	return &Message{Code: CodeUnauthorized, Data: msg}
}

// NewForbidden creates a 403 Forbidden response message.
func NewForbidden(msg string) *Message {
	return &Message{Code: CodeForbidden, Data: msg}
}

// NewNotFound creates a 404 Not Found response message.
func NewNotFound(msg string) *Message {
	return &Message{Code: CodeNotFound, Data: msg}
}

// NewMethodNotAllowed creates a 405 Method Not Allowed response message.
func NewMethodNotAllowed(msg string) *Message {
	return &Message{Code: CodeMethodNotAllowed, Data: msg}
}

// NewTooManyRequests creates a 429 Too Many Requests response message.
func NewTooManyRequests(msg string) *Message {
	return &Message{Code: CodeTooManyRequests, Data: msg}
}

// NewInternalError creates a 500 Internal Server Error response message.
func NewInternalError(msg string) *Message {
	return &Message{Code: CodeInternalError, Data: msg}
}

// NewServiceUnavailable creates a 503 Service Unavailable response message.
func NewServiceUnavailable(msg string) *Message {
	return &Message{Code: CodeServiceUnavailable, Data: msg}
}

// NewValidationError creates a 1001 Validation Failed response message.
func NewValidationError(msg string) *Message {
	return &Message{Code: CodeValidationFailed, Data: msg}
}

// NewDuplicateEntry creates a 1002 Duplicate Entry response message.
func NewDuplicateEntry(msg string) *Message {
	return &Message{Code: CodeDuplicateEntry, Data: msg}
}

// NewTokenExpired creates a 1003 Token Expired response message.
func NewTokenExpired(msg string) *Message {
	return &Message{Code: CodeTokenExpired, Data: msg}
}

// NewTokenInvalid creates a 1004 Token Invalid response message.
func NewTokenInvalid(msg string) *Message {
	return &Message{Code: CodeTokenInvalid, Data: msg}
}

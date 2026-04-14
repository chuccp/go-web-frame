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
	CodeTokenInvalid   = 1004
)

// ErrorCode is a structured error that can be used as Message.Data
type ErrorCode struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

func (e *ErrorCode) Error() string {
	return e.Message
}

// NewErrorCode creates a new ErrorCode
func NewErrorCode(code int, msg string) *ErrorCode {
	return &ErrorCode{Code: code, Message: msg}
}

func (e *ErrorCode) WithDetail(detail string) *ErrorCode {
	e.Detail = detail
	return e
}

// ToMessage converts ErrorCode to *Message
func (e *ErrorCode) ToMessage() *Message {
	return &Message{
		Code: e.Code,
		Msg:  e.Message,
		Data: e,
	}
}

// Constructor functions that return *Message directly

func NewBadRequest(msg string) *Message {
	return &Message{Code: CodeBadRequest, Msg: msg}
}

func NewBadRequestWithDetail(msg, detail string) *Message {
	return &Message{Code: CodeBadRequest, Msg: msg, Data: NewErrorCode(CodeBadRequest, msg).WithDetail(detail)}
}

func NewUnauthorized(msg string) *Message {
	return &Message{Code: CodeUnauthorized, Msg: msg}
}

func NewForbidden(msg string) *Message {
	return &Message{Code: CodeForbidden, Msg: msg}
}

func NewNotFound(msg string) *Message {
	return &Message{Code: CodeNotFound, Msg: msg}
}

func NewMethodNotAllowed(msg string) *Message {
	return &Message{Code: CodeMethodNotAllowed, Msg: msg}
}

func NewTooManyRequests(msg string) *Message {
	return &Message{Code: CodeTooManyRequests, Msg: msg}
}

func NewInternalError(msg string) *Message {
	return &Message{Code: CodeInternalError, Msg: msg}
}

func NewServiceUnavailable(msg string) *Message {
	return &Message{Code: CodeServiceUnavailable, Msg: msg}
}

func NewValidationError(msg string) *Message {
	return &Message{Code: CodeValidationFailed, Msg: msg}
}

func NewDuplicateEntry(msg string) *Message {
	return &Message{Code: CodeDuplicateEntry, Msg: msg}
}

func NewTokenExpired(msg string) *Message {
	return &Message{Code: CodeTokenExpired, Msg: msg}
}

func NewTokenInvalid(msg string) *Message {
	return &Message{Code: CodeTokenInvalid, Msg: msg}
}

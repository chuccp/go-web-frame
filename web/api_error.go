package web

import "strings"

// APIErrorType identifies every standardized API error response.
const APIErrorType = "api_error"

// IsErrorCode reports whether a message code represents an HTTP or business error.
func IsErrorCode(code int) bool {
	return code >= CodeBadRequest
}

// NewErrorMessage creates the standard API error response:
// {"code": 400, "data": null, "msg": "...", "type": "api_error"}.
func NewErrorMessage(code int, message string) *Message {
	if code < CodeBadRequest && code < CodeValidationFailed {
		code = CodeInternalError
	}
	return &Message{
		Code: code,
		Data: nil,
		Msg:  NormalizeErrorMessage(code, message),
		Type: APIErrorType,
	}
}

// HTTPStatusForCode maps application/business error codes to a valid HTTP
// status while preserving the original code in the response body.
func HTTPStatusForCode(code int) int {
	if code >= CodeBadRequest && code <= 599 {
		return code
	}
	switch code {
	case CodeValidationFailed:
		return CodeBadRequest
	case CodeDuplicateEntry:
		return 409
	case CodeTokenExpired, CodeTokenInvalid:
		return CodeUnauthorized
	default:
		return CodeInternalError
	}
}

// NormalizeErrorMessage keeps specific business messages while replacing
// framework defaults, empty messages, technical 5xx details and the old
// "template rendering failed" response with user-facing text.
func NormalizeErrorMessage(code int, message string) string {
	message = strings.TrimSpace(message)
	if (code >= CodeInternalError && code < 600) || message == "" || isFrameworkDefaultMessage(code, message) {
		return defaultErrorMessage(code)
	}
	return message
}

func defaultErrorMessage(code int) string {
	switch code {
	case CodeBadRequest:
		return "请求参数不正确"
	case CodeUnauthorized:
		return "登录状态已失效，请重新登录"
	case CodeForbidden:
		return "没有访问权限"
	case CodeNotFound:
		return "请求的资源不存在"
	case CodeMethodNotAllowed:
		return "请求方法不支持"
	case CodeTooManyRequests:
		return "请求过于频繁，请稍后重试"
	case CodeValidationFailed:
		return "提交内容校验失败"
	case CodeDuplicateEntry:
		return "数据已存在，请勿重复提交"
	case CodeTokenExpired:
		return "登录状态已失效，请重新登录"
	case CodeTokenInvalid:
		return "登录凭证无效，请重新登录"
	case CodeServiceUnavailable:
		return "服务暂时不可用，请稍后重试"
	default:
		return "请求处理失败，请稍后重试"
	}
}

func isFrameworkDefaultMessage(code int, message string) bool {
	var expected string
	switch code {
	case CodeBadRequest:
		expected = "bad request"
	case CodeUnauthorized:
		expected = "unauthorized"
	case CodeForbidden:
		expected = "forbidden"
	case CodeNotFound:
		expected = "not found"
	case CodeMethodNotAllowed:
		expected = "method not allowed"
	case CodeTooManyRequests:
		expected = "too many requests"
	case CodeInternalError:
		expected = "internal server error"
	case CodeServiceUnavailable:
		expected = "service unavailable"
	case CodeValidationFailed:
		expected = "validation failed"
	case CodeDuplicateEntry:
		expected = "duplicate entry"
	case CodeTokenExpired:
		expected = "token expired"
	case CodeTokenInvalid:
		expected = "token invalid"
	}
	return expected != "" && strings.EqualFold(message, expected) || strings.EqualFold(message, "error")
}

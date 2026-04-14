package web

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorCode_NewErrorCode(t *testing.T) {
	err := NewErrorCode(CodeBadRequest, "invalid input")
	assert.Equal(t, CodeBadRequest, err.Code)
	assert.Equal(t, "invalid input", err.Message)
	assert.Empty(t, err.Detail)
}

func TestErrorCode_WithDetail(t *testing.T) {
	err := NewErrorCode(CodeBadRequest, "invalid input").WithDetail("field 'age' must be positive")
	assert.Equal(t, "field 'age' must be positive", err.Detail)
}

func TestErrorCode_Error(t *testing.T) {
	err := NewErrorCode(CodeInternalError, "something went wrong")
	assert.Equal(t, "something went wrong", err.Error())
}

func TestErrorCode_ToMessage(t *testing.T) {
	err := NewErrorCode(CodeNotFound, "user not found")
	msg := err.ToMessage()
	assert.Equal(t, CodeNotFound, msg.Code)
	assert.Equal(t, "user not found", msg.Msg)
	assert.Equal(t, err, msg.Data)
}

func TestNewBadRequest(t *testing.T) {
	msg := NewBadRequest("bad request")
	assert.Equal(t, CodeBadRequest, msg.Code)
	assert.Equal(t, "bad request", msg.Msg)
}

func TestNewBadRequestWithDetail(t *testing.T) {
	msg := NewBadRequestWithDetail("bad request", "detail info")
	assert.Equal(t, CodeBadRequest, msg.Code)
	assert.Equal(t, "bad request", msg.Msg)
	err, ok := msg.Data.(*ErrorCode)
	assert.True(t, ok)
	assert.Equal(t, "detail info", err.Detail)
}

func TestNewUnauthorized(t *testing.T) {
	msg := NewUnauthorized("unauthorized")
	assert.Equal(t, CodeUnauthorized, msg.Code)
}

func TestNewForbidden(t *testing.T) {
	msg := NewForbidden("forbidden")
	assert.Equal(t, CodeForbidden, msg.Code)
}

func TestNewNotFound(t *testing.T) {
	msg := NewNotFound("not found")
	assert.Equal(t, CodeNotFound, msg.Code)
}

func TestNewMethodNotAllowed(t *testing.T) {
	msg := NewMethodNotAllowed("method not allowed")
	assert.Equal(t, CodeMethodNotAllowed, msg.Code)
}

func TestNewTooManyRequests(t *testing.T) {
	msg := NewTooManyRequests("too many requests")
	assert.Equal(t, CodeTooManyRequests, msg.Code)
}

func TestNewInternalError(t *testing.T) {
	msg := NewInternalError("internal error")
	assert.Equal(t, CodeInternalError, msg.Code)
}

func TestNewServiceUnavailable(t *testing.T) {
	msg := NewServiceUnavailable("service unavailable")
	assert.Equal(t, CodeServiceUnavailable, msg.Code)
}

func TestNewValidationError(t *testing.T) {
	msg := NewValidationError("validation failed")
	assert.Equal(t, CodeValidationFailed, msg.Code)
}

func TestNewDuplicateEntry(t *testing.T) {
	msg := NewDuplicateEntry("duplicate entry")
	assert.Equal(t, CodeDuplicateEntry, msg.Code)
}

func TestNewTokenExpired(t *testing.T) {
	msg := NewTokenExpired("token expired")
	assert.Equal(t, CodeTokenExpired, msg.Code)
}

func TestNewTokenInvalid(t *testing.T) {
	msg := NewTokenInvalid("token invalid")
	assert.Equal(t, CodeTokenInvalid, msg.Code)
}

func TestErrorCodeConstants(t *testing.T) {
	assert.Equal(t, 200, CodeOK)
	assert.Equal(t, 400, CodeBadRequest)
	assert.Equal(t, 401, CodeUnauthorized)
	assert.Equal(t, 403, CodeForbidden)
	assert.Equal(t, 404, CodeNotFound)
	assert.Equal(t, 405, CodeMethodNotAllowed)
	assert.Equal(t, 429, CodeTooManyRequests)
	assert.Equal(t, 500, CodeInternalError)
	assert.Equal(t, 503, CodeServiceUnavailable)
	assert.Equal(t, 1001, CodeValidationFailed)
	assert.Equal(t, 1002, CodeDuplicateEntry)
	assert.Equal(t, 1003, CodeTokenExpired)
	assert.Equal(t, 1004, CodeTokenInvalid)
}

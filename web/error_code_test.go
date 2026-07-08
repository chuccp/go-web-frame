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

func TestNewBadRequest(t *testing.T) {
	err := NewBadRequest()
	assert.Equal(t, CodeBadRequest, err.Code)
	assert.Equal(t, "bad request", err.Message)
	assert.Empty(t, err.Detail)
}

func TestNewBadRequestWithDetail(t *testing.T) {
	err := NewBadRequest().WithDetail("detail info")
	assert.Equal(t, CodeBadRequest, err.Code)
	assert.Equal(t, "detail info", err.Detail)
}

func TestNewUnauthorized(t *testing.T) {
	err := NewUnauthorized()
	assert.Equal(t, CodeUnauthorized, err.Code)
	assert.Equal(t, "unauthorized", err.Message)
}

func TestNewForbidden(t *testing.T) {
	err := NewForbidden()
	assert.Equal(t, CodeForbidden, err.Code)
	assert.Equal(t, "forbidden", err.Message)
}

func TestNewNotFound(t *testing.T) {
	err := NewNotFound()
	assert.Equal(t, CodeNotFound, err.Code)
	assert.Equal(t, "not found", err.Message)
}

func TestNewMethodNotAllowed(t *testing.T) {
	err := NewMethodNotAllowed()
	assert.Equal(t, CodeMethodNotAllowed, err.Code)
	assert.Equal(t, "method not allowed", err.Message)
}

func TestNewTooManyRequests(t *testing.T) {
	err := NewTooManyRequests()
	assert.Equal(t, CodeTooManyRequests, err.Code)
	assert.Equal(t, "too many requests", err.Message)
}

func TestNewInternalError(t *testing.T) {
	err := NewInternalError()
	assert.Equal(t, CodeInternalError, err.Code)
	assert.Equal(t, "internal server error", err.Message)
}

func TestNewServiceUnavailable(t *testing.T) {
	err := NewServiceUnavailable()
	assert.Equal(t, CodeServiceUnavailable, err.Code)
	assert.Equal(t, "service unavailable", err.Message)
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError()
	assert.Equal(t, CodeValidationFailed, err.Code)
	assert.Equal(t, "validation failed", err.Message)
}

func TestNewDuplicateEntry(t *testing.T) {
	err := NewDuplicateEntry()
	assert.Equal(t, CodeDuplicateEntry, err.Code)
	assert.Equal(t, "duplicate entry", err.Message)
}

func TestNewTokenExpired(t *testing.T) {
	err := NewTokenExpired()
	assert.Equal(t, CodeTokenExpired, err.Code)
	assert.Equal(t, "token expired", err.Message)
}

func TestNewTokenInvalid(t *testing.T) {
	err := NewTokenInvalid()
	assert.Equal(t, CodeTokenInvalid, err.Code)
	assert.Equal(t, "token invalid", err.Message)
}

func TestNewReturnsCopy(t *testing.T) {
	a := NewNotFound()
	b := NewNotFound()
	a.WithDetail("a detail")
	assert.Empty(t, b.Detail, "clone should be independent")
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

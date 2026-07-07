package web

import (
	"fmt"
	"net/http"
)

type Message struct {
	Code int `json:"code"`
	Data any `json:"data"`
}

// IsOK reports whether the message represents a successful response (code 200).
func (m *Message) IsOK() bool {
	return m.Code == http.StatusOK
}

// Data creates a success response with the given data.
func Data(data any) *Message {
	return &Message{Code: http.StatusOK, Data: data}
}

// DataCode creates a response with the given code and data.
func DataCode(code int, data any) *Message {
	return &Message{Code: code, Data: data}
}

// Errors creates an error response. If msg is provided, it is used as the error detail;
// otherwise err.Error() is used.
func Errors(data any, msg ...error) *Message {
	m := &Message{Code: http.StatusInternalServerError, Data: data}
	if len(msg) > 0 && msg[0] != nil {
		m.Data = fmt.Sprintf("%v", msg[0])
	}
	return m
}

// ErrorMessage creates an error response with the given message string.
func ErrorMessage(msg ...string) *Message {
	m := &Message{Code: http.StatusInternalServerError}
	if len(msg) > 0 {
		m.Data = msg[0]
	}
	return m
}

// Error creates an error response from an error value.
func Error(err ...error) *Message {
	m := &Message{Code: http.StatusInternalServerError}
	if len(err) > 0 && err[0] != nil {
		m.Data = err[0].Error()
	}
	return m
}

// Redirect creates a redirect response.
func Redirect(url string) *Message {
	return &Message{
		Code: http.StatusMovedPermanently,
		Data: url,
	}
}

// Unauthorized creates a 401 response.
func Unauthorized(data any) *Message {
	return &Message{
		Code: http.StatusUnauthorized,
		Data: data,
	}
}

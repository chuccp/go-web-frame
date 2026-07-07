package web

import "net/http"

// Message is the standard response format for the framework.
type Message struct {
	Code int    `json:"code"`
	Data any    `json:"data"`
	Msg  string `json:"msg"`
	Type string `json:"type"`
}

// IsOK reports whether the message code indicates success (200).
func (msg *Message) IsOK() bool {
	return msg.Code == 200
}

// Ok creates a success response with an optional message.
func Ok(msg ...string) *Message {
	if len(msg) > 0 {
		return &Message{Code: 200, Msg: msg[0]}
	}
	return &Message{Code: 200, Msg: "ok"}
}

// Data creates a success response with the given data.
func Data(data any) *Message {
	return &Message{Code: 200, Msg: "ok", Data: data}
}

// DataType creates a success response with a custom type and data.
func DataType(t string, data any) *Message {
	return &Message{Type: t, Code: 200, Msg: "ok", Data: data}
}

// DataCode creates a response with a custom status code and data.
func DataCode(code int, data any) *Message {
	return &Message{Code: code, Msg: "ok", Data: data}
}

// ErrorMessage creates an error response with code 500 and the given message.
func ErrorMessage(msg ...string) *Message {
	m := "error"
	if len(msg) > 0 {
		m = msg[0]
	}
	return &Message{Code: 500, Msg: m}
}

// Error creates an error response with code 500 from the given error.
func Error(err ...error) *Message {
	m := "error"
	if len(err) > 0 {
		m = err[0].Error()
	}
	return &Message{Code: 500, Msg: m}
}

// Errors creates an error response with code 500, data, and an optional error message.
func Errors(data any, msg ...error) *Message {
	m := "error"
	if len(msg) > 0 {
		m = msg[0].Error()
	}
	return &Message{Code: 500, Msg: m, Data: data}
}

// Unauthorized creates a response with code 401, data, and an optional error message.
func Unauthorized(data any, msg ...error) *Message {
	m := "error"
	if len(msg) > 0 {
		m = msg[0].Error()
	}
	return &Message{Code: 401, Msg: m, Data: data}
}

// Redirect creates a redirect response with the given URL.
func Redirect(url string) *Message {
	return &Message{Code: http.StatusMovedPermanently, Msg: "redirect", Data: url}
}

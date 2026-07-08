// Package web: standard response message format with code, data, msg, and type.
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

// Redirect creates a redirect response with the given URL.
func Redirect(url string) *Message {
	return &Message{Code: http.StatusMovedPermanently, Msg: "redirect", Data: url}
}

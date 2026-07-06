package web2

import "net/http"

type Message struct {
	Code int `json:"code"`
	Data any `json:"data"`
}

func Redirect(url string) *Message {
	return &Message{
		Code: http.StatusMovedPermanently,
		Data: url,
	}
}

func Unauthorized(data any) *Message {
	return &Message{
		Code: http.StatusUnauthorized,
		Data: data,
	}
}

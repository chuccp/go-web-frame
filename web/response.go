package web

import (
	"github.com/gin-gonic/gin"
)

type Response gin.ResponseWriter

type ResponseWriteCloser struct {
	response Response
}

func (w *ResponseWriteCloser) Write(p []byte) (n int, err error) {
	return w.response.Write(p)
}
func (w *ResponseWriteCloser) Close() error {
	w.response.Flush()
	return nil
}

func CreateResponseWriteCloser(response Response) *ResponseWriteCloser {
	return &ResponseWriteCloser{
		response: response,
	}
}

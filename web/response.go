package web

import (
	"github.com/gin-gonic/gin"
)

type Response interface {
	gin.ResponseWriter
	SetAttachmentFileName(fileName string)
}

type response struct {
	gin.ResponseWriter
}

func (r *response) SetAttachmentFileName(fileName string) {
	r.Header().Set("Content-Disposition", `attachment; filename="`+fileName+`"`)
}
func newResponse(responseWriter gin.ResponseWriter) *response {
	return &response{
		ResponseWriter: responseWriter,
	}
}

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

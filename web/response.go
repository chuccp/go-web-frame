package web

import (
	"github.com/gin-gonic/gin"
)

type Response interface {
	gin.ResponseWriter
	SetAttachmentFileName(fileName string)
	JSON(code int, value any)
	Abort()
	Redirect(code int, location string)
	FileAttachment(path string, name string)
}

type response struct {
	gin.ResponseWriter
	ctx *gin.Context
}

func (r *response) SetAttachmentFileName(fileName string) {
	r.Header().Set("Content-Disposition", `attachment; filename="`+fileName+`"`)
}

func (r *response) JSON(code int, value any) {
	r.ctx.JSON(code, value)
}
func (r *response) Abort() {
	r.ctx.Abort()
}
func (r *response) Redirect(code int, location string) {
	r.ctx.Redirect(code, location)
}
func (r *response) FileAttachment(path string, name string) {
	r.ctx.FileAttachment(path, name)
}

func newResponse(ctx *gin.Context) *response {
	return &response{
		ResponseWriter: ctx.Writer,
		ctx:            ctx,
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

package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response is the interface for writing HTTP responses.
// It extends gin.ResponseWriter with framework-specific methods.
type Response interface {
	gin.ResponseWriter
	// SetAttachmentFileName sets the Content-Disposition header for file downloads.
	SetAttachmentFileName(fileName string)
	// JSON writes a JSON response with the given status code and value.
	JSON(code int, value any)
	// Abort stops further handler execution.
	Abort()
	// Redirect sends an HTTP redirect to the client.
	Redirect(code int, location string)
	// FileAttachment sends a file as an attachment download.
	FileAttachment(path string, name string)
	// WriteStatus writes only the HTTP status code.
	WriteStatus(code int)
	// Message writes a Message as a JSON response.
	Message(t *Message)
	// AbortWithMessage writes a Message and aborts the handler chain.
	AbortWithMessage(t *Message)
	// AbortWithStatusJSON writes a JSON response with the given status and aborts.
	AbortWithStatusJSON(i int, value any)
	// AbortWithError writes an error response and returns the error.
	AbortWithError(err error) error
}

type response struct {
	gin.ResponseWriter
	ctx *gin.Context
}

func (r *response) WriteStatus(code int) {
	r.ctx.Status(code)
}

func (r *response) AbortWithError(err error) error {
	er := r.ctx.AbortWithError(http.StatusInternalServerError, err)
	return er
}

func (r *response) AbortWithStatusJSON(i int, value any) {
	r.ctx.AbortWithStatusJSON(i, value)
}

func (r *response) Message(t *Message) {
	if t.Code == http.StatusMovedPermanently {
		r.ctx.Redirect(http.StatusMovedPermanently, t.Data.(string))
		r.ctx.Abort()
		return
	}
	r.ctx.JSON(t.Code, t)
}

func (r *response) IsAborted() bool {
	return r.ctx.IsAborted()
}

func (r *response) AbortWithMessage(t *Message) {
	if t.Code == http.StatusMovedPermanently {
		r.ctx.Redirect(http.StatusMovedPermanently, t.Data.(string))
		r.ctx.Abort()
		return
	}
	r.ctx.JSON(t.Code, t)
	r.ctx.Abort()
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

// ResponseWriteCloser wraps a Response to implement io.WriteCloser.
// Close flushes the response.
type ResponseWriteCloser struct {
	response Response
}

// Write writes bytes to the underlying response.
func (w *ResponseWriteCloser) Write(p []byte) (n int, err error) {
	return w.response.Write(p)
}

// Close flushes the response writer.
func (w *ResponseWriteCloser) Close() error {
	w.response.Flush()
	return nil
}

// CreateResponseWriteCloser creates a ResponseWriteCloser from a Response.
func CreateResponseWriteCloser(response Response) *ResponseWriteCloser {
	return &ResponseWriteCloser{
		response: response,
	}
}

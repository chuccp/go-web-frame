package web2

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/util"
	"github.com/coder/websocket"
	"go.uber.org/zap"
)

// Converter transforms handler return values into HTTP responses.
// It is called after the filter chain completes.
type Converter interface {
	// Request converts the handler result to an HTTP response.
	Request(filterChain FilterChain, request *Request)
}

var defaultConverter = &DefaultConverter{}

// DefaultConverter is the built-in converter that handles Message, string,
// *FileResponse, *os.File, and arbitrary values (converted to JSON).
type DefaultConverter struct {
}

// Request converts the handler result to the appropriate HTTP response.
// It handles errors, Message, string, file download, and JSON responses.
func (c *DefaultConverter) Request(filterChain FilterChain, request *Request) {
	value, err := filterChain.Next()
	if err == nil {
		if value != nil {
			switch t := value.(type) {
			case *Message:
				c.Message(request, t)
			case *FileResponse:
				c.FileResponse(request, t)
			case *FileSystemResponse:
				c.FileSystemResponse(request, t)
			case *SSEResponse:
				c.SSEResponse(request, t)
			case *WSResponse:
				c.WSResponse(request, t)
			case *os.File:
				c.RawFile(request, t)
			case string:
				c.String(request, t)
			case error:
				c.Error(request, value, err)
			}
		}
	} else {
		log.Debug("converter: handler error", zap.Error(err))
		c.Error(request, value, err)
		return
	}
}
func (c *DefaultConverter) Message(request *Request, value *Message) {
	if value.Code == http.StatusMovedPermanently {
		request.response.Redirect(http.StatusMovedPermanently, value.Data.(string))
		request.response.Abort()
		return
	}
	request.response.JSON(value.Code, value.Data)
}
func (c *DefaultConverter) FileResponse(request *Request, value *FileResponse) {
	if len(value.FileName) == 0 {
		_, filename := path.Split(value.Path)
		value.FileName = filename
	}
	if util.IsNotBlank(value.Suffix) && !strings.HasSuffix(value.FileName, value.Suffix) {
		if !strings.HasPrefix(value.Suffix, ".") {
			value.Suffix = "." + value.Suffix
		}
		value.FileName = value.FileName + value.Suffix
	}
	request.response.FileAttachment(value.Path, value.FileName)
}
func (c *DefaultConverter) FileSystemResponse(request *Request, value *FileSystemResponse) {
	request.GinContext().FileFromFS(value.Filepath, value.FS)
}
func (c *DefaultConverter) SSEResponse(request *Request, value *SSEResponse) {
	stream := NewSSEStream(request)
	stream.setHeaders()
	if err := value.Handler(stream); err != nil {
		log.Debug("converter: SSE handler error", zap.Error(err))
	}
}
func (c *DefaultConverter) WSResponse(request *Request, value *WSResponse) {
	conn, err := websocket.Accept(request.GinContext().Writer, request.Request(), nil)
	if err != nil {
		log.Debug("converter: WebSocket accept error", zap.Error(err))
		if abortErr := request.response.AbortWithError(err); abortErr != nil {
		log.Debug("converter: WebSocket abort error", zap.Error(abortErr))
	}
		return
	}
	stream := newWebSocketStream(request, conn)
	defer stream.Close()
	if err := value.Handler(stream); err != nil {
		log.Debug("converter: WebSocket handler error", zap.Error(err))
	}
}
func (c *DefaultConverter) RawFile(request *Request, value *os.File) {
	defer func() {
		if err := value.Close(); err != nil {
			log.Debug("converter: RawFile close error", zap.Error(err))
		}
	}()
	filename := filepath.Base(value.Name())
	request.response.FileAttachment(value.Name(), filename)
}
func (c *DefaultConverter) String(request *Request, value string) {
	if _, err := request.response.WriteString(value); err != nil {
		log.Debug("converter: WriteString error", zap.Error(err))
	}
}
func (c *DefaultConverter) Error(request *Request, value any, err error) {
	if abortErr := request.response.AbortWithError(err); abortErr != nil {
		log.Error("emptyConverter AbortWithError", zap.Error(abortErr))
	}
}

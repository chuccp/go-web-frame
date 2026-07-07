package web2

import (
	"errors"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/util"
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
			case *os.File:
				c.RawFile(request, t)
			case string:
				c.String(request, t)
			case error:
				c.Error(request, value, err)
			}
		}
	} else {
		c.Error(request, value, err)
		return
	}
}
func (c *DefaultConverter) Message(request *Request, value *Message) {
	resp := request.response
	if value.Code == http.StatusMovedPermanently {
		resp.Redirect(http.StatusMovedPermanently, value.Data.(string))
		resp.Abort()
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
func (c *DefaultConverter) RawFile(request *Request, value *os.File) {
	defer value.Close()
	fileInfo, err := value.Stat()
	if err != nil {
		c.Error(request, value, err)
		return
	}
	if fileInfo.IsDir() {
		c.Error(request, value, errors.New("cannot serve a directory as file"))
		return
	}
	filename := filepath.Base(value.Name())
	request.response.FileAttachment(value.Name(), filename)
}
func (c *DefaultConverter) String(request *Request, value string) {
	_, err := request.response.WriteString(value)
	if err != nil {
		return
	}
}
func (c *DefaultConverter) Error(request *Request, value any, err error) {
	resp := request.response
	err = resp.AbortWithError(err)
	if err != nil {
		log.Error("emptyConverter AbortWithError", zap.Error(err))
	}
}

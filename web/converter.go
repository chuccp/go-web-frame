package web

import (
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/chuccp/go-web-frame/util"
)

// Converter transforms handler return values into HTTP responses.
// It is called after the filter chain completes.
type Converter interface {
	// Request converts the handler result to an HTTP response.
	Request(filterChain FilterChain, request *Request)
}

// DefaultConverter is the built-in converter that handles Message, string,
// *File, *os.File, and arbitrary values (converted to JSON).
type DefaultConverter struct {
}

// Request converts the handler result to the appropriate HTTP response.
// It handles errors, Message, string, file download, and JSON responses.
func (c *DefaultConverter) Request(filterChain FilterChain, request *Request) {
	value, err := filterChain.Next()
	if err == nil {
		if e, ok := value.(error); ok {
			err = e
		}
	}
	resp := request.Response()
	if err != nil {
		err0 := Errors(value, err)
		resp.JSON(err0.Code, err0)
		resp.Abort()
	} else {
		if value != nil {
			switch t := value.(type) {
			case *Message:
				if t.Code == http.StatusMovedPermanently {
					resp.Redirect(http.StatusMovedPermanently, t.Data.(string))
					resp.Abort()
					return
				}
				resp.JSON(t.Code, value)
			case string:
				_, err2 := resp.Write([]byte(t))
				if err2 != nil {
					resp.Abort()
					return
				}
			case *File:
				if len(t.FileName) == 0 {
					_, filename := path.Split(t.Path)
					t.FileName = filename
				}
				if util.IsNotBlank(t.Suffix) && !strings.HasSuffix(t.FileName, t.Suffix) {
					if !strings.HasPrefix(t.Suffix, ".") {
						t.Suffix = "." + t.Suffix
					}
					t.FileName = t.FileName + t.Suffix
				}
				resp.FileAttachment(t.Path, t.FileName)
			case *os.File:
				resp.FileAttachment(t.Name(), t.Name())
			default:
				resp.JSON(200, Data(value))
			}
		}
	}
}

package web

import (
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/chuccp/go-web-frame/util"
)

type Converter func(value any, err error, ctx *HttpContext)

var DefaultConverter Converter = func(value any, err error, ctx *HttpContext) {
	if err != nil {
		err0 := Errors(value, err)
		ctx.JSON(err0.Code, err0)
		ctx.Abort()
	} else {
		if value != nil {
			switch t := value.(type) {
			case *Message:
				if t.Code == http.StatusMovedPermanently {
					ctx.Redirect(http.StatusMovedPermanently, t.Data.(string))
					ctx.Abort()
					return
				}
				ctx.JSON(t.Code, value)
			case string:
				_, err2 := ctx.Write([]byte(t))
				if err2 != nil {
					ctx.Abort()
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
				ctx.FileAttachment(t.Path, t.FileName)
			case *os.File:
				ctx.FileAttachment(t.Name(), t.Name())
			default:
				ctx.JSON(200, Data(value))
			}
		}
	}
}

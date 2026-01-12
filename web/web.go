package web

import (
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/chuccp/go-web-frame/util"
	"github.com/gin-gonic/gin"
)

type HandlersChain []HandlerFunc

func (c HandlersChain) GetFuncName() string {
	return runtime.FuncForPC(reflect.ValueOf(c.Last()).Pointer()).Name()
}

func (c HandlersChain) Last() HandlerFunc {
	if length := len(c); length > 0 {
		return c[length-1]
	}
	return nil
}

func Of(handlerFunc ...HandlerFunc) HandlersChain {
	return HandlersChain(handlerFunc)
}

type HandlersRawChain []HandlerRawFunc

func (c HandlersRawChain) GetFuncName() string {
	return runtime.FuncForPC(reflect.ValueOf(c.Last()).Pointer()).Name()
}

func (c HandlersRawChain) Last() HandlerRawFunc {
	if length := len(c); length > 0 {
		return c[length-1]
	}
	return nil
}
func OfRaw(handlerRawFunc ...HandlerRawFunc) HandlersRawChain {
	return HandlersRawChain(handlerRawFunc)
}

type HandlerFunc func(*Request) (any, error)

type HandlerRawFunc func(*Request, Response) error

func ToGinHandlerFunc(digestAuth *DigestAuth, handlers ...HandlerFunc) []gin.HandlerFunc {
	var handlerFunc = make([]gin.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		handlerFunc[i] = toGinHandlerFunc(digestAuth, handler)
	}
	return handlerFunc
}
func ToGinHandlerRawFunc(digestAuth *DigestAuth, handlers ...HandlerRawFunc) []gin.HandlerFunc {
	var handlerFunc = make([]gin.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		handlerFunc[i] = toGinHandlerRawFunc(digestAuth, handler)
	}
	return handlerFunc
}

func AuthChecks(handlers ...HandlerFunc) []HandlerFunc {
	var hs = make([]HandlerFunc, len(handlers))
	for i, handler := range handlers {
		hs[i] = func(req *Request) (any, error) {
			check, err := req.GetDigestAuth().User(req)
			if err != nil || check == nil {
				return Unauthorized("", err), nil
			}
			return handler(req)
		}
	}
	return hs
}

func AuthRawChecks(handlers ...HandlerRawFunc) []HandlerRawFunc {
	var hs = make([]HandlerRawFunc, len(handlers))
	for i, handler := range handlers {
		hs[i] = func(req *Request, response Response) error {
			check, err := req.GetDigestAuth().User(req)
			if err != nil || check == nil {
				err0 := Unauthorized("", err)
				req.c.JSON(err0.Code, err0)
				req.c.Abort()
				return nil
			}

			return handler(req, response)
		}
	}
	return hs
}

func toGinHandlerFunc(digestAuth *DigestAuth, handler HandlerFunc) gin.HandlerFunc {
	handlerFunc := func(context *gin.Context) {
		value, err := handler(NewRequest(context, digestAuth))
		if err != nil {
			err0 := Errors(value, err)
			context.JSON(err0.Code, err0)
			context.Abort()
		} else {
			if value != nil {
				switch t := value.(type) {
				case *Message:
					if t.Code == http.StatusMovedPermanently {
						context.Redirect(http.StatusMovedPermanently, t.Data.(string))
						context.Abort()
						return
					}
					context.JSON(t.Code, value)
				case string:
					_, err2 := context.Writer.Write([]byte(t))
					if err2 != nil {
						context.Abort()
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
					context.FileAttachment(t.Path, t.FileName)
				case *os.File:
					context.FileAttachment(t.Name(), t.Name())

				default:
					context.JSON(200, Data(value))
				}
			}
		}

	}
	return handlerFunc
}
func toGinHandlerRawFunc(digestAuth *DigestAuth, handler HandlerRawFunc) gin.HandlerFunc {
	handlerFunc := func(context *gin.Context) {
		err := handler(NewRequest(context, digestAuth), newResponse(context.Writer))
		if err != nil {
			err0 := Error(err)
			context.JSON(err0.Code, err0)
			context.Abort()
		}
	}
	return handlerFunc
}

func SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	// 打开上传的临时文件
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer func() {
		// 确保临时文件关闭，并捕获可能的错误
		closeErr := src.Close()
		if err == nil {
			err = closeErr
		}
	}()

	// 创建目标目录
	if err = os.MkdirAll(filepath.Dir(dst), 0750); err != nil {
		return err
	}

	// 创建目标文件
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		// 确保目标文件关闭，并捕获可能的错误
		closeErr := out.Close()
		if err == nil {
			err = closeErr
		}
	}()

	// 复制文件内容
	if _, err = io.Copy(out, src); err != nil {
		return err
	}

	// 强制将数据刷新到磁盘，确保数据写入完成
	if err = out.Sync(); err != nil {
		return err
	}

	return nil
}

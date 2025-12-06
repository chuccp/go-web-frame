package web

import (
	"os"
	"path"

	"github.com/gin-gonic/gin"
)

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
			} else {
				return handler(req, response)
			}
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
					context.JSON(t.Code, value)
				case string:
					_, err2 := context.Writer.Write([]byte(t))
					if err2 != nil {
						context.Abort()
						return
					}
				case *File:
					if len(t.GetFilename()) == 0 {
						_, filename := path.Split(t.Path)
						t.FileName = filename
					}
					context.FileAttachment(t.GetPath(), t.GetFilename())

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
		err := handler(NewRequest(context, digestAuth), context.Writer)
		if err != nil {
			err0 := Error(err)
			context.JSON(err0.Code, err0)
			context.Abort()
		}
	}
	return handlerFunc
}

package web

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"reflect"
	"runtime"

	"github.com/gin-gonic/gin"
)

type HandlerConfig struct {
	//digestAuth   *DigestAuth
	converter    Converter
	handlerInfos []*HandlerInfo
	routeTree    RouteTree
}

func (h *HandlerConfig) Handle(httpMethods []string, relativePath string, handlers ...HandlerFunc) *HandlerInfo {
	handlerInfo := NewHandlerInfo(httpMethods, relativePath, handlers...)
	h.handlerInfos = append(h.handlerInfos, handlerInfo)
	for _, httpMethod := range httpMethods {
		h.routeTree.Set(httpMethod, relativePath, handlerInfo.HandlerMeta)
	}
	return handlerInfo
}
func (h *HandlerConfig) HasHandler(httpMethod string, fullPath string) bool {
	return h.routeTree.Has(httpMethod, fullPath)
}
func (h *HandlerConfig) HandlerMeta(httpMethod string, fullPath string) *HandlerMeta {
	return h.routeTree.GetHandlerMeta(httpMethod, fullPath)
}

func (h *HandlerConfig) HandlerInfos() []*HandlerInfo {
	return h.handlerInfos
}
func NewHandlerConfig(converter Converter) *HandlerConfig {
	return &HandlerConfig{
		//digestAuth:   digestAuth,
		converter:    converter,
		handlerInfos: make([]*HandlerInfo, 0),
		routeTree:    make(RouteTree),
	}
}

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
	return handlerFunc
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

func ToGinHandlerFunc(handlerConfig *HandlerConfig, handlers ...HandlerFunc) []gin.HandlerFunc {
	var handlerFunc = make([]gin.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		handlerFunc[i] = toGinHandlerFunc(handlerConfig, handler)
	}
	return handlerFunc
}

func toGinHandlerFunc(handlerConfig *HandlerConfig, handler HandlerFunc) gin.HandlerFunc {
	handlerFunc := func(ctx *gin.Context) {
		req := NewRequest(ctx, handlerConfig.HandlerMeta(ctx.Request.Method, ctx.FullPath()), handlerConfig)
		value, err := handler(req)
		handlerConfig.converter(value, err, req, newResponse(ctx))
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

package web

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"reflect"
	"runtime"

	"github.com/chuccp/go-web-frame/log"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type FilterChain interface {
	Next() (any, error)
}

type Filter interface {
	Handle(filterChain FilterChain, request *Request) (any, error)
}

type HandlerConfig struct {
	converter    Converter
	handlerInfos []*HandlerInfo
	routeTree    RouteTree
	httpServer   *HttpServer
	filters      []Filter
}

type mockFilterChain struct {
	index      int
	request    *Request
	filters    []Filter
	lastFilter Filter
}

func (c *mockFilterChain) Next() (any, error) {
	if c.index < len(c.filters)-1 {
		c.index++
		return c.filters[c.index].Handle(c, c.request)
	}
	return c.lastFilter.Handle(c, c.request)
}
func newMockFilterChain(request *Request, filters []Filter, lastFilter Filter) *mockFilterChain {

	return &mockFilterChain{
		filters:    filters,
		index:      -1,
		request:    request,
		lastFilter: lastFilter,
	}

}
func (h *HandlerConfig) Handle(httpMethods []string, relativePath string, handlers ...HandlerFunc) *HandlerInfo {
	handlerInfo := NewHandlerInfo(relativePath)
	h.handlerInfos = append(h.handlerInfos, handlerInfo)
	for _, httpMethod := range httpMethods {
		h.routeTree.Set(httpMethod, handlerInfo)
		log.Debug("handle", zap.String("method", httpMethod), zap.String("path", relativePath), zap.Any("handlers", Of(handlers...).GetFuncName()))
		h.httpServer.Handle(httpMethod, relativePath, h.ToGinHandlerFunc(handlers...)...)
	}
	return handlerInfo
}

func (h *HandlerConfig) ToGinHandlerFunc(handlers ...HandlerFunc) []gin.HandlerFunc {
	var handlerFunc = make([]gin.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		handlerFunc[i] = h.toGinHandlerFunc(handler)
	}
	return handlerFunc
}

type lastFilter struct {
	handler HandlerFunc
}

func (last *lastFilter) Handle(filterChain FilterChain, request *Request) (any, error) {
	return last.handler(request)
}
func (h *HandlerConfig) toGinHandlerFunc(handler HandlerFunc) gin.HandlerFunc {
	handlerFunc := func(ctx *gin.Context) {
		resp := newResponse(ctx)
		req := NewRequest(ctx, resp, h.HandlerMeta(ctx.Request.Method, ctx.FullPath()))
		mock := newMockFilterChain(req, h.filters, &lastFilter{handler})
		value, err := mock.Next()
		h.converter(value, err, req, resp)
	}
	return handlerFunc
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

func (h *HandlerConfig) Use(handlers ...Filter) {
	h.filters = append(h.filters, handlers...)
}

func NewHandlerConfig(httpServer *HttpServer, converter Converter) *HandlerConfig {
	return &HandlerConfig{
		converter:    converter,
		handlerInfos: make([]*HandlerInfo, 0),
		routeTree:    make(RouteTree),
		httpServer:   httpServer,
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

type HandlerFunc func(*Request) (any, error)

type HandlerRawFunc func(*Request, Response) error

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

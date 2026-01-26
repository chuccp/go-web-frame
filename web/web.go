package web

import (
	"io"
	"mime/multipart"
	"net/http"
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
	converter Converter
	//handlerInfos []*HandlerInfo
	routeTree RouteTree
	//httpServer   *HttpServer
	filters []Filter
}

type mockFilterChain struct {
	index       int
	request     *Request
	filters     []Filter
	lastFilter  Filter
	firstFilter Filter
	converter   Converter
}

type emptyConverter struct {
}

var empty = &emptyConverter{}

func (c *emptyConverter) Request(filterChain FilterChain, request *Request) {
	next, err := filterChain.Next()
	if err != nil {
		err := request.Response().AbortWithError(err)
		if err != nil {
			log.Error("emptyConverter AbortWithError", zap.Error(err))
		}
		return
	}
	if next != nil {
		request.Response().AbortWithStatusJSON(http.StatusOK, next)
	}
}

func (c *mockFilterChain) Converter() {
	if c.converter != nil {
		c.converter.Request(c, c.request)
	} else {
		empty.Request(c, c.request)
	}

}

func (c *mockFilterChain) Next() (any, error) {
	if c.index < len(c.filters)-1 {
		c.index++
		return c.filters[c.index].Handle(c, c.request)
	}
	return c.lastFilter.Handle(c, c.request)
}
func newMockFilterChain(request *Request, converter Converter, filters []Filter, lastFilter Filter) *mockFilterChain {
	return &mockFilterChain{
		filters:    filters,
		index:      -1,
		request:    request,
		lastFilter: lastFilter,
		converter:  converter,
	}
}
func (h *HandlerConfig) Handle(httpMethods []string, relativePath string, handlers ...HandlerFunc) *HandlerInfo {
	handlerInfo := NewHandlerInfo(relativePath, h.ToGinHandlerFunc(handlers...)...)
	for _, httpMethod := range httpMethods {
		h.routeTree.Set(httpMethod, handlerInfo)
		log.Debug("handle", zap.String("method", httpMethod), zap.String("path", relativePath), zap.Any("handlers", Of(handlers...).GetFuncName()))
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
		request := NewRequest(ctx, resp, h.HandlerMeta(ctx.Request.Method, ctx.FullPath()))
		mock := newMockFilterChain(request, h.converter, h.filters, &lastFilter{handler})
		mock.Converter()
	}
	return handlerFunc
}

func (h *HandlerConfig) HasHandler(httpMethod string, fullPath string) bool {
	return h.routeTree.Has(httpMethod, fullPath)
}
func (h *HandlerConfig) HandlerMeta(httpMethod string, fullPath string) *HandlerMeta {
	return h.routeTree.GetHandlerMeta(httpMethod, fullPath)
}

func (h *HandlerConfig) RouteTree() RouteTree {
	return h.routeTree
}

func (h *HandlerConfig) Use(handlers ...Filter) {
	h.filters = append(h.filters, handlers...)
}

func NewHandlerConfig(converter Converter) *HandlerConfig {
	return &HandlerConfig{
		converter: converter,
		//handlerInfos: make([]*HandlerInfo, 0),
		routeTree: make(RouteTree),
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

type HandlerFunc func(*Request) (any, error)

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

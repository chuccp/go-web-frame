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
	"go.uber.org/zap"
)

type FilterChain interface {
	Next() (any, error)
}

type Filter interface {
	Handle(filterChain FilterChain, request *Request) (any, error)
}

type Handles struct {
	routeTree RouteTree
}

func (h *Handles) Handles(httpMethods []string, relativePath string, handlers ...HandlerFunc) *HandlerInfo {
	handlerInfo := NewHandlerInfo(relativePath, handlers...)
	for _, httpMethod := range httpMethods {
		h.routeTree.Set(httpMethod, handlerInfo)
		log.Debug("handle", zap.String("method", httpMethod), zap.String("path", relativePath), zap.Any("handlers", Of(handlers...).GetFuncName()))
	}
	return handlerInfo
}
func (h *Handles) Handle(httpMethod string, relativePath string, handlers ...HandlerFunc) *HandlerInfo {
	return h.Handles([]string{httpMethod}, relativePath, handlers...)
}
func NewHandles() *Handles {
	return &Handles{
		routeTree: make(RouteTree),
	}
}

type HandlerConfig struct {
	converter Converter
	handles   *Handles
	filters   []Filter
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
	return h.handles.Handles(httpMethods, relativePath, handlers...)
}

type lastFilter struct {
	handler HandlerFunc
}

func (last *lastFilter) Handle(filterChain FilterChain, request *Request) (any, error) {
	return last.handler(request)
}

func (h *Handles) HasHandler(httpMethod string, fullPath string) bool {
	return h.routeTree.Has(httpMethod, fullPath)
}
func (h *Handles) HandlerMeta(httpMethod string, fullPath string) *HandlerMeta {
	return h.routeTree.GetHandlerMeta(httpMethod, fullPath)
}

func (h *Handles) RouteTree() RouteTree {
	return h.routeTree
}
func (h *HandlerConfig) HasHandler(httpMethod string, fullPath string) bool {
	return h.handles.HasHandler(httpMethod, fullPath)
}
func (h *HandlerConfig) HandlerMeta(httpMethod string, fullPath string) *HandlerMeta {
	return h.handles.HandlerMeta(httpMethod, fullPath)
}

func (h *HandlerConfig) Handles() *Handles {
	return h.handles
}

func (h *HandlerConfig) Use(handlers ...Filter) {
	h.filters = append(h.filters, handlers...)
}

func NewHandlerConfig(converter Converter, handles *Handles) *HandlerConfig {
	return &HandlerConfig{
		converter: converter,
		filters:   make([]Filter, 0),
		handles:   handles,
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

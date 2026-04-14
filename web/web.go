// Package web provides the HTTP routing, request handling, and response layer
// for the go-web-frame framework, built on top of Gin.
//
// It includes routing with HTTP method support, filter/middleware chain,
// WebSocket, SSE, static file serving, reverse proxy, and JSON/form request binding.
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
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var (
	// anyMethods for RouterGroup Any method
	anyMethods = []string{
		http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodHead, http.MethodOptions, http.MethodDelete, http.MethodConnect,
		http.MethodTrace,
	}
)

// FilterChain 是过滤器链接口，用于在过滤器之间传递控制权
//
// 过滤器链支持责任链模式，每个过滤器可以决定是否继续执行后续过滤器。
//
// 使用示例：
//
//	func (f *AuthFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
//	    if !isAuthenticated(req) {
//	        return nil, errors.New("unauthorized")
//	    }
//	    return fc.Next()
//	}
type FilterChain interface {
	// Next 执行下一个过滤器或最终处理器
	// 返回 (any, error): 处理结果和可能的错误
	Next() (any, error)
}

// Filter 是 HTTP 过滤器接口，用于处理请求和响应
//
// 过滤器可以实现认证、日志、限流、缓存等横切关注点。
// 过滤器按添加顺序执行，每个过滤器可以决定是否继续执行后续过滤器。
//
// 使用示例：
//
//	type LoggingFilter struct {
//	    core.IFilter
//	}
//
//	func (f *LoggingFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
//	    start := time.Now()
//	    result, err := fc.Next()
//	    log.Info("request completed", zap.Duration("elapsed", time.Since(start)))
//	    return result, err
//	}
type Filter interface {
	// Handle 处理请求
	// fc: 过滤器链，用于调用下一个过滤器
	// req: HTTP 请求对象
	// 返回 (any, error): 处理结果和可能的错误
	Handle(filterChain FilterChain, request *Request) (any, error)
}

// Handles manages the route tree and provides methods to register
// HTTP handlers, static files, WebSocket, SSE, and reverse proxy routes.
type Handles struct {
	routeTree RouteTree
}

// Empty reports whether the Handles has no registered routes.
func (h *Handles) Empty() bool {
	return len(h.routeTree) == 0
}

// Handles registers a handler for multiple HTTP methods at the given relative path.
// Returns a HandlerInfo that can be used to attach metadata.
func (h *Handles) Handles(httpMethods []string, relativePath string, handlers ...HandlerFunc) *HandlerInfo {
	handlerInfo := NewHandlerInfo(relativePath, handlers...)
	for _, httpMethod := range httpMethods {
		h.routeTree.Set(httpMethod, handlerInfo)
	}
	return handlerInfo
}

// Handle registers a handler for a single HTTP method at the given relative path.
// Returns a HandlerInfo that can be used to attach metadata.
func (h *Handles) Handle(httpMethod string, relativePath string, handlers ...HandlerFunc) *HandlerInfo {
	return h.Handles([]string{httpMethod}, relativePath, handlers...)
}

// AddStaticFs registers a static file server at the given relative path.
// Both GET and HEAD methods are supported.
func (h *Handles) AddStaticFs(relativePath string, fs http.FileSystem) *HandlerInfo {
	handlerInfo := NewStaticFsHandlerInfo(relativePath, fs)
	h.routeTree.Set(http.MethodGet, handlerInfo)
	log.Debug("staticFs", zap.String("path", relativePath))
	return handlerInfo
}

// AddReverseProxy registers a reverse proxy at the given relative path.
// All HTTP methods are proxied to the target URL.
func (h *Handles) AddReverseProxy(relativePath string, targetUrl string) *HandlerInfo {
	handlerInfo := NewReverseProxyHandlerInfo(relativePath, targetUrl)
	// 反向代理需要处理所有 HTTP 方法
	for _, method := range anyMethods {
		h.routeTree.Set(method, handlerInfo)
	}
	log.Debug("reverseProxy", zap.String("path", relativePath), zap.String("target", targetUrl))
	return handlerInfo
}

// AddWebSocket registers a WebSocket handler at the given relative path.
// If no upgrader is provided, the default upgrader is used.
func (h *Handles) AddWebSocket(relativePath string, handler WebSocketHandler, upgrader *websocket.Upgrader) *HandlerInfo {
	if upgrader == nil {
		upgrader = DefaultWebSocketUpgrader()
	}
	handlerInfo := NewWebSocketHandlerInfo(relativePath, handler, upgrader)
	h.routeTree.Set(http.MethodGet, handlerInfo)
	log.Debug("webSocket", zap.String("path", relativePath))
	return handlerInfo
}

// AddSSE registers a Server-Sent Events handler at the given relative path.
func (h *Handles) AddSSE(relativePath string, handler SSEHandler) *HandlerInfo {
	handlerInfo := NewSSEHandlerInfo(relativePath, handler)
	h.routeTree.Set(http.MethodGet, handlerInfo)
	log.Debug("sse", zap.String("path", relativePath))
	return handlerInfo
}

// NewHandles creates a new empty Handles with an initialized route tree.
func NewHandles() *Handles {
	return &Handles{
		routeTree: make(RouteTree),
	}
}

// HandlerConfig holds the configuration for a group of handlers,
// including the response converter, route handles, filter chain, and context path prefix.
type HandlerConfig struct {
	converter   Converter
	handles     *Handles
	filters     []Filter
	contextPath string
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

// lastFilter wraps the final handler in the filter chain.
type lastFilter struct {
	handler HandlerFunc
}

// Handle executes the final handler in the chain.
func (last *lastFilter) Handle(filterChain FilterChain, request *Request) (any, error) {
	return last.handler(request)
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
// Handle registers handlers for the given HTTP methods and relative path.
func (h *HandlerConfig) Handle(httpMethods []string, relativePath string, handlers ...HandlerFunc) *HandlerInfo {
	return h.handles.Handles(httpMethods, relativePath, handlers...)
}

// HasHandler reports whether a handler is registered for the given method and path.
func (h *Handles) HasHandler(httpMethod string, fullPath string) bool {
	return h.routeTree.Has(httpMethod, fullPath)
}

// HandlerMeta returns the HandlerMeta for the given method and path.
// Returns an empty HandlerMeta if no handler is found.
func (h *Handles) HandlerMeta(httpMethod, fullPath string) *HandlerMeta {
	return h.routeTree.GetHandlerMeta(httpMethod, fullPath)
}

// RouteTree returns the underlying route tree.
func (h *Handles) RouteTree() RouteTree {
	return h.routeTree
}

// HasHandler reports whether a handler is registered in this config.
func (h *HandlerConfig) HasHandler(httpMethod string, fullPath string) bool {
	return h.handles.HasHandler(httpMethod, fullPath)
}

// HandlerMeta returns the HandlerMeta for the given method and path.
func (h *HandlerConfig) HandlerMeta(httpMethod, fullPath string) *HandlerMeta {
	return h.handles.HandlerMeta(httpMethod, fullPath)
}

// Handles returns the Handles associated with this config.
func (h *HandlerConfig) Handles() *Handles {
	return h.handles
}

// Use appends filters to the filter chain.
func (h *HandlerConfig) Use(handlers ...Filter) {
	h.filters = append(h.filters, handlers...)
}

// NewHandlerConfig creates a new HandlerConfig with the given converter, handles, and server config.
func NewHandlerConfig(converter Converter, handles *Handles, serverConfig *ServerConfig) *HandlerConfig {
	return &HandlerConfig{
		converter:   converter,
		filters:     make([]Filter, 0),
		handles:     handles,
		contextPath: serverConfig.ContextPath,
	}
}

// HandlersChain is a slice of HandlerFunc, representing a chain of handlers.
type HandlersChain []HandlerFunc

// GetFuncName returns the name of the last handler function in the chain.
func (c HandlersChain) GetFuncName() string {
	return runtime.FuncForPC(reflect.ValueOf(c.Last()).Pointer()).Name()
}

// Last returns the last handler in the chain, or nil if the chain is empty.
func (c HandlersChain) Last() HandlerFunc {
	if length := len(c); length > 0 {
		return c[length-1]
	}
	return nil
}

// Of creates a HandlersChain from the given handler functions.
func Of(handlerFunc ...HandlerFunc) HandlersChain {
	return handlerFunc
}

// HandlerFunc is a function that handles an HTTP request and returns a result and error.
type HandlerFunc func(*Request) (any, error)

// SaveUploadedFile saves an uploaded file to the destination path.
// It creates the destination directory if it does not exist.
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
	if err = os.MkdirAll(filepath.Dir(dst), 0775); err != nil {
		return err
	}
	err = os.Chmod(filepath.Dir(dst), 0775)
	if err != nil {
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

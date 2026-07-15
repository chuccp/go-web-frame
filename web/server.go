// Package web: HTTP server with routing, filters, and static file support.
package web

import (
	"context"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/util"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Servers manages multiple HTTP servers that can be started together.
type Servers struct {
	servers []*Server
	ctx     context.Context
}

// NewServers creates a new Servers instance with release mode and background context.
func NewServers() *Servers {
	gin.SetMode(gin.ReleaseMode)
	return NewServerWithContext(context.Background())
}

// NewServerWithContext creates a new Servers instance with the given context.
func NewServerWithContext(ctx context.Context) *Servers {
	return &Servers{
		ctx:     ctx,
		servers: make([]*Server, 0),
	}
}

// CreateServerWithContext creates a new Server with the given config and context, checking for port conflicts.
func (servers *Servers) CreateServerWithContext(serverConfig *ServerConfig, ctx context.Context) (*Server, error) {
	for _, server := range servers.servers {
		if server.serverConfig.Port == serverConfig.Port {
			return nil, errors.New("port already in use")
		}
	}
	server := &Server{
		serverConfig: serverConfig,
		ctx:          ctx,
		engine:       defaultEngine(),
		routes:       make([]*Route, 0),
		filters:      make([]Filter, 0),
		converter:    defaultConverter,
	}
	servers.servers = append(servers.servers, server)
	return server, nil
}

// CreateServer creates a new Server with the given config using the Servers' context.
func (servers *Servers) CreateServer(serverConfig *ServerConfig) (*Server, error) {
	return servers.CreateServerWithContext(serverConfig, servers.ctx)
}

// GetServers returns all managed Server instances.

func (servers *Servers) GetServers() []*Server {
	return servers.servers
}

// GetHandler returns an http.Handler that dispatches requests to the correct
// Server based on the port in the request's Host header. Each Server's routes,
// filters, and ContextPath are fully independent.
func (servers *Servers) GetHandler() http.Handler {
	if len(servers.servers) == 0 {
		return nil
	}
	m := make(map[string]http.Handler, len(servers.servers))
	for _, s := range servers.servers {
		s.initRoute()
		m[strconv.Itoa(s.serverConfig.Port)] = s.engine
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, port, err := net.SplitHostPort(r.Host)
		if err != nil {
			// Host without explicit port — try default HTTP port
			port = "80"
		}
		if h, ok := m[port]; ok {
			h.ServeHTTP(w, r)
			return
		}
		http.Error(w, "no server for port "+port, http.StatusNotFound)
	})
}

func defaultEngine() *gin.Engine {
	engine := gin.New()
	engine.ForwardedByClientIP = true
	return engine
}

// Start starts all managed servers with TLS and auto-cert support.
func (servers *Servers) Start() error {
	certServer := newServerRunner(servers.ctx, "./auto_cert", servers.servers)
	return certServer.Start()
}

// Server is an HTTP server with routing, filters, and static file support.
type Server struct {
	serverConfig *ServerConfig
	routes       []*Route
	ctx          context.Context
	engine       *gin.Engine
	converter    Converter
	filters      []Filter
}

func DefaultServer() *Server {
	gin.SetMode(gin.ReleaseMode)
	server := &Server{
		serverConfig: DefaultServerConfig(),
		ctx:          context.Background(),
		engine:       defaultEngine(),
		routes:       make([]*Route, 0),
		filters:      make([]Filter, 0),
		converter:    defaultConverter,
	}
	return server
}
func (server *Server) GetHandler() http.Handler {
	server.justInitRoute()
	server.optionsMiddleware()
	return server.engine
}

// Port returns the configured listen port for this server.
func (server *Server) Port() int {
	return server.serverConfig.Port
}
func (server *Server) isAuto(host string) bool {
	if server.isTls() {
		if len(server.serverConfig.SSL.Certs) <= 0 {
			if len(server.serverConfig.SSL.Hosts) > 0 {
				if util.EqualsAnyIgnoreCase(host, server.serverConfig.SSL.Hosts...) {
					return true
				}
			}
		}
	}
	return false
}
func (server *Server) isTls() bool {
	if server.serverConfig.SSLEnabled() {
		return true
	}
	return false
}

// AddFilter appends a filter (middleware) to the server's filter chain.
func (server *Server) AddFilter(filter Filter) {
	server.filters = append(server.filters, filter)
}

// SetConverter sets a custom converter that transforms handler results to HTTP responses.
func (server *Server) SetConverter(converter Converter) {
	server.converter = converter
}

// AddHandles transfers all routes from the given Handles into this Server.
func (server *Server) AddHandles(handles *Handles) {
	for _, route := range handles.routes {
		server.routes = append(server.routes, route)
	}
}

// Get registers a GET route handler on this server.
func (server *Server) Get(relativePath string, handlers ...HandlerFunc) *Route {
	return server.Handle(http.MethodGet, relativePath, handlers...)
}

// Post registers a POST route handler on this server.
func (server *Server) Post(relativePath string, handlers ...HandlerFunc) *Route {
	return server.Handle(http.MethodPost, relativePath, handlers...)
}

// Delete registers a DELETE route handler on this server.
func (server *Server) Delete(relativePath string, handlers ...HandlerFunc) *Route {
	return server.Handle(http.MethodDelete, relativePath, handlers...)
}

// Put registers a PUT route handler on this server.
func (server *Server) Put(relativePath string, handlers ...HandlerFunc) *Route {
	return server.Handle(http.MethodPut, relativePath, handlers...)
}

// Patch registers a PATCH route handler on this server.
func (server *Server) Patch(relativePath string, handlers ...HandlerFunc) *Route {
	return server.Handle(http.MethodPatch, relativePath, handlers...)
}

// Any registers a route handler for all HTTP methods on this server.
func (server *Server) Any(relativePath string, handlers ...HandlerFunc) *Route {
	return server.Handlers(allMethods, relativePath, handlers...)
}

// Handle registers a handler for a single HTTP method on this server.
func (server *Server) Handle(httpMethod string, relativePath string, handlers ...HandlerFunc) *Route {
	return server.Handlers([]string{httpMethod}, relativePath, handlers...)
}

// AddSSE registers a Server-Sent Events endpoint on this server.
func (server *Server) AddSSE(relativePath string, handler SSEHandler) *Route {
	return server.Get(relativePath, func(r *Request) (any, error) {
		return &SSEResponse{Handler: handler}, nil
	})
}

// AddWebSocket registers a WebSocket endpoint on this server.
func (server *Server) AddWebSocket(relativePath string, handler WebSocketHandler) *Route {
	return server.Get(relativePath, func(r *Request) (any, error) {
		return &WSResponse{Handler: handler}, nil
	})
}

var allMethods = []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodHead, http.MethodOptions, http.MethodConnect, http.MethodTrace}

// AddReverseProxy registers a reverse proxy route on this server for all HTTP methods.
// The relativePath should be a prefix (e.g. "/api"); a wildcard is appended automatically
// so that "/api/hello" matches in addition to "/api".
func (server *Server) AddReverseProxy(relativePath string, target string) *Route {
	proxyPath := strings.TrimSuffix(relativePath, "/") + "/*path"
	return server.Handlers(allMethods, proxyPath, func(r *Request) (any, error) {
		return &ReverseProxyResponse{Target: target}, nil
	})
}

// AddStaticFs registers a static file server at the given path on this server.
func (server *Server) AddStaticFs(relativePath string, fs http.FileSystem) *Route {
	return server.Get(relativePath+"/*filepath", func(r *Request) (any, error) {
		return &FileSystemResponse{Filepath: r.Param("filepath"), FS: fs}, nil
	})
}

// Handlers registers a handler for multiple HTTP methods on this server.
func (server *Server) Handlers(httpMethods []string, relativePath string, handlers ...HandlerFunc) *Route {
	route := route(relativePath, httpMethods, handlers...)
	server.routes = append(server.routes, route)
	return route
}

func (server *Server) initRoute() {
	server.justInitRoute()
	server.optionsMiddleware()
	server.noRoute()
}
func (server *Server) justInitRoute() {
	for _, route := range server.routes {
		for _, httpMethod := range route.httpMethods {
			server.addHandler(httpMethod, route)
		}
	}
}

// optionsMiddleware handles OPTIONS preflight for routes that have no
// explicit OPTIONS handler. Because filters are only invoked when a route
// matches, and OPTIONS does not match method-specific routes, we intercept
// here at gin level and run the filter chain before the 404 is returned.
func (server *Server) optionsMiddleware() {
	if len(server.filters) == 0 {
		return
	}
	server.engine.Use(func(c *gin.Context) {
		fullPath := c.FullPath()
		if len(fullPath) == 0 {
			req := &Request{
				c:           c,
				cookie:      NewCookie(c),
				handlerMeta: NewHandlerMeta(),
				response:    newResponse(c),
			}
			mock := newFilterChain(req, server.converter, server.filters, func(req *Request) (any, error) {
				return nil, nil
			})
			mock.next()
		}
	})
}

func (server *Server) noRoute() {
	if len(server.serverConfig.Locations) == 0 {
		return
	}
	fs := DefaultMemFileSystem(server.serverConfig.Locations)
	server.engine.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		// 去除 ContextPath 前缀，memfs 中的路径不带 ContextPath
		if cp := server.serverConfig.ContextPath; len(cp) > 0 {
			var ok bool
			if path, ok = stripContextPath(cp, path); !ok {
				return
			}
		}
		if server.tryServeFile(c, fs, path) {
			return
		}
		// SPA fallback: 非图片请求返回 404 页面
		accept := c.Request.Header.Get("Accept")
		if strings.Contains(accept, "html") && !util.IsImagePath(path) {
			if exists, _ := fs.Exists(server.serverConfig.Page404); exists {
				c.FileFromFS(server.serverConfig.Page404, fs)
			}
		}
	})
}

// tryServeFile 尝试从 memfs 提供静态文件，成功返回 true
func (server *Server) tryServeFile(c *gin.Context, fs *MemFileSystem, path string) bool {
	info, err := fs.Stat(path)
	if err != nil || info == nil {
		return false
	}
	if !info.IsDir() {
		c.FileFromFS(path, fs)
		return true
	}
	// 目录：有 index.html 时才提供目录服务（http.FileServer 会自动 serve index.html）
	index := filepath.Join(path, "index.html")
	if exists, _ := fs.Exists(index); exists {
		c.FileFromFS(path, fs)
		return true
	}
	return false
}
func (server *Server) addHandler(httpMethod string, route *Route) {
	log.Debug("handle", zap.String("method", httpMethod), zap.String("path", route.relativePath), zap.Any("handlers", route.LastFuncName()))
	relativePath := route.relativePath
	if len(server.serverConfig.ContextPath) > 0 {
		relativePath = joinContextPath(server.serverConfig.ContextPath, relativePath)
	}
	server.engine.Handle(httpMethod, relativePath, server.toGinHandlerFunc(route)...)
}

// stripContextPath 去除 path 的 contextPath 前缀，返回去除后的路径和是否匹配。
// 处理多种格式：尾部斜杠、大小写不敏感、边界匹配（/app 不匹配 /application）。
// 匹配时返回的路径始终以 / 开头；contextPath 为空时匹配所有路径。
func stripContextPath(contextPath, path string) (string, bool) {
	if contextPath == "" {
		return path, true
	}
	// 统一格式：contextPath 去尾部 /，path 保持原样用于截取
	cp := strings.TrimSuffix(contextPath, "/")
	if !strings.HasPrefix(path, cp) {
		return "", false
	}
	// 长度相等 → 精确匹配，返回 /
	if len(path) == len(cp) {
		return "/", true
	}
	// 边界检查：下一个字符必须是 /
	if path[len(cp)] != '/' {
		return "", false
	}
	return path[len(cp):], true
}

func joinContextPath(contextPath string, relativePath string) string {

	if contextPath == "" {
		return relativePath
	}
	// Ensure contextPath starts with /
	if !strings.HasPrefix(contextPath, "/") {
		contextPath = "/" + contextPath
	}
	// Remove trailing slash from contextPath
	contextPath = strings.TrimSuffix(contextPath, "/")

	// Handle root path
	if relativePath == "/" {
		return contextPath + "/"
	}

	// Ensure relativePath starts with /
	if !strings.HasPrefix(relativePath, "/") {
		relativePath = "/" + relativePath
	}

	return contextPath + relativePath
}

func (server *Server) toGinHandlerFunc(route *Route) []gin.HandlerFunc {
	handlers := route.handlers
	var handlerFunc = make([]gin.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		handlerFunc[i] = server.toSingleGinHandlerFunc(route, handler)
	}
	return handlerFunc
}
func (server *Server) toSingleGinHandlerFunc(route *Route, handler HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request(ctx, route, server.serverConfig)
		mock := newFilterChain(req, server.converter, server.filters, handler)
		mock.next()
	}
}

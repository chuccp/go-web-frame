package web2

import (
	"context"
	"net/http"
	"path/filepath"
	"strings"

	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/util"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Servers struct {
	servers []*Server
	ctx     context.Context
}

func NewServers() *Servers {
	gin.SetMode(gin.ReleaseMode)
	return NewServerWithContext(context.Background())
}
func NewServerWithContext(ctx context.Context) *Servers {
	return &Servers{
		ctx:     ctx,
		servers: make([]*Server, 0),
	}
}
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
	}
	servers.servers = append(servers.servers, server)
	return server, nil
}
func (servers *Servers) CreateServer(serverConfig *ServerConfig) (*Server, error) {
	return servers.CreateServerWithContext(serverConfig, servers.ctx)
}

func defaultEngine() *gin.Engine {
	engine := gin.Default()
	engine.ForwardedByClientIP = true
	return engine
}

func (servers *Servers) Start() error {
	certServer := newServerRunner(servers.ctx, "./auto_cert", servers.servers)
	return certServer.Start()
}

type Server struct {
	serverConfig *ServerConfig
	routes       []*Route
	ctx          context.Context
	engine       *gin.Engine
	converter    Converter
	filters      []Filter
}

func (server *Server) isAuto() bool {
	if server.isTls() {
		if len(server.serverConfig.SSL.Certs) <= 0 {
			if len(server.serverConfig.SSL.Hosts) > 0 {
				return true
			}
		}
	}
	return false
}
func (server *Server) isTls() bool {
	if server.serverConfig.SSL != nil && server.serverConfig.SSL.Enabled {
		return true
	}
	return false
}

func (server *Server) AddFilter(filter Filter) {
	server.filters = append(server.filters, filter)
}
func (server *Server) SetConverter(converter Converter) {
	server.converter = converter
}

func (server *Server) Get(relativePath string, handlers ...HandlerFunc) *Route {
	return server.Handle(http.MethodGet, relativePath, handlers...)
}
func (server *Server) Post(relativePath string, handlers ...HandlerFunc) *Route {
	return server.Handle(http.MethodPost, relativePath, handlers...)
}
func (server *Server) Handle(httpMethod string, relativePath string, handlers ...HandlerFunc) *Route {
	return server.Handlers([]string{httpMethod}, relativePath, handlers...)
}

func (server *Server) AddSSE(relativePath string, handler SSEHandler) *Route {
	return server.Get(relativePath, func(r *Request) (any, error) {
		return &SSEResponse{Handler: handler}, nil
	})
}

func (server *Server) AddWebSocket(relativePath string, handler WebSocketHandler) *Route {
	return server.Get(relativePath, func(r *Request) (any, error) {
		return &WSResponse{Handler: handler}, nil
	})
}

var allMethods = []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodHead, http.MethodOptions, http.MethodConnect, http.MethodTrace}

func (server *Server) AddReverseProxy(relativePath string, target string) *Route {
	return server.Handlers(allMethods, relativePath, func(r *Request) (any, error) {
		return &ReverseProxyResponse{Target: target}, nil
	})
}

func (server *Server) AddStaticFs(relativePath string, fs http.FileSystem) *Route {
	return server.Get(relativePath+"/*filepath", func(r *Request) (any, error) {
		return &FileSystemResponse{Filepath: r.Param("filepath"), FS: fs}, nil
	})
}

func (server *Server) Handlers(httpMethods []string, relativePath string, handlers ...HandlerFunc) *Route {
	route := newHandlerRoute(relativePath, httpMethods, handlers...)
	server.routes = append(server.routes, route)
	return route
}

func (server *Server) initRoute() {
	for _, route := range server.routes {
		for _, httpMethod := range route.httpMethods {
			server.addHandler(httpMethod, route)
		}
	}
	server.noRoute()
}
func (server *Server) noRoute() {
	if len(server.serverConfig.Locations) == 0 {
		return
	}
	fs := DefaultMemFileSystem(server.serverConfig.Locations)
	server.engine.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
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
	server.engine.Handle(httpMethod, route.relativePath, server.toGinHandlerFunc(route)...)
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
	handlerFunc := func(ctx *gin.Context) {
		req := request(ctx, route)
		mock := newFilterChain(req, server.converter, server.filters, handler)
		mock.next()
	}
	return handlerFunc
}

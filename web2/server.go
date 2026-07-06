package web2

import (
	"context"
	"net/http"

	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/log"
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
		routeTree:    newRouteTree(),
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
	certServer := newCertServer(servers.ctx, "./auto_cert", servers.servers)
	return certServer.Start()
}

type Server struct {
	serverConfig *ServerConfig
	routeTree    *routeTree
	ctx          context.Context
	engine       *gin.Engine
	converter    Converter
	filters      []Filter
}

func (server *Server) AddFilter(filter Filter) {
	server.filters = append(server.filters, filter)
}
func (server *Server) SetConverter(converter Converter) {
	server.converter = converter
}

func (server *Server) Get(relativePath string, handlers ...HandlerFunc) *HandlerInfo {
	return server.Handle(http.MethodGet, relativePath, handlers...)
}
func (server *Server) Handle(httpMethod string, relativePath string, handlers ...HandlerFunc) *HandlerInfo {
	return server.Handlers([]string{httpMethod}, relativePath, handlers...)
}
func (server *Server) Handlers(httpMethods []string, relativePath string, handlers ...HandlerFunc) *HandlerInfo {
	handlerInfo := NewHandlerInfo(relativePath, handlers...)
	for _, httpMethod := range httpMethods {
		server.routeTree.add(httpMethod, handlerInfo)
	}
	return handlerInfo

}

func (server *Server) initRoute() {
	server.routeTree.each(func(httpMethod string, handlerInfos []*HandlerInfo) {
		for _, handlerInfo := range handlerInfos {
			server.addHandler(httpMethod, handlerInfo)
		}
	})
}

func (server *Server) addHandler(httpMethod string, handlerInfo *HandlerInfo) {
	log.Debug("handle", zap.String("method", httpMethod), zap.String("path", handlerInfo.relativePath), zap.Any("handlers", Of(handlerInfo).GetFuncName()))
	server.engine.Handle(httpMethod, handlerInfo.relativePath, server.toGinHandlerFunc(handlerInfo)...)
}

func (server *Server) toGinHandlerFunc(handlerInfo *HandlerInfo) []gin.HandlerFunc {
	handlers := handlerInfo.handlers
	var handlerFunc = make([]gin.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		handlerFunc[i] = server.toSingleGinHandlerFunc(handlerInfo.relativePath, handlerInfo.handlerMeta, handler)
	}
	return handlerFunc
}
func (server *Server) toSingleGinHandlerFunc(relativePath string, handlerMeta *HandlerMeta, handler HandlerFunc) gin.HandlerFunc {
	handlerFunc := func(ctx *gin.Context) {
		req := request(ctx)
		mock := newMockFilterChain(req, server.converter, server.filters, handler)
		mock.next()
	}
	return handlerFunc
}

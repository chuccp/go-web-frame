package web2

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"time"

	"emperror.dev/errors"
	"github.com/gin-gonic/gin"
	"github.com/sourcegraph/conc/pool"
)

type Servers struct {
	tslServers []*Server
	ctx        context.Context
}

func NewServers() *Servers {
	return NewServerWithContext(context.Background())
}
func NewServerWithContext(ctx context.Context) *Servers {
	return &Servers{
		ctx:        ctx,
		tslServers: make([]*Server, 0),
	}
}
func (servers *Servers) CreateServerWithContext(serverConfig *ServerConfig, ctx context.Context) (*Server, error) {
	for _, server := range servers.tslServers {
		if server.serverConfig.Port == serverConfig.Port {
			return nil, errors.New("port already in use")
		}
	}
	tslServer := &Server{
		serverConfig: serverConfig,
		ctx:          ctx,
		engine:       defaultEngine(),
	}
	servers.tslServers = append(servers.tslServers, tslServer)
	return tslServer, nil
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
	errorsPool := pool.New().WithContext(servers.ctx).WithFirstError()
	for _, server := range servers.tslServers {
		errorsPool.Go(func(ctx context.Context) error {
			return server.listen()
		})
	}
	return errorsPool.Wait()
}

const MaxHeaderBytes = 8192
const MaxReadHeaderTimeout = time.Second * 30
const MaxReadTimeout = time.Minute * 10

type Server struct {
	serverConfig *ServerConfig
	routeTree    routeTree
	ctx          context.Context
	engine       *gin.Engine
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
	server.routeTree.each(func(httpMethod string, handler []*HandlerInfo) {
		for _, info := range handler {
			if info.IsHandler() {

			}
		}
	})
}

func (server *Server) addHandler(httpMethods []string, relativePath string, handlers ...HandlerFunc) {
	for _, httpMethod := range httpMethods {
		server.engine.Handle(httpMethod, relativePath, server.toGinHandlerFunc(handlers...))
	}
}

//	func (server *Server) ToGinHandlerFunc(handlers ...HandlerFunc) []gin.HandlerFunc {
//		var handlerFunc = make([]gin.HandlerFunc, len(handlers))
//		//for i, handler := range handlers {
//		//	handlerFunc[i] = httpServer.toGinHandlerFunc(handlerConfig, handler)
//		//}
//		return handlerFunc
//	}
func (server *Server) toGinHandlerFunc(handler ...HandlerFunc) gin.HandlerFunc {
	handlerFunc := func(ctx *gin.Context) {
		//resp := newResponse(ctx)
		//handlerMeta := handlerConfig.HandlerMeta(ctx.Request.Method, ctx.FullPath())
		//request := newRequest(ctx, resp, handlerMeta, handlerConfig)
		//mock := newMockFilterChain(request, handlerConfig.converter, handlerConfig.filters, &lastFilter{handler})
		//mock.Converter()
	}
	return handlerFunc
}

func (server *Server) listen() error {
	server.initRoute()
	httpServer := &http.Server{
		BaseContext: func(listener net.Listener) context.Context {
			return server.ctx
		},
		Addr:              ":" + strconv.Itoa(server.serverConfig.Port),
		Handler:           server.engine,
		ReadHeaderTimeout: MaxReadHeaderTimeout,
		MaxHeaderBytes:    MaxHeaderBytes,
		ReadTimeout:       MaxReadTimeout,
	}
	return httpServer.ListenAndServe()
}

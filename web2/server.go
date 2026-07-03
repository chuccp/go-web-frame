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

func NewServer() *Servers {
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
func (server *Server) listen() error {
	var engine2 http.Handler = defaultEngine()

	httpServer := &http.Server{
		BaseContext: func(listener net.Listener) context.Context {
			return server.ctx
		},
		Addr:              ":" + strconv.Itoa(server.serverConfig.Port),
		Handler:           engine2,
		ReadHeaderTimeout: MaxReadHeaderTimeout,
		MaxHeaderBytes:    MaxHeaderBytes,
		ReadTimeout:       MaxReadTimeout,
	}
	return httpServer.ListenAndServe()
}

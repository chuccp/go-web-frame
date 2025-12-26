package core

import (
	"sync"

	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/web"
	"github.com/gin-gonic/gin"
	"github.com/sourcegraph/conc/pool"
)

type Server struct {
	certManager *web.CertManager
	restGroups  []*RestGroup
	httpServers map[int]*web.HttpServer
	lock        *sync.RWMutex
}

func (server *Server) getHttpServer(serverConfig *web.ServerConfig) *web.HttpServer {
	server.lock.Lock()
	defer server.lock.Unlock()
	if httpServer, ok := server.httpServers[serverConfig.Port]; ok {
		return httpServer
	}
	httpServer := web.NewHttpServer(serverConfig, server.certManager)
	server.httpServers[serverConfig.Port] = httpServer
	return httpServer
}
func (server *Server) Init(context *Context) error {
	for _, restGroup := range server.restGroups {
		serverConfig := restGroup.serverConfig
		httpServer := server.getHttpServer(serverConfig)
		restContext := context.Copy(restGroup.digestAuth, httpServer)
		for _, middlewareFunc := range restGroup.middlewareFunc {
			httpServer.Use(func(ctx *gin.Context) {
				middlewareFunc(web.NewRequest(ctx, restGroup.digestAuth), restContext)
			})
		}
		for _, rest := range restGroup.rests {
			err := rest.Init(restContext)
			if err != nil {
				return errors.WithStackIf(err)
			}
		}
	}
	return nil
}
func (server *Server) Run() error {
	var wg = pool.New()
	wg.WithMaxGoroutines(len(server.httpServers))
	errorsPool := wg.WithErrors()
	for _, httpServer := range server.httpServers {
		errorsPool.Go(func() error {
			return errors.WithStackIf(httpServer.Run())
		})
	}
	server.certManager.Start()
	return errorsPool.Wait()
}
func NewServer(restGroups []*RestGroup) *Server {
	return &Server{
		certManager: web.NewCertManager(),
		restGroups:  restGroups,
		httpServers: make(map[int]*web.HttpServer),
		lock:        new(sync.RWMutex),
	}
}

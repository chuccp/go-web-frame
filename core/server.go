package core

import (
	"context"
	"sync"

	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/web"
	"github.com/sourcegraph/conc/pool"
)

type Server struct {
	certManager *web.CertManager
	restGroups  []*RestGroup
	httpServers map[int]*web.HttpServer
	lock        *sync.RWMutex
	runners     []IRunner
}

func (server *Server) getHttpServer(serverConfig *web.ServerConfig) *web.HttpServer {
	server.lock.Lock()
	defer server.lock.Unlock()
	if httpServer, ok := server.httpServers[serverConfig.Port]; ok {
		return httpServer
	}
	if serverConfig.SSLEnabled() {
		for _, host := range serverConfig.SSL.Hosts {
			server.certManager.AddHost(host)
		}
	}
	server.certManager.AddPort(serverConfig.Port)
	httpServer := web.NewHttpServer(serverConfig, server.certManager)
	server.httpServers[serverConfig.Port] = httpServer
	return httpServer
}
func (server *Server) Init(ctx *Context) error {
	for _, runner := range server.runners {
		err := runner.Init(ctx)
		if err != nil {
			return errors.WithStackIf(err)
		}
	}
	for _, restGroup := range server.restGroups {
		serverConfig := restGroup.serverConfig
		httpServer := server.getHttpServer(serverConfig)
		handlerConfig := web.NewHandlerConfig(restGroup.converter, restGroup.handles)
		restContext := ctx.Copy(handlerConfig, restGroup.filters)
		err := restGroup.converter.Init(restContext)
		if err != nil {
			return errors.WithStackIf(err)
		}
		for _, filter := range restGroup.filters {
			err := filter.Init(restContext)
			if err != nil {
				return errors.WithStackIf(err)
			}
		}
		for _, filter := range restGroup.filters {
			handlerConfig.Use(filter)
		}
		for _, rest := range restGroup.rests {
			err := rest.Init(restContext)
			if err != nil {
				return errors.WithStackIf(err)
			}
		}
		httpServer.Handle(handlerConfig)
	}
	return nil
}
func (server *Server) Run(ctx context.Context) error {
	var wg = pool.New()
	wg.WithMaxGoroutines(len(server.httpServers) + len(server.runners))
	errorsPool := wg.WithContext(ctx).WithFirstError()
	for _, httpServer := range server.httpServers {
		errorsPool.Go(func(ctx context.Context) error {
			return errors.WithStackIf(httpServer.Run(ctx))
		})
	}
	for _, runner := range server.runners {
		errorsPool.Go(func(ctx context.Context) error {
			return errors.WithStackIf(runner.Run(ctx))
		})
	}
	errorsPool.Go(func(ctx context.Context) error {
		return errors.WithStackIf(server.certManager.Run(ctx))
	})
	return errors.WithStackIf(errorsPool.Wait())
}
func NewServer(restGroups []*RestGroup, runners []IRunner) *Server {
	return &Server{
		certManager: web.NewCertManager(),
		restGroups:  restGroups,
		httpServers: make(map[int]*web.HttpServer),
		lock:        new(sync.RWMutex),
		runners:     runners,
	}
}

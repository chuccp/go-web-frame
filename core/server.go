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
	context     context.Context
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
	for _, runner := range server.runners {
		err := runner.Init(context)
		if err != nil {
			return errors.WithStackIf(err)
		}
	}
	for _, restGroup := range server.restGroups {
		serverConfig := restGroup.serverConfig
		httpServer := server.getHttpServer(serverConfig)
		handlerConfig := web.NewHandlerConfig(restGroup.converter)
		restContext := context.Copy(handlerConfig, restGroup.filters)
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
func (server *Server) Run() error {
	var wg = pool.New()
	wg.WithMaxGoroutines(len(server.httpServers) + len(server.runners))
	errorsPool := wg.WithErrors()
	for _, httpServer := range server.httpServers {
		errorsPool.Go(func() error {
			return errors.WithStackIf(httpServer.Run())
		})
	}
	for _, runner := range server.runners {
		errorsPool.Go(func() error {
			return errors.WithStackIf(runner.Run(server.context))
		})
	}
	server.certManager.Start()
	return errorsPool.Wait()
}
func NewServer(context context.Context, restGroups []*RestGroup, runners []IRunner) *Server {
	return &Server{
		certManager: web.NewCertManager(),
		restGroups:  restGroups,
		httpServers: make(map[int]*web.HttpServer),
		lock:        new(sync.RWMutex),
		runners:     runners,
	}
}

package core

import (
	"context"
	"sync"

	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/util"
	"github.com/chuccp/go-web-frame/web"
	"github.com/sourcegraph/conc/pool"
	"go.uber.org/zap"
)

type Server struct {
	certManager *web.CertManager
	restGroups  []*RestGroup
	httpServers map[int]*web.HttpServer
	lock        *sync.RWMutex
	runners     []IRunner
	ctx         *Context
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
	server.ctx = ctx
	server.certManager = ctx.CertManager()
	for _, runner := range server.runners {
		log.Debug("Init", zap.String("runner", util.GetStructFullQualifiedName(runner)))
		err := runner.Init(ctx)
		if err != nil {
			return errors.WithStackIf(err)
		}
	}
	for _, restGroup := range server.restGroups {
		serverConfig := restGroup.serverConfig
		httpServer := server.getHttpServer(serverConfig)
		handlerConfig := web.NewHandlerConfig(restGroup.converter, restGroup.handles, serverConfig)
		restContext := ctx.Copy(restGroup.handles, restGroup.filters)
		err := restGroup.converter.Init(restContext)
		if err != nil {
			return errors.WithStackIf(err)
		}
		for _, filter := range restGroup.filters {
			log.Debug("Init", zap.String("filter", util.GetStructFullQualifiedName(filter)))
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
		httpServer.AddHandle(handlerConfig)
	}
	return nil
}
func (server *Server) Run() error {
	var wg = pool.New()
	wg.WithMaxGoroutines(len(server.httpServers) + len(server.runners))
	errorsPool := wg.WithContext(server.ctx).WithFirstError()

	for _, runner := range server.runners {
		r := runner
		errorsPool.Go(func(poolCtx context.Context) error {
			log.Info("runner", zap.String("runner", util.GetStructFullName(r)))
			err := errors.WithStackIf(r.Run())
			if err != nil {
				log.Error("runner", zap.String("runner", util.GetStructFullName(r)), zap.Error(err))
				log.PrintPanic(err)
			}
			return err
		})
	}

	errorsPool.Go(func(ctx context.Context) error {
		return errors.WithStackIf(server.certManager.Run(ctx))
	})

	for _, httpServer := range server.httpServers {
		httpServer.Handle()
		errorsPool.Go(func(ctx context.Context) error {
			return errors.WithStackIf(httpServer.Run(ctx))
		})
	}

	return errors.WithStackIf(errorsPool.Wait())
}
func NewServer(restGroups []*RestGroup, runners []IRunner) *Server {
	return &Server{
		restGroups:  restGroups,
		httpServers: make(map[int]*web.HttpServer),
		lock:        new(sync.RWMutex),
		runners:     runners,
	}
}

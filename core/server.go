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

// Server manages HTTP servers and background runners.
// It initializes REST groups, filters, converters, and runs them concurrently.
type Server struct {
	servers    *web.Servers
	restGroups []*RestGroup
	lock       *sync.RWMutex
	runners    []IRunner
	ctx        *Context
}

func (server *Server) getOrCreateServer(serverConfig *web.ServerConfig) *web.Server {
	server.lock.Lock()
	defer server.lock.Unlock()
	for _, s := range server.servers.GetServers() {
		if s.Port() == serverConfig.Port {
			return s
		}
	}
	s, err := server.servers.CreateServerWithContext(serverConfig, server.ctx)
	if err != nil {
		log.Error("create server", zap.Error(err))
		return nil
	}
	return s
}

// Init initializes all runners, filters, converters, and REST controllers.
// It also registers route handlers on the HTTP servers.
func (server *Server) Init(ctx *Context) error {
	server.ctx = ctx
	for _, runner := range server.runners {
		log.Debug("Init", zap.String("runner", util.GetStructFullQualifiedName(runner)))
		err := runner.Init(ctx)
		if err != nil {
			return errors.WithStackIf(err)
		}
	}
	for _, restGroup := range server.restGroups {
		serverConfig := restGroup.serverConfig
		webServer := server.getOrCreateServer(serverConfig)
		if webServer == nil {
			return errors.New("failed to create server for port")
		}
		webServer.SetConverter(restGroup.converter)
		restContext := ctx.Copy(webServer, restGroup.filters)
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
			webServer.AddFilter(filter)
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

// Run starts all HTTP servers and background runners concurrently.
// It uses a goroutine pool with the server's context for lifecycle management.
// Returns when any component fails or the context is cancelled.
func (server *Server) Run() error {
	var wg = pool.New()
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
		return errors.WithStackIf(server.servers.Start())
	})

	err := errorsPool.Wait()
	log.Error("server Run", zap.Error(err))
	return errors.WithStackIf(err)
}

// NewServer creates a new Server with the given REST groups and runners.
func NewServer(restGroups []*RestGroup, runners []IRunner) *Server {
	return &Server{
		servers:    web.NewServers(),
		restGroups: restGroups,
		lock:       new(sync.RWMutex),
		runners:    runners,
	}
}

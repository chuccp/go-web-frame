package core

import (
	"context"
	"net/http"
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

func (server *Server) initServer(restGroup *RestGroup) error {
	server.lock.Lock()
	defer server.lock.Unlock()
	ser := func(serverConfig *web.ServerConfig) *web.Server {
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
	}(restGroup.serverConfig)
	for _, filter := range restGroup.filters {
		err := filter.Init(server.ctx)
		if err != nil {
			return err
		}
		ser.AddFilters(filter)
	}
	ser.AddHandles(restGroup.handles)
	for _, rest := range restGroup.rests {
		ctx := server.ctx.Copy(ser, restGroup.filters)
		err := rest.Init(ctx)
		if err != nil {
			return err
		}
	}
	ser.SetConverter(restGroup.converter)
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
		err := runner.Init(server.ctx)
		if err != nil {
			return errors.WithStackIf(err)
		}
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

	for _, restGroup := range server.restGroups {
		err := server.initServer(restGroup)
		if err != nil {
			return errors.WithStackIf(err)
		}
	}
	errorsPool.Go(func(ctx context.Context) error {
		return errors.WithStackIf(server.servers.Start())
	})

	err := errorsPool.Wait()
	log.Error("server Run", zap.Error(err))
	return errors.WithStackIf(err)
}

func (server *Server) AddIRunner(runner ...IRunner) {
	server.lock.Lock()
	defer server.lock.Unlock()
	server.runners = append(server.runners, runner...)
}
func (server *Server) AddRestGroup(restGroups ...*RestGroup) {
	server.lock.Lock()
	defer server.lock.Unlock()
	server.restGroups = append(server.restGroups, restGroups...)
}

// GetHandler returns an http.Handler for testing. Routes, filters, and
// ContextPath of each underlying Server are fully preserved.
func (server *Server) GetHandler() http.Handler {
	return server.servers.GetHandler()
}

// NewServer creates a new Server with the given REST groups and runners.
func NewServer(ctx *Context) *Server {
	return &Server{
		ctx:        ctx,
		servers:    web.NewServers(),
		restGroups: make([]*RestGroup, 0),
		lock:       new(sync.RWMutex),
		runners:    make([]IRunner, 0),
	}
}

package core

import (
	"context"
	"net/http"
	"sync"
	"time"

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
	ser, err := func(serverConfig *web.ServerConfig) (*web.Server, error) {
		for _, s := range server.servers.GetServers() {
			if s.Port() == serverConfig.Port {
				return s, nil
			}
		}
		s, err := server.servers.CreateServerWithContext(serverConfig, server.ctx)
		if err != nil {
			return nil, errors.WithStackIf(err)
		}
		return s, nil
	}(restGroup.serverConfig)
	if err != nil {
		return errors.WithStackIf(err)
	}
	for _, filter := range restGroup.filters {
		err := filter.Init(server.ctx)
		if err != nil {
			return errors.WithStackIf(err)
		}
		ser.AddFilters(filter)
	}
	ser.AddHandles(restGroup.handles)
	for _, rest := range restGroup.rests {
		ctx := server.ctx.Copy(ser, restGroup.filters)
		err := rest.Init(ctx)
		if err != nil {
			return errors.WithStackIf(err)
		}
	}
	ser.SetConverter(restGroup.converter)
	return nil
}

// runnerShutdownGrace bounds how long Run keeps waiting for background runners
// after the context is cancelled. A runner that ignores the context must not
// be able to block shutdown (or WebFrame.ReStart) forever.
var runnerShutdownGrace = 5 * time.Second

// Run starts all HTTP servers and background runners concurrently.
// It uses a goroutine pool with the server's context for lifecycle management.
// Returns when any component fails, when every component has stopped, or when
// the context is cancelled — waiting at most runnerShutdownGrace for runners
// that do not watch the context.
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

	for _, restGroup := range server.restGroups {
		err := server.initServer(restGroup)
		if err != nil {
			return errors.WithStackIf(err)
		}
	}
	errorsPool.Go(func(ctx context.Context) error {
		return errors.WithStackIf(server.servers.Start())
	})

	// Wait in a goroutine: errorsPool.Wait() returns only after every task has
	// returned, and a runner that ignores the context may never return.
	waitErr := make(chan error, 1)
	go func() { waitErr <- errorsPool.Wait() }()

	select {
	case err := <-waitErr:
		if err != nil && !isShutdownErr(err) {
			log.Error("server Run", zap.Error(err))
		}
		return errors.WithStackIf(err)
	case <-server.ctx.Done():
		// Listeners watch the context and stop on their own; give runners a
		// grace period to return before abandoning the ones that ignore it.
		select {
		case err := <-waitErr:
			if err != nil && !isShutdownErr(err) {
				log.Error("server Run", zap.Error(err))
				return errors.WithStackIf(err)
			}
		case <-time.After(runnerShutdownGrace):
			log.Warn("server Run: stopping with runners still running",
				zap.Duration("grace", runnerShutdownGrace))
		}
		return errors.WithStackIf(server.ctx.Err())
	}
}

// isShutdownErr reports whether err is the expected result of stopping the
// server (a cancelled context or a closed listener) rather than a failure.
func isShutdownErr(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, http.ErrServerClosed)
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
	for _, rg := range server.restGroups {
		_ = server.initServer(rg)
	}
	return server.servers.GetHandler()
}

// NewServer creates a new Server with the given REST groups and runners.
func NewServer(ctx *Context) *Server {
	return &Server{
		ctx:        ctx,
		servers:    web.NewServerWithContext(ctx),
		restGroups: make([]*RestGroup, 0),
		lock:       new(sync.RWMutex),
		runners:    make([]IRunner, 0),
	}
}

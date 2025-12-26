package core

import (
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
		restContext := context.Copy(restGroup.digestAuth, httpServer)
		restContext.Use(restGroup.middlewareFunc...)
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
	wg.WithMaxGoroutines(len(server.httpServers) + len(server.runners))
	errorsPool := wg.WithErrors()
	for _, httpServer := range server.httpServers {
		errorsPool.Go(func() error {
			return errors.WithStackIf(httpServer.Run())
		})
	}
	for _, runner := range server.runners {
		errorsPool.Go(func() error {
			return errors.WithStackIf(runner.Run())
		})
	}
	server.certManager.Start()
	return errorsPool.Wait()
}
func (server *Server) Stop() error {
	errs := make([]error, 0)
	for _, httpServer := range server.httpServers {
		err := httpServer.Close()
		errs = append(errs, err)
	}
	for _, runner := range server.runners {
		err := runner.Stop()
		errs = append(errs, err)
	}
	return errors.Combine(errs...)
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

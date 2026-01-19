package core

import (
	"sync"

	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/web"
	"github.com/gin-gonic/gin"
	"github.com/sourcegraph/conc/pool"
	"go.uber.org/zap"
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
		handlerConfig := web.NewHandlerConfig(restGroup.digestAuth, restGroup.converter)
		restContext := context.Copy(handlerConfig)
		for _, rest := range restGroup.rests {
			err := rest.Init(restContext)
			if err != nil {
				return errors.WithStackIf(err)
			}
		}
		for _, filter := range restGroup.filters {
			err := filter.Init(restContext)
			if err != nil {
				return err
			}
		}
		for _, filter := range restGroup.filters {
			httpServer.Use(func(ctx *gin.Context) {
				if handlerConfig.HasHandler(ctx.Request.Method, ctx.FullPath()) {
					filter.Handle(web.NewRequest(ctx, handlerConfig.HandlerMeta(ctx.Request.Method, ctx.FullPath()), handlerConfig))
				}
			})
		}
		handlerInfos := handlerConfig.HandlerInfos()
		for _, handlerInfo := range handlerInfos {
			for _, httpMethod := range handlerInfo.HttpMethod {
				log.Debug("handle", zap.String("method", httpMethod), zap.String("path", handlerInfo.RelativePath), zap.Any("handlers", web.Of(handlerInfo.Handlers...).GetFuncName()))
				httpServer.Handle(httpMethod, handlerInfo.RelativePath, web.ToGinHandlerFunc(handlerConfig, handlerInfo.Handlers...)...)
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
func (server *Server) Destroy() error {
	errs := make([]error, 0)
	for _, httpServer := range server.httpServers {
		err := httpServer.Close()
		errs = append(errs, err)
	}
	for _, runner := range server.runners {
		err := runner.Destroy()
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

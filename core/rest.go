package core

import (
	"github.com/chuccp/go-web-frame/web"
	"github.com/gin-gonic/gin"
)

type IRest interface {
	Init(context *Context)
	Name() string
}

type RestGroup struct {
	rests          []IRest
	port           int
	name           string
	httpServer     *web.HttpServer
	digestAuth     *web.DigestAuth
	serverConfig   *web.ServerConfig
	middlewareFunc []MiddlewareFunc
}

func (rg *RestGroup) AddRest(rest ...IRest) *RestGroup {
	rg.rests = append(rg.rests, rest...)
	return rg
}

func (rg *RestGroup) AddMiddlewares(middlewareFunc ...MiddlewareFunc) *RestGroup {
	rg.middlewareFunc = append(rg.middlewareFunc, middlewareFunc...)
	return rg
}

func (rg *RestGroup) merge(restGroup *RestGroup) *RestGroup {
	rg.rests = append(rg.rests, restGroup.rests...)
	if rg.digestAuth == nil {
		rg.digestAuth = restGroup.digestAuth
	}
	if rg.port == 0 {
		rg.port = restGroup.port
	}
	if rg.port == restGroup.port {
		if rg.serverConfig == nil || (!rg.serverConfig.SSLEnabled()) {
			rg.serverConfig = restGroup.serverConfig
		}
	}
	rg.middlewareFunc = append(rg.middlewareFunc, restGroup.middlewareFunc...)
	return rg
}
func (rg *RestGroup) Authentication(authentication web.Authentication) *RestGroup {
	if rg.digestAuth == nil {
		rg.digestAuth = web.NewDigestAuth(authentication)
	}
	return rg
}

func (rg *RestGroup) Init(context *Context) {
	rg.httpServer = web.NewHttpServer(rg.serverConfig, context.GetCertManager())
	context.httpServer = rg.httpServer
	for _, middlewareFunc := range rg.middlewareFunc {
		rg.httpServer.Use(func(ctx *gin.Context) {
			middlewareFunc(web.NewRequest(ctx, rg.digestAuth), context)
		})
	}
	for _, rest := range rg.rests {
		rest.Init(context)
	}
}
func (rg *RestGroup) Run() error {
	return rg.httpServer.Run()
}

func newRestGroup(serverConfig *web.ServerConfig) *RestGroup {

	return &RestGroup{
		rests:        make([]IRest, 0),
		port:         serverConfig.Port,
		serverConfig: serverConfig,
	}
}

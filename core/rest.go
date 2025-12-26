package core

import (
	"github.com/chuccp/go-web-frame/web"
	"github.com/gin-gonic/gin"
)

type IRest interface {
	IService
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

func (rg *RestGroup) DigestAuth() *web.DigestAuth {
	return rg.digestAuth
}
func (rg *RestGroup) HttpServer() *web.HttpServer {
	return rg.httpServer
}
func (rg *RestGroup) Port() int {
	return rg.port
}
func (rg *RestGroup) AddRest(rest ...IRest) *RestGroup {
	rg.rests = append(rg.rests, rest...)
	return rg
}

func (rg *RestGroup) AddMiddlewares(middlewareFunc ...MiddlewareFunc) *RestGroup {
	rg.middlewareFunc = append(rg.middlewareFunc, middlewareFunc...)
	return rg
}

func (rg *RestGroup) Merge(restGroup *RestGroup) *RestGroup {
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

func NewRestGroup(serverConfig *web.ServerConfig, certManager *web.CertManager) *RestGroup {

	return &RestGroup{
		rests:        make([]IRest, 0),
		port:         serverConfig.Port,
		serverConfig: serverConfig,
		httpServer:   web.NewHttpServer(serverConfig, certManager),
	}
}

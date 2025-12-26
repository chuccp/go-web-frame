package core

import (
	"github.com/chuccp/go-web-frame/web"
)

type RestGroup struct {
	rests          []IRest
	port           int
	name           string
	digestAuth     *web.DigestAuth
	middlewareFunc []MiddlewareFunc
	serverConfig   *web.ServerConfig
}

func (rg *RestGroup) DigestAuth() *web.DigestAuth {
	return rg.digestAuth
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
func NewRestGroup(serverConfig *web.ServerConfig) *RestGroup {

	return &RestGroup{
		rests:        make([]IRest, 0),
		port:         serverConfig.Port,
		serverConfig: serverConfig,
	}
}

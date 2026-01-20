package core

import (
	"github.com/chuccp/go-web-frame/web"
)

type RestGroup struct {
	rests []IRest
	port  int
	name  string
	//digestAuth   *web.DigestAuth
	converter    web.Converter
	filters      []IFilter
	serverConfig *web.ServerConfig
}

//	func (rg *RestGroup) DigestAuth() *web.DigestAuth {
//		return rg.digestAuth
//	}
func (rg *RestGroup) Port() int {
	return rg.port
}
func (rg *RestGroup) AddRest(rest ...IRest) *RestGroup {
	rg.rests = append(rg.rests, rest...)
	return rg
}

func (rg *RestGroup) AddFilter(filter ...IFilter) *RestGroup {
	rg.filters = append(rg.filters, filter...)
	return rg
}
func (rg *RestGroup) Converter(converter web.Converter) *RestGroup {
	rg.converter = converter
	return rg
}

func (rg *RestGroup) Merge(restGroup *RestGroup) *RestGroup {
	rg.rests = append(rg.rests, restGroup.rests...)
	//if rg.digestAuth == nil {
	//	rg.digestAuth = restGroup.digestAuth
	//}
	if rg.port == 0 {
		rg.port = restGroup.port
	}
	if rg.port == restGroup.port {
		if rg.serverConfig == nil || (!rg.serverConfig.SSLEnabled()) {
			rg.serverConfig = restGroup.serverConfig
		}
	}
	rg.filters = append(rg.filters, restGroup.filters...)
	return rg
}

//	func (rg *RestGroup) Authentication(authentication web.Authentication) *RestGroup {
//		if rg.digestAuth == nil {
//			rg.digestAuth = web.NewDigestAuth(authentication)
//		}
//		return rg
//	}
func NewRestGroup(serverConfig *web.ServerConfig) *RestGroup {
	filters := make([]IFilter, 0)
	//filters = append(filters, &LoginFilter{})
	return &RestGroup{
		rests:        make([]IRest, 0),
		port:         serverConfig.Port,
		serverConfig: serverConfig,
		//digestAuth:   nil,
		converter: web.DefaultConverter,
		filters:   filters,
	}
}

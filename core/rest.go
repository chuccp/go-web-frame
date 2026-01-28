package core

import (
	"github.com/chuccp/go-web-frame/web"
)

type RestGroup struct {
	rests        []IRest
	port         int
	converter    IConverter
	filters      []IFilter
	serverConfig *web.ServerConfig
	handles      *web.Handles
}

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
func (rg *RestGroup) Converter(converter IConverter) *RestGroup {
	rg.converter = converter
	return rg
}

func (rg *RestGroup) Merge(restGroup *RestGroup) *RestGroup {
	rg.rests = append(rg.rests, restGroup.rests...)
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

func NewRestGroup(serverConfig *web.ServerConfig, converter IConverter, handles *web.Handles) *RestGroup {
	return &RestGroup{
		rests:        make([]IRest, 0),
		port:         serverConfig.Port,
		serverConfig: serverConfig,
		converter:    converter,
		filters:      make([]IFilter, 0),
		handles:      handles,
	}
}

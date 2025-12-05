package core

import (
	"github.com/chuccp/go-web-frame/web"
)

type IRest interface {
	Init(context *Context)
	Name() string
}

type RestGroup struct {
	rests      []IRest
	port       int
	name       string
	engine     *webEngine
	digestAuth *web.DigestAuth
}

func (rg *RestGroup) AddRest(rest ...IRest) *RestGroup {
	rg.rests = append(rg.rests, rest...)
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
	return rg
}
func (rg *RestGroup) Authentication(authentication web.Authentication) *RestGroup {
	if rg.digestAuth != nil {
		rg.digestAuth = web.NewDigestAuth(authentication)
	}
	return rg
}
func (rg *RestGroup) Run() error {
	return nil
}
func newRestGroup(port int) *RestGroup {

	return &RestGroup{
		rests: make([]IRest, 0),
		port:  port,
	}
}

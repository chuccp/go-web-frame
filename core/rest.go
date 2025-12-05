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
func (rg *RestGroup) Authentication(authentication web.Authentication) *RestGroup {
	rg.digestAuth = web.NewDigestAuth(authentication)
	return rg
}
func (rg *RestGroup) Run() error {
	return nil
}
func NewRestGroup(port int) *RestGroup {

	return &RestGroup{
		rests: make([]IRest, 0),
		port:  port,
	}
}

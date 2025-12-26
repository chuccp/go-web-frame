package main

import (
	"emperror.dev/errors"
	wf "github.com/chuccp/go-web-frame"
	"github.com/chuccp/go-web-frame/cache"
	"github.com/chuccp/go-web-frame/config"
	log2 "github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/web"
)

type Authentication struct {
}

func (a *Authentication) SignIn(user any, request *web.Request) (any, error) {
	return nil, errors.New("This method is not implemented.")
}
func (a *Authentication) SignOut(request *web.Request) (any, error) {
	return nil, errors.New("This method is not implemented.")
}
func (a *Authentication) User(request *web.Request) (any, error) {
	return nil, errors.New("This method is not implemented.")
}
func (a *Authentication) NewUser() any {
	return nil
}

func main() {
	loadConfig, err := config.LoadConfig("application.yml")
	if err != nil {
		return
	}
	webFrame := wf.New(loadConfig)
	webFrame.AddComponent(&cache.Component{})
	webFrame.Authentication(&Authentication{})
	webFrame.AddRest(&Api{})
	err = webFrame.Start()
	if err != nil {
		log2.Errors("启动失败", err)
		return
	}
}

package main

import (
	"fmt"

	"github.com/chuccp/go-web-frame/cache"
	"github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/core"
	log2 "github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/web"
)

type Authentication struct {
}

func (a *Authentication) SignIn(user any, request *web.Request) (any, error) {
	return nil, fmt.Errorf("未实现")
}
func (a *Authentication) SignOut(request *web.Request) (any, error) {
	return nil, fmt.Errorf("未实现")
}
func (a *Authentication) User(request *web.Request) (any, error) {
	return nil, fmt.Errorf("未实现")
}
func (a *Authentication) NewUser() any {
	return nil
}

func main() {
	loadConfig, err := config.LoadConfig("application.yml")
	if err != nil {
		return
	}
	webFrame := core.New(loadConfig)
	webFrame.AddComponent(&cache.Component{})
	webFrame.Authentication(&Authentication{})
	webFrame.AddRest(&Api{})
	err = webFrame.Start()
	if err != nil {
		log2.Errors("启动失败", err)
		return
	}
}

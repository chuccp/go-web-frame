package main

import (
	"fmt"
	"log"

	"github.com/chuccp/go-web-frame/cache"
	"github.com/chuccp/go-web-frame/core"
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
	web := core.CreateWebFrame("application.yml")
	web.AddComponent(&cache.Component{})
	web.GetRestGroup().Authentication(&Authentication{})
	web.AddRest(&Api{})
	err := web.Start()
	if err != nil {
		log.Printf("启动失败 %v", err)
		return
	}
}

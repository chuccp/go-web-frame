package main

import (
	"log"

	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
)

type Authentication struct {
}

func (a *Authentication) SignIn(user any, request *web.Request) (any, error) {
	return nil, nil
}
func (a *Authentication) SignOut(request *web.Request) (any, error) {
	return nil, nil
}
func (a *Authentication) User(request *web.Request) (any, error) {
	return nil, nil
}
func (a *Authentication) NewUser() any {
	return nil
}

func main() {
	web := core.CreateWeb("application.yml")
	web.GetRestGroup().Authentication(&Authentication{})
	web.AddRest(&Api{})
	err := web.Start()
	if err != nil {
		log.Printf("启动失败 %v", err)
		return
	}
}

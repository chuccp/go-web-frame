package main

import (
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
)

type System struct {
	Password string
}

func (s *System) Key() string {
	return "db"
}

type Api struct {
	context *core.Context
}

func (api *Api) test(request *web.Request) (any, error) {
	system := core.GetConfig[*System](api.context)
	return system.Password, nil
}
func (api *Api) Init(context *core.Context) {
	api.context = context

	api.context.Get("/test", api.test)
}
func (api *Api) Name() string {
	return "api"
}

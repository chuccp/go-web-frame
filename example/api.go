package main

import (
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
)

type Api struct {
	context *core.Context
}

func (api *Api) test(request *web.Request) (any, error) {
	return web.Ok("ok"), nil
}
func (api *Api) Init(context *core.Context) {
	api.context = context
	api.context.Get("/test", api.test)
}
func (api *Api) Name() string {
	return "api"
}

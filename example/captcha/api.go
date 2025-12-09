package main

import (
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
)

type Api struct {
	context *core.Context
}

func (api *Api) test(request *web.Request, response web.Response) error {

	return nil
}
func (api *Api) Init(context *core.Context) {
	api.context = context

	api.context.GetRaw("/test", api.test)
}
func (api *Api) Name() string {
	return "api"
}

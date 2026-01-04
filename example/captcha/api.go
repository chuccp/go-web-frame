package main

import (
	wf "github.com/chuccp/go-web-frame"
	"github.com/chuccp/go-web-frame/component/captcha"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
)

type Api struct {
	context *core.Context
}

func (api *Api) test(request *web.Request) (any, error) {

	getCaptcha := wf.GetComponent[*captcha.captcha](api.context).GetCaptcha()
	generate, err := getCaptcha.Generate()
	if err != nil {
		return nil, err
	}
	return generate.GetData(), nil
}
func (api *Api) Init(context *core.Context) error {
	api.context = context

	api.context.Get("/test", api.test)
	return nil
}
func (api *Api) Name() string {
	return "api"
}

package main

import (
	"github.com/chuccp/go-web-frame/cache"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/util"
	"github.com/chuccp/go-web-frame/web"
	"github.com/yeqown/go-qrcode/writer/standard"
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
func (api *Api) test2(request *web.Request, response web.Response) error {
	localCache := core.GetComponent[*cache.Component](cache.Name, api.context).GetLocalCache()
	err := localCache.GetFileResponseWrite(response, func(fileResponseWriteCloser *cache.FileResponseWriteCloser, value ...any) error {
		err := util.GenerateQrcode(
			util.JoinValues(value...),
			fileResponseWriteCloser,
			standard.WithBorderWidth(5),
			standard.WithQRWidth(uint8(5)),
			util.WithRoundedSquareShape(),
		)
		return err
	}, "11111111111")
	return err
}
func (api *Api) Init(context *core.Context) {
	api.context = context
	api.context.Get("/test", api.test)
	api.context.GetRaw("/test2", api.test2)
}
func (api *Api) Name() string {
	return "api"
}

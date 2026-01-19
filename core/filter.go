package core

import "github.com/chuccp/go-web-frame/web"

type LoginFilter struct {
	ctx *Context
}

func (l *LoginFilter) Init(ctx *Context) error {
	l.ctx = ctx
	return nil
}
func (l *LoginFilter) Handle(request *web.Request) {
	if request.HandlerMeta().Has(web.LoginKey) {
		user, err := request.User()
		if err == nil && user != nil {
			request.Next()
		} else {
			response := request.Response()
			err0 := web.Unauthorized("", err)
			response.JSON(err0.Code, err0)
			response.Abort()
		}
	} else {
		request.Next()
	}
}

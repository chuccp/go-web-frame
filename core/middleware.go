package core

import "github.com/chuccp/go-web-frame/web"

type MiddlewareFunc func(request *web.Request, ctx *Context)

func LoginMiddlewareFunc(request *web.Request, ctx *Context) {
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

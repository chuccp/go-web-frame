package core

import "github.com/chuccp/go-web-frame/web"

type MiddlewareFunc func(request *web.Request, ctx *Context)

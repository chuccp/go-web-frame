package core

import "github.com/chuccp/go-web-frame/web"

type MiddlewareFunc func(ctx *web.Request)

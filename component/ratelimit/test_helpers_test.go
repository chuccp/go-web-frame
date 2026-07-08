package ratelimit

import (
	"context"

	config2 "github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
)

func newTestContext(cfg config2.IConfig, ctx context.Context) *core.Context {
	servers := web.NewServers()
	server, _ := servers.CreateServerWithContext(web.DefaultServerConfig(), ctx)
	return core.NewContext(server, cfg, ctx)
}

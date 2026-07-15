package ratelimit

import (
	"context"

	config2 "github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/core"
)

func newTestContext(cfg config2.IConfig, ctx context.Context) *core.Context {
	return core.NewContext(cfg, ctx)
}

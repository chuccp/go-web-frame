package redis

import (
	"github.com/chuccp/go-web-frame/core"
	redis2 "github.com/redis/go-redis/v9"
)

type Component struct {
	client *redis2.Client
}

func (l *Component) Init(ctx *core.Context) error {
	var options = &redis2.Options{}
	err := ctx.GetConfig().UnmarshalKey("web.redis", options)
	if err != nil {
		return err
	}
	l.client = redis2.NewClient(options)
	return nil
}
func (l *Component) GetClient() *redis2.Client {
	return l.client
}

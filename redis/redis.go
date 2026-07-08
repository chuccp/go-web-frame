package redis

import (
	"github.com/chuccp/go-web-frame/core"
	redis2 "github.com/redis/go-redis/v9"
)

// ConfigKey is the configuration key under which Redis settings are stored.
const ConfigKey = "web.redis"

type Component struct {
	client *redis2.Client
}

func (l *Component) Init(ctx *core.Context) error {
	var options = &redis2.Options{}
	err := ctx.GetConfig().UnmarshalKey(ConfigKey, options)
	if err != nil {
		return err
	}
	l.client = redis2.NewClient(options)
	return nil
}
func (l *Component) GetClient() *redis2.Client {
	return l.client
}

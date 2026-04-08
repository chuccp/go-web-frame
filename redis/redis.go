package redis

import (
	"context"

	config2 "github.com/chuccp/go-web-frame/config"
	redis2 "github.com/redis/go-redis/v9"
)

type Component struct {
	client *redis2.Client
}

func (l *Component) Init(ctx context.Context, config config2.IConfig) error {
	var options = redis2.Options{}
	err := config.UnmarshalKey("web.redis", options)
	if err != nil {
		return err
	}
	l.client = redis2.NewClient(&options)
	return nil
}
func (l *Component) GetClient() *redis2.Client {
	return l.client
}

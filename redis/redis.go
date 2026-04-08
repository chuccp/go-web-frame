package redis

import (
	"context"

	config2 "github.com/chuccp/go-web-frame/config"
	"github.com/redis/go-redis/v9"
)

const Name = "redis_component"

type Component struct {
	client *redis.Client
}

func (l *Component) Init(ctx context.Context, config config2.IConfig) error {
	var options = redis.Options{}
	err := config.UnmarshalKey("web.redis", options)
	if err != nil {
		return err
	}
	l.client = redis.NewClient(&options)
	return nil
}
func (l *Component) GetClient() *redis.Client {
	return l.client
}
func (l *Component) Name() string {
	return Name
}

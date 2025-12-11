package redis

import (
	config2 "github.com/chuccp/go-web-frame/config"
	"github.com/redis/go-redis/v9"
)

const Name = "redis_component"

type Component struct {
	client *redis.Client
}

func (l *Component) Init(config *config2.Config) error {
	var options = redis.Options{}
	err := config.Unmarshal("web.redis", &options)
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

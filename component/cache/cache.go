package cache

import (
	"time"

	"emperror.dev/errors"
	config2 "github.com/chuccp/go-web-frame/config"
	"github.com/maypok86/otter/v2"
	"github.com/maypok86/otter/v2/stats"
)

type Cache struct {
	cache *otter.Cache[string, any]
}

// Get 获取缓存值（如果不存在，返回 nil 和 false）
func (c *Cache) Get(key string) (any, bool) {
	// GetIfPresent 不触发加载，直接检查是否存在
	return c.cache.GetIfPresent(key)
}

// Set 设置缓存值（无过期时间，除非全局配置了 ExpiryCalculator）
func (c *Cache) Set(key string, value any) {
	c.cache.Set(key, value)
}

func (c *Cache) Invalidate(key string) (any, bool) {
	return c.cache.Invalidate(key)
}

func (c *Cache) Stats() stats.Stats {
	return c.cache.Stats()
}

func (c *Cache) Init(Config config2.IConfig) error {
	counter := stats.NewCounter()
	cache, err := otter.New(&otter.Options[string, any]{
		MaximumSize:      100_000,
		ExpiryCalculator: otter.ExpiryAccessing[string, any](time.Hour),
		StatsRecorder:    counter,
	})
	if err != nil {
		return errors.WithStackIf(err)
	}
	c.cache = cache
	return nil
}

func (c *Cache) Destroy() error {
	if stopped := c.cache.StopAllGoroutines(); stopped {
	}
	c.cache.CleanUp()
	return nil
}

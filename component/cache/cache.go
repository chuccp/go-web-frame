package cache

import (
	"time"

	"emperror.dev/errors"
	config2 "github.com/chuccp/go-web-frame/config"
	"github.com/maypok86/otter/v2"
	"github.com/maypok86/otter/v2/stats"
)

type Config struct {
	MaxSize int // 最大缓存数量
	Expiry  int // 缓存过期时间 单位秒
}

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

func (c *Cache) Init(config config2.IConfig) error {
	lConfig := &Config{
		MaxSize: 1000_000,
		Expiry:  3600,
	}
	counter := stats.NewCounter()
	cache, err := otter.New(&otter.Options[string, any]{
		MaximumSize:      lConfig.MaxSize,
		ExpiryCalculator: otter.ExpiryAccessing[string, any](time.Duration(lConfig.Expiry) * time.Second),
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

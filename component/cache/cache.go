package cache

import (
	"time"

	"emperror.dev/errors"
	config2 "github.com/chuccp/go-web-frame/config"
	"github.com/maypok86/otter/v2"
	"github.com/maypok86/otter/v2/stats"
	"github.com/sourcegraph/conc/panics"
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

// SetNX 设置缓存值（有过期时间）, 如果已存在则返回 false, 否则返回 true
func (c *Cache) SetNX(key string, value any, expire time.Duration) (any, bool) {
	v, ok := c.cache.SetIfAbsent(key, value)
	if !ok {
		return v, false
	}
	if expire > 0 {
		c.cache.SetExpiresAfter(key, expire)
	}
	return value, true
}

func (c *Cache) ComputeIfAbsent(key string, f func()) bool {
	_, fa := c.cache.ComputeIfAbsent(key, func() (any, bool) {
		defer c.cache.Invalidate(key) // 執行完自動刪除標記，讓下次可以重新進入
		var catcher panics.Catcher
		catcher.Try(f)
		return struct{}{}, false // 佔位值，無實際意義
	})
	return fa
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

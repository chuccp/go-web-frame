package cache

import (
	"time"

	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/log"
	"github.com/maypok86/otter/v2"
	"github.com/maypok86/otter/v2/stats"
)

// Config holds cache configuration settings.
type Config struct {
	MaxSize int // Maximum number of cached entries
	Expiry  int // Cache TTL in seconds
}

// ConfigKey is the configuration key under which cache settings are stored.
const ConfigKey = "cache"

// Cache provides a high-performance in-memory cache backed by Otter.
type Cache struct {
	cache *otter.Cache[string, any]
}

// Get returns the cached value for the given key, or (nil, false) if not found.
func (c *Cache) Get(key string) (any, bool) {
	// GetIfPresent 不触发加载，直接检查是否存在
	return c.cache.GetIfPresent(key)
}

// Set stores a value in the cache.
func (c *Cache) Set(key string, value any) {
	c.cache.Set(key, value)
}

// SetNX stores a value with an expiration if the key does not already exist.
// Returns (value, true) on success, or (existing, false) if the key exists.
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

// GetOrSet returns the cached value or computes and stores it using f.
func (c *Cache) GetOrSet(key string, f func() any) any {
	v, _ := c.cache.ComputeIfAbsent(key, func() (newValue any, cancel bool) {
		return f(), false
	})
	return v
}

// ComputeIfAbsent atomically computes a value if the key is not present.
// f should return (value, cancel); if cancel is true the operation is aborted.
func (c *Cache) ComputeIfAbsent(key string, f func() (any, bool)) (any, bool) {
	return c.cache.ComputeIfAbsent(key, f)
}

// Invalidate removes a key from the cache.
func (c *Cache) Invalidate(key string) (any, bool) {
	return c.cache.Invalidate(key)
}

// Stats returns cache performance statistics.
func (c *Cache) Stats() stats.Stats {
	return c.cache.Stats()
}

func (c *Cache) Init(context *core.Context) error {
	lConfig := &Config{
		MaxSize: 1000_000,
		Expiry:  3600,
	}
	err := context.GetConfig().UnmarshalKey(ConfigKey, lConfig)
	if err != nil {
		return errors.WithStackIf(err)
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
	go func() {
		<-context.Done()
		err := c.destroy()
		log.Errors("cache destroy", err)
	}()
	return nil
}

func (c *Cache) destroy() error {
	if stopped := c.cache.StopAllGoroutines(); stopped {
	}
	c.cache.CleanUp()
	return nil
}

// SetIfAbsent stores a value only if the key does not already exist.
func (c *Cache) SetIfAbsent(key string, value any) (any, bool) {
	return c.cache.SetIfAbsent(key, value)
}

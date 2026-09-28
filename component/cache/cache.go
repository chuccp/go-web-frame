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
	// Expiry 是**从写入算起**的存活秒数：读一次不给它续命。
	// 别换回 ExpiryAccessing（距上次访问）——那样 SetNX 显式给的过期会被一次读顶掉，
	// 「到点就重取」变成「一直有人读就永不重取」（2026-09-28 发现并修的，
	// 见 cache_bug_test.go 的 TestSetNXExpiryHonored）。
	Expiry int
}

// ConfigKey is the configuration key under which cache settings are stored.
const ConfigKey = "cache"

// 没配（或配了非法值）时的默认：100 万条、1 小时。
const (
	defaultMaxSize = 1000_000
	defaultExpiry  = time.Hour
)

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
	cache, err := newStore(lConfig.MaxSize, time.Duration(lConfig.Expiry)*time.Second)
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

// New 直接构造一个缓存（不经过 DI 容器）：给单测、以及"我就想自己持有一个缓存"的用法。
// 过期语义和 Init 一致——从写入算起、读不续命；maxSize/expiry 非正时用上面那两个默认值。
func New(maxSize int, expiry time.Duration) (*Cache, error) {
	if maxSize <= 0 {
		maxSize = defaultMaxSize
	}
	if expiry <= 0 {
		expiry = defaultExpiry
	}
	store, err := newStore(maxSize, expiry)
	if err != nil {
		return nil, err
	}
	return &Cache{cache: store}, nil
}

// newStore 建底层 Otter 实例：Init 和 New 都走这里，免得过期策略两处各写一遍走岔。
func newStore(maxSize int, expiry time.Duration) (*otter.Cache[string, any], error) {
	return otter.New(&otter.Options[string, any]{
		MaximumSize: maxSize,
		// ExpiryWriting：TTL 从写入算起、**读不续命**，这样 SetNX 第三参给的硬过期才算数；
		// 换成 ExpiryAccessing 的话，读一次就把过期推后一次，高频键永不过期。
		ExpiryCalculator: otter.ExpiryWriting[string, any](expiry),
		StatsRecorder:    stats.NewCounter(),
	})
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

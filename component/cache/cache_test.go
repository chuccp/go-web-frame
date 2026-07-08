package cache

import (
	"context"
	"testing"
	"time"

	config2 "github.com/chuccp/go-web-frame/config"
	"github.com/maypok86/otter/v2"
	"github.com/maypok86/otter/v2/stats"
	"github.com/stretchr/testify/assert"
)

func newTestCache(t *testing.T) *Cache {
	t.Helper()
	cfg := config2.NewConfig()
	ctx := context.Background()
	c := &Cache{}
	err := c.initForTest(cfg, ctx)
	assert.NoError(t, err)
	return c
}

// initForTest is a test helper that initializes the cache without requiring a full core.Context.
func (c *Cache) initForTest(cfg config2.IConfig, ctx context.Context) error {
	lConfig := &Config{
		MaxSize: 100,
		Expiry:  3600,
	}
	_ = cfg.UnmarshalKey(ConfigKey, lConfig)
	return c.initWithConfig(lConfig, ctx)
}

func (c *Cache) initWithConfig(lConfig *Config, ctx context.Context) error {
	cache, err := newCacheForTest(lConfig)
	if err != nil {
		return err
	}
	c.cache = cache
	go func() {
		<-ctx.Done()
		c.cache.CleanUp()
	}()
	return nil
}

func newCacheForTest(lConfig *Config) (*otter.Cache[string, any], error) {
	counter := stats.NewCounter()
	return otter.New(&otter.Options[string, any]{
		MaximumSize:      lConfig.MaxSize,
		ExpiryCalculator: otter.ExpiryAccessing[string, any](time.Duration(lConfig.Expiry) * time.Second),
		StatsRecorder:    counter,
	})
}

func TestCache_SetAndGet(t *testing.T) {
	c := newTestCache(t)

	c.Set("key1", "value1")
	v, ok := c.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", v)

	_, ok = c.Get("missing")
	assert.False(t, ok)
}

func TestCache_SetNX(t *testing.T) {
	c := newTestCache(t)

	v, ok := c.SetNX("key1", "value1", 0)
	assert.True(t, ok)
	assert.Equal(t, "value1", v)

	// Second SetNX on same key should fail and return existing value
	v, ok = c.SetNX("key1", "value2", 0)
	assert.False(t, ok)
	assert.Equal(t, "value1", v)
}

func TestCache_SetNX_WithExpiry(t *testing.T) {
	c := newTestCache(t)

	// SetNX returns the value and true on first insert
	v, ok := c.SetNX("expirable", "data", 50*time.Millisecond)
	assert.True(t, ok)
	assert.Equal(t, "data", v)

	// Value is retrievable
	got, ok := c.Get("expirable")
	assert.True(t, ok)
	assert.Equal(t, "data", got)

	// Second SetNX on same key returns existing value and false
	v, ok = c.SetNX("expirable", "newdata", 0)
	assert.False(t, ok)
	assert.Equal(t, "data", v)
}

func TestCache_GetOrSet(t *testing.T) {
	c := newTestCache(t)

	called := false
	v := c.GetOrSet("key1", func() any {
		called = true
		return "computed"
	})
	assert.True(t, called)
	assert.Equal(t, "computed", v)

	// Second call should return cached value without calling f
	called = false
	v = c.GetOrSet("key1", func() any {
		called = true
		return "should-not-be-called"
	})
	assert.False(t, called)
	assert.Equal(t, "computed", v)
}

func TestCache_ComputeIfAbsent(t *testing.T) {
	c := newTestCache(t)

	v, ok := c.ComputeIfAbsent("key1", func() (any, bool) {
		return "value1", false
	})
	assert.True(t, ok)
	assert.Equal(t, "value1", v)

	// Cancel should abort the operation
	v, ok = c.ComputeIfAbsent("key2", func() (any, bool) {
		return "cancelled", true
	})
	assert.False(t, ok)
}

func TestCache_Invalidate(t *testing.T) {
	c := newTestCache(t)

	c.Set("key1", "value1")
	v, ok := c.Invalidate("key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", v)

	_, ok = c.Get("key1")
	assert.False(t, ok)
}

func TestCache_SetIfAbsent(t *testing.T) {
	c := newTestCache(t)

	_, ok := c.SetIfAbsent("key1", "value1")
	assert.True(t, ok)

	_, ok = c.SetIfAbsent("key1", "value2")
	assert.False(t, ok)

	v, ok := c.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", v)
}

func TestCache_Stats(t *testing.T) {
	c := newTestCache(t)

	c.Set("key1", "value1")
	c.Get("key1")
	c.Get("missing")

	s := c.Stats()
	assert.NotNil(t, s)
}

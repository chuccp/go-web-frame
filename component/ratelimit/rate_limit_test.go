package ratelimit

import (
	"context"
	"testing"
	"time"

	config2 "github.com/chuccp/go-web-frame/config"
	"github.com/stretchr/testify/assert"
)

func newTestRateLimit(t *testing.T, cfgOverrides ...func(*Config)) *RateLimit {
	t.Helper()
	cfg := config2.NewConfig()
	cfg.Put(ConfigKey+".Limit", 1)
	cfg.Put(ConfigKey+".Burst", 2)
	cfg.Put(ConfigKey+".MaxSize", 100)
	cfg.Put(ConfigKey+".Expiry", 3600)
	for _, fn := range cfgOverrides {
		fn(nil)
	}

	rl := &RateLimit{}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	coreCtx := newTestContext(cfg, ctx)
	err := rl.Init(coreCtx)
	assert.NoError(t, err)
	return rl
}

func TestRateLimit_Allow_BurstExhaustion(t *testing.T) {
	rl := newTestRateLimit(t)

	// Burst=2: first two requests should be allowed
	assert.True(t, rl.Allow("user1"))
	assert.True(t, rl.Allow("user1"))
	// Third request within burst window should be denied
	assert.False(t, rl.Allow("user1"))
}

func TestRateLimit_Allow_DifferentKeys(t *testing.T) {
	rl := newTestRateLimit(t)

	// Each key has its own limiter
	assert.True(t, rl.Allow("user1"))
	assert.True(t, rl.Allow("user2"))
	assert.True(t, rl.Allow("user1"))
	assert.True(t, rl.Allow("user2"))
	// Both exhausted now (burst=2)
	assert.False(t, rl.Allow("user1"))
	assert.False(t, rl.Allow("user2"))
}

func TestRateLimit_AllowSBurst(t *testing.T) {
	rl := newTestRateLimit(t)

	// Custom burst=5
	assert.True(t, rl.AllowSBurst("vip", 5))
	assert.True(t, rl.AllowSBurst("vip", 5))
	assert.True(t, rl.AllowSBurst("vip", 5))
	assert.True(t, rl.AllowSBurst("vip", 5))
	assert.True(t, rl.AllowSBurst("vip", 5))
	assert.False(t, rl.AllowSBurst("vip", 5))
}

func TestRateLimit_Wait(t *testing.T) {
	rl := newTestRateLimit(t)

	// Exhaust burst
	assert.True(t, rl.Allow("user1"))
	assert.True(t, rl.Allow("user1"))

	// Wait should block until a token is available (Limit=1s, but we use a short timeout)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Override the rate limiter's ctx for the Wait call
	rl.ctx = ctx

	err := rl.Wait("user1")
	assert.NoError(t, err)
}

func TestRateLimit_Wait_ContextCancelled(t *testing.T) {
	rl := newTestRateLimit(t)

	// Exhaust burst
	rl.Allow("user1")
	rl.Allow("user1")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately cancelled

	rl.ctx = ctx
	err := rl.Wait("user1")
	assert.Error(t, err)
}

func TestRateLimit_Stats(t *testing.T) {
	rl := newTestRateLimit(t)

	rl.Allow("user1")
	rl.Allow("user1")

	s := rl.Stats()
	assert.NotNil(t, s)
}

func TestRateLimit_Init_ReadsConfig(t *testing.T) {
	cfg := config2.NewConfig()
	cfg.Put(ConfigKey+".Limit", 1)
	cfg.Put(ConfigKey+".Burst", 1)
	cfg.Put(ConfigKey+".MaxSize", 10)
	cfg.Put(ConfigKey+".Expiry", 60)

	rl := &RateLimit{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	coreCtx := newTestContext(cfg, ctx)
	err := rl.Init(coreCtx)
	assert.NoError(t, err)
	assert.NotNil(t, rl.config)
	assert.Equal(t, 1, rl.config.Burst)

	// With burst=1, only first request allowed
	assert.True(t, rl.Allow("k"))
	assert.False(t, rl.Allow("k"))
}

func TestRateLimit_Init_DefaultConfig(t *testing.T) {
	cfg := config2.NewConfig()

	rl := &RateLimit{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	coreCtx := newTestContext(cfg, ctx)
	err := rl.Init(coreCtx)
	assert.NoError(t, err)
	// Default Burst=5
	assert.Equal(t, 5, rl.config.Burst)
	assert.Equal(t, 600, rl.config.Limit)
}

func TestRateLimit_AllowRefill(t *testing.T) {
	cfg := config2.NewConfig()
	cfg.Put(ConfigKey+".Limit", 1) // refill 1 token every 1 second
	cfg.Put(ConfigKey+".Burst", 1) // burst=1
	cfg.Put(ConfigKey+".MaxSize", 100)
	cfg.Put(ConfigKey+".Expiry", 3600)

	rl := &RateLimit{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	coreCtx := newTestContext(cfg, ctx)
	err := rl.Init(coreCtx)
	assert.NoError(t, err)

	// Use up the single token
	assert.True(t, rl.Allow("k"))
	assert.False(t, rl.Allow("k"))

	// Wait for refill
	time.Sleep(1100 * time.Millisecond)
	assert.True(t, rl.Allow("k"))
}

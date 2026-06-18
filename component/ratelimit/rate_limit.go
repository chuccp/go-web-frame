package ratelimit

import (
	"context"
	"time"

	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/core"
	"github.com/maypok86/otter/v2"
	"github.com/maypok86/otter/v2/stats"
	"golang.org/x/time/rate"
)

// Config holds rate limiter configuration.
type Config struct {
	Limit   int // Requests per second
	Burst   int // Maximum burst size
	MaxSize int // Maximum number of cached limiters
	Expiry  int // Limiter cache TTL in seconds
}

// RateLimit provides per-key rate limiting using a token bucket algorithm.
type RateLimit struct {
	cache         *otter.Cache[string, *rate.Limiter]
	limiterLoader otter.Loader[string, *rate.Limiter]
	ctx           context.Context
	config        *Config
}

// Allow checks if a request for the given key is permitted.
func (r *RateLimit) Allow(key string) bool {
	limiter, err := r.cache.Get(r.ctx, key, r.limiterLoader)
	if err != nil {
		return false
	}
	return limiter.Allow()
}
// AllowSBurst checks if a request is permitted with a custom burst size.
func (r *RateLimit) AllowSBurst(key string, burst int) bool {
	limiter, err := r.cache.Get(r.ctx, key, r._limiterLoader(burst))
	if err != nil {
		return false
	}
	return limiter.Allow()
}

// Wait blocks until a request for the given key is permitted.
func (r *RateLimit) Wait(key string) error {
	limiter, err := r.cache.Get(r.ctx, key, r.limiterLoader)
	if err != nil {
		return err
	}
	return limiter.Wait(r.ctx)
}
func (r *RateLimit) _limiterLoader(burst int) otter.Loader[string, *rate.Limiter] {
	return otter.LoaderFunc[string, *rate.Limiter](func(ctx context.Context, key string) (*rate.Limiter, error) {
		// 每 15 分钟允许 3 次请求 → 每 5 分钟填充 1 个令牌，burst = 3
		return rate.NewLimiter(rate.Every(time.Duration(r.config.Limit)*time.Second), burst), nil
	})
}
func (r *RateLimit) Init(ctx *core.Context) error {
	lConfig := &Config{
		Limit:   600,
		Burst:   5,
		MaxSize: 1000_000,
		Expiry:  3600,
	}

	r.ctx = ctx
	err := ctx.GetConfig().UnmarshalKey("rate_limit", lConfig)
	if err != nil {
		return errors.WithStackIf(err)
	}
	r.config = lConfig
	r.limiterLoader = otter.LoaderFunc[string, *rate.Limiter](func(ctx context.Context, key string) (*rate.Limiter, error) {
		// 每 15 分钟允许 3 次请求 → 每 5 分钟填充 1 个令牌，burst = 3
		return rate.NewLimiter(rate.Every(time.Duration(lConfig.Limit)*time.Second), lConfig.Burst), nil
	})
	counter := stats.NewCounter()
	cache, err := otter.New[string, *rate.Limiter](&otter.Options[string, *rate.Limiter]{
		MaximumSize:      lConfig.MaxSize,
		ExpiryCalculator: otter.ExpiryAccessing[string, *rate.Limiter](time.Duration(lConfig.Expiry) * time.Second), // 最后访问后 1 小时过期
		StatsRecorder:    counter,
	})
	if err != nil {
		return errors.WithStackIf(err)
	}
	r.cache = cache
	go func() {
		<-ctx.Done()
		cache.CleanUp()
	}()
	return nil
}

// Stats returns rate limiter cache statistics.
func (r *RateLimit) Stats() stats.Stats {
	return r.cache.Stats()
}

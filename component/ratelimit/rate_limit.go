package rate_limit

import (
	"context"
	"time"

	"emperror.dev/errors"
	config2 "github.com/chuccp/go-web-frame/config"
	"github.com/maypok86/otter/v2"
	"github.com/maypok86/otter/v2/stats"
	"golang.org/x/time/rate"
)

type Config struct {
	Limit   int // 每秒限制
	Burst   int // 最大令牌数
	MaxSize int // 最大缓存数量
	Expiry  int // 缓存过期时间 单位秒
}

type RateLimit struct {
	cache         *otter.Cache[string, *rate.Limiter]
	limiterLoader otter.Loader[string, *rate.Limiter]
}

// Allow 瞬间检查是否允许（不阻塞，直接返回 false 拒绝）
func (r *RateLimit) Allow(ctx context.Context, key string) bool {
	limiter, err := r.cache.Get(ctx, key, r.limiterLoader)
	if err != nil {
		return false
	}
	return limiter.Allow()
}

// Wait 阻塞等待直到允许通过（推荐用于严格限流）
func (r *RateLimit) Wait(ctx context.Context, key string) error {
	limiter, err := r.cache.Get(ctx, key, r.limiterLoader)
	if err != nil {
		return err
	}
	return limiter.Wait(ctx)
}
func (r *RateLimit) Init(config config2.IConfig) error {
	lConfig := &Config{
		Limit:   600,
		Burst:   3,
		MaxSize: 1000_000,
		Expiry:  3600,
	}
	err := config.Unmarshal("rate_limit", lConfig)
	if err != nil {
		return errors.WithStackIf(err)
	}
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
	return nil
}

func (r *RateLimit) Destroy() error {
	// otter v2 支持主动关闭，释放内部资源
	if stopped := r.cache.StopAllGoroutines(); stopped {
		// 可选：记录日志 "otter cache goroutines stopped"
	}
	r.cache.CleanUp()
	return nil
}

// Stats o可选：获取缓存统计（命中率、驱逐数等）
func (r *RateLimit) Stats() stats.Stats {
	return r.cache.Stats()
}

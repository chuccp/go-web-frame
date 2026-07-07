package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	wf "github.com/chuccp/go-web-frame"
	config2 "github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/web"
	"go.uber.org/zap"
)

// ========== 常用中间件示例 ==========

// 1. CORS 中间件 - 跨域资源共享
type CorsFilter struct {
	core.IFilter
	allowedOrigins []string
}

func NewCorsFilter(origins ...string) *CorsFilter {
	return &CorsFilter{
		allowedOrigins: origins,
	}
}

func (f *CorsFilter) Init(ctx *core.Context) error {
	log.Info("CORS filter initialized")
	return nil
}

func (f *CorsFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
	origin := req.GetHeader("Origin")

	// 检查是否允许该来源
	allowed := false
	for _, allowedOrigin := range f.allowedOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			allowed = true
			break
		}
	}

	if allowed {
		resp := req.Response()
		resp.Header().Set("Access-Control-Allow-Origin", origin)
		resp.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		resp.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		resp.Header().Set("Access-Control-Allow-Credentials", "true")
		resp.Header().Set("Access-Control-Max-Age", "86400")
	}

	// 处理 OPTIONS 预检请求
	if req.Request().Method == "OPTIONS" {
		return nil, nil
	}

	return fc.Next()
}

// 2. 简单内存缓存中间件
type MemoryCacheFilter struct {
	core.IFilter
	cache map[string]cacheEntry
	ttl   time.Duration
	mu    sync.RWMutex
}

type cacheEntry struct {
	value     any
	expiresAt time.Time
}

func NewMemoryCacheFilter(ttl time.Duration) *MemoryCacheFilter {
	return &MemoryCacheFilter{
		cache: make(map[string]cacheEntry),
		ttl:   ttl,
	}
}

func (f *MemoryCacheFilter) Init(ctx *core.Context) error {
	log.Info("Memory cache filter initialized")
	return nil
}

func (f *MemoryCacheFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
	// 只缓存 GET 请求
	if req.Request().Method != "GET" {
		return fc.Next()
	}

	// 生成缓存键
	cacheKey := fmt.Sprintf("cache:%s:%s", req.Request().Method, req.FullPath())

	// 尝试从缓存获取
	f.mu.RLock()
	entry, exists := f.cache[cacheKey]
	f.mu.RUnlock()

	if exists && time.Now().Before(entry.expiresAt) {
		log.Debug("Cache hit", zap.String("key", cacheKey))
		return entry.value, nil
	}

	// 执行后续处理
	result, err := fc.Next()
	if err != nil {
		return nil, err
	}

	// 缓存结果
	if result != nil {
		f.mu.Lock()
		f.cache[cacheKey] = cacheEntry{
			value:     result,
			expiresAt: time.Now().Add(f.ttl),
		}
		f.mu.Unlock()
		log.Debug("Cache set", zap.String("key", cacheKey))
	}

	return result, nil
}

// 3. 限流中间件
// 使用方式: 安装 github.com/chuccp/go-web-frame/component/ratelimit
// 然后在 Init 中通过 core.GetService 获取 ratelimit.RateLimit 实例
type RateLimitFilter struct {
	core.IFilter
}

func NewRateLimitFilter() *RateLimitFilter {
	return &RateLimitFilter{}
}

func (f *RateLimitFilter) Init(ctx *core.Context) error {
	log.Info("Rate limit filter initialized")
	return nil
}

func (f *RateLimitFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
	// 使用客户端 IP 作为限流键
	// 实际项目中，通过 core.GetService[*ratelimit.RateLimit](ctx) 获取限流器
	// allowed := rateLimiter.Allow(req.ClientIP())
	return fc.Next()
}

// 4. 安全头中间件
type SecurityFilter struct {
	core.IFilter
}

func NewSecurityFilter() *SecurityFilter {
	return &SecurityFilter{}
}

func (f *SecurityFilter) Init(ctx *core.Context) error {
	log.Info("Security filter initialized")
	return nil
}

func (f *SecurityFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
	result, err := fc.Next()

	// 添加安全头
	resp := req.Response()
	resp.Header().Set("X-Content-Type-Options", "nosniff")
	resp.Header().Set("X-Frame-Options", "DENY")
	resp.Header().Set("X-XSS-Protection", "1; mode=block")
	resp.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	resp.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

	return result, err
}

// 5. 请求日志中间件
type LoggingFilter struct {
	core.IFilter
}

func NewLoggingFilter() *LoggingFilter {
	return &LoggingFilter{}
}

func (f *LoggingFilter) Init(ctx *core.Context) error {
	log.Info("Logging filter initialized")
	return nil
}

func (f *LoggingFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
	start := time.Now()

	// 记录请求信息
	log.Info("Incoming request",
		zap.String("method", req.Request().Method),
		zap.String("path", req.FullPath()),
		zap.String("remote_addr", req.RemoteAddr()),
		zap.String("user_agent", req.GetHeader("User-Agent")),
	)

	// 执行请求
	result, err := fc.Next()

	// 记录响应信息
	elapsed := time.Since(start)
	if err != nil {
		log.Error("Request failed",
			zap.String("method", req.Request().Method),
			zap.String("path", req.FullPath()),
			zap.Duration("elapsed", elapsed),
			zap.Error(err),
		)
	} else {
		log.Info("Request completed",
			zap.String("method", req.Request().Method),
			zap.String("path", req.FullPath()),
			zap.Duration("elapsed", elapsed),
		)
	}

	return result, err
}

// 6. 恢复中间件 - 捕获 panic
type RecoveryFilter struct {
	core.IFilter
}

func NewRecoveryFilter() *RecoveryFilter {
	return &RecoveryFilter{}
}

func (f *RecoveryFilter) Init(ctx *core.Context) error {
	log.Info("Recovery filter initialized")
	return nil
}

func (f *RecoveryFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("Panic recovered",
				zap.Any("error", r),
				zap.String("path", req.FullPath()),
			)
		}
	}()

	return fc.Next()
}

// 7. 请求体大小限制中间件
type BodySizeLimitFilter struct {
	core.IFilter
	maxSize int64 // 最大字节数
}

func NewBodySizeLimitFilter(maxSize int64) *BodySizeLimitFilter {
	return &BodySizeLimitFilter{
		maxSize: maxSize,
	}
}

func (f *BodySizeLimitFilter) Init(ctx *core.Context) error {
	log.Info("Body size limit filter initialized", zap.Int64("max_size", f.maxSize))
	return nil
}

func (f *BodySizeLimitFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
	if req.Request().ContentLength > f.maxSize {
		log.Warn("Request body too large",
			zap.Int64("size", req.Request().ContentLength),
			zap.Int64("max", f.maxSize),
		)
		return nil, errors.New("request body too large")
	}

	return fc.Next()
}

// 8. API 版本控制中间件
type APIVersionFilter struct {
	core.IFilter
	version string
}

func NewAPIVersionFilter(version string) *APIVersionFilter {
	return &APIVersionFilter{
		version: version,
	}
}

func (f *APIVersionFilter) Init(ctx *core.Context) error {
	log.Info("API version filter initialized", zap.String("version", f.version))
	return nil
}

func (f *APIVersionFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
	// 在响应头中添加 API 版本
	req.Response().Header().Set("API-Version", f.version)

	// 检查客户端请求的版本
	clientVersion := req.GetHeader("API-Version")
	if clientVersion != "" && clientVersion != f.version {
		return nil, fmt.Errorf("unsupported API version: %s, current version: %s", clientVersion, f.version)
	}

	return fc.Next()
}

// ========== 使用示例 ==========

func main() {
	builder := wf.NewBuilder(config2.LoadAutoConfig())

	// 添加组件（用于限流中间件）
	// 需要安装: go get github.com/chuccp/go-web-frame/component/ratelimit
	// 然后: builder.Service(&ratelimit.RateLimit{})

	// 添加中间件（顺序很重要）
	// 1. 恢复中间件 - 最先执行，捕获 panic
	builder.Filter(NewRecoveryFilter())

	// 2. 日志中间件 - 记录所有请求
	builder.Filter(NewLoggingFilter())

	// 3. 安全头中间件 - 添加安全响应头
	builder.Filter(NewSecurityFilter())

	// 4. CORS 中间件 - 处理跨域
	builder.Filter(NewCorsFilter("*"))

	// 5. API 版本控制中间件
	builder.Filter(NewAPIVersionFilter("v1.0"))

	// 6. 限流中间件 - 限制请求频率
	builder.Filter(NewRateLimitFilter())

	// 7. 请求体大小限制中间件
	builder.Filter(NewBodySizeLimitFilter(10 * 1024 * 1024)) // 10MB

	// 8. 缓存中间件 - 缓存 GET 请求
	builder.Filter(NewMemoryCacheFilter(5*time.Minute))

	// 注册路由
	builder.Get("/", func(c *web.Request) (any, error) {
		return map[string]string{
			"message": "Hello, World!",
			"version": "v1.0",
		}, nil
	})

	builder.Get("/data", func(c *web.Request) (any, error) {
		// 这个接口会被缓存 5 分钟
		return map[string]any{
			"data":      []string{"item1", "item2", "item3"},
			"timestamp": time.Now().Unix(),
		}, nil
	})

	builder.Post("/api/test", func(c *web.Request) (any, error) {
		return map[string]string{
			"status": "success",
		}, nil
	})

	// 创建一个需要认证的 REST 组
	serverConfig := web.DefaultServerConfig()
	authGroup := core.NewRestGroupBuilder().ServerConfig(serverConfig).Build()
	authGroup.AddFilter(&AuthFilter{})

	// 添加 REST 控制器
	authGroup.AddRest(&ProtectedController{})

	// Add the rest group to builder
	builder.RestGroup(authGroup)

	app := builder.Build()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := app.Run(ctx); err != nil {
		log.PrintPanic(err)
	}
}

// AuthFilter 认证中间件
type AuthFilter struct {
	core.IFilter
}

func (f *AuthFilter) Init(ctx *core.Context) error {
	log.Info("Auth filter initialized")
	return nil
}

func (f *AuthFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
	token := req.GetHeader("Authorization")
	if token == "" {
		return nil, errors.New("authorization required")
	}

	// 这里应该验证 token
	// 简化示例：假设 token 格式为 "Bearer <token>"
	if len(token) < 7 || token[:7] != "Bearer " {
		return nil, errors.New("invalid authorization format")
	}

	log.Info("User authenticated", zap.String("token", token[:20]+"..."))
	return fc.Next()
}

// ProtectedController 受保护的控制器
type ProtectedController struct {
	core.IRest
}

func (c *ProtectedController) Init(ctx *core.Context) error {
	ctx.Get("/api/protected/data", func(c *web.Request) (any, error) {
		return map[string]string{
			"message": "This is protected data",
		}, nil
	})
	return nil
}
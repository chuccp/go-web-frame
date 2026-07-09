# 过滤器/中间件

过滤器用于处理 HTTP 请求的横切关注点，如认证、日志、限流、CORS 等。

## 创建过滤器

嵌入 `core.IFilter` 接口：

```go
package filter

import (
    "errors"
    "github.com/chuccp/go-web-frame/core"
    "github.com/chuccp/go-web-frame/web"
)

type AuthFilter struct {
    core.IFilter
}

func (f *AuthFilter) Init(ctx *core.Context) error {
    return nil
}

func (f *AuthFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    // 前置处理
    token := req.GetHeader("Authorization")
    if token == "" {
        return nil, errors.New("unauthorized")
    }

    // 调用下一个处理器
    result, err := fc.Next()

    // 后置处理（可选）
    return result, err
}
```

### 注册过滤器

```go
package main

import (
    "context"
    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/config"
    "myapp/filter"
)

func main() {
    cfg, _ := config.LoadSingleFileConfig("application.yml")
    builder := wf.NewBuilder(cfg)

    builder.Filter(&filter.LoggingFilter{}, &filter.AuthFilter{})

    builder.Build().Run(context.Background())
}
```

## 过滤器链

多个过滤器形成链式调用：

```
请求 → Filter1 → Filter2 → ... → Handler → 响应
```

执行顺序按照注册顺序：

```go
builder.Filter(&filter.LoggingFilter{})   // 第 1 个执行
builder.Filter(&filter.AuthFilter{})      // 第 2 个执行
builder.Filter(&filter.RateLimitFilter{}) // 第 3 个执行
```

## 常见过滤器示例

### 日志过滤器

```go
type LoggingFilter struct {
    core.IFilter
}

func (f *LoggingFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    start := time.Now()
    log.Info("request started",
        zap.String("method", req.Request().Method),
        zap.String("path", req.Request().URL.Path),
    )

    result, err := fc.Next()

    log.Info("request completed",
        zap.Duration("duration", time.Since(start)),
        zap.Error(err),
    )

    return result, err
}
```

### 认证过滤器

```go
type AuthFilter struct {
    core.IFilter
}

func (f *AuthFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    token := req.GetHeader("Authorization")
    if token == "" {
        return nil, errors.New("unauthorized")
    }
    // 验证 token...
    return fc.Next()
}
```

### 内置 CORS 过滤器

```go
import "github.com/chuccp/go-web-frame/component/cors"

builder.Filter(&cors.Filter{})
```

## 路由元数据（WithMeta）

### 定义元数据选项

```go
package controller

import "github.com/chuccp/go-web-frame/web"

// 不需要认证
func Public() web.MetaOption {
    return web.WithValue("public", true)
}

// 需要认证
func RequireAuth() web.MetaOption {
    return web.WithValue("auth", true)
}

// 需要特定权限
func RequirePermission(perm string) web.MetaOption {
    return web.WithValue("permission", perm)
}
```

### 注册带元数据的路由

```go
func (c *UserController) Init(ctx *core.Context) error {
    ctx.Get("/login", c.Login).WithMeta(Public())
    ctx.Get("/dashboard", c.Dashboard).WithMeta(RequireAuth())
    ctx.Delete("/users/:id", c.DeleteUser).
        WithMeta(RequireAuth(), RequirePermission("admin:delete"))
    return nil
}
```

### 在过滤器中使用元数据

```go
func (f *AuthFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    if req.HasMeta(Public()) {
        return fc.Next()
    }

    if req.HasMeta(RequireAuth()) {
        token := req.GetHeader("Authorization")
        if token == "" {
            return nil, errors.New("unauthorized")
        }
        // 验证 token、检查 permission...
    }

    return fc.Next()
}
```

## 限流过滤器示例

使用框架内置限流组件：

```go
package main

import (
    "context"
    "errors"
    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/component/ratelimit"
    "github.com/chuccp/go-web-frame/config"
    "github.com/chuccp/go-web-frame/core"
    "github.com/chuccp/go-web-frame/web"
)

type DataController struct {
    core.IService
    rateLimit *ratelimit.RateLimit
}

func (c *DataController) Init(ctx *core.Context) error {
    c.rateLimit = wf.GetService[*ratelimit.RateLimit](ctx)
    ctx.Get("/api/data", c.Handle)
    return nil
}

func (c *DataController) Handle(req *web.Request) (any, error) {
    if !c.rateLimit.Allow(req.ClientIP()) {
        return nil, errors.New("rate limited")
    }
    return "ok", nil
}

func main() {
    cfg, _ := config.LoadSingleFileConfig("application.yml")
    builder := wf.NewBuilder(cfg)
    builder.Service(&ratelimit.RateLimit{})
    builder.Rest(&DataController{})
    builder.Build().Run(context.Background())
}
```

`application.yml`：

```yaml
rate_limit:
  limit: 600    # 每多少秒补充 1 个令牌
  burst: 5      # 最大 burst 大小
  maxSize: 1000000
  expiry: 3600  # 缓存过期时间（秒）
```

## 复用 Gin 生态中间件

通过 `req.GinContext()` 暴露底层 `*gin.Context`，可以直接使用 `gin-contrib` 中间件：

```go
import (
    "github.com/gin-contrib/gzip"
    "github.com/chuccp/go-web-frame/core"
    "github.com/chuccp/go-web-frame/web"
)

type GzipFilter struct{ core.IFilter }

func (f *GzipFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    gzip.Gzip(gzip.DefaultCompression)(req.GinContext())
    return fc.Next()
}
```

## 完整示例

```go
package main

import (
    "context"
    "errors"
    "time"
    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/config"
    "github.com/chuccp/go-web-frame/core"
    "github.com/chuccp/go-web-frame/web"
    "go.uber.org/zap"
)

type LoggingFilter struct{ core.IFilter }

func (f *LoggingFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    start := time.Now()
    log.Info("request started",
        zap.String("method", req.Request().Method),
        zap.String("path", req.Request().URL.Path),
    )
    result, err := fc.Next()
    log.Info("request completed",
        zap.Duration("duration", time.Since(start)),
        zap.Error(err),
    )
    return result, err
}

type AuthFilter struct{ core.IFilter }

func Public() web.MetaOption { return web.WithValue("public", true) }

func (f *AuthFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    if req.HasMeta(Public()) {
        return fc.Next()
    }
    if req.GetHeader("Authorization") == "" {
        return nil, errors.New("unauthorized")
    }
    return fc.Next()
}

type UserController struct{ core.IService }

func (c *UserController) Init(ctx *core.Context) error {
    ctx.Get("/login", c.Login).WithMeta(Public())
    ctx.Get("/users", c.List).WithMeta(RequireAuth())
    return nil
}

func RequireAuth() web.MetaOption { return web.WithValue("auth", true) }

func (c *UserController) Login(req *web.Request) (any, error) {
    return map[string]any{"token": "xxx"}, nil
}

func (c *UserController) List(req *web.Request) (any, error) {
    return []map[string]any{{"id": 1, "name": "Alice"}}, nil
}

func main() {
    cfg, _ := config.LoadSingleFileConfig("application.yml")
    builder := wf.NewBuilder(cfg)
    builder.Filter(&LoggingFilter{}, &AuthFilter{})
    builder.Rest(&UserController{})
    builder.Build().Run(context.Background())
}
```

## 下一步

- [配置](configuration.md) - 了解配置管理
- [服务](service.md) - 业务逻辑层与依赖注入
- [高级主题](../advanced/database.md) - 了解更多高级功能

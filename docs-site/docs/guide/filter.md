# 过滤器/中间件

过滤器用于处理 HTTP 请求的横切关注点（如认证、日志、限流等）。

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

func (f *AuthFilter) Init(context *core.Context) error {
    // 初始化逻辑
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

在 `main.go` 中注册：

```go
package main

import (
    config2 "github.com/chuccp/go-web-frame/config"
    wf "github.com/chuccp/go-web-frame"
    "go.uber.org/zap"
    "myapp/filter"
)

func createApp() (*wf.WebFrame, error) {
    // 加载配置
    fileConfig, err := config2.LoadSingleFileConfig("config.ini")
    if err != nil {
        return nil, err
    }
    
    // 创建 Builder
    builder := wf.NewBuilder(fileConfig)
    
    // 注册过滤器
    builder.Filter(&filter.LoggingFilter{}, &filter.AuthFilter{})
    
    // 构建应用
    app := builder.Build()
    return app, nil
}

func main() {
    app, err := createApp()
    if err != nil {
        zap.L().Fatal("创建应用失败", zap.Error(err))
        return
    }
    err = app.Start()
    if err != nil {
        zap.L().Fatal("启动应用失败", zap.Error(err))
    }
}
```

## 过滤器链

多个过滤器形成过滤器链：

```
请求 → Filter1 → Filter2 → ... → Handler → 响应
```

### 执行顺序

过滤器的执行顺序按照注册顺序：

```go
builder := wf.NewBuilder(cfg)
builder.Filter(&filter.LoggingFilter{})   // 第 1 个执行
builder.Filter(&filter.AuthFilter{})      // 第 2 个执行
builder.Filter(&filter.RateLimitFilter{}) // 第 3 个执行
app := builder.Build()
```

## 常见过滤器示例

### 日志过滤器

```go
type LoggingFilter struct {
    core.IFilter
}

func (f *LoggingFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    start := time.Now()
    
    // 记录请求开始
    log.Info("request started",
        zap.String("method", req.Request().Method),
        zap.String("path", req.Request().URL.Path),
    )
    
    // 执行下一个处理器
    result, err := fc.Next()
    
    // 记录请求结束
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
    jwtSecret string
}

func (f *AuthFilter) Init(context *core.Context) error {
    f.jwtSecret = context.GetConfig().GetString("jwt.secret")
    return nil
}

func (f *AuthFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    // 获取 Token
    token := req.GetHeader("Authorization")
    if token == "" {
        return nil, errors.New("unauthorized")
    }
    
    // 验证 Token
    claims, err := f.validateToken(token)
    if err != nil {
        return nil, errors.New("invalid token")
    }
    
    // 存储用户信息到请求上下文
    req.Set("user", claims)
    
    return fc.Next()
}

func (f *AuthFilter) validateToken(token string) (map[string]any, error) {
    // 实现 JWT 验证逻辑
    return nil, nil
}
```

### CORS 过滤器

```go
type CORSFilter struct {
    core.IFilter
}

func (f *CORSFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    // 设置 CORS 头
    req.Response().Header().Set("Access-Control-Allow-Origin", "*")
    req.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    req.Response().Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
    
    // 处理 OPTIONS 请求
    if req.Request().Method == "OPTIONS" {
        return nil, nil
    }
    
    return fc.Next()
}
```

### 限流过滤器

使用框架内置的限流组件：

```go
package main

import (
    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/component/ratelimit"
    "github.com/chuccp/go-web-frame/core"
    "github.com/chuccp/go-web-frame/web"
)

type MyController struct {
    core.IService
    rateLimit *ratelimit.RateLimit
}

func (c *MyController) Init(ctx *core.Context) error {
    c.rateLimit = wf.GetComponent[*ratelimit.RateLimit](ctx)
    ctx.Get("/api/data", c.Handle)
    return nil
}

func (c *MyController) Handle(req *web.Request) (any, error) {
    // 使用限流检查
    if !c.rateLimit.AllowSBurst(req.ClientIP(), 5) {
        return nil, errors.New("请求过于频繁")
    }
    return "ok", nil
}

func main() {
    builder := wf.NewBuilder(cfg)
    builder.Component(&ratelimit.RateLimit{})
    builder.Rest(&MyController{})
    app := builder.Build()
    app.Start()
}
```

在配置文件中设置限流参数（`application.yml`）：

```yaml
rate_limit:
  limit: 600     # 每秒限制
  burst: 5       # 最大令牌数
  maxSize: 1000000
  expiry: 3600   # 缓存过期时间（秒）
```

## 路由元数据

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

// 需要特定角色
func RequireRole(role string) web.MetaOption {
    return web.WithValue("role", role)
}
```

### 注册带元数据的路由

```go
func (c *UserController) Init(context *core.Context) error {
    context.Get("/login", c.Login).WithMeta(Public())
    context.Get("/dashboard", c.Dashboard).WithMeta(RequireAuth())
    context.Delete("/users/:id", c.DeleteUser).WithMeta(RequireAuth(), RequireRole("admin"))
    return nil
}
```

### 在过滤器中使用元数据

```go
func (f *AuthFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    meta := req.HandlerMeta()
    
    // 跳过公开路由
    if meta.Has("public") {
        return fc.Next()
    }
    
    // 检查认证
    if meta.Has("auth") {
        token := req.GetHeader("Authorization")
        if token == "" {
            return nil, errors.New("unauthorized")
        }
        
        // 验证 Token
        claims, err := f.validateToken(token)
        if err != nil {
            return nil, errors.New("unauthorized")
        }
        
        // 检查角色
        requiredRole, _ := meta.Get("role").(string)
        if requiredRole != "" && claims["role"] != requiredRole {
            return nil, errors.New("forbidden")
        }
    }
    
    return fc.Next()
}
```

## 完整示例

```go
package main

import (
    "errors"
    "time"
    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/core"
    "github.com/chuccp/go-web-frame/web"
    "go.uber.org/zap"
)

// 日志过滤器
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

// 认证过滤器
type AuthFilter struct {
    core.IFilter
}

func (f *AuthFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    meta := req.HandlerMeta()
    
    // 跳过公开路由
    if meta.Has("public") {
        return fc.Next()
    }
    
    token := req.GetHeader("Authorization")
    if token == "" {
        return nil, errors.New("unauthorized")
    }
    
    return fc.Next()
}

// 控制器
type UserController struct {
    core.IService
}

func (c *UserController) Init(context *core.Context) error {
    context.Get("/login", c.Login).WithMeta(Public())
    context.Get("/users", c.List).WithMeta(RequireAuth())
    return nil
}

func Public() web.MetaOption {
    return web.WithValue("public", true)
}

func RequireAuth() web.MetaOption {
    return web.WithValue("auth", true)
}

func (c *UserController) Login(req *web.Request) (any, error) {
    return map[string]any{"token": "xxx"}, nil
}

func (c *UserController) List(req *web.Request) (any, error) {
    return []map[string]any{{"id": float64(1), "name": "Alice"}}, nil
}

func main() {
    // 使用 Builder 注册组件
    builder := wf.NewBuilder(config.LoadAutoConfig())
    
    // 注册过滤器（顺序很重要）
    builder.Filter(&LoggingFilter{}, &AuthFilter{})
    
    // 注册控制器
    builder.Rest(&UserController{})
    
    app := builder.Build()
    app.Start()
}
```

## 下一步

- [配置](configuration.md) - 了解配置管理
- [高级主题](advanced/database.md) - 了解更多高级功能

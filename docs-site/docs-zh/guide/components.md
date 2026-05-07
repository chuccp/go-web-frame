# 组件

组件是框架中独立的功能模块，实现 `core.IService` 接口，通过 `Service()` 注册，可以注册到应用中供其他组件使用。

## 创建组件

实现 `core.IService` 接口：

```go
package main

import (
    "github.com/chuccp/go-web-frame/core"
)

type CacheComponent struct {
    core.IService
    cache map[string]string
}

func (c *CacheComponent) Init(ctx *core.Context) error {
    c.cache = make(map[string]string)
    return nil
}

func (c *CacheComponent) Get(key string) string {
    return c.cache[key]
}

func (c *CacheComponent) Set(key, value string) {
    c.cache[key] = value
}
```

### 注册组件

```go
builder := wf.NewBuilder(cfg)
builder.Service(&CacheComponent{})
app := builder.Build()
app.Start()
```

### 在控制器中使用

```go
type MyController struct {
    core.IService
    cache *CacheComponent
}

func (c *MyController) Init(ctx *core.Context) error {
    c.cache = wf.GetService[*CacheComponent](ctx)
    ctx.Get("/cache", c.Handle)
    return nil
}
```

## 框架内置组件

### 限流组件（ratelimit）

基于令牌桶算法的限流组件，支持单机限流：

```go
package main

import (
    "errors"
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
    c.rateLimit = wf.GetService[*ratelimit.RateLimit](ctx)
    ctx.Get("/api/data", c.Handle)
    return nil
}

func (c *MyController) Handle(req *web.Request) (any, error) {
    // 检查限流（每秒最多 5 次，突发 5 次）
    if !c.rateLimit.AllowSBurst(req.ClientIP(), 5) {
        return nil, errors.New("请求过于频繁")
    }
    return "ok", nil
}

func main() {
    builder := wf.NewBuilder(cfg)
    builder.Service(&ratelimit.RateLimit{})
    builder.Rest(&MyController{})
    app := builder.Build()
    app.Start()
}
```

配置（`application.yml`）：

```yaml
rate_limit:
  limit: 600     # 每秒限制
  burst: 5       # 最大令牌数
  maxSize: 1000000
  expiry: 3600   # 缓存过期时间（秒）
```

### 认证组件（auth）

提供基于 Token 的认证过滤器和登录/登出功能：

```go
import auth2 "github.com/chuccp/go-web-frame/component/auth"
```

主要功能：

- `auth2.NewAuthenticationFilter(authenticator)` - 创建认证过滤器
- `auth2.WithLogin()` - 路由元数据，标记需要登录的路由
- `auth.User(req, ctx)` - 从请求中获取当前登录用户

完整示例：

```go
type UserController struct {
    core.IService
    authFilter *auth2.AuthenticationFilter
}

func (c *UserController) Init(ctx *core.Context) error {
    c.authFilter = wf.GetFilter[*auth2.AuthenticationFilter](ctx)

    // 登录接口（不需要认证）
    ctx.Post("/signIn", c.signIn)
    ctx.Post("/logout", c.logout)

    // 业务接口（需要认证）
    ctx.Get("/users", c.List).WithMeta(auth2.WithLogin())
    return nil
}

func (c *UserController) signIn(req *web.Request) (any, error) {
    // 验证用户名密码后登录
    loginUser := &LoginUser{Id: user.Id, IsAdmin: user.IsAdmin}
    return c.authFilter.SignIn(loginUser, req)
}

func (c *UserController) logout(req *web.Request) (any, error) {
    return c.authFilter.SignOut(req)
}
```

注册认证过滤器到 RestGroup：

```go
restGroupBuilder := core.NewRestGroupBuilder()
restGroupBuilder.Rest(&UserController{})
restGroupBuilder.Port(8081)
restGroupBuilder.Filter(auth2.NewAuthenticationFilter(&MyAuthenticator{}))
builder.RestGroup(restGroupBuilder.Build())
```

### 定时任务调度（schedule）

支持 cron 表达式的定时任务调度组件：

```go
import "github.com/chuccp/go-web-frame/component/schedule"
```

主要功能：

- `schedule.NewScheduleWithSeconds()` - 创建秒级精度的调度器
- `schedule.AddFunc(cronExpr, func)` - 添加定时任务
- `schedule.AddIdOrReplaceKeyFunc(id, key, cronExpr, func)` - 按 ID 添加或按 key 替换
- `schedule.StopIdFunc(id)` - 停止指定任务
- `schedule.GetIds()` - 获取所有已注册的任务 ID

使用示例（见 [Runner 文档](runner.md)）：

```go
// 注册调度组件
builder.Runner(schedule.NewScheduleWithSeconds())

// 在 Runner 中使用
func (r *MyRunner) Init(c *core.Context) error {
    r.schedule = core.GetRunner[*schedule.Schedule](c)
    _, err := r.schedule.AddFunc("0 */5 * * * ?", func(c *core.Context) {
        // 每 5 分钟执行
    })
    return err
}
```

### 其他组件

| 组件 | 包路径 | 说明 |
|------|--------|------|
| 内存缓存 | `component/cache` | 基于 Otter 的高性能内存缓存，支持过期时间 |
| 本地缓存 | `component/localcache` | 基于文件的本地缓存，支持 Base64 文件存储 |
| 验证码 | `component/captcha` | 滑块验证码生成和验证 |
| 输入验证 | `component/validator` | 基于 go-playground/validator 的输入验证 |
| CORS | `component/cors` | 跨域资源共享过滤器 |

### 内存缓存组件（cache）

基于 [Otter](https://github.com/maypok86/otter) 的高性能内存缓存组件：

```go
import "github.com/chuccp/go-web-frame/component/cache"

type MyController struct {
    core.IService
    cache *cache.Cache
}

func (c *MyController) Init(ctx *core.Context) error {
    c.cache = wf.GetService[*cache.Cache](ctx)
    ctx.Get("/api/data", c.Handle)
    return nil
}

func (c *MyController) Handle(req *web.Request) (any, error) {
    // 设置缓存
    c.cache.Set("key", "value")

    // 获取缓存
    val, exists := c.cache.Get("key")

    // 仅在不存在时设置（带过期时间）
    c.cache.SetNX("temp", "data", 5*time.Minute)

    // 如果不存在则计算并缓存
    c.cache.ComputeIfAbsent("expensive_key", func() {
        // 执行耗时计算，结果自动缓存
    })

    // 使缓存失效
    c.cache.Invalidate("key")

    // 获取缓存统计
    stats := c.cache.Stats()

    return val, nil
}
```

注册缓存组件：

```go
builder.Service(&cache.Cache{})
```

### 本地缓存组件（localcache）

基于文件的本地缓存，适合缓存生成的文件（如图片、PDF 等）：

```go
import "github.com/chuccp/go-web-frame/component/localcache"

type MyController struct {
    core.IService
    localCache *localcache.LocalCache
}

func (c *MyController) Init(ctx *core.Context) error {
    c.localCache = wf.GetService[*localcache.LocalCache](ctx)
    ctx.Get("/report", c.GetReport)
    return nil
}

func (c *MyController) GetReport(req *web.Request) (any, error) {
    // 生成文件并缓存（如果已缓存则直接返回）
    file, err := c.localCache.GetFile(func(value ...any) ([]byte, error) {
        // 生成文件内容
        return generateReport()
    }, "report", "2024")
    if err != nil {
        return nil, err
    }
    return file, nil
}
```

配置（`application.yml`）：

```yaml
local_cache:
  open: true        # 是否启用
  path: ./cache     # 缓存文件存储目录
```

注册组件：

```go
builder.Service(&localcache.LocalCache{})
```

### 滑块验证码组件（captcha）

提供滑块验证码的生成和验证功能：

```go
import "github.com/chuccp/go-web-frame/component/captcha"

type CaptchaController struct {
    core.IService
    captcha *captcha.Captcha
}

func (c *CaptchaController) Init(ctx *core.Context) error {
    c.captcha = wf.GetService[*captcha.Captcha](ctx)
    ctx.Get("/captcha", c.GetCaptcha)
    ctx.Post("/captcha/verify", c.Verify)
    return nil
}

// 获取验证码
func (c *CaptchaController) GetCaptcha(req *web.Request) (any, error) {
    data, err := c.captcha.GetCaptchaData()
    if err != nil {
        return nil, err
    }
    // 返回给前端：data.TileImage（滑块图）、data.MasterImage（背景图）、data.ThumbCode（加密验证码）
    return data, nil
}

// 验证用户滑块位置
func (c *CaptchaController) Verify(req *web.Request) (any, error) {
    thumbCode := req.Query("thumbCode")  // 前端传回的加密验证码
    x := req.Query("x")                  // 用户滑动的 x 坐标

    data, valid := c.captcha.ValidateThumb(thumbCode, x)
    if !valid {
        return nil, errors.New("验证码验证失败")
    }
    // data.Code 为验证通过的凭证码，可用于后续业务验证
    return web.Data(data), nil
}
```

配置（`application.yml`）：

```yaml
captcha:
  codeKey: "your-encryption-key"   # AES 加密密钥（16 字节）
  codeIv: "your-encryption-iv"     # AES 加密 IV（16 字节）
```

注册组件：

```go
builder.Service(&captcha.Captcha{})
```

### 输入验证组件（validator）

基于 [go-playground/validator](https://github.com/go-playground/validator) 的输入验证，内置自定义规则：

```go
import "github.com/chuccp/go-web-frame/component/validator"

type MyController struct {
    core.IService
    validator *validator.Validator
}

func (c *MyController) Init(ctx *core.Context) error {
    c.validator = wf.GetService[*validator.Validator](ctx)
    ctx.Post("/users", c.Create)
    return nil
}

type CreateUserInput struct {
    Name     string `json:"name" validate:"required"`
    Phone    string `json:"phone" validate:"mobile"`      // 中国手机号验证
    Password string `json:"password" validate:"password"`  // 强密码验证
}

func (c *MyController) Create(req *web.Request) (any, error) {
    var input CreateUserInput
    if err := req.BindJSON(&input); err != nil {
        return nil, err
    }
    // 使用验证器验证
    if err := c.validator.Validate(input); err != nil {
        return nil, err
    }
    // 创建用户...
    return web.Ok(), nil
}
```

内置验证规则：

| 规则 | 说明 |
|------|------|
| `mobile` | 验证中国手机号（支持 +86 前缀、空格、连字符） |
| `password` | 强密码验证（最少 8 位，需包含大写、小写和数字） |

注册组件：

```go
builder.Service(&validator.Validator{})
```

### CORS 过滤器（cors）

跨域资源共享过滤器，基于 [gin-contrib/cors](https://github.com/gin-contrib/cors)：

```go
import "github.com/chuccp/go-web-frame/component/cors"

// 注册 CORS 过滤器
builder.Filter(cors.NewCrosFilter())
```

默认策略：

- 允许所有来源（`AllowOriginFunc: returns true`）
- 允许凭证（`AllowCredentials: true`）
- 允许的请求头：`Origin`、`Content-Length`、`Content-Type`、`Authorization`
- 自动处理 OPTIONS 预检请求

如需自定义 CORS 策略，可以创建自己的过滤器并在其中配置 `gin-contrib/cors`。

### Redis 组件（redis）

基于 [go-redis/v9](https://github.com/redis/go-redis) 的 Redis 客户端组件：

```go
import "github.com/chuccp/go-web-frame/redis"

type MyController struct {
    core.IService
    redisClient *redis.Component
}

func (c *MyController) Init(ctx *core.Context) error {
    c.redisClient = wf.GetService[*redis.Component](ctx)
    ctx.Get("/cache", c.GetCache)
    ctx.Post("/cache", c.SetCache)
    return nil
}

func (c *MyController) GetCache(req *web.Request) (any, error) {
    // 获取 go-redis 客户端
    client := c.redisClient.GetClient()

    // 基本操作
    val, err := client.Get(req.Request().Context(), "key").Result()
    if err != nil {
        return nil, err
    }
    return val, nil
}

func (c *MyController) SetCache(req *web.Request) (any, error) {
    client := c.redisClient.GetClient()

    // 设置键值（带过期时间）
    err := client.Set(req.Request().Context(), "key", "value", 10*time.Minute).Err()
    if err != nil {
        return nil, err
    }
    return web.Ok(), nil
}
```

注册组件：

```go
builder.Service(&redis.Component{})
```

配置（`application.yml`）：

```yaml
redis:
  addr: localhost:6379
  password: ""
  db: 0
  pool_size: 10           # 连接池大小
  min_idle_conns: 5        # 最小空闲连接数
```

`GetClient()` 返回 `*github.com/redis/go-redis/v9.Client`，支持 go-redis 的全部 API（String、Hash、List、Set、Sorted Set、Stream、Pipeline 等）。

## 下一步

- [Runner（后台任务）](runner.md) - 后台任务运行器
- [服务](service.md) - 业务逻辑层
- [过滤器/中间件](filter.md) - HTTP 请求过滤

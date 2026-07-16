# Components

Components are independent functional modules in the framework. They implement the `core.IService` interface and are registered via `Service()` or `Filter()`.

> **Since v1.0.14**, all packages under `component/` are independent Go modules with their own `go.mod` and version tags. You need to `go get` the specific component package before using it.

## Creating a Component

Implement the `core.IService` interface:

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

### Register a Component

```go
builder := wf.NewBuilder(cfg)
builder.Service(&CacheComponent{})
app := builder.Build()
app.Start()
```

### Use in a Controller

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

## Built-in Components

### Rate Limiting (ratelimit)

Token bucket-based rate limiting component:

```go
package main

import (
    "errors"
    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/component/ratelimit"
    "github.com/chuccp/go-web-frame/config"
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
    // Check rate limit
    if !c.rateLimit.AllowSBurst(req.ClientIP(), 5) {
        return nil, errors.New("rate limited")
    }
    return "ok", nil
}

func main() {
    cfg, _ := config.LoadSingleFileConfig("application.yml")
    builder := wf.NewBuilder(cfg)
    builder.Service(&ratelimit.RateLimit{})
    builder.Rest(&MyController{})
    app := builder.Build()
    app.Start()
}
```

Configuration (`application.yml`):

```yaml
rate_limit:
  limit: 600     # token fill interval (seconds)
  burst: 5       # max bucket size
  maxSize: 1000000
  expiry: 3600   # cache expiry (seconds)
```

### Authentication (auth)

Token-based authentication filter with login/logout:

```go
import auth2 "github.com/chuccp/go-web-frame/component/auth"
```

Main features:

- `auth2.NewAuthenticationFilter(authenticator)` - Create auth filter
- `auth2.WithLogin()` - Route metadata, marks routes requiring login
- `auth.User(req, ctx)` - Get current logged-in user from request

Full example:

```go
type UserController struct {
    core.IService
    authFilter *auth2.AuthenticationFilter
}

func (c *UserController) Init(ctx *core.Context) error {
    c.authFilter = wf.GetFilter[*auth2.AuthenticationFilter](ctx)

    // Login endpoints (no auth required)
    ctx.Post("/signIn", c.signIn)
    ctx.Post("/logout", c.logout)

    // Business endpoints (auth required)
    ctx.Get("/users", c.List).WithMeta(auth2.WithLogin())
    return nil
}

func (c *UserController) signIn(req *web.Request) (any, error) {
    // Verify credentials, then sign in
    loginUser := &LoginUser{Id: user.Id, IsAdmin: user.IsAdmin}
    return c.authFilter.SignIn(loginUser, req)
}

func (c *UserController) logout(req *web.Request) (any, error) {
    return c.authFilter.SignOut(req)
}
```

Register the auth filter to a RestGroup:

```go
restGroupBuilder := core.NewRestGroupBuilder()
restGroupBuilder.Rest(&UserController{})
restGroupBuilder.Port(8081)
restGroupBuilder.Filter(auth2.NewAuthenticationFilter(&MyAuthenticator{}))
builder.RestGroup(restGroupBuilder.Build())
```

### Scheduled Tasks (schedule)

Cron expression-based task scheduling:

```go
import "github.com/chuccp/go-web-frame/component/schedule"
```

Main features:

- `schedule.NewScheduleWithSeconds()` - Create second-precision scheduler
- `schedule.AddFunc(cronExpr, func)` - Add a scheduled task
- `schedule.AddIdOrReplaceKeyFunc(id, key, cronExpr, func)` - Add by ID or replace by key
- `schedule.StopIdFunc(id)` - Stop a specific task
- `schedule.GetIds()` - Get all registered task IDs

See the [Runner documentation](runner.md) for usage examples.

### Other Components

All components below are independent Go modules (since v1.0.14) and must be installed separately:

| Component | Package | Role | Install |
|-----------|---------|------|---------|
| Auth | `component/auth` | Filter | `go get .../component/auth@v1.0.14` |
| In-memory Cache | `component/cache` | Service | `go get .../component/cache@v1.0.14` |
| Captcha | `component/captcha` | Service | `go get .../component/captcha@v1.0.14` |
| CORS | `component/cors` | Filter | `go get .../component/cors@v1.0.14` |
| Local Cache | `component/localcache` | Service | `go get .../component/localcache@v1.0.14` |
| QR Code | `component/qrcode` | Utility | `go get .../component/qrcode@v1.0.14` |
| Rate Limit | `component/ratelimit` | Service | `go get .../component/ratelimit@v1.0.14` |
| Schedule | `component/schedule` | Runner | `go get .../component/schedule@v1.0.14` |
| Validator | `component/validator` | Service | `go get .../component/validator@v1.0.14` |

### In-Memory Cache (cache)

High-performance cache based on [Otter](https://github.com/maypok86/otter):

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
    // Set cache
    c.cache.Set("key", "value")

    // Get cache
    val, exists := c.cache.Get("key")

    // Set if not exists (with TTL)
    c.cache.SetNX("temp", "data", 5*time.Minute)

    // Compute if absent
    c.cache.ComputeIfAbsent("expensive_key", func() {
        // Expensive computation, result auto-cached
    })

    // Invalidate cache
    c.cache.Invalidate("key")

    // Get cache stats
    stats := c.cache.Stats()

    return val, nil
}
```

Register:

```go
builder.Service(&cache.Cache{})
```

### Local Cache (localcache)

File-based local cache for caching generated files (images, PDFs, etc.):

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
    file, err := c.localCache.GetFile(func(value ...any) ([]byte, error) {
        // Generate file content
        return generateReport()
    }, "report", "2024")
    if err != nil {
        return nil, err
    }
    return file, nil
}
```

Configuration:

```yaml
local_cache:
  open: true        # enable
  path: ./cache     # cache file storage directory
```

Register:

```go
builder.Service(&localcache.LocalCache{})
```

### Slider Captcha (captcha)

Slider captcha generation and verification:

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

func (c *CaptchaController) GetCaptcha(req *web.Request) (any, error) {
    data, err := c.captcha.GetCaptchaData()
    if err != nil {
        return nil, err
    }
    return data, nil
}

func (c *CaptchaController) Verify(req *web.Request) (any, error) {
    thumbCode := req.Query("thumbCode")
    x := req.Query("x")

    data, valid := c.captcha.ValidateThumb(thumbCode, x)
    if !valid {
        return nil, errors.New("captcha verification failed")
    }
    return web.Data(data), nil
}
```

Configuration:

```yaml
captcha:
  codeKey: "your-encryption-key"   # AES key (16 bytes)
  codeIv: "your-encryption-iv"     # AES IV (16 bytes)
```

Register:

```go
builder.Service(&captcha.Captcha{})
```

### Input Validation (validator)

Based on [go-playground/validator](https://github.com/go-playground/validator) with custom rules:

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
    Phone    string `json:"phone" validate:"mobile"`      // Chinese mobile validation
    Password string `json:"password" validate:"password"`  // Strong password validation
}

func (c *MyController) Create(req *web.Request) (any, error) {
    var input CreateUserInput
    if err := req.BindJSON(&input); err != nil {
        return nil, err
    }
    if err := c.validator.Validate(input); err != nil {
        return nil, err
    }
    return web.Ok(), nil
}
```

Built-in validation rules:

| Rule | Description |
|------|-------------|
| `mobile` | Validate Chinese mobile numbers (supports +86 prefix, spaces, hyphens) |
| `password` | Strong password (min 8 chars, requires upper, lower, and digit) |

Register:

```go
builder.Service(&validator.Validator{})
```

### CORS Filter (cors)

Cross-origin resource sharing filter based on [gin-contrib/cors](https://github.com/gin-contrib/cors):

```go
import "github.com/chuccp/go-web-frame/component/cors"

// Register CORS filter
builder.Filter(cors.NewCorsFilter())
```

Default policy:

- Allow all origins (`AllowOriginFunc: returns true`)
- Allow credentials (`AllowCredentials: true`)
- Allowed headers: `Origin`, `Content-Length`, `Content-Type`, `Authorization`
- Auto-handle OPTIONS preflight requests

### Redis Component (redis)

Redis client component based on [go-redis/v9](https://github.com/redis/go-redis):

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
    client := c.redisClient.GetClient()
    val, err := client.Get(req.Request().Context(), "key").Result()
    if err != nil {
        return nil, err
    }
    return val, nil
}

func (c *MyController) SetCache(req *web.Request) (any, error) {
    client := c.redisClient.GetClient()
    err := client.Set(req.Request().Context(), "key", "value", 10*time.Minute).Err()
    if err != nil {
        return nil, err
    }
    return web.Ok(), nil
}
```

Register:

```go
builder.Service(&redis.Component{})
```

Configuration:

```yaml
redis:
  addr: localhost:6379
  password: ""
  db: 0
  pool_size: 10
  min_idle_conns: 5
```

`GetClient()` returns `*github.com/redis/go-redis/v9.Client`, supporting the full go-redis API.

## Next Steps

- [Runner](runner.md) - Background task runners
- [Service](service.md) - Business logic layer
- [Filter/Middleware](filter.md) - HTTP request filtering

# 组件

组件是框架中独立的功能模块，通过 `core.IComponent` 接口初始化，可以注册到应用中供其他组件使用。

## 创建组件

嵌入 `core.IComponent` 接口：

```go
package main

import (
    "context"
    "github.com/chuccp/go-web-frame/config"
    "github.com/chuccp/go-web-frame/core"
)

type CacheComponent struct {
    core.IComponent
    cache map[string]string
}

func (c *CacheComponent) Init(ctx context.Context, cfg config.IConfig) error {
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
builder.Component(&CacheComponent{})
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
    c.cache = wf.GetComponent[*CacheComponent](ctx)
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
    c.rateLimit = wf.GetComponent[*ratelimit.RateLimit](ctx)
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
    builder.Component(&ratelimit.RateLimit{})
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
| 本地缓存 | `component/localcache` | 基于内存的本地缓存，支持过期时间 |
| 验证码 | `component/captcha` | 图形验证码生成和验证 |
| 输入验证 | `component/validator` | 基于 go-playground/validator 的输入验证 |
| CORS | `component/cors` | 跨域资源共享中间件 |

## 下一步

- [Runner（后台任务）](runner.md) - 后台任务运行器
- [服务](service.md) - 业务逻辑层
- [过滤器/中间件](filter.md) - HTTP 请求过滤

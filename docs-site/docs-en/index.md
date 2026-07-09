# Go Web Frame User Guide

> An integrated Go backend toolkit: zero-boilerplate CRUD, declarative route metadata, explicit dependency injection, and seamless reuse of the Gin ecosystem.

## What is Go Web Frame?

Go Web Frame is an integrated backend toolkit. Routing, ORM, caching, logging, configuration, filters, and background runners are pre-integrated — no need to pick and wire them separately.

It focuses on **simplicity, transparency, and extensibility**:

- **Generic Model layer**: `Model[T]` / `EntryModel[T, PK]` eliminate CRUD boilerplate with full type safety from the database to the handler.
- **Declarative route metadata**: Tag routes with `.WithMeta()` and handle auth, permissions, rate limiting, etc. in one global filter.
- **Explicit dependency injection**: Register components through a fluent Builder, making initialization order fully transparent and controllable.
- **Gin ecosystem compatible**: Wrap HTTP requests/responses while exposing `req.GinContext()` so `gin-contrib` middlewares like CORS and Gzip work out of the box.

## 30-Second Hello World

```bash
go get github.com/chuccp/go-web-frame
```

```go
package main

import (
    "context"
    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/config"
    "github.com/chuccp/go-web-frame/web"
)

func main() {
    cfg, _ := config.LoadSingleFileConfig("application.yml")
    builder := wf.NewBuilder(cfg)
    builder.Get("/", func(c *web.Request) (any, error) {
        return "Hello, World!", nil
    })
    builder.Build().Run(context.Background())
}
```

```bash
go run main.go
# → http://localhost:19009
```

## Framework Highlights

### 1. Zero-Boilerplate CRUD with Generic Models (`Model[T]`)

**Pain point**: Traditional frameworks force you to run CLI commands over and over to generate `model.go` or `dao.go` files, cluttering the project.

**Design**: Go Web Frame uses Go generics to bind structs to the database at runtime. You define a plain struct, and create, read, update, delete, and advanced pagination (`Page` / `PageForWeb`) are immediately available — no code generation.

```go
type UserModel struct {
    *model.EntryModel[*User, uint]
}

// Queries
user, err := userModel.Query().Where("email = ?", email).One()
users, total, err := userModel.Query().Where("status = ?", 1).Page(page)

// Writes
err := userModel.Save(&User{Name: "alice"})
err := userModel.UpdateByPK(&user)
err := userModel.DeleteByPK(1)

// Request context propagates to the database automatically
m := userModel.WithContext(req.Ctx())
users, err := m.FindAll()
```

| Type | Capabilities | Use when |
|---|---|---|
| `Model[T]` | `Save`, `Query()`, `Update()`, `Delete()`, `CreateTable()`, `WithContext()` | Full control over query building |
| `EntryModel[T, PK]` | Everything in `Model[T]` + `FindByPK`, `FindAll`, `DeleteByPK`, `UpdateByPK`, `Page` | Entity has a primary key (most common) |

### 2. Declarative Route Metadata (`WithMeta`)

**Pain point**: In Gin, configuring auth, rate limiting, or public routes usually means attaching many middlewares in many places — scattered and hard to manage.

**Design**: Routes are tagged declaratively with `.WithMeta()`. A single top-level filter checks the metadata, keeping business handlers clean.

```go
func RequireAuth() web.MetaOption      { return web.WithValue("require_auth", true) }
func SkipAuth() web.MetaOption          { return web.WithValue("skip_auth", true) }
func RequirePermission(p string) web.MetaOption { return web.WithValue("require_permission", p) }

func (c *ApiController) Init(ctx *core.Context) error {
    ctx.Get("/api/login", c.Login).WithMeta(SkipAuth())
    ctx.Get("/api/profile", c.Profile).WithMeta(RequireAuth())
    ctx.Post("/api/admin/users", c.CreateUser).
        WithMeta(RequireAuth(), RequirePermission("admin:create_user"))
    return nil
}
```

```go
// One filter handles all auth logic
func (f *AuthFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    if !req.HasMeta(RequireAuth()) || req.HasMeta(SkipAuth()) {
        return fc.Next()
    }
    token := req.GetHeader("Authorization")
    if token == "" {
        return nil, errors.New("unauthorized")
    }
    // verify token, check permission...
    return fc.Next()
}
```

### 3. Explicit Dependency Injection with the Builder Pattern

**Pain point**: Implicit global scanning, like in Java Spring, can lead to hidden initialization order and hard-to-debug "magic" in Go.

**Design**: Components are registered explicitly through a fluent Builder. Dependencies are retrieved transparently via `GetService[T]` and `GetModel[T]`, and initialization order is fully visible and controllable.

```go
builder := wf.NewBuilder(cfg)

// Infrastructure runs first
builder.Filter(&cors.Filter{})
builder.Filter(&AuthFilter{})

// Data layer
builder.Model(&UserModel{})
builder.Model(&OrderModel{})

// Business layer
builder.Service(&UserService{})

// HTTP layer
builder.Rest(&UserController{})

// Background work
builder.Runner(&CleanupTask{})

app := builder.Build()
app.Run(ctx)
```

```go
// In any Init() or handler, get dependencies by type
userModel   := wf.GetModel[*UserModel](ctx)
userService := wf.GetService[*UserService](ctx)
```

### 4. Seamless Gin Ecosystem Reuse (`GinContext` Compatible)

**Pain point**: Many custom frameworks cannot use Gin community plugins, so common features have to be reimplemented.

**Design**: Go Web Frame wraps HTTP requests and responses, but exposes the underlying `*gin.Context` through `req.GinContext()`. Hundreds of `gin-contrib` middlewares — CORS, Gzip, Secure, and more — work out of the box.

```go
import "github.com/gin-contrib/gzip"

type GzipFilter struct{ core.IFilter }

func (f *GzipFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    gzip.Gzip(gzip.DefaultCompression)(req.GinContext())
    return fc.Next()
}
```

Common concerns are also built in:

```go
builder.Filter(&cors.Filter{})              // cross-origin
builder.Service(&ratelimit.RateLimit{})     // rate limiting
builder.Service(&cache.Cache{})             // in-memory cache
builder.Service(&captcha.Captcha{})         // slide-puzzle captcha
```

## Tech Stack

| Layer | Library | Role |
|---|---|---|
| HTTP | Gin | Router, middleware chain, parameter binding |
| ORM | GORM | Underlying SQL driver, migrations, joins/preload |
| Config | Viper | Multi-format, multi-path loading |
| Logging | Zap | Structured, leveled, rotated logging |
| JSON | Sonic | High-performance marshal/unmarshal |
| Cache | Otter | Local in-memory cache |
| Redis | go-redis | Pub/sub, caching |
| SQLite | modernc/sqlite | Pure Go, zero CGO |
| Validation | go-playground/validator | Struct tag validation |
| WebSocket | coder/websocket | Upgrade, read/write |
| Cron | robfig/cron | Expression-based scheduling |

## Quick Links

### Getting Started

- [Installation](getting-started/installation.md) - Requirements and installation
- [Quick Start](getting-started/quick-start.md) - Create your first application
- [Hello World](getting-started/hello-world.md) - The simplest example

### User Guide

- [Routing](guide/routing.md) - HTTP routing system
- [Controller](guide/controller.md) - REST controllers
- [Service](guide/service.md) - Business logic and dependency injection
- [Model](guide/model.md) - Type-safe ORM
- [Filter/Middleware](guide/filter.md) - Auth, logging, CORS, rate limiting, route metadata
- [Configuration](guide/configuration.md) - Configuration management
- [Logging](guide/logging.md) - Structured logging
- [Runner](guide/runner.md) - Runners and scheduled tasks
- [Components](guide/components.md) - Rate limiting, cache, captcha, etc.

### Advanced and Reference

- [Database](advanced/database.md) - Transactions, model groups, migrations, raw SQL
- [Deployment](advanced/deployment.md) - HTTPS, SSL, graceful shutdown
- [Core API](api/core.md) / [Web API](api/web.md) / [Model API](api/model.md)
- [Best Practices](best-practices.md)
- [Changelog](changelog.md)

## Community

- [GitHub](https://github.com/chuccp/go-web-frame)
- [Issue Tracker](https://github.com/chuccp/go-web-frame/issues)

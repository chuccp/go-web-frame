# Go Web Frame

**Build a CRUD backend in Go with zero ORM boilerplate. Define a struct, embed a generic Model — typed queries, pagination, and context propagation come with it.**

---

**🌐 Language**
[English](README.md) • [中文](README_ZH.md) • [繁體中文](README_ZH_TW.md) • [日本語](README_JA.md)

---

## Overview

Go Web Frame is an integrated backend toolkit. Routing, ORM, and caching are pre-integrated — no need to pick and wire them separately.

The core is **a generic Model layer that eliminates CRUD boilerplate.** Define an entity struct, embed `Model[T]`, and the compiler checks types from database to handler. No `interface{}`, no code generation.

```go
// Define once, use everywhere
type User struct {
    Id   uint   `gorm:"primaryKey;autoIncrement"`
    Name string
}

type UserModel struct {
    *model.EntryModel[*User, uint]   // ← all CRUD comes from this
}

func (m *UserModel) Init(db *db.DB, c *core.Context) error {
    m.EntryModel = model.NewEntryModel[*User, uint](db, "t_user")
    return m.CreateTable()
}

// That's your model. Now use it anywhere:
user, _ := userModel.FindByPK(1)          // → *User
users, _ := userModel.FindAll()            // → []*User
users, _ := userModel.Query().Where("age > ?", 18).Order("id desc").All()
total, _ := userModel.Query().Where("status = ?", 1).Count()
users, total, _ := userModel.Page(&web.Page{PageNo: 1, PageSize: 10})
userModel.Update().Where("id = ?", 1).UpdateColumn("name", "new name")
userModel.DeleteByPK(1)
```

Also included: routing (Gin), auth filters, WebSocket, SSE, CORS, rate limiting, cron, validation, caching, HTTPS (Let's Encrypt auto-cert or local certificates), multi-DB (MySQL / PostgreSQL / SQLite / Redis) — all wired up and configured from one YAML file.

---

## Quick Start

### 30 seconds: Hello World

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
    builder := wf.NewBuilder(config.LoadAutoConfig())
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

### 5 minutes: REST + Database

Create `application.yml`:

```yaml
web:
  db:
    type: sqlite
    file_path: ./data.db
```

Create `main.go`:

```go
package main

import (
    "context"
    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/config"
    "github.com/chuccp/go-web-frame/core"
    "github.com/chuccp/go-web-frame/db"
    "github.com/chuccp/go-web-frame/model"
    "github.com/chuccp/go-web-frame/web"
)

// ── Entity ──
type User struct {
    Id   uint   `gorm:"primaryKey;autoIncrement"`
    Name string `gorm:"size:255"`
}

// ── Model (zero CRUD boilerplate) ──
type UserModel struct {
    *model.EntryModel[*User, uint]
}

func (m *UserModel) Init(database *db.DB, ctx *core.Context) error {
    m.EntryModel = model.NewEntryModel[*User, uint](database, "t_user")
    return m.CreateTable()
}

// ── Controller ──
type UserController struct {
    core.IService
    userModel *UserModel
}

func (c *UserController) Init(ctx *core.Context) error {
    c.userModel = wf.GetModel[*UserModel](ctx)

    ctx.Get("/users", c.List)
    ctx.Get("/users/:id", c.Get)
    ctx.Post("/users", c.Create)
    ctx.Put("/users/:id", c.Update)
    ctx.Delete("/users/:id", c.Delete)
    return nil
}

func (c *UserController) List(req *web.Request) (any, error) {
    return c.userModel.FindAll()
}

func (c *UserController) Get(req *web.Request) (any, error) {
    return c.userModel.FindByPK(req.ParamUint("id"))
}

func (c *UserController) Create(req *web.Request) (any, error) {
    var user User
    if err := req.BindJSON(&user); err != nil {
        return nil, err
    }
    return &user, c.userModel.Save(&user)
}

func (c *UserController) Update(req *web.Request) (any, error) {
    var user User
    if err := req.BindJSON(&user); err != nil {
        return nil, err
    }
    user.Id = req.ParamUint("id")
    return nil, c.userModel.UpdateByPK(&user)
}

func (c *UserController) Delete(req *web.Request) (any, error) {
    return nil, c.userModel.DeleteByPK(req.ParamUint("id"))
}

// ── Main ──
func main() {
    builder := wf.NewBuilder(config.LoadAutoConfig())
    builder.Model(&UserModel{})
    builder.Rest(&UserController{})
    builder.Build().Run(context.Background())
}
```

```bash
curl http://localhost:19009/users              # → [{"Id":1,"Name":"alice"}]
curl http://localhost:19009/users/1            # → {"Id":1,"Name":"alice"}
curl -X POST ... -d '{"Name":"bob"}'          # → {"Id":2,"Name":"bob"}
```

The table is auto-created. All CRUD works. No SQL written, no ORM wiring code needed.

---

## Working with Data

### Model Tier

| Type | What you get | Use when |
|---|---|---|
| `Model[T]` | `Save`, `Query()`, `Update()`, `Delete()`, `CreateTable()`, `WithContext()` | Full control with fluent builders |
| `EntryModel[T, PK]` | Everything in `Model` + `FindByPK`, `FindAll`, `DeleteByPK`, `UpdateByPK`, `UpdateColumn`, `Page` | Entity has a primary key (most common) |

`PK` can be `uint`, `int`, `string`, or any type satisfying `~uint | ~int | ~string`.

### Common Queries

```go
m := userModel.WithContext(req.Ctx())   // context auto-propagates to all DB calls

// Fetch
user, err := m.FindByPK(1)                          // by primary key
users, err := m.FindAll()                            // all records
user, err := m.Query().Where("email = ?", email).One()
users, err := m.Query().Where("status = ?", 1).Order("id desc").List(100)

// Pagination
page := &web.Page{PageNo: 1, PageSize: 10}
users, total, err := m.Query().Where("age > ?", 18).Page(page)
pageAble, err := m.Query().Where("age > ?", 18).PageForWeb(page)  // returns PageAble[*User]

// Counting
count, err := m.Query().Where("status = ?", 1).Count()

// Associations (GORM Preload)
users, err := m.Query().Preload("Orders").Preload("Profile").All()
user, err := m.Query().Where("id = ?", 1).Preload("Orders").One()

// Joins
users, err := m.Query().Joins("JOIN orders ON orders.user_id = t_user.id").All()
```

### Common Writes

```go
// Insert
err := m.Save(&User{Name: "alice"})

// Update by primary key
user.Name = "new name"
err := m.UpdateByPK(user)

// Update single column
err := m.UpdateColumn(1, "status", 0)

// Conditional update
err := m.Update().Where("status = ?", 0).UpdateForMap(map[string]any{"status": 1})

// Delete
err := m.DeleteByPK(1)
err := m.Delete().Where("status = ?", -1).Delete()
```

### Compared to Raw GORM

```go
// Go Web Frame
m := userModel.WithContext(req.Ctx())
user, _ := m.FindByPK(1)
users, total, _ := m.Query().Where("age > ?", 18).Page(&web.Page{PageNo: 1, PageSize: 10})

// Raw GORM — manual table, manual ctx, no pagination
var user User
db.WithContext(ctx).Table("t_user").Where("id = ?", 1).First(&user)
var users []User
db.WithContext(ctx).Table("t_user").Where("age > ?", 18).Offset(0).Limit(10).Find(&users)
var total int64
db.WithContext(ctx).Table("t_user").Where("age > ?", 18).Count(&total)
```

Every call in raw GORM repeats the table name, the context, and the type. Go Web Frame binds those once at construction. The difference adds up fast when you have 20+ models.

---

## Organizing a Real Project

### The Builder: One Place to Register Everything

```go
builder := wf.NewBuilder(config.LoadAutoConfig())

// Infrastructure (runs first)
builder.Filter(&cors.Filter{})        // CORS headers
builder.Filter(&AuthFilter{})         // auth check

// Data layer
builder.Model(&UserModel{})
builder.Model(&OrderModel{})
builder.Model(&ProductModel{})

// Business logic
builder.Service(&UserService{})       // shared business logic
builder.Service(&PaymentService{})

// HTTP layer
builder.Rest(&UserController{})
builder.Rest(&OrderController{})

// Background work
builder.Runner(&CleanupTask{})

// Done
app := builder.Build()
app.Run(ctx)
```

Registration order matters: filters initialize first, then models, then services, then controllers, then runners. Dependencies resolve automatically through `Init()`.

### Getting Dependencies at Runtime

```go
// In any Init() or handler, get what you need by type:
userModel := wf.GetModel[*UserModel](ctx)
userService := wf.GetService[*UserService](ctx)
authFilter := wf.GetFilter[*AuthFilter](ctx)
cleanupTask := wf.GetRunner[*CleanupTask](ctx)
```

No string keys, no type assertions. The generic function handles it.

### Service Layer: Shared Business Logic

When multiple controllers need the same logic, extract a service:

```go
type UserService struct {
    core.IService
    userModel *UserModel
}

func (s *UserService) Init(ctx *core.Context) error {
    s.userModel = wf.GetModel[*UserModel](ctx)
    return nil
}

func (s *UserService) GetActiveUsers() ([]*User, error) {
    return s.userModel.Query().Where("status = ?", 1).All()
}

// Register:
builder.Service(&UserService{})

// Use in controller:
userService := wf.GetService[*UserService](ctx)
users, _ := userService.GetActiveUsers()
```

### Model Groups: Multiple Databases

```go
// Default database
builder.Model(&UserModel{}, &LogModel{})

// Separate database for analytics
analyticsDB := db.NewDBFromConfig(analyticsConfig)
analyticsGroup := app.NewModelGroup(analyticsDB, "analytics")
analyticsGroup.AddModel(&ReportModel{})
analyticsGroup.AutoCreateTable(true)
```

---

## Auth and Middleware

### Per-Route Metadata (WithMeta)

Tag routes declaratively — the filter checks once, not in every handler:

```go
func (c *ApiController) Init(ctx *core.Context) error {
    // Public
    ctx.Get("/api/login", login).WithMeta(SkipAuth())

    // Requires login
    ctx.Get("/api/profile", profile).WithMeta(RequireAuth())

    // Requires login + specific permission
    ctx.Post("/api/admin/users", createUser).
        WithMeta(RequireAuth(), RequirePermission("admin:create_user"))
    return nil
}
```

```go
// Define metadata factories
func RequireAuth() web.MetaOption      { return web.WithValue("require_auth", true) }
func SkipAuth() web.MetaOption          { return web.WithValue("skip_auth", true) }
func RequirePermission(p string) web.MetaOption { return web.WithValue("require_permission", p) }
```

```go
// One filter handles all auth logic
func (f *AuthFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    meta := req.HandlerMeta()

    if !meta.Has("require_auth") || meta.Has("skip_auth") {
        return fc.Next()
    }

    token := req.Request().Header.Get("Authorization")
    if token == "" {
        return nil, errors.New("unauthorized")
    }

    // Verify token, check permission from meta...
    return fc.Next()
}
```

### Global Filters

Filters registered via `builder.Filter()` run on every request:

```go
// Logging
type LoggingFilter struct { core.IFilter }

func (f *LoggingFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    start := time.Now()
    result, err := fc.Next()
    log.Info("request", zap.String("path", req.FullPath()), zap.Duration("elapsed", time.Since(start)))
    return result, err
}

// CORS (built-in, just register it)
builder.Filter(&cors.Filter{})
```

### RestGroup: Filters Scoped to a Route Group

```go
apiGroup := core.NewRestGroupBuilder().
    ServerConfig(web.DefaultServerConfig()).
    ContextPath("/api/v1").
    Build()

apiGroup.AddFilter(&AuthFilter{})     // only affects routes in this group
apiGroup.AddRest(&UserController{})   // all routes in this controller need auth

builder.RestGroup(apiGroup)
```

---

## Common Recipes

### File Upload

```go
ctx.Post("/upload", func(req *web.Request) (any, error) {
    file, header, err := req.File("file")   // form field name
    if err != nil {
        return nil, err
    }
    defer file.Close()

    // Save to disk
    dst := "./uploads/" + header.Filename
    if err := req.SaveUploadedFile(header, dst); err != nil {
        return nil, err
    }
    return map[string]string{"path": dst}, nil
})
```

### WebSocket

```go
ctx.WebSocket("/ws", func(stream *web.WebSocketStream) error {
    defer stream.Close()
    for {
        typ, msg, err := stream.Read(stream.Context())
        if err != nil {
            return err
        }
        stream.Write(stream.Context(), typ, msg)  // echo
    }
})
```

### Server-Sent Events

```go
ctx.SSE("/events", func(stream *web.SSEStream) error {
    defer stream.Close()
    stream.SendRetry(3000)  // reconnect after 3s

    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
    for range ticker.C {
        stream.Send("update", fmt.Sprintf("time: %s", time.Now()))
    }
    return nil
})
```

### Static Files & SPA

```go
ctx.Static("/assets", "./public")
ctx.Static("/", "./frontend/dist")   // SPA build output
```

```yaml
# Or via config — serves multiple directories, auto 404 fallback:
web:
  server:
    locations:
      - view/dist
      - www
    page404: 404.html
```

### Reverse Proxy

```go
ctx.ReverseProxy("/api/legacy", "http://127.0.0.1:8081")
```

### Background Tasks

```go
type CleanupTask struct { core.IRunner }

func (t *CleanupTask) Init(ctx *core.Context) error { return nil }

func (t *CleanupTask) Run() error {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    for range ticker.C {
        // clean up expired sessions...
    }
    return nil  // returned error kills the runner
}

builder.Runner(&CleanupTask{})
```

### Validation

```go
type CreateUserInput struct {
    Name  string `validate:"required,min=2,max=50"`
    Email string `validate:"required,email"`
}

func (c *Controller) Create(req *web.Request) (any, error) {
    var input CreateUserInput
    if err := req.BindJSON(&input); err != nil {
        return nil, err
    }

    validator := wf.GetService[*validator.Validator](req.Context())
    if err := validator.Validate(input); err != nil {
        return nil, web.NewValidationError(err.Error())
    }

    // input is valid, proceed...
}
```

### Custom HTTP Response

```go
// Return a plain struct → auto-wrapped: {"code":200, "data":{...}, "msg":"ok"}
return &User{Id: 1, Name: "alice"}, nil

// Return a string → plain text response
return "ok", nil

// Return a specific status code
return web.DataCode(http.StatusCreated, &user), nil

// Return an error
return nil, errors.New("something went wrong")

// Return a business error code
return nil, web.NewValidationError("name is required")

// Redirect
return web.Redirect("/new-url"), nil

// File download
return &web.File{Path: "/path/to/report.pdf", FileName: "report.pdf"}, nil
```

---

## Configuration

One YAML file covers everything. Auto-loaded from `./config/`, `~/.<appname>/`, `/etc/<appname>/`.

```yaml
web:
  server:
    port: 8080
    context_path: /api            # all routes prefixed with /api
    ssl:                          # HTTPS: auto-cert or local certificates
      enabled: true
      hosts:                      # Let's Encrypt auto-cert for these domains
        - example.com
      # certs:                    # Or use local certificate files (optional)
      #   - host: example.com
      #     cert-file: /path/to/fullchain.pem
      #     key-file: /path/to/privkey.pem

  db:
    type: mysql                   # mysql | postgres | sqlite
    host: localhost
    port: 3306
    user: root
    password: your_password
    database: mydb
    max_open_conns: 100
    max_idle_conns: 10
    conn_max_lifetime: 3600

  log:
    level: info                   # debug | info | warn | error
    path: ./logs/app.log
    max_size: 500                 # MB per file before rotation
    max_backups: 3
    max_age: 30                   # days
    compress: true

  redis:
    addr: localhost:6379
    password: ""
    db: 0
```

PostgreSQL and SQLite variants:

```yaml
# PostgreSQL
db:
  type: postgres
  host: localhost
  port: 5432
  user: postgres
  password: ""
  database: mydb
  sslmode: disable

# SQLite
db:
  type: sqlite
  file_path: ./data.db
```

Format support: JSON, YAML, TOML. Call `config.LoadAutoConfig()` for zero-config auto-discovery, or pass a path explicitly.

---

## Tech Stack

Pre-integrated, production-proven components:

| Layer | Library | Role |
|---|---|---|
| HTTP | Gin | Router, middleware chain, parameter binding |
| ORM | GORM | Underlying SQL driver, migrations, join/preload |
| Config | Viper | Multi-format, multi-path auto-loading |
| Logging | Zap | Structured, leveled, with rotation |
| JSON | Sonic | High-performance marshal/unmarshal |
| Cache | Otter | Local in-memory cache |
| Concurrency | Conc | Structured concurrency pool for server lifecycle |
| Redis | go-redis | Pub/sub, caching |
| SQLite | modernc/sqlite | Pure Go, zero CGO |
| Validation | go-playground/validator | Struct tag validation |
| WebSocket | coder/websocket | Upgrade + read/write |
| Cron | robfig/cron | Expression-based scheduler |

---

## Project Layout

```
├── web_frame.go        # Builder, WebFrame factory
├── core/               # Interfaces (IService, IModel, IRest, IRunner, IFilter), DI context
├── web/                # Request, response, routing, filters, SSE, WebSocket, static files
├── model/              # Model[T], EntryModel[T, PK], Query[T], Update[T], Delete[T]
├── db/                 # MySQL, PostgreSQL, SQLite connection management
├── redis/              # Redis client wrapper
├── config/             # Viper auto-loading
├── log/                # Zap + lumberjack rotation
├── component/          # cors, cache, rate-limit, captcha, qrcode, cron, validator
├── util/               # Crypto, file, string helpers
└── example/            # Runnable examples
    ├── helloworld/     # Minimal app
    ├── rest/           # REST controller
    ├── model/          # Generic ORM usage
    ├── filter/         # Auth + logging filters
    ├── withmeta/       # Route metadata
    └── background/     # Background runners
```

---

## Optional Modules

Heavy dependencies are split into independent sub-modules. Install only what you need:

```bash
# Core framework (no captcha/qrcode/cron/otter)
go get github.com/chuccp/go-web-frame

# Optional — install as needed
go get github.com/chuccp/go-web-frame/component/captcha@v1.0.1
go get github.com/chuccp/go-web-frame/component/schedule@v1.0.1
go get github.com/chuccp/go-web-frame/component/qrcode@v1.0.1
go get github.com/chuccp/go-web-frame/component/cache@v1.0.1
go get github.com/chuccp/go-web-frame/component/ratelimit@v1.0.1
```

### Publishing Sub-Modules

Each sub-module is versioned independently using tag prefixes:

```bash
# 1. Publish the main module
git tag v1.0.1
git push origin v1.0.1

# 2. Update each sub-module's dependency on the main module
for mod in captcha schedule qrcode cache ratelimit; do
  cd component/$mod
  go get github.com/chuccp/go-web-frame@v1.0.1
  go mod edit -dropreplace github.com/chuccp/go-web-frame
  go mod tidy
  cd ../..
done

# 3. Tag each sub-module (format: component/<name>/vX.Y.Z)
git tag component/captcha/v1.0.1
git tag component/schedule/v1.0.1
git tag component/qrcode/v1.0.1
git tag component/cache/v1.0.1
git tag component/ratelimit/v1.0.1

# 4. Push all tags
git push origin --tags
```

Go's module proxy resolves `component/captcha/v1.0.1` to the `component/captcha/` directory automatically.

---

## Getting Help

- **[User Guide (English)](./docs-site/docs-en/index.md)** — Full documentation
- **[User Guide (中文)](./docs-site/docs-zh/index.md)** — Chinese documentation
- **[Architecture Notes](./ARCHITECTURE.md)** — Design decisions and internals
- **[CLAUDE.md](./CLAUDE.md)** — Guide for AI coding assistants
- **[Go Reference](https://pkg.go.dev/github.com/chuccp/go-web-frame)** — Package API docs

---

## Contributing

PRs welcome. Run `go test ./...` before submitting.

## License

MIT — see [LICENSE](./LICENSE)

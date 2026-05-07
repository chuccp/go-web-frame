# Go Web Frame

A modern, feature-rich web framework for Go, providing a structured approach to building enterprise-grade web applications.

---

**🌐 Language / 语言**
[English](README.md) • [中文](README_ZH.md) • [繁體中文](README_ZH_TW.md) • [日本語](README_JA.md)

---

**📖 [Documentation →](./docs-site/docs-en/index.md)**

---

## Project Overview

Go Web Frame combines the best open-source components from the Go ecosystem: Gin (HTTP), GORM (ORM), Viper (config), Zap (logging), Sonic (JSON), Otter (cache), and more. It provides declarative route metadata, type-safe generic ORM, and dependency injection out of the box.

**Key Features:**
- **WithMeta**: Declare metadata per route (auth, permissions, rate limit flags) and handle uniformly in filters
- **Builder Pattern**: Explicit registration, controllable initialization order, no implicit scanning
- **Generic ORM**: Zero-boilerplate CRUD with `Model[T]`, no code generation
- **Transparent Context**: All dependencies injectable via `GetService[T]`, debug-friendly

## 🧩 Tech Stack

This framework carefully selects and integrates the following excellent open-source components, deeply combined and configured with best practices:

### Core Framework
| Component | Description |
|-----------|-------------|
| [Gin](https://github.com/gin-gonic/gin) | High-performance HTTP web framework with excellent API performance |
| [GORM](https://gorm.io/) | Powerful ORM library with multi-database support |
| [Viper](https://github.com/spf13/viper) | Complete configuration solution supporting multiple formats |
| [Zap](https://go.uber.org/zap) | Uber's high-performance structured logging library |

### Data Storage
| Component | Description |
|-----------|-------------|
| [go-redis](https://github.com/redis/go-redis) | Redis client recommended by Redis |
| [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) | Pure Go SQLite implementation, no CGO dependency |
| [gorm-driver/mysql](https://gorm.io/docs/connecting_to_the_database.html) | MySQL database driver |
| [gorm-driver/postgres](https://gorm.io/docs/connecting_to_the_database.html) | PostgreSQL database driver |

### Caching & Performance
| Component | Description |
|-----------|-------------|
| [Otter](https://github.com/maypok86/otter) | High-performance Go local cache library |
| [golang.org/x/time/rate](https://pkg.go.dev/golang.org/x/time/rate) | Token bucket rate limiter |

### Utilities
| Component | Description |
|-----------|-------------|
| [Cron](https://github.com/robfig/cron) | Scheduled task library |
| [go-qrcode](https://github.com/yeqown/go-qrcode) | QR code generation |
| [go-captcha](https://github.com/wenlng/go-captcha) | Behavioral captcha generation |
| [validator](https://github.com/go-playground/validator) | Struct field validation |
| [UUID](https://github.com/google/uuid) | UUID generation |
| [Lumberjack](https://pkg.go.dev/gopkg.in/natefinch/lumberjack.v2) | Log rotation |
| [Conc](https://github.com/sourcegraph/conc) | Better concurrency primitives |
| [Emperror](https://emperror.dev/errors) | Production-grade error handling |

### Why These Components?

- **Production Proven**: All components are widely verified in large-scale production environments
- **High Performance**: Gin, Zap, Otter are the best performers in their respective domains
- **Best Practices**: Carefully integrated, ready to use out of the box, no complex configuration needed
- **Mature Ecosystem**: Active community support with continuous updates

## 🤖 Why Choose Go Web Frame?

### Comparison with Other Frameworks

| Feature | Go Web Frame | Gin | Beego | Echo |
|---------|-------------|-----|-------|------|
| Ready to Use | ✅ Complete Solution | ❌ Need Integration | ✅ Complete Solution | ⚠️ Partial |
| Generic ORM | ✅ Zero Boilerplate | ❌ Choose Yourself | ❌ No Generics | ❌ Choose Yourself |
| Dependency Injection | ✅ Built-in DI | ❌ Implement Yourself | ⚠️ Basic Support | ❌ Implement Yourself |
| Learning Curve | 🟢 Moderate | 🟢 Easy | 🟡 Steep | 🟢 Easy |
| Feature Completeness | 🟢 High | 🟡 Medium | 🟢 High | 🟡 Medium |
| Performance | 🟢 Excellent (42k QPS) | 🟢 Best (45k QPS) | 🟡 Average | 🟢 Excellent |

### When Should You Choose Go Web Frame?

**Highly Recommended For:**

- 🚀 **Rapid Prototyping**: Need to complete project prototypes quickly with all necessary components pre-integrated
- 🏢 **Enterprise Applications**: Require clean architecture, dependency injection, unified error handling
- 📊 **Admin Dashboard Systems**: Built-in CRUD operations, pagination, validation
- 🔌 **RESTful API Services**: Simplified controller implementation, automatic route registration
- ⚙️ **Microservices**: Lightweight yet feature-complete
- 🛠️ **Full-stack Go Projects**: One-stop solution from frontend to backend to database

**Especially Suitable For:**

- Go Beginners: Learn best practices without trial and error
- Solo Developers: Complete projects efficiently, reduce technology selection time
- Small Teams: Unified tech stack, lower collaboration costs
- AI-Assisted Development: Clean architecture makes it easier for AI to understand and generate code

### Selection Guide

If you need:
- A feature-complete Go web framework with declarative route metadata
- Production-proven components pre-integrated
- Type-safe generic ORM without code generation
- Clean architecture with explicit initialization

Go Web Frame may be a good fit.

## Features

- **Route Metadata (WithMeta)**: Declare metadata per route (auth, permissions, rate limit) and handle uniformly in filters - no repetitive auth checks in each handler
- **Builder Pattern**: Explicit registration, controllable initialization order, no implicit scanning or reflection
- **Type-safe Generic ORM**: Zero-boilerplate CRUD operations with `Model[T]`, no code generation
- **Dependency Injection**: Built-in DI container via Context - `GetService[T]`, `GetModel[T]`
- **MVC-like Architecture**: Clean separation of concerns with services, controllers, and models
- **Database Integration**: SQLite, MySQL, PostgreSQL, Redis support with connection pool configuration
- **Component System**: Reusable components including caching, rate limiting, captcha, QR code, cron, validation
- **RESTful Support**: Simplified REST controller implementation
- **Auto-Configuration**: Auto-loading config from JSON, YAML, or TOML files
- **Advanced Logging**: Structured logging powered by Zap with rotation support
- **Background Tasks**: Built-in runner system for background processing
- **Request Filtering**: HTTP middleware/filter system for cross-cutting concerns
- **Unified Error Handling**: Automatic conversion of service errors to standardized HTTP responses
- **Gin Ecosystem Compatibility**: `GinContext()` to reuse gin-contrib middleware (CORS, Gzip, Secure, etc.)
- **Built-in CORS Component**: Pre-configured CORS filter

## Quick Start

### Installation

```bash
go get github.com/chuccp/go-web-frame
```

### Hello World Example

```go
package main

import (
    "context"
    "time"

    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/config"
    "github.com/chuccp/go-web-frame/log"
    "github.com/chuccp/go-web-frame/web"
)

func main() {
    // Create builder with auto config loading
    builder := wf.NewBuilder(config.LoadAutoConfig())

    // Register a simple route
    builder.Get("/", func(c *web.Request) (any, error) {
        return "Hello, World!", nil
    })

    // Build the application
    app := builder.Build()

    // Run with context for graceful shutdown
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Auto shutdown after 10 seconds (example)
    go func() {
        time.Sleep(time.Second * 10)
        cancel()
    }()

    if err := app.Run(ctx); err != nil {
        log.PrintPanic(err)
    }
}
```

### 🔌 REST Controller Example

```go
package main

import (
    "context"

    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/config"
    "github.com/chuccp/go-web-frame/core"
    "github.com/chuccp/go-web-frame/log"
    "github.com/chuccp/go-web-frame/web"
)

type UserController struct {
    // Embed core.IService interface
    core.IService
}

// Init controller and register routes
func (u *UserController) Init(ctx *core.Context) error {
    // Register routes through context
    ctx.Get("/users", u.GetUsers)
    ctx.Get("/users/:id", u.GetUser)
    ctx.Post("/users", u.CreateUser)

    return nil
}

// Handler: get all users
func (u *UserController) GetUsers(c *web.Request) (any, error) {
    // Access query parameters
    page := c.Query("page")
    limit := c.Query("limit")

    return map[string]any{
        "users": []string{"alice", "bob"},
        "page":  page,
        "limit": limit,
    }, nil
}

// Handler: get a single user by ID
func (u *UserController) GetUser(c *web.Request) (any, error) {
    id := c.Param("id")
    return map[string]any{
        "id":   id,
        "name": "alice",
    }, nil
}

// Handler: create a new user
func (u *UserController) CreateUser(c *web.Request) (any, error) {
    var user struct {
        Name string `json:"name"`
    }
    if err := c.BindJSON(&user); err != nil {
        return nil, err
    }

    return map[string]any{
        "id":   1,
        "name": user.Name,
    }, nil
}

func main() {
    builder := wf.NewBuilder(config.LoadAutoConfig())
    builder.Rest(&UserController{})
    app := builder.Build()

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    if err := app.Run(ctx); err != nil {
        log.PrintPanic(err)
    }
}
```

### Static Files and Reverse Proxy

```go
func (c *AssetsController) Init(ctx *core.Context) error {
    ctx.Static("/assets", "./public")
    ctx.ReverseProxy("/api", "http://127.0.0.1:8081")
    return nil
}
```

`Context.Static()` serves local files through the current server, while `Context.ReverseProxy()` forwards a route prefix to an upstream service.

### Context Path (Route Prefix)

Similar to Tomcat's context path, you can set a global prefix for all routes:

```yaml
# application.yml
web:
  server:
    port: 8080
    context_path: /api
```

With this configuration:
- Registered route `/users` → Accessible at `/api/users`
- Registered route `/orders` → Accessible at `/api/orders`
- WebSocket `/ws` → Accessible at `/api/ws`
- Static files `/assets` → Accessible at `/api/assets`

### WebSocket Support

```go
// Simple echo server
ctx.WebSocket("/ws", func(conn *websocket.Conn) error {
    for {
        messageType, message, err := conn.ReadMessage()
        if err != nil {
            return err
        }
        err = conn.WriteMessage(messageType, message)
        if err != nil {
            return err
        }
    }
})

// With custom upgrader
upgrader := &websocket.Upgrader{
    ReadBufferSize:  4096,
    WriteBufferSize: 4096,
    CheckOrigin: func(r *http.Request) bool {
        return r.Header.Get("Origin") == "https://example.com"
    },
}
ctx.WebSocket("/ws/chat", handler, upgrader)
```

### Server-Sent Events (SSE) Support

```go
ctx.SSE("/events", func(stream *web.SSEStream) error {
    // Set headers
    stream.SetHeaders()
    
    // Send retry interval (reconnect after 3 seconds)
    stream.SendRetry(3000)
    
    // Send events
    for i := 0; i < 10; i++ {
        // Send with event name
        stream.Send("update", fmt.Sprintf("Count: %d", i))
        
        // Or send plain message
        // stream.SendMessage("plain message")
        
        // Or send with ID
        // stream.SendWithID("123", "event", "data")
        
        time.Sleep(time.Second)
    }
    return nil
})
```

**SSE Stream Methods:**
| Method | Description |
|--------|-------------|
| `Send(event, data)` | Send message with event name |
| `SendMessage(data)` | Send plain message |
| `SendWithID(id, event, data)` | Send message with ID |
| `SendRetry(ms)` | Set reconnection interval |
| `Heartbeat()` | Send heartbeat comment |
| `StartHeartbeat(interval)` | Start heartbeat goroutine |

### 🏷️ Route Metadata with `.WithMeta()`

The `.WithMeta()` feature allows you to attach arbitrary metadata to individual routes, which can then be accessed by filters for flexible cross-cutting concerns like authentication, permission checks, feature flags, caching configuration, and more.

**Basic Usage:**
```go
// Create meta options
func RequireAuth() web.MetaOption {
    return web.WithValue("require_auth", true)
}

func RequirePermission(perm string) web.MetaOption {
    return web.WithValue("require_permission", perm)
}

func SkipAuth() web.MetaOption {
    return web.WithValue("skip_auth", true)
}

// In route registration
func (c *ApiController) Init(ctx *core.Context) error {
    // Public route - no auth needed
    ctx.Get("/api/login", loginHandler).WithMeta(SkipAuth())

    // Protected route - requires authentication
    ctx.Get("/api/profile", profileHandler).WithMeta(RequireAuth())

    // Protected route with multiple metadata options
    ctx.Post("/api/admin/users", createUserHandler).WithMeta(RequireAuth(), RequirePermission("admin:create_user"))

    return nil
}
```

**Access metadata in a filter:**
```go
func (f *AuthFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    meta := req.HandlerMeta()

    // Check if this route requires authentication
    requireAuth, ok := meta.Get("require_auth").(bool)
    if ok && requireAuth {
        // Skip auth if marked as public
        if meta.Has("skip_auth") {
            return fc.Next()
        }
        // Get token and verify...
        token := req.Request().Header.Get("Authorization")
        if token == "" {
            return nil, errors.New("missing authorization token")
        }
    }

    return fc.Next()
}
```

See the complete example at [example/withmeta/withmeta.go](./example/withmeta/withmeta.go)

### 🔗 Gin Ecosystem Integration (CORS, Gzip, etc.)

The framework exposes `Request.GinContext()` to seamlessly wrap and reuse existing gin middleware from the gin-contrib ecosystem. This enables compatibility with hundreds of battle-tested middleware without rewriting them.

**Built-in CORS Filter:**

```go
package main

import (
    "context"
    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/component/cors"
    "github.com/chuccp/go-web-frame/config"
    "github.com/chuccp/go-web-frame/log"
    "github.com/chuccp/go-web-frame/web"
)

func main() {
    builder := wf.NewBuilder(config.LoadAutoConfig())
    
    // Add CORS filter - handles preflight OPTIONS requests and sets CORS headers
    builder.Filter(&cors.Filter{})
    
    builder.Get("/", func(c *web.Request) (any, error) {
        return "CORS-enabled endpoint", nil
    })
    
    app := builder.Build()
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    if err := app.Run(ctx); err != nil {
        log.PrintPanic(err)
    }
}
```

**Wrapping Other Gin Middleware:**

You can wrap any `gin.HandlerFunc` middleware using the same pattern. For example, to wrap Gzip compression:

```go
import (
    "github.com/gin-contrib/gzip"
    "github.com/chuccp/go-web-frame/core"
    "github.com/chuccp/go-web-frame/web"
)

type GzipFilter struct {
    core.IFilter
    handler gin.HandlerFunc
}

func (f *GzipFilter) Init(ctx *core.Context) error {
    f.handler = gzip.Gzip(gzip.DefaultCompression)
    return nil
}

func (f *GzipFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    f.handler(req.GinContext())
    return fc.Next()
}

// Usage: builder.Filter(&GzipFilter{})
```

The same pattern works for:
- **Secure**: Security headers (github.com/gin-contrib/secure)
- **Session**: Session management (github.com/gin-contrib/sessions)
- **Logger**: Custom logging (github.com/gin-contrib/logger)
- **Recovery**: Panic recovery with custom logic
- And hundreds of other gin middleware

### ⚡ Generic ORM Example

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

// Define your entity struct
type User struct {
    Id   uint   `gorm:"primaryKey"`
    Name string
    Age  int
}

// UserModel extends generic Model
type UserModel struct {
    *model.Model[*User]
}

func (u *UserModel) Init(database *db.DB, ctx *core.Context) error {
    u.Model = model.NewModel[*User](database, "t_user")
    // Auto create table if not exists
    return u.CreateTable()
}

func main() {
    builder := wf.NewBuilder(config.LoadAutoConfig())
    builder.Model(&UserModel{})

    // Example ORM operations
    builder.Get("/users", func(c *web.Request) (any, error) {
        userModel := wf.GetModel[*UserModel](c.Context())

        // Query with chain API
        users, err := userModel.Query().
            Where("age > ?", 18).
            Order("id desc").
            All()

        return users, err
    })

    builder.Post("/users", func(c *web.Request) (any, error) {
        userModel := wf.GetModel[*UserModel](c.Context())

        // Create user
        user := &User{Name: "John", Age: 25}
        err := userModel.Save(user)
        return user.Id, err
    })

    builder.Put("/users/:id", func(c *web.Request) (any, error) {
        userModel := wf.GetModel[*UserModel](c.Context())
        id := c.ParamInt("id")

        // Update user
        return nil, userModel.Update().
            Where("id = ?", id).
            UpdateColumn("name", "John Updated")
    })

    builder.Delete("/users/:id", func(c *web.Request) (any, error) {
        userModel := wf.GetModel[*UserModel](c.Context())
        id := c.ParamInt("id")

        // Delete user
        return nil, userModel.Delete().
            Where("id = ?", id).
            Delete()
    })

    app := builder.Build()
    ctx := context.Background()
    app.Run(ctx)
}
```

## Architecture Overview

### Core Layers

1. **Core Abstraction Layer (`./core`)**: Defines fundamental interfaces and DI container
   - `IService`: Base interface for all services requiring initialization
   - `IModel`: Data access layer interface with CRUD and table management
   - `IRest`: REST controller interface (extends IService)
   - `IService`: Base interface for all services and components
   - `IRunner`: Background task runners (extends IService)
   - `IFilter`: HTTP request filters/middleware (extends IService)
   - `Context`: Dependency injection container that manages all components (embeds `context.Context` for lifecycle control)

2. **Web Layer (`./web`)**: HTTP handling with request/response abstraction
   - Request/response abstraction with helper methods
   - Routing with support for all HTTP methods
   - Filter/middleware system
   - Automatic conversion of service responses to standardized HTTP responses

3. **Data Access Layer (`./db`, `./model`)**: Database abstraction and ORM
   - `./db`: Multi-database abstraction (MySQL, SQLite, PostgreSQL) powered by GORM
   - `./model`: Type-safe generic base model with zero-boilerplate CRUD
   - `./sqlite`: SQLite-specific configuration
   - `./redis`: Redis integration for caching and messaging
   - Configurable connection pool settings for production performance tuning

4. **Infrastructure Components**:
   - `./config`: Configuration management with Viper (JSON/YAML/TOML)
   - `./log`: Structured logging with Zap
   - `./component`: Reusable components (cache, rate limiting, captcha, QR code, cron, validation)
   - `./util`: Comprehensive utilities (strings, time, crypto, networking, etc.)

### Application Lifecycle

1. **Create**: Initialize `Builder` with `NewBuilder(config)` or `NewBuilder(config.LoadAutoConfig())`
2. **Configure**: Add routes, controllers, models, services, components, runners, and filters via builder methods
3. **Build**: Create `WebFrame` instance with `builder.Build()`
4. **Run**: Start the server with `app.Run(ctx)`

## Configuration Example

### Complete Configuration Example

```yaml
web:
  # Server configuration
  server:
    port: 8080                    # Server port, default 19009
    context_path: /api            # Global route prefix (optional), like Tomcat's context path
    locations:                    # Static file directories (optional)
      - view/dist
      - www
    page404: 404.html             # 404 page (optional)
    # HTTPS/SSL configuration (optional)
    ssl:
      enabled: true               # Enable HTTPS
      hosts:                      # Domain list (auto Let's Encrypt certificate)
        - example.com
        - api.example.com

  # Database configuration
  db:
    type: mysql                   # Database type: mysql, postgres, sqlite
    host: localhost
    port: 3306
    user: root                    # Username (also supports username)
    password: your_password
    database: your_database       # Database name (also supports dbname)
    charset: utf8mb4
    # Connection pool settings (optional, with defaults)
    max_open_conns: 100           # Max open connections, default 100
    max_idle_conns: 10            # Max idle connections, default 10
    conn_max_lifetime: 3600       # Connection max lifetime (seconds), default 3600

  # Log configuration
  log:
    level: info                   # Log level: debug, info, warn, error
    path: ./logs/app.log          # Log file path
    write: false                  # Background write mode
    # Log rotation configuration (optional, with defaults)
    max_size: 100                 # Max size of a single log file (MB), default 500
    max_backups: 5                # Max number of old log files to retain, default 3
    max_age: 7                    # Max number of days to retain old log files, default 30
    compress: true                # Whether to compress old log files, default true
    local_time: false             # Whether to use local time, default false

  # Redis configuration (optional)
  redis:
    addr: localhost:6379          # Redis address
    password: ""                  # Password
    db: 0                         # Database number

# Local cache configuration (optional)
local_cache:
  path: ./cache                   # Cache file storage path
  open: true                      # Enable file cache

# Rate limit configuration (optional)
rate_limit:
  limit: 600                      # Token fill interval (seconds)
  burst: 5                        # Token bucket capacity
  max_size: 1000000               # Max cache entries
  expiry: 3600                    # Cache expiry time (seconds)
```

### Database Configuration

#### MySQL Configuration

```yaml
web:
  db:
    type: mysql
    host: localhost
    port: 3306
    user: root                    # Username (also supports username)
    password: your_password
    database: your_database       # Database name (also supports dbname)
    charset: utf8mb4              # Optional, default utf8
    max_open_conns: 100           # Optional, default 100
    max_idle_conns: 10            # Optional, default 10
    conn_max_lifetime: 3600       # Optional, default 3600 seconds
```

#### PostgreSQL Configuration

```yaml
web:
  db:
    type: postgres                # or postgresql
    host: localhost
    port: 5432
    user: postgres
    password: your_password
    database: your_database
    sslmode: disable              # Optional: disable, require, verify-ca, verify-full
    timezone: Asia/Shanghai       # Optional
    max_open_conns: 100
    max_idle_conns: 10
    conn_max_lifetime: 3600
```

#### SQLite Configuration

```yaml
web:
  db:
    type: sqlite
    file_path: ./data/app.db      # Database file path
    max_open_conns: 10            # Optional, default 10
    max_idle_conns: 5             # Optional, default 5
    conn_max_lifetime: 3600       # Optional, default 3600 seconds
```

### HTTPS Configuration

The framework supports automatic Let's Encrypt SSL certificate application and management, no manual certificate configuration required.

```yaml
web:
  server:
    port: 443                     # HTTPS default port
    ssl:
      enabled: true               # Enable HTTPS
      hosts:                      # Domain list
        - example.com
        - api.example.com
```

**HTTPS Configuration Notes:**

- `enabled: true` - Enable HTTPS mode
- `hosts` - List of domains for certificate application
- Certificates are automatically applied and cached in `./certs` directory
- HTTP/2 protocol support
- Port 80 automatically sets up HTTP to HTTPS redirect

**Important Notes:**

1. Domain must be correctly resolved to server IP
2. Server needs external network access (Let's Encrypt validation)
3. Recommended to use port 443, other ports also work

### Static Files Configuration

```yaml
web:
  server:
    port: 8080
    locations:                    # Static file directory list
      - view/dist                 # Frontend build output
      - www                       # Static resources directory
    page404: 404.html             # SPA 404 fallback page
```

**Static Files Notes:**

- `locations` - Static file lookup directories, searched in order
- `page404` - 404 page returned when HTML page is requested but file doesn't exist
- Supports SPA route fallback

### Redis Configuration

```yaml
web:
  redis:
    addr: localhost:6379          # Redis address
    password: ""                  # Password (optional)
    db: 0                         # Database number
    pool_size: 10                 # Connection pool size (optional)
```

### Local Cache Configuration

```yaml
local_cache:
  path: ./cache                   # Cache file storage path
  open: true                      # Enable file cache
```

### Rate Limit Configuration

```yaml
rate_limit:
  limit: 600                      # Token fill interval (seconds), 1 token per limit seconds
  burst: 5                        # Token bucket capacity
  max_size: 1000000               # Max cache entries
  expiry: 3600                    # Cache expiry time (seconds)
```

## Project Structure

```
├── web_frame.go         # Main entry point - factory methods for WebFrame
├── core/                # Core abstractions and DI container
│   ├── interface.go     # Core interfaces (IService, IModel, IRest, etc.)
│   ├── context.go       # Dependency injection context
│   ├── server.go        # Server implementation managing REST groups and runners
│   └── db.go            # DB wrapper
├── web/                 # Web layer with HTTP handling
│   ├── handles.go       # Route registration
│   ├── request.go       # Request abstraction with helper methods
│   ├── response.go      # Response conversion
│   └── filter.go        # Filter/middleware interface
├── db/                  # Database abstraction layer
│   ├── db.go            # DB creation and config parsing
│   ├── mysql.go         # MySQL configuration and connection
│   └── sqlite.go        # SQLite configuration and connection
├── model/               # Generic ORM implementation
│   └── model.go         # Base Model with CRUD operations
├── sqlite/              # SQLite driver
├── redis/               # Redis integration
├── config/              # Configuration management
├── log/                 # Logging with Zap
├── component/           # Reusable components
│   ├── cors/            # CORS cross-origin resource sharing filter
│   ├── cache.go         # Cache component
│   ├── localcache.go    # Local in-memory cache
│   ├── rate_limit.go    # Rate limiting
│   ├── captcha.go       # Captcha generation
│   ├── qrcode.go        # QR code generation
│   ├── cron.go          # Cron scheduled tasks
│   └── validate.go      # Input validation
├── util/                # Utility functions
└── example/             # Example applications
    ├── helloworld/      # Basic hello world example
    ├── rest/            # REST controller example
    ├── model/           # ORM model example
    ├── filter/          # Custom HTTP filters example
    ├── background/      # Background tasks/runners example
    └── withmeta/        # Route metadata .WithMeta() example
```

## Common Development Commands

### Build and Run Examples

```bash
# Run the hello world example
go run example/helloworld/helloworld.go

# Run the REST example
go run example/rest/rest.go

# Run the ORM model example
go run example/model/model.go

# Run the filters example
go run example/filter/filter.go

# Run the background tasks example
go run example/background/background.go

# Run the route metadata .WithMeta() example
go run example/withmeta/withmeta.go

# Build the framework (library only)
go build
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests in a specific package
go test ./core
go test ./web

# Run tests with verbose output
go test -v ./core

# Run a specific test case
go test -v ./core -run TestSpecificFunction
```

### Formatting and Linting

```bash
# Format all code with gofmt
gofmt -w ./...

# Alternative formatting with gofumpt (if installed)
gofumpt -w ./...

# Install linter
go install golang.org/x/lint/golint@latest

# Run linter
golint ./...
```

### Dependency Management

```bash
# Add a new dependency
go get github.com/example/package

# Update dependencies
go get -u ./...

# Tidy up go.mod and go.sum
go mod tidy
```


## Development Notes

- The framework follows Go conventions and uses standard Go tooling
- All components implement the `IService` interface with `Init(ctx)` method
- Dependency injection is done through the context - use `wf.GetService[T](ctx)` to retrieve services
- Connection pool has reasonable defaults that work for most applications
- SQLite is recommended for development and small applications, MySQL for production

## Documentation

- **[📖 User Guide (中文)](./docs-site/docs-zh/index.md)** - Complete usage manual: installation, routing, controllers, models, filters, configuration, logging, runners, components, API reference
- **[📖 User Guide (English)](./docs-site/docs-en/index.md)** - English version of the user guide
- **[Architecture Design](./ARCHITECTURE.md)** - Internal architecture and design decisions
- **[Best Practices](./BEST_PRACTICES.md)** - Recommended patterns and practices
- **[Changelog](./CHANGELOG.md)** - Version history and changes
- **[CLAUDE.md](./CLAUDE.md)** - Guide for AI-assisted development
- [Go Reference Documentation](https://pkg.go.dev/github.com/chuccp/go-web-frame)
- [Example Applications](./example/)

## Contributing

Contributions are welcome! Please feel free to submit Issues and Pull Requests. Before submitting a PR:

1. Run tests to ensure everything passes
2. Keep the code style consistent with the project
3. Add appropriate tests for new features
4. Update related documentation

## License

MIT License - see [LICENSE](./LICENSE) for details

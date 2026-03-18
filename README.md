# Go Web Frame

A modern, feature-rich web framework for Go, providing a structured approach to building enterprise-grade web applications.

---

**🌐 Language / 语言**
[English](README.md) • [中文](README_ZH.md)

---

## Project Overview

Go Web Frame is an opinionated web framework that enforces clean architecture through component-based design. It provides built-in dependency injection, type-safe ORM, and database integration out of the box.

**Core Advantage**: This framework integrates the best open-source packages and frameworks from the Go ecosystem, applying best practices in an organic combination. It represents the optimal solution for rapidly building production-ready applications, significantly boosting development efficiency.

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

If you're looking for:
- A **feature-complete** and **ready-to-use** Go web framework
- **Production-proven** excellent components integrated without repeated research
- **Generic-powered** type-safe ORM with minimal boilerplate
- **Clean architecture** for maintainability and team collaboration
- **Rapid development** to shorten time from idea to product

**Go Web Frame is your best choice!**

## Features

- **MVC-like Architecture**: Clean separation of concerns with services, controllers, and models
- **⚡ Type-safe Generic ORM**: Zero-boilerplate CRUD operations with generics, no reflection overhead
- **Dependency Injection**: Built-in DI container for managing component lifecycle
- **Database Integration**: SQLite, MySQL, PostgreSQL, Redis support with extensible database abstraction layer (powered by GORM)
- **Connection Pool Configuration**: Configurable connection pool settings (`max_open_conns`, `max_idle_conns`, `conn_max_lifetime`)
- **Component System**: Reusable components including caching, rate limiting, captcha, QR code generation, cron scheduled tasks, and input validation
- **RESTful Support**: Simplified REST controller implementation
- **Auto-Configuration**: Auto-loading config from JSON, YAML, or TOML files
- **Advanced Logging**: Structured logging powered by Zap with rotation support
- **Background Tasks**: Built-in runner system for background processing
- **Request Filtering**: HTTP middleware/filter system for cross-cutting concerns
- **Route Metadata**: `.WithMeta()` support for attaching custom metadata to routes (enables flexible per-route authentication, permissions, etc.)
- **Unified Error Handling**: Automatic conversion of service errors to standardized HTTP responses

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
    "github.com/chuccp/go-web-frame/log"
    "github.com/chuccp/go-web-frame/web"
)

func main() {
    // Create web framework instance with auto config loading
    app := wf.NewWithAutoConfig()

    // Register a simple route
    app.Get("/", func(c *web.Request) (any, error) {
        return "Hello, World!", nil
    })

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
    app := wf.NewWithAutoConfig()
    app.AddRest(&UserController{})

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    if err := app.Run(ctx); err != nil {
        log.PrintPanic(err)
    }
}
```

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

// In route registration
func (c *ApiController) Init(ctx *core.Context) error {
    // Public route - no auth needed
    ctx.Get("/api/login", loginHandler).WithMeta(SkipAuth())

    // Protected route - requires authentication
    ctx.Get("/api/profile", profileHandler, RequireAuth())

    // Protected route with multiple metadata options
    ctx.Post("/api/admin/users", createUserHandler, RequireAuth(), RequirePermission("admin:create_user"))

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

### ⚡ Generic ORM Example

```go
package main

import (
    "context"
    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/core"
    "github.com/chuccp/go-web-frame/model"
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

func (u *UserModel) Init(db *core.DB, ctx *core.Context) error {
    u.Model = model.NewModel[*User](db, "t_user")
    // Auto create table if not exists
    return u.CreateTable()
}

func main() {
    app := wf.NewWithAutoConfig()
    // Register model to DI container
    app.AddModel(&UserModel{})

    // Example ORM operations
    app.Get("/users", func(c *web.Request) (any, error) {
        userModel := wf.GetModel[*UserModel](c.Context())

        // Query with chain API
        users, err := userModel.Query().
            Where("age > ?", 18).
            Order("id desc").
            All()

        return users, err
    })

    app.Post("/users", func(c *web.Request) (any, error) {
        userModel := wf.GetModel[*UserModel](c.Context())

        // Create user
        user := &User{Name: "John", Age: 25}
        err := userModel.Save(user)
        return user.Id, err
    })

    app.Put("/users/:id", func(c *web.Request) (any, error) {
        userModel := wf.GetModel[*UserModel](c.Context())
        id := c.ParamInt("id")

        // Update user
        return nil, userModel.Update().
            Where("id = ?", id).
            UpdateColumn("name", "John Updated")
    })

    app.Delete("/users/:id", func(c *web.Request) (any, error) {
        userModel := wf.GetModel[*UserModel](c.Context())
        id := c.ParamInt("id")

        // Delete user
        return nil, userModel.Delete().
            Where("id = ?", id).
            Delete()
    })

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
   - `IComponent`: Independent components initialized with config
   - `IRunner`: Background task runners (extends IService)
   - `IFilter`: HTTP request filters/middleware (extends IService)
   - `Context`: Dependency injection container that manages all components

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

1. **Create**: Initialize `WebFrame` with `NewWithAutoConfig()` or `New(config)`
2. **Register**: Add routes, controllers, models, services, components, and runners
3. **Configure**: Customize settings, add middleware, configure logging
4. **Run**: Start the server with `Run(ctx)`

## Configuration Example

Example YAML configuration with database connection pool:

```yaml
server:
  port: 8080
  mode: debug # or release

web:
  db:
    type: mysql
    host: localhost
    port: 3306
    username: root
    password: your_password
    database: your_db
    charset: utf8mb4
    # Connection pool settings (optional)
    max_open_conns: 100
    max_idle_conns: 10
    conn_max_lifetime: 3600 # seconds

log:
  level: info
  path: ./logs/app.log
  # Log rotation configuration (optional, with defaults)
  max_size: 100      # Max size of a single log file (MB), default 500
  max_backups: 5     # Max number of old log files to retain, default 3
  max_age: 7         # Max number of days to retain old log files, default 30
  compress: true     # Whether to compress old log files, default true
  local_time: false  # Whether to use local time, default false (uses UTC)
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

- [Go Reference Documentation](https://pkg.go.dev/github.com/chuccp/go-web-frame)
- [Example Applications](./example/)
- [CLAUDE.md](./CLAUDE.md) - Detailed developer guide for Claude Code

## Contributing

Contributions are welcome! Please feel free to submit Issues and Pull Requests. Before submitting a PR:

1. Run tests to ensure everything passes
2. Keep the code style consistent with the project
3. Add appropriate tests for new features
4. Update related documentation

## License

MIT License - see [LICENSE](./LICENSE) for details

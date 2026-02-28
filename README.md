# Go Web Frame

A modern, feature-rich web framework for Go built on top of Gin, providing a structured approach to building enterprise-grade web applications.

## Features

- **MVC-like Architecture**: Clean separation of concerns with services, controllers, and models
- **Dependency Injection**: Built-in DI container for managing component lifecycle
- **Database Integration**: SQLite, Redis, and extensible database abstraction layer
- **Component System**: Reusable components including caching, rate limiting, and local cache
- **RESTful Support**: Simplified REST controller implementation
- **Daemon/Service Mode**: Run applications as system services on Windows, Linux, and macOS
- **Auto-Configuration**: Auto-loading config from JSON, YAML, or TOML files
- **Advanced Logging**: Structured logging powered by Zap
- **Background Tasks**: Built-in runner system for background processing
- **Request Filtering**: HTTP middleware/filter system for cross-cutting concerns

## Quick Start

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
    // Create web frame instance with auto-config
    app := wf.NewWithAutoConfig()

    // Register simple route
    app.Get("/", func(c *web.Request) (any, error) {
        return "Hello, World!", nil
    })

    // Run with context
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Example: auto-shutdown after 10 seconds
    go func() {
        time.Sleep(time.Second * 10)
        cancel()
    }()

    if err := app.Run(ctx); err != nil {
        log.PrintPanic(err)
    }
}
```

### REST Controller Example

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

// Initialize controller and register routes
func (u *UserController) Init(ctx *core.Context) error {
    // Register routes directly through context
    ctx.Get("/users", u.GetUsers)
    ctx.Get("/users/:id", u.GetUser)
    ctx.Post("/users", u.CreateUser)

    return nil
}

// Handler: Get all users
func (u *UserController) GetUsers(c *web.Request) (any, error) {
    // Example: Access query parameters
    page := c.Query("page")
    limit := c.Query("limit")

    return map[string]any{
        "users": []string{"alice", "bob"},
        "page":  page,
        "limit": limit,
    }, nil
}

// Handler: Get single user by ID
func (u *UserController) GetUser(c *web.Request) (any, error) {
    id := c.Param("id")
    return map[string]any{
        "id":   id,
        "name": "alice",
    }, nil
}

// Handler: Create new user
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

## Installation

```bash
go get github.com/chuccp/go-web-frame
```

## Architecture

### Core Components

- **Core Abstractions** (`./core`): Interfaces defining component model (IService, IModel, IRest, etc.)
- **Web Layer** (`./web`): Gin-based HTTP handling, routing, and middleware
- **Data Access** (`./db`, `./model`, `./sqlite`, `./redis`): Database integration and ORM-like capabilities
- **Configuration** (`./config`): Auto-loading config system with Viper
- **Logging** (`./log`): Zap-powered structured logging
- **Components** (`./component`): Reusable components (cache, rate limiting, local cache)
- **Utilities** (`./util`): Helper functions and utilities

### Application Lifecycle

1. **Create**: Initialize WebFrame with `NewWithAutoConfig()` or `New(config)`
2. **Register**: Add routes, controllers, models, services, components, and runners
3. **Configure**: Customize settings, add middleware, set up logging
4. **Run**: Start the server with `Run()` or run as a service with daemon mode

## Development Commands

```bash
# Run hello world example
go run example/helloworld/helloworld.go

# Run rest example
go run example/rest/rest.go

# Run all tests
go test ./...

# Run specific package tests
go test ./core

# Format code
gofmt -w ./...

# Lint code
golint ./...
```

## Documentation

- [API Documentation](https://pkg.go.dev/github.com/chuccp/go-web-frame)
- [Example Applications](./example/)
- [CLAUDE.md](./CLAUDE.md) - Detailed developer guide

## License

MIT License - see [LICENSE](./LICENSE) for details


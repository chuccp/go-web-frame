# Go Web Frame

A modern, feature-rich web framework for Go built on top of Gin, providing a structured approach to building enterprise-grade web applications.

---

**🌐 Language / 语言**
[English](README.md) • [中文](README_ZH.md)

---

## Features

- **MVC-like Architecture**: Clean separation of concerns with services, controllers, and models
- **⚡ Type-safe Generic ORM**: Zero-boilerplate CRUD operations with generics, no reflection overhead
- **Dependency Injection**: Built-in DI container for managing component lifecycle
- **Database Integration**: SQLite, Redis, MySQL and extensible database abstraction layer
- **Component System**: Reusable components including caching, rate limiting, and local cache
- **RESTful Support**: Simplified REST controller implementation
- **Daemon/Service Mode**: Run applications as system services on Windows, Linux, and macOS
- **Auto-Configuration**: Auto-loading config from JSON, YAML, or TOML files
- **Advanced Logging**: Structured logging powered by Zap
- **Background Tasks**: Built-in runner system for background processing
- **Request Filtering**: HTTP middleware/filter system for cross-cutting concerns
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
    // 创建web框架实例，自动加载配置
    app := wf.NewWithAutoConfig()

    // 注册简单路由
    app.Get("/", func(c *web.Request) (any, error) {
        return "Hello, World!", nil
    })

    // 使用上下文运行服务
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // 示例：10秒后自动关闭
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
    // 嵌入core.IService接口
    core.IService
}

// 初始化控制器并注册路由
func (u *UserController) Init(ctx *core.Context) error {
    // 通过上下文注册路由
    ctx.Get("/users", u.GetUsers)
    ctx.Get("/users/:id", u.GetUser)
    ctx.Post("/users", u.CreateUser)

    return nil
}

// 处理器：获取所有用户
func (u *UserController) GetUsers(c *web.Request) (any, error) {
    // 示例：访问查询参数
    page := c.Query("page")
    limit := c.Query("limit")

    return map[string]any{
        "users": []string{"alice", "bob"},
        "page":  page,
        "limit": limit,
    }, nil
}

// 处理器：根据ID获取单个用户
func (u *UserController) GetUser(c *web.Request) (any, error) {
    id := c.Param("id")
    return map[string]any{
        "id":   id,
        "name": "alice",
    }, nil
}

// 处理器：创建新用户
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

### ⚡ Generic ORM Example

```go
package main

import (
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

## 🏗️ Architecture

### Core Components

#### [1. Core Abstractions](./core)
Key interfaces that define the component model:
- **IService**: Base interface for all services requiring initialization
- **IModel**: Data access layer interface with CRUD and table management
- **IRest**: REST controller interface (extends IService)
- **IComponent**: Independent components initialized with config
- **IRunner**: Background task runners (extends IService and IRun)
- **IFilter**: HTTP request filters/middleware (extends IService and web.Filter)

#### [2. Web Layer](./web)
HTTP handling layer built on Gin:
- Request/response abstraction
- Routing with support for all HTTP methods (GET, POST, PUT, DELETE, etc.)
- Filter/middleware system
- Conversion between service responses and standardized HTTP responses

#### [3. Data Access](./db)
- **./db**: Multi-database abstraction layer (MySQL, SQLite) powered by GORM
- **./model**: Type-safe generic base model with zero-boilerplate CRUD operations
- **./sqlite**: SQLite-specific configuration and initialization
- **./redis**: Redis integration for caching and messaging

#### 4. Other Key Packages
- **./config**: Configuration management with Viper (supports JSON, YAML, TOML)
- **./log**: Structured logging powered by Zap with rotation support
- **./component**: Reusable components: cache, local cache, rate limiting, captcha, QR code generation, cron scheduled tasks, input validation
- **./util**: Comprehensive utility functions for strings, time, crypto, networking, and more

### Application Lifecycle

1. **Create**: Initialize `WebFrame` with `NewWithAutoConfig()` or `New(config)`
2. **Register**: Add routes, controllers, models, services, components, and runners
3. **Configure**: Customize settings, add middleware, configure logging
4. **Run**: Start the server with `Run()` or run in daemon/service mode

## 🛠️ Development Commands

```bash
# Run the hello world example
go run example/helloworld/helloworld.go

# Run the REST example
go run example/rest/rest.go

# Run the ORM model example
go run example/model/model.go

# Build the framework (library only)
go build

# Run all tests
go test ./...

# Run tests in a specific package
go test ./core

# Run tests with verbose output
go test -v ./core

# Run a specific test case
go test -v ./core -run TestSpecificFunction

# Format code
gofmt -w ./...

# Alternative formatting with gofumpt (if installed)
gofumpt -w ./...

# Lint code
golint ./...

# Run application as a daemon/service
# (Requires implementing AppService interface)
go run your_app.go
# Stop the daemon
go run your_app.go -stop
```

## 📚 文档

- [API文档](https://pkg.go.dev/github.com/chuccp/go-web-frame)
- [示例应用](./example/)
- [CLAUDE.md](./CLAUDE.md) - 详细的开发者指南

## 🤝 贡献

欢迎提交Issue和Pull Request！请确保遵循以下准则：

1. 提交PR前请运行测试
2. 保持代码风格一致
3. 为新功能添加适当的测试
4. 更新相关文档

## 📄 许可证

MIT License - 查看[LICENSE](./LICENSE)了解详情


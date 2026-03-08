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

### 核心组件

#### [1. 核心抽象层](./core)
定义了组件模型的关键接口：
- **IService**：所有需要初始化的服务基接口
- **IModel**：数据访问层接口，包含CRUD和表管理
- **IRest**：REST控制器接口（继承IService）
- **IComponent**：独立组件，通过配置初始化
- **IRunner**：后台任务运行器（继承IService和IRun）
- **IFilter**：HTTP请求过滤器（继承IService和web.Filter）

#### [2. Web层](./web)
基于Gin的HTTP处理层：
- 请求/响应处理
- 支持HTTP方法的路由（GET、POST、PUT、DELETE等）
- 过滤器/中间件支持
- 服务响应与HTTP响应的转换

#### [3. 数据访问](./db)
- **./db**：支持多数据库的数据库抽象层
- **./model**：基础模型实现和工具
- **./sqlite**：SQLite特定实现
- **./redis**：Redis集成

#### [4. 其他关键包](.)
- **./config**：配置管理
- **./log**：基于Zap的日志系统
- **./component**：可重用组件（缓存、限流、本地缓存）
- **./util**：工具函数

### 应用生命周期

1. **创建**：使用`NewWithAutoConfig()`或`New(config)`初始化WebFrame
2. **注册**：添加路由、控制器、模型、服务、组件和运行器
3. **配置**：自定义设置、添加中间件、配置日志
4. **运行**：使用`Run()`启动服务器，或使用守护进程模式运行

## 🛠️ 开发命令

```bash
# 运行hello world示例
go run example/helloworld/helloworld.go

# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./core

# 格式化代码
gofmt -w ./...

# 代码检查
golint ./...

# 运行应用作为守护进程
go run your_app.go
# 停止守护进程
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


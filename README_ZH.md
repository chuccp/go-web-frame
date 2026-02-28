<div align="center">
  <h1>Go Web Frame</h1>
  <p>一个基于Gin构建的现代化Go Web框架，提供结构化的方式来构建企业级Web应用。</p>
  <div style="border: 1px solid #e0e0e0; border-radius: 4px; padding: 8px; display: inline-block;">
    <a href="#" onclick="switchLang('en'); return false;" style="margin: 0 10px; text-decoration: none; color: #337ab7;">English</a>
    <span style="color: #ccc;">|</span>
    <a href="#" onclick="switchLang('zh'); return false;" style="margin: 0 10px; text-decoration: none; font-weight: bold; color: #333;">中文</a>
  </div>
</div>

<div id="zh-content">

# Go Web Frame

一个基于Gin构建的现代化Go Web框架，提供结构化的方式来构建企业级Web应用。

## 🌟 特性

- 🎯 **类MVC架构**：清晰的关注点分离，包含服务、控制器和模型
- 🧩 **依赖注入**：内置DI容器，管理组件生命周期
- 📦 **数据库集成**：SQLite、Redis支持，可扩展的数据库抽象层
- 🛠️ **组件系统**：可重用组件，包括缓存、限流和本地缓存
- 🌐 **RESTful支持**：简化的REST控制器实现
- 🚀 **守护进程模式**：在Windows、Linux和macOS上作为系统服务运行
- ⚙️ **自动配置**：从JSON、YAML或TOML文件自动加载配置
- 📝 **高级日志**：基于Zap的结构化日志
- 🔄 **后台任务**：内置的后台任务运行器系统
- 🛡️ **请求过滤**：HTTP中间件/过滤器系统，处理横切关注点

## 🚀 快速开始

### 安装

```bash
go get github.com/chuccp/go-web-frame
```

### 🏠 Hello World 示例

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

### 🔌 REST控制器示例

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

## 🏗️ 架构

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

</div>

<div id="en-content" style="display: none;">

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

## Architecture

### Core Components

#### 1. Core Abstractions (./core)
The framework is built around several key interfaces that define the component model:
- **IService**: Base interface for services that need initialization
- **IModel**: Data access layer interface with CRUD and table management
- **IRest**: REST controller interface (extends `IService`)
- **IComponent**: Independent components that initialize with config
- **IRunner**: Background task runners (extends `IService` and `IRun`)
- **IFilter**: HTTP request filters (extends `IService` and `web.Filter`)

#### 2. Web Layer (./web)
Built on Gin, provides:
- Request/Response handling
- Routing with HTTP method support (GET, POST, PUT, DELETE, etc.)
- Filter/middleware support
- Conversion between service responses and HTTP responses

#### 3. Data Access
- **./db**: Database abstraction layer supporting multiple databases
- **./model**: Base model implementations and utilities
- **./sqlite**: SQLite-specific implementation
- **./redis**: Redis integration

#### 4. Other Key Packages
- **./config**: Configuration management
- **./log**: Logging based on Zap
- **./component**: Reusable components (cache, rate limiting, local cache)
- **./util**: Utility functions

### Application Lifecycle

1. **Create**: Initialize WebFrame with `NewWithAutoConfig()` or `New(config)`
2. **Register**: Add routes, controllers, models, services, components, and runners
3. **Configure**: Customize settings, add middleware, set up logging
4. **Run**: Start the server with `Run()` or run as a service with daemon mode

## Development Commands

```bash
# Run hello world example
go run example/helloworld/helloworld.go

# Run all tests
go test ./...

# Run specific package tests
go test ./core

# Format code
gofmt -w ./...

# Lint code
golint ./...

# Run application as a daemon/service
go run your_app.go
# Stop the daemon
go run your_app.go -stop
```

## Documentation

- [API Documentation](https://pkg.go.dev/github.com/chuccp/go-web-frame)
- [Example Applications](./example/)
- [CLAUDE.md](./CLAUDE.md) - Detailed developer guide

## License

MIT License - see [LICENSE](./LICENSE) for details

</div>

<script>
// 持久化切换状态
const savedLang = localStorage.getItem('preferredLang');
if (savedLang === 'en') {
  document.getElementById('en-content').style.display = 'block';
  document.getElementById('zh-content').style.display = 'none';
}

// 语言切换函数
window.switchLang = function(lang) {
  if (lang === 'en') {
    document.getElementById('en-content').style.display = 'block';
    document.getElementById('zh-content').style.display = 'none';
    localStorage.setItem('preferredLang', 'en');
  } else {
    document.getElementById('zh-content').style.display = 'block';
    document.getElementById('en-content').style.display = 'none';
    localStorage.setItem('preferredLang', 'zh');
  }
};
</script>

## 🌟 特性

- 🎯 **类MVC架构**：清晰的关注点分离，包含服务、控制器和模型
- 🧩 **依赖注入**：内置DI容器，管理组件生命周期
- 📦 **数据库集成**：SQLite、Redis支持，可扩展的数据库抽象层
- 🛠️ **组件系统**：可重用组件，包括缓存、限流和本地缓存
- 🌐 **RESTful支持**：简化的REST控制器实现
- 🚀 **守护进程模式**：在Windows、Linux和macOS上作为系统服务运行
- ⚙️ **自动配置**：从JSON、YAML或TOML文件自动加载配置
- 📝 **高级日志**：基于Zap的结构化日志
- 🔄 **后台任务**：内置的后台任务运行器系统
- 🛡️ **请求过滤**：HTTP中间件/过滤器系统，处理横切关注点

## 🚀 快速开始

### 安装

```bash
go get github.com/chuccp/go-web-frame
```

### 🏠 Hello World 示例

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

### 🔌 REST控制器示例

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

## 🏗️ 架构

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

---

**使用提示**：如果您在使用过程中遇到问题或有改进建议，请随时创建Issue或联系维护者。

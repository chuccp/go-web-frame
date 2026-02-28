# Go Web Frame

一个基于Gin构建的现代化Go Web框架，提供结构化的方式来构建企业级Web应用。

---

**🌐 Language / 语言**
[English](README.md) • [中文](README_ZH.md)

---

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

## 📄 许可证

MIT License - 查看[LICENSE](./LICENSE)了解详情

---

**🌐 Language / 语言**
[English](README.md) • [中文](README_ZH.md)

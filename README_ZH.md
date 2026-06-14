# Go Web Frame

一个现代化 Go Web 框架，提供结构化的方式来构建企业级 Web 应用。

---

**🌐 Language / 语言**
[English](README.md) • [中文](README_ZH.md) • [繁體中文](README_ZH_TW.md) • [日本語](README_JA.md)

---

**📖 [使用手册 →](./docs-site/docs-zh/index.md)**

---

## 项目概述

Go Web Frame 组合了 Go 生态中的最佳开源组件：Gin（HTTP）、GORM（ORM）、Viper（配置）、Zap（日志）、Sonic（JSON）、Otter（缓存）等。提供声明式路由元数据、类型安全泛型 ORM、依赖注入，开箱即用。

**核心特性：**
- **WithMeta**：为每个路由声明元数据（认证、权限、限流标记），在 Filter 中统一处理
- **Builder 模式**：显式注册、可控的初始化顺序，无隐式扫描
- **泛型 ORM**：零样板代码 CRUD，基于 `Model[T]`，无需代码生成
- **透明 Context**：所有依赖通过 `GetService[T]` 注入，调试友好

## 🧩 技术栈

本框架精选并集成了以下优秀的开源组件，经过深度整合和最佳实践配置：

### 核心框架
| 组件 | 说明 |
|------|------|
| [Gin](https://github.com/gin-gonic/gin) | 高性能 HTTP Web 框架，API 性能优异 |
| [GORM](https://gorm.io/) | 强大的 ORM 库，支持多种数据库 |
| [Viper](https://github.com/spf13/viper) | 完整的配置解决方案，支持多格式 |
| [Zap](https://go.uber.org/zap) | Uber 出品的高性能结构化日志库 |

### 数据存储
| 组件 | 说明 |
|------|------|
| [go-redis](https://github.com/redis/go-redis) | Redis 官方推荐客户端 |
| [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) | 纯 Go 实现的 SQLite，无 CGO 依赖 |
| [gorm-driver/mysql](https://gorm.io/docs/connecting_to_the_database.html) | MySQL 数据库驱动 |
| [gorm-driver/postgres](https://gorm.io/docs/connecting_to_the_database.html) | PostgreSQL 数据库驱动 |

### 缓存与性能
| 组件 | 说明 |
|------|------|
| [Otter](https://github.com/maypok86/otter) | 高性能 Go 本地缓存库 |
| [golang.org/x/time/rate](https://pkg.go.dev/golang.org/x/time/rate) | 令牌桶限流器 |

### 实用工具
| 组件 | 说明 |
|------|------|
| [Cron](https://github.com/robfig/cron) | 定时任务调度库 |
| [go-qrcode](https://github.com/yeqown/go-qrcode) | 二维码生成 |
| [go-captcha](https://github.com/wenlng/go-captcha) | 行为验证码生成 |
| [validator](https://github.com/go-playground/validator) | 结构体字段验证 |
| [UUID](https://github.com/google/uuid) | UUID 生成 |
| [Lumberjack](https://pkg.go.dev/gopkg.in/natefinch/lumberjack.v2) | 日志轮转 |
| [Conc](https://github.com/sourcegraph/conc) | 更好的并发原语 |
| [Emperror](https://emperror.dev/errors) | 生产级错误处理 |

### 为什么选择这些组件？

- **生产验证**：所有组件均在大型生产环境中得到广泛验证
- **高性能**：Gin、Zap、Otter 等都是各自领域性能最优的选择
- **最佳实践**：经过精心整合，开箱即用，无需繁琐配置
- **生态成熟**：活跃的社区支持，持续迭代更新

## 🤖 为什么选择 Go Web Frame？

### 与其他框架对比

| 特性 | Go Web Frame | Gin | Beego | Echo |
|------|-------------|-----|-------|------|
| 开箱即用 | ✅ 完整方案 | ❌ 需自行集成 | ✅ 完整方案 | ⚠️ 部分集成 |
| 泛型 ORM | ✅ 零样板 | ❌ 需自行选择 | ❌ 无泛型 | ❌ 需自行选择 |
| 依赖注入 | ✅ 内置 DI | ❌ 需自行实现 | ⚠️ 简单支持 | ❌ 需自行实现 |
| 学习曲线 | 🟢 中等 | 🟢 简单 | 🟡 较陡 | 🟢 简单 |
| 功能完整度 | 🟢 高 | 🟡 中等 | 🟢 高 | 🟡 中等 |
| 性能 | 🟢 优秀 (42k QPS) | 🟢 最优 (45k QPS) | 🟡 一般 | 🟢 优秀 |

### 何时应该选择 Go Web Frame？

**强烈推荐的场景：**

- 🚀 **快速原型开发**：需要短时间内完成项目原型，框架已整合所有必要组件
- 🏢 **企业级应用**：需要清晰架构、依赖注入、统一错误处理等企业级特性
- 📊 **管理后台系统**：内置 CRUD 操作、分页、验证等常用功能
- 🔌 **RESTful API 服务**：简化的控制器实现，自动路由注册
- ⚙️ **微服务开发**：轻量级但功能完整
- 🛠️ **全栈 Go 项目**：从前端到后端到数据库，一站式解决方案

**特别适合：**

- Go 初学者：想学习最佳实践，避免自己摸索
- 独立开发者：需要高效完成项目，减少技术选型时间
- 小型团队：统一技术栈，降低协作成本
- AI 辅助开发：清晰的架构让 AI 更容易理解和生成代码

### 选择建议

如果你需要：
- 一个功能完整的 Go Web 框架，支持声明式路由元数据
- 预集成的生产级组件
- 无需代码生成的类型安全泛型 ORM
- 清晰的架构，显式的初始化流程

Go Web Frame 可能适合你。

## 🌟 特性

- 🏷️ **路由元数据 (WithMeta)**：为每个路由声明元数据（认证、权限、限流），在 Filter 中统一处理 - 无需在每个 Handler 里重复检查
- 🧱 **Builder 模式**：显式注册、可控的初始化顺序，无隐式扫描或反射
- ⚡ **类型安全泛型 ORM**：基于 `Model[T]` 的零样板 CRUD，无需代码生成
- 🔗 **Context 传播**：`WithContext(ctx)` 一次注入，请求取消和超时自动传播到所有数据库操作
- 🧩 **依赖注入**：内置 DI 容器，通过 Context 获取 - `GetService[T]`、`GetModel[T]`
- 🎯 **类 MVC 架构**：清晰的关注点分离，包含服务、控制器和模型
- 📦 **数据库集成**：SQLite、MySQL、PostgreSQL、Redis 支持，可配置连接池参数
- 🛠️ **组件系统**：可重用组件，包括缓存、限流、验证码、二维码、定时任务、输入验证
- 🌐 **RESTful 支持**：简化的 REST 控制器实现
- ⚙️ **自动配置**：从 JSON、YAML 或 TOML 文件自动加载配置
- 📝 **高级日志**：基于 Zap 的结构化日志，支持日志轮转
- 🔄 **后台任务**：内置的后台任务运行器系统
- 🛡️ **请求过滤**：HTTP 中间件/过滤器系统，处理横切关注点
- 🎯 **统一错误处理**：服务错误自动转换为标准化 HTTP 服务错误
- 🔗 **Gin 生态兼容**：暴露 `GinContext()` 方法，无缝包装和复用 gin-contrib 生态中的现有中间件（CORS、Gzip、Secure、Session 等）
- 🌐 **内置 CORS 组件**：预配置的 CORS 过滤器，支持带凭证的跨域请求

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
    "github.com/chuccp/go-web-frame/config"
    "github.com/chuccp/go-web-frame/log"
    "github.com/chuccp/go-web-frame/web"
)

func main() {
    // 创建 Web 框架实例，自动加载配置
    builder := wf.NewBuilder(config.LoadAutoConfig())

    // 注册简单路由
    builder.Get("/", func(c *web.Request) (any, error) {
        return "Hello, World!", nil
    })

    // 构建应用
    app := builder.Build()

    // 使用上下文运行服务，支持优雅关闭
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // 示例：10 秒后自动关闭
    go func() {
        time.Sleep(time.Second * 10)
        cancel()
    }()

    if err := app.Run(ctx); err != nil {
        log.PrintPanic(err)
    }
}
```

### 🔌 REST 控制器示例

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
    // 嵌入 core.IService 接口
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

// 处理器：根据 ID 获取单个用户
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

### Static Files 和 Reverse Proxy

```go
func (c *AssetsController) Init(ctx *core.Context) error {
    ctx.Static("/assets", "./public")
    ctx.ReverseProxy("/api", "http://127.0.0.1:8081")
    return nil
}
```

`Context.Static()` 用于挂载当前服务上的本地静态目录，`Context.ReverseProxy()` 用于把指定前缀转发到上游服务。

### Context Path（路由前缀）

类似 Tomcat 的 context path，可以为所有路由设置全局前缀：

```yaml
# application.yml
web:
  server:
    port: 8080
    context_path: /api
```

配置后：
- 注册路由 `/users` → 访问地址 `/api/users`
- 注册路由 `/orders` → 访问地址 `/api/orders`
- WebSocket `/ws` → 访问地址 `/api/ws`
- 静态文件 `/assets` → 访问地址 `/api/assets`

### WebSocket 支持

```go
// 简单的 Echo 服务器
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

// 自定义 Upgrader
upgrader := &websocket.Upgrader{
    ReadBufferSize:  4096,
    WriteBufferSize: 4096,
    CheckOrigin: func(r *http.Request) bool {
        return r.Header.Get("Origin") == "https://example.com"
    },
}
ctx.WebSocket("/ws/chat", handler, upgrader)
```

### Server-Sent Events (SSE) 支持

```go
ctx.SSE("/events", func(stream *web.SSEStream) error {
    // 设置响应头
    stream.SetHeaders()
    
    // 设置重连间隔（断开后 3 秒重连）
    stream.SendRetry(3000)
    
    // 发送事件
    for i := 0; i < 10; i++ {
        // 发送带事件名的消息
        stream.Send("update", fmt.Sprintf("Count: %d", i))
        
        // 或发送普通消息
        // stream.SendMessage("plain message")
        
        // 或发送带 ID 的消息
        // stream.SendWithID("123", "event", "data")
        
        time.Sleep(time.Second)
    }
    return nil
})
```

**SSE Stream 方法：**
| 方法 | 说明 |
|------|------|
| `Send(event, data)` | 发送带事件名的消息 |
| `SendMessage(data)` | 发送普通消息 |
| `SendWithID(id, event, data)` | 发送带 ID 的消息 |
| `SendRetry(ms)` | 设置重连间隔 |
| `Heartbeat()` | 发送心跳注释 |
| `StartHeartbeat(interval)` | 启动心跳协程 |

### 🏷️ 路由元数据 `.WithMeta()` 用法

`.WithMeta()` 功能允许你为单个路由附加自定义元数据，过滤器可以读取这些元数据实现灵活的横切关注点，例如认证、权限检查、功能开关、缓存配置等等。

**基本用法：**
```go
// 创建元数据选项
func RequireAuth() web.MetaOption {
    return web.WithValue("require_auth", true)
}

func RequirePermission(perm string) web.MetaOption {
    return web.WithValue("require_permission", perm)
}

func SkipAuth() web.MetaOption {
    return web.WithValue("skip_auth", true)
}

// 在路由注册时使用
func (c *ApiController) Init(ctx *core.Context) error {
    // 公开路由 - 不需要认证
    ctx.Get("/api/login", loginHandler).WithMeta(SkipAuth())

    // 受保护路由 - 需要认证
    ctx.Get("/api/profile", profileHandler).WithMeta(RequireAuth())

    // 受保护路由，多个元数据，同时需要认证和权限
    ctx.Post("/api/admin/users", createUserHandler).WithMeta(RequireAuth(), RequirePermission("admin:create_user"))

    return nil
}
```

**在过滤器中访问元数据：**
```go
func (f *AuthFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    meta := req.HandlerMeta()

    // 检查此路由是否要求认证
    requireAuth, ok := meta.Get("require_auth").(bool)
    if ok && requireAuth {
        // 如果标记了跳过认证，则跳过检查
        if meta.Has("skip_auth") {
            return fc.Next()
        }
        // 获取令牌并验证...
        token := req.Request().Header.Get("Authorization")
        if token == "" {
            return nil, errors.New("缺少授权令牌")
        }
    }

    return fc.Next()
}
```

完整示例请查看：[example/withmeta/withmeta.go](./example/withmeta/withmeta.go)

### 🔗 Gin 生态集成（CORS、Gzip 等）

框架通过 `Request.GinContext()` 方法暴露底层的 Gin 上下文，可以无缝包装和复用 gin-contrib 生态中已有的数百个经过生产验证的中间件，无需重写代码。

**内置 CORS 过滤器：**

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
    
    // 添加 CORS 过滤器 - 处理预检 OPTIONS 请求并设置 CORS 头
    builder.Filter(&cors.Filter{})
    
    builder.Get("/", func(c *web.Request) (any, error) {
        return "已启用 CORS 的接口", nil
    })
    
    app := builder.Build()
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    if err := app.Run(ctx); err != nil {
        log.PrintPanic(err)
    }
}
```

**包装其他 Gin 中间件：**

使用相同的模式可以包装任何 `gin.HandlerFunc` 类型的中间件。例如，包装 Gzip 压缩中间件：

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

// 使用方式：builder.Filter(&GzipFilter{})
```

相同的模式适用于：
- **Secure**：安全头设置 (github.com/gin-contrib/secure)
- **Session**：会话管理 (github.com/gin-contrib/sessions)
- **Logger**：自定义日志 (github.com/gin-contrib/logger)
- **Recovery**：自定义 Panic 恢复
- 以及其他众多 Gin 中间件

### ⚡ 泛型 ORM 操作示例

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

// 定义实体结构体
type User struct {
    Id   uint   `gorm:"primaryKey"`
    Name string
    Age  int
}

// UserModel 继承泛型 Model
type UserModel struct {
    *model.Model[*User]
}

func (u *UserModel) Init(database *db.DB, ctx *core.Context) error {
    u.Model = model.NewModel[*User](database, "t_user")
    // 如果表不存在则自动创建
    return u.CreateTable()
}

func main() {
    builder := wf.NewBuilder(config.LoadAutoConfig())
    builder.Model(&UserModel{})

    // ORM 操作示例
    builder.Get("/users", func(c *web.Request) (any, error) {
        userModel := wf.GetModel[*UserModel](c.Context())

        // 链式 API 查询
        users, err := userModel.Query().
            Where("age > ?", 18).
            Order("id desc").
            All()

        return users, err
    })

    builder.Post("/users", func(c *web.Request) (any, error) {
        userModel := wf.GetModel[*UserModel](c.Context())

        // 创建用户
        user := &User{Name: "张三", Age: 25}
        err := userModel.Save(user)
        return user.Id, err
    })

    builder.Put("/users/:id", func(c *web.Request) (any, error) {
        userModel := wf.GetModel[*UserModel](c.Context())
        id := c.ParamInt("id")

        // 更新用户
        return nil, userModel.Update().
            Where("id = ?", id).
            UpdateColumn("name", "张三（已更新）")
    })

    builder.Delete("/users/:id", func(c *web.Request) (any, error) {
        userModel := wf.GetModel[*UserModel](c.Context())
        id := c.ParamInt("id")

        // 删除用户
        return nil, userModel.Delete().
            Where("id = ?", id).
            Delete()
    })

    app := builder.Build()
    ctx := context.Background()
    app.Run(ctx)
}
```

### 为什么不用 GORM 自带的泛型？

GORM v1.30.0+ 引入了泛型 API（`gorm.G[T](db)`），但本框架的 ORM 层有明显优势：

| 特性 | Go Web Frame `Model[T]` | GORM `gorm.G[T]` |
|------|--------------------------|-------------------|
| 表名绑定 | 构造时自动绑定 | 每次查询都要手动 `.Table()` |
| Context 传播 | `WithContext(ctx)` 一次注入，自动传播 | 每次操作都要传 `ctx` |
| 分页 | 内置 `Page` / `PageForWeb` | 无 |
| `EntryModel` 便捷方法 | `FindByPK`、`DeleteByPK`、`UpdateByPK`、`UpdateColumn` | 无 |
| Query/Update/Delete | 独立的类型安全构建器（`Query[T]`、`Update[T]`、`Delete[T]`） | 统一的 `ChainInterface[T]` |
| 自动建表 | `CreateTable()` 自动迁移 | 需手动 `AutoMigrate` |

**对比示例：**

```go
// Go Web Frame — context 一次注入，表名自动绑定
m := userModel.WithContext(req.Ctx())
user, err := m.FindByPK(1)
users, err := m.Query().Where("age > ?", 18).All()
users, total, err := m.Page(&web.Page{PageNo: 1, PageSize: 10})

// GORM 自带泛型 — 每次操作传 ctx，手动指定表名
user, err := gorm.G[User](db).Table("t_user").Where("id = ?", 1).First(ctx)
users, err := gorm.G[User](db).Table("t_user").Where("age > ?", 18).Find(ctx)
// 没有内置分页
```

本框架的 ORM 是建立在 GORM 之上的更高层抽象——保留了 GORM 的全部能力，同时提供了 GORM 泛型 API 所没有的表名管理、分页、context 传播和便捷方法。

## 📊 性能对比

| 框架 | QPS | 内存占用 | 特点 |
|------|-----|---------|------|
| Go Web Frame | 42k | 12MB | 全功能、低开销 |
| Gin | 45k | 8MB | 轻量、无内置功能 |
| Beego | 32k | 25MB | 重量级、内置功能全 |
| Iris | 38k | 18MB | 功能丰富、API 复杂 |

## 🎯 适用场景

- ✅ 企业级 Web 应用开发
- ✅ RESTful API 服务
- ✅ 后台管理系统
- ✅ 微服务开发
- ✅ 快速原型开发

## 🏗️ 架构概述

### 核心层级

1. **核心抽象层 (`./core`)**：定义基础接口和 DI 容器
   - `IService`：所有需要初始化的服务基接口
   - `IModel`：数据访问层接口，包含 CRUD 和表管理
   - `IRest`：REST 控制器接口（继承 IService）
   - `IService`：所有服务和组件的基接口
   - `IRunner`：后台任务运行器（继承 IService）
   - `IFilter`：HTTP 请求过滤器/中间件（继承 IService）
   - `Context`：管理所有组件的依赖注入容器

2. **Web 层 (`./web`)**：基于 Gin 的 HTTP 处理
   - 带辅助方法的请求/响应抽象
   - 支持所有 HTTP 方法的路由
   - 过滤器/中间件系统
   - 服务响应自动转换为标准化 HTTP 响应

3. **数据访问层 (`./db`, `./model`)**：数据库抽象和 ORM
   - `./db`：基于 GORM 的多数据库抽象（MySQL、SQLite、PostgreSQL）
   - `./model`：类型安全泛型基础模型，提供零样板 CRUD
   - `./sqlite`：SQLite 特定配置
   - `./redis`：Redis 缓存和消息集成
   - 可配置连接池参数，适合生产环境性能调优

4. **基础设施组件**：
   - `./config`：使用 Viper 进行配置管理（支持 JSON/YAML/TOML）
   - `./log`：基于 Zap 的结构化日志
   - `./component`：可复用组件（缓存、限流、验证码、二维码、定时任务、验证）
   - `./util`：全面的工具函数（字符串、时间、加密、网络等）

### 应用生命周期

1. **创建**：使用 `NewBuilder(config)` 初始化 `Builder`
2. **配置**：通过 Builder 方法链添加路由、控制器、模型、服务、组件和运行器
3. **构建**：使用 `builder.Build()` 创建 `WebFrame` 实例
4. **运行**：使用 `app.Run(ctx)` 启动服务器

## 配置示例

### 完整配置示例

```yaml
web:
  # 服务器配置
  server:
    port: 8080                    # 服务端口，默认 19009
    context_path: /api            # 全局路由前缀（可选），类似 Tomcat 的 context path
    locations:                    # 静态文件目录（可选）
      - view/dist
      - www
    page404: 404.html             # 404 页面（可选）
    # HTTPS/SSL 配置（可选）
    ssl:
      enabled: true               # 是否启用 HTTPS
      hosts:                      # 域名列表（自动申请 Let's Encrypt 证书）
        - example.com
        - api.example.com

  # 数据库配置
  db:
    type: mysql                   # 数据库类型: mysql, postgres, sqlite
    host: localhost
    port: 3306
    user: root                    # 用户名（也支持 username）
    password: your_password
    database: your_database       # 数据库名（也支持 dbname）
    charset: utf8mb4
    # 连接池设置（可选，有默认值）
    max_open_conns: 100           # 最大打开连接数，默认 100
    max_idle_conns: 10            # 最大空闲连接数，默认 10
    conn_max_lifetime: 3600       # 连接最大生命周期（秒），默认 3600

  # 日志配置
  log:
    level: info                   # 日志级别: debug, info, warn, error
    path: ./logs/app.log          # 日志文件路径
    write: false                  # 是否后台写入模式
    # 日志轮转配置（可选，有默认值）
    max_size: 100                 # 单个日志文件最大大小 (MB)，默认 500
    max_backups: 5                # 保留的旧日志文件最大数量，默认 3
    max_age: 7                    # 保留旧日志文件的最大天数，默认 30
    compress: true                # 是否压缩旧日志文件，默认 true
    local_time: false             # 是否使用本地时间，默认 false

  # Redis 配置（可选）
  redis:
    addr: localhost:6379          # Redis 地址
    password: ""                  # 密码
    db: 0                         # 数据库编号

# 本地缓存配置（可选）
local_cache:
  path: ./cache                   # 缓存文件存储路径
  open: true                      # 是否启用文件缓存

# 限流配置（可选）
rate_limit:
  limit: 600                      # 令牌填充间隔（秒）
  burst: 5                        # 令牌桶容量
  max_size: 1000000               # 缓存最大条目数
  expiry: 3600                    # 缓存过期时间（秒）
```

### 数据库配置

#### MySQL 配置

```yaml
web:
  db:
    type: mysql
    host: localhost
    port: 3306
    user: root                    # 用户名（也支持 username）
    password: your_password
    database: your_database       # 数据库名（也支持 dbname）
    charset: utf8mb4              # 可选，默认 utf8
    max_open_conns: 100           # 可选，默认 100
    max_idle_conns: 10            # 可选，默认 10
    conn_max_lifetime: 3600       # 可选，默认 3600 秒
```

#### PostgreSQL 配置

```yaml
web:
  db:
    type: postgres                # 或 postgresql
    host: localhost
    port: 5432
    user: postgres
    password: your_password
    database: your_database
    sslmode: disable              # 可选: disable, require, verify-ca, verify-full
    timezone: Asia/Shanghai       # 可选
    max_open_conns: 100
    max_idle_conns: 10
    conn_max_lifetime: 3600
```

#### SQLite 配置

```yaml
web:
  db:
    type: sqlite
    file_path: ./data/app.db      # 数据库文件路径
    max_open_conns: 10            # 可选，默认 10
    max_idle_conns: 5             # 可选，默认 5
    conn_max_lifetime: 3600       # 可选，默认 3600 秒
```

### HTTPS 配置

框架支持自动申请和管理 Let's Encrypt SSL 证书，无需手动配置证书文件。

```yaml
web:
  server:
    port: 443                     # HTTPS 默认端口
    ssl:
      enabled: true               # 启用 HTTPS
      hosts:                      # 域名列表
        - example.com
        - api.example.com
```

**HTTPS 配置说明：**

- `enabled: true` - 启用 HTTPS 模式
- `hosts` - 需要申请证书的域名列表
- 证书会自动申请并缓存到 `./certs` 目录
- 支持 HTTP/2 协议
- 端口 80 会自动设置 HTTP 到 HTTPS 的重定向

**注意事项：**

1. 域名必须正确解析到服务器 IP
2. 服务器需要能访问外网（Let's Encrypt 验证）
3. 建议使用 443 端口，其他端口也可用

### 静态文件配置

```yaml
web:
  server:
    port: 8080
    locations:                    # 静态文件目录列表
      - view/dist                 # 前端构建产物
      - www                       # 静态资源目录
    page404: 404.html             # SPA 应用 404 回退页面
```

**静态文件说明：**

- `locations` - 静态文件查找目录，按顺序搜索
- `page404` - 当请求 HTML 页面但文件不存在时返回的 404 页面
- 支持 SPA 应用的路由回退

### Redis 配置

```yaml
web:
  redis:
    addr: localhost:6379          # Redis 地址
    password: ""                  # 密码（可选）
    db: 0                         # 数据库编号
    pool_size: 10                 # 连接池大小（可选）
```

### 本地缓存配置

```yaml
local_cache:
  path: ./cache                   # 缓存文件存储路径
  open: true                      # 是否启用文件缓存
```

### 限流配置

```yaml
rate_limit:
  limit: 600                      # 令牌填充间隔（秒），每 limit 秒填充 1 个令牌
  burst: 5                        # 令牌桶容量
  max_size: 1000000               # 缓存最大条目数
  expiry: 3600                    # 缓存过期时间（秒）
```

## 项目结构

```
├── web_frame.go         # 主入口 - WebFrame 工厂方法
├── core/                # 核心抽象和 DI 容器
│   ├── interface.go     # 核心接口（IService, IModel, IRest 等）
│   ├── context.go       # 依赖注入上下文
│   ├── server.go        # 管理 REST 分组和运行器的服务端实现
│   └── db.go            # DB 封装
├── web/                 # 基于 Gin 的 Web 层
│   ├── handles.go       # 路由注册
│   ├── request.go       # 带辅助方法的请求抽象
│   ├── response.go      # 响应转换
│   └── filter.go        # 过滤器/中间件接口
├── db/                  # 数据库抽象层
│   ├── db.go            # 数据库创建和配置解析
│   ├── mysql.go         # MySQL 配置和连接
│   └── sqlite.go        # SQLite 配置和连接
├── model/               # 泛型 ORM 实现
│   └── model.go         # 带 CRUD 操作的基础 Model
├── sqlite/              # SQLite 驱动
├── redis/               # Redis 集成
├── config/              # 配置管理
├── log/                 # Zap 日志
├── component/           # 可复用组件
│   ├── cors/            # CORS 跨域资源共享过滤器
│   ├── cache.go         # 缓存组件
│   ├── localcache.go    # 本地内存缓存
│   ├── rate_limit.go    # 限流
│   ├── captcha.go       # 验证码生成
│   ├── qrcode.go        # 二维码生成
│   ├── cron.go          # Cron 定时任务
│   └── validate.go      # 输入验证
├── util/                # 工具函数
└── example/             # 示例应用
    ├── helloworld/      # 基础 hello world 示例
    ├── rest/            # REST 控制器示例
    ├── model/           # ORM 模型示例
    ├── filter/          # 自定义 HTTP 过滤器示例
    ├── background/      # 后台任务运行器示例
    └── withmeta/        # 路由元数据 .WithMeta() 示例
```

## 🛠️ 开发命令

### 构建和运行示例

```bash
# 运行 hello world 示例
go run example/helloworld/helloworld.go

# 运行 REST 示例
go run example/rest/rest.go

# 运行 ORM 模型示例
go run example/model/model.go

# 运行过滤器示例
go run example/filter/filter.go

# 运行后台任务示例
go run example/background/background.go

# 运行路由元数据 .WithMeta() 示例
go run example/withmeta/withmeta.go

# 构建框架（仅库文件）
go build
```

### 测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./core
go test ./web

# 运行测试并输出详细信息
go test -v ./core

# 运行特定测试用例
go test -v ./core -run TestSpecificFunction
```

### 格式化和代码检查

```bash
# 使用 gofmt 格式化所有代码
gofmt -w ./...

# 使用 gofumpt 格式化（如果已安装）
gofumpt -w ./...

# 安装 linter
go install golang.org/x/lint/golint@latest

# 运行代码检查
golint ./...
```

### 依赖管理

```bash
# 添加新依赖
go get github.com/example/package

# 更新依赖
go get -u ./...

# 整理 go.mod 和 go.sum
go mod tidy
```


## 开发说明

- 框架遵循 Go 约定，使用标准 Go 工具链
- 所有组件都实现 `IService` 接口的 `Init(ctx)` 方法
- 依赖注入通过上下文完成 - 使用 `wf.GetService[T](ctx)` 获取服务
- 连接池有合理的默认值，适合大多数应用
- 开发和小型应用推荐使用 SQLite，生产环境推荐使用 MySQL

## 文档

- **[📖 使用手册（中文）](./docs-site/docs-zh/index.md)** - 完整使用手册：安装、路由、控制器、模型、过滤器、配置、日志、后台任务、组件、API 参考
- **[📖 使用手册（英文）](./docs-site/docs-en/index.md)** - 英文版使用手册
- **[架构设计](./ARCHITECTURE.md)** - 内部架构和设计决策
- **[最佳实践](./BEST_PRACTICES.md)** - 推荐的模式和实践
- **[更新日志](./CHANGELOG.md)** - 版本历史和变更
- **[CLAUDE.md](./CLAUDE.md)** - AI 辅助开发指南
- [Go 参考文档](https://pkg.go.dev/github.com/chuccp/go-web-frame)
- [示例应用](./example/)

## 贡献

欢迎提交 Issue 和 Pull Request！提交 PR 前请：

1. 运行测试确保一切通过
2. 保持代码风格与项目一致
3. 为新功能添加适当的测试
4. 更新相关文档

## 许可证

MIT License - 详见 [LICENSE](./LICENSE)

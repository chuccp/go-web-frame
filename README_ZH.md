# Go Web Frame

一个现代化 Go Web 框架，提供结构化的方式来构建企业级 Web 应用。

---

**🌐 Language / 语言**
[English](README.md) • [中文](README_ZH.md)

---

## 项目概述

Go Web Frame 是一个 opinionated 的 Web 框架，通过基于组件的设计强制实现清晰的架构。它内置了依赖注入、类型安全 ORM、数据库集成，以及生产环境就绪的守护进程模式等特性。

## 🌟 特性

- 🎯 **类 MVC 架构**：清晰的关注点分离，包含服务、控制器和模型
- ⚡ **类型安全泛型 ORM**：零样板代码 CRUD 操作，基于泛型实现，无反射性能损耗
- 🧩 **依赖注入**：内置 DI 容器，管理组件生命周期
- 📦 **数据库集成**：SQLite、MySQL、PostgreSQL、Redis 支持，可扩展的数据库抽象层（基于 GORM）
- 🏊 **连接池配置**：可配置连接池参数（`max_open_conns`、`max_idle_conns`、`conn_max_lifetime`）
- 🛠️ **组件系统**：可重用组件，包括缓存、限流、验证码、二维码生成、定时任务、输入验证
- 🌐 **RESTful 支持**：简化的 REST 控制器实现
- 🚀 **守护进程模式**：在 Windows、Linux 和 macOS 上作为系统服务运行
- ⚙️ **自动配置**：从 JSON、YAML 或 TOML 文件自动加载配置
- 📝 **高级日志**：基于 Zap 的结构化日志，支持日志轮转
- 🔄 **后台任务**：内置的后台任务运行器系统
- 🛡️ **请求过滤**：HTTP 中间件/过滤器系统，处理横切关注点
- 🏷️ **路由元数据**：支持 `.WithMeta()` 为路由附加自定义元数据，实现灵活的路由级别认证、权限检查等
- 🎯 **统一错误处理**：服务错误自动转换为标准化 HTTP 服务错误

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
    // 创建 web 框架实例，自动加载配置
    app := wf.NewWithAutoConfig()

    // 注册简单路由
    app.Get("/", func(c *web.Request) (any, error) {
        return "Hello, World!", nil
    })

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
    app := wf.NewWithAutoConfig()
    app.AddRest(&UserController{})

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    if err := app.Run(ctx); err != nil {
        log.PrintPanic(err)
    }
}
```

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

// 在路由注册时使用
func (c *ApiController) Init(ctx *core.Context) error {
    // 公开路由 - 不需要认证
    ctx.Get("/api/login", loginHandler).WithMeta(SkipAuth())

    // 受保护路由 - 需要认证
    ctx.Get("/api/profile", profileHandler, RequireAuth())

    // 受保护路由，多个元数据，同时需要认证和权限
    ctx.Post("/api/admin/users", createUserHandler, RequireAuth(), RequirePermission("admin:create_user"))

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

### ⚡ 泛型 ORM 操作示例

```go
package main

import (
    "context"
    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/core"
    "github.com/chuccp/go-web-frame/model"
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

func (u *UserModel) Init(db *core.DB, ctx *core.Context) error {
    u.Model = model.NewModel[*User](db, "t_user")
    // 如果表不存在则自动创建
    return u.CreateTable()
}

func main() {
    app := wf.NewWithAutoConfig()
    // 注册模型到 DI 容器
    app.AddModel(&UserModel{})

    // ORM 操作示例
    app.Get("/users", func(c *web.Request) (any, error) {
        userModel := wf.GetModel[*UserModel](c.Context())

        // 链式 API 查询
        users, err := userModel.Query().
            Where("age > ?", 18).
            Order("id desc").
            All()

        return users, err
    })

    app.Post("/users", func(c *web.Request) (any, error) {
        userModel := wf.GetModel[*UserModel](c.Context())

        // 创建用户
        user := &User{Name: "张三", Age: 25}
        err := userModel.Save(user)
        return user.Id, err
    })

    app.Put("/users/:id", func(c *web.Request) (any, error) {
        userModel := wf.GetModel[*UserModel](c.Context())
        id := c.ParamInt("id")

        // 更新用户
        return nil, userModel.Update().
            Where("id = ?", id).
            UpdateColumn("name", "张三（已更新）")
    })

    app.Delete("/users/:id", func(c *web.Request) (any, error) {
        userModel := wf.GetModel[*UserModel](c.Context())
        id := c.ParamInt("id")

        // 删除用户
        return nil, userModel.Delete().
            Where("id = ?", id).
            Delete()
    })

    ctx := context.Background()
    _ = app.Run(ctx)
}
```

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
   - `IComponent`：通过配置初始化的独立组件
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

1. **创建**：使用 `NewWithAutoConfig()` 或 `New(config)` 初始化 `WebFrame`
2. **注册**：添加路由、控制器、模型、服务、组件和运行器
3. **配置**：自定义设置、添加中间件、配置日志
4. **运行**：使用 `Run(ctx)` 启动服务器，或以守护进程模式运行

## 配置示例

带数据库连接池的 YAML 配置示例：

```yaml
server:
  port: 8080
  mode: debug # 或者 release

web:
  db:
    type: mysql
    host: localhost
    port: 3306
    username: root
    password: 你的密码
    database: 你的数据库
    charset: utf8mb4
    # 连接池设置（可选）
    max_open_conns: 100
    max_idle_conns: 10
    conn_max_lifetime: 3600 # 秒

log:
  level: info
  path: ./logs/app.log
```

## 项目结构

```
├── web_frame.go         # 主入口 - WebFrame 工厂方法
├── daemon.go            # Windows/Linux/macOS 守护进程/服务支持
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

### 守护进程模式

```bash
# 以系统守护进程/服务方式运行应用
# 需要实现 AppService 接口
go run your_app.go

# 停止正在运行的守护进程
go run your_app.go -stop
```


## 开发说明

- 框架遵循 Go 约定，使用标准 Go 工具链
- 所有组件都实现 `IService` 接口的 `Init(ctx)` 方法
- 依赖注入通过上下文完成 - 使用 `wf.GetService[T](ctx)` 获取服务
- 连接池有合理的默认值，适合大多数应用
- 开发和小型应用推荐使用 SQLite，生产环境推荐使用 MySQL

## 文档

- [Go 参考文档](https://pkg.go.dev/github.com/chuccp/go-web-frame)
- [示例应用](./example/)
- [CLAUDE.md](./CLAUDE.md) - Claude Code 详细开发者指南

## 贡献

欢迎提交 Issue 和 Pull Request！提交 PR 前请：

1. 运行测试确保一切通过
2. 保持代码风格与项目一致
3. 为新功能添加适当的测试
4. 更新相关文档

## 许可证

MIT License - 详见 [LICENSE](./LICENSE)

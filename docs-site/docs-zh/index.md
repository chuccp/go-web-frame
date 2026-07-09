# Go Web Frame 使用手册

> 用 Go 编写 CRUD 后端，零 ORM 样板代码。定义结构体、嵌入泛型 `Model`，即可获得类型安全的查询、分页和请求上下文传播。

## 什么是 Go Web Frame？

Go Web Frame 是一个集成式后端开发工具包。路由、ORM、缓存、日志、配置等已预先集成，无需单独挑选和组装。

核心亮点是**泛型 Model 层消除 CRUD 样板**。定义实体结构体，嵌入 `Model[T]` 或 `EntryModel[T, PK]`，从数据库到 Handler 全链路类型安全，无 `interface{}`，无需代码生成。

```go
type User struct {
    Id   uint   `gorm:"primaryKey;autoIncrement"`
    Name string
}

type UserModel struct {
    *model.EntryModel[*User, uint]
}

func (m *UserModel) Init(db *db.DB, c *core.Context) error {
    m.EntryModel = model.NewEntryModel[*User, uint](db, "t_user")
    return m.CreateTable()
}

// 这就是你的模型层。然后随处使用：
user, _  := userModel.FindByPK(1)
users, _ := userModel.FindAll()
users, total, _ := userModel.Query().Where("age > ?", 18).Page(&web.Page{PageNo: 1, PageSize: 10})
```

## 30 秒 Hello World

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
    cfg, _ := config.LoadSingleFileConfig("application.yml")
    builder := wf.NewBuilder(cfg)
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

## 框架亮点

### 1. 泛型 Model 实现零模板 CRUD（Model[T]）

**痛点**：传统框架为了解决 CRUD 重复代码，需要开发者在终端频繁运行 CLI 命令去“生成”一大堆 `model.go` 或 `dao.go` 文件，污染项目目录。

**精巧设计**：Go Web Frame 利用 Go 语言的泛型特性，直接在运行时完成结构体与数据库的动态绑定。你只需定义一个基础 struct，增删改查、高级分页（`Page` / `PageForWeb`）直接可用，实现真正的零代码生成。

```go
type UserModel struct {
    *model.EntryModel[*User, uint]
}

// 查询
user, err := userModel.Query().Where("email = ?", email).One()
users, total, err := userModel.Query().Where("status = ?", 1).Page(page)

// 写入
err := userModel.Save(&User{Name: "alice"})
err := userModel.UpdateByPK(&user)
err := userModel.DeleteByPK(1)

// 请求上下文自动透传到数据库
m := userModel.WithContext(req.Ctx())
users, err := m.FindAll()
```

| 类型 | 能力 | 适用场景 |
|---|---|---|
| `Model[T]` | `Save`、`Query()`、`Update()`、`Delete()`、`CreateTable()`、`WithContext()` | 需要完全控制查询构建 |
| `EntryModel[T, PK]` | 继承 `Model[T]` 全部能力 + `FindByPK`、`FindAll`、`DeleteByPK`、`UpdateByPK`、`Page` | 实体有主键，最常见场景 |

### 2. 声明式路由元数据（WithMeta）

**痛点**：在 Gin 中，如果要为不同路由配置权限、限流、免密，通常需要套用大量中间件，配置散乱，难以动态管理。

**精巧设计**：允许在声明路由时通过 `.WithMeta()` 直接给路由绑定元数据。开发者只需在顶层写一个全局 Filter，即可一处校验所有路由的 Meta 标记，保持业务 Handler 的极致干净。

```go
func RequireAuth() web.MetaOption      { return web.WithValue("require_auth", true) }
func SkipAuth() web.MetaOption          { return web.WithValue("skip_auth", true) }
func RequirePermission(p string) web.MetaOption { return web.WithValue("require_permission", p) }

func (c *ApiController) Init(ctx *core.Context) error {
    ctx.Get("/api/login", c.Login).WithMeta(SkipAuth())
    ctx.Get("/api/profile", c.Profile).WithMeta(RequireAuth())
    ctx.Post("/api/admin/users", c.CreateUser).
        WithMeta(RequireAuth(), RequirePermission("admin:create_user"))
    return nil
}
```

```go
// 一个 Filter 处理所有认证逻辑
func (f *AuthFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    if !req.HasMeta(RequireAuth()) || req.HasMeta(SkipAuth()) {
        return fc.Next()
    }
    token := req.GetHeader("Authorization")
    if token == "" {
        return nil, errors.New("unauthorized")
    }
    // 校验 token、检查权限...
    return fc.Next()
}
```

### 3. 显式依赖注入（Builder 模式）

**痛点**：像 Java Spring 那样的隐式全局扫描，在 Go 中容易导致“暗箱操作”，初始化顺序混乱且调试困难。

**精巧设计**：采用显式的 Builder 模式进行组件注册。所有依赖组件都可通过透明上下文的 `GetService[T]`、`GetModel[T]` 动态获取。初始化顺序完全透明、可控。

```go
builder := wf.NewBuilder(cfg)

// 基础设施（最先初始化）
builder.Filter(&cors.Filter{})
builder.Filter(&AuthFilter{})

// 数据层
builder.Model(&UserModel{})
builder.Model(&OrderModel{})

// 业务层
builder.Service(&UserService{})

// HTTP 层
builder.Rest(&UserController{})

// 后台任务
builder.Runner(&CleanupTask{})

app := builder.Build()
app.Run(ctx)
```

```go
// 在任意 Init() 或 Handler 中按类型获取依赖
userModel   := wf.GetModel[*UserModel](ctx)
userService := wf.GetService[*UserService](ctx)
```

### 4. 无缝复用 Gin 生态（GinContext 兼容）

**痛点**：许多自研框架无法使用 Gin 社区插件，导致很多基础功能需要重新手写。

**精巧设计**：框架封装了自己的请求/响应抽象，但通过 `req.GinContext()` 暴露底层 `*gin.Context`。这意味着社区沉淀多年的数百个 `gin-contrib` 中间件（如 CORS、Gzip、Secure 等）可以无缝拿来即用。

```go
import "github.com/gin-contrib/gzip"

type GzipFilter struct{ core.IFilter }

func (f *GzipFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    gzip.Gzip(gzip.DefaultCompression)(req.GinContext())
    return fc.Next()
}
```

框架也已内置常用组件，开箱即用：

```go
builder.Filter(&cors.Filter{})              // 跨域
builder.Service(&ratelimit.RateLimit{})     // 限流
builder.Service(&cache.Cache{})             // 本地缓存
builder.Service(&captcha.Captcha{})         // 滑块验证码
```

## 技术栈

| 层级 | 库 | 作用 |
|---|---|---|
| HTTP | Gin | 路由、中间件链、参数绑定 |
| ORM | GORM | 底层 SQL 驱动、迁移、Join/Preload |
| 配置 | Viper | 多格式、多路径加载 |
| 日志 | Zap | 结构化、分级、文件轮转 |
| JSON | Sonic | 高性能序列化 |
| 缓存 | Otter | 本地内存缓存 |
| Redis | go-redis | 发布订阅、缓存 |
| SQLite | modernc/sqlite | 纯 Go、零 CGO |
| 校验 | go-playground/validator | 结构体标签校验 |
| WebSocket | coder/websocket | 连接升级与读写 |
| 定时任务 | robfig/cron | 表达式定时调度 |

## 快速链接

### 快速开始

- [安装](getting-started/installation.md) - 环境要求与安装方式
- [快速开始](getting-started/quick-start.md) - 创建第一个应用
- [Hello World](getting-started/hello-world.md) - 最简单的示例

### 用户指南

- [路由](guide/routing.md) - HTTP 路由系统
- [控制器](guide/controller.md) - REST 控制器
- [服务](guide/service.md) - 业务逻辑层与依赖注入
- [模型](guide/model.md) - 类型安全 ORM
- [过滤器/中间件](guide/filter.md) - 认证、日志、CORS、限流、路由元数据
- [配置](guide/configuration.md) - 配置管理
- [日志](guide/logging.md) - 结构化日志
- [后台任务](guide/runner.md) - Runner 与定时调度
- [组件](guide/components.md) - 限流、缓存、验证码等内置组件

### 高级与参考

- [数据库高级用法](advanced/database.md) - 事务、模型组、迁移、原生 SQL
- [部署](advanced/deployment.md) - HTTPS、SSL、优雅关闭
- [核心 API](api/core.md) / [Web API](api/web.md) / [模型 API](api/model.md)
- [最佳实践](best-practices.md)
- [更新日志](changelog.md)

## 社区

- [GitHub](https://github.com/chuccp/go-web-frame)
- [问题反馈](https://github.com/chuccp/go-web-frame/issues)

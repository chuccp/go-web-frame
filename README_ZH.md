# Go Web Frame

**用 Go 写 CRUD 后端，零 ORM 样板代码。定义 struct，嵌入泛型 Model ——类型安全的查询、分页、context 传播都包含在内。**

---

**🌐 Language / 语言**
[English](README.md) • [中文](README_ZH.md) • [繁體中文](README_ZH_TW.md) • [日本語](README_JA.md)

---

## 框架概述

Go Web Frame 是一个集成好的后端工具箱。路由、ORM、缓存等组件已预先集成，不需要分别选型再手动组装。

核心是**一个消除 CRUD 样板代码的泛型 Model 层。** 定义实体 struct，嵌入 `Model[T]`，编译器从数据库到 handler 全程检查数据类型。没有 `interface{}`，不用代码生成。

```go
// 定义一次，到处使用
type User struct {
    Id   uint   `gorm:"primaryKey;autoIncrement"`
    Name string
}

type UserModel struct {
    *model.EntryModel[*User, uint]   // ← 所有 CRUD 方法都来自这里
}

func (m *UserModel) Init(db *db.DB, c *core.Context) error {
    m.EntryModel = model.NewEntryModel[*User, uint](db, "t_user")
    return m.CreateTable()
}

// 就这些。现在可以在任何地方用了：
user, _ := userModel.FindByPK(1)          // → *User
users, _ := userModel.FindAll()            // → []*User
users, _ := userModel.Query().Where("age > ?", 18).Order("id desc").All()
total, _ := userModel.Query().Where("status = ?", 1).Count()
users, total, _ := userModel.Page(&web.Page{PageNo: 1, PageSize: 10})
userModel.Update().Where("id = ?", 1).UpdateColumn("name", "新名字")
userModel.DeleteByPK(1)
```

还附带：路由（Gin）、认证过滤器、WebSocket、SSE、CORS、限流、定时任务、校验、缓存、Let's Encrypt HTTPS、多数据库（MySQL / PostgreSQL / SQLite / Redis）——一个 YAML 配好全跑起来。

---

## 快速开始

### 30 秒：Hello World

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
    builder := wf.NewBuilder(config.LoadAutoConfig())
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

### 5 分钟：REST + 数据库

创建 `application.yml`：

```yaml
web:
  db:
    type: sqlite
    path: ./data.db
```

创建 `main.go`：

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

// ── 实体 ──
type User struct {
    Id   uint   `gorm:"primaryKey;autoIncrement"`
    Name string `gorm:"size:255"`
}

// ── Model（零 CRUD 样板代码）──
type UserModel struct {
    *model.EntryModel[*User, uint]
}

func (m *UserModel) Init(database *db.DB, ctx *core.Context) error {
    m.EntryModel = model.NewEntryModel[*User, uint](database, "t_user")
    return m.CreateTable()
}

// ── Controller ──
type UserController struct {
    core.IService
    userModel *UserModel
}

func (c *UserController) Init(ctx *core.Context) error {
    c.userModel = wf.GetModel[*UserModel](ctx)

    ctx.Get("/users", c.List)
    ctx.Get("/users/:id", c.Get)
    ctx.Post("/users", c.Create)
    ctx.Put("/users/:id", c.Update)
    ctx.Delete("/users/:id", c.Delete)
    return nil
}

func (c *UserController) List(req *web.Request) (any, error) {
    return c.userModel.FindAll()
}

func (c *UserController) Get(req *web.Request) (any, error) {
    return c.userModel.FindByPK(req.ParamUint("id"))
}

func (c *UserController) Create(req *web.Request) (any, error) {
    var user User
    if err := req.BindJSON(&user); err != nil {
        return nil, err
    }
    return &user, c.userModel.Save(&user)
}

func (c *UserController) Update(req *web.Request) (any, error) {
    var user User
    if err := req.BindJSON(&user); err != nil {
        return nil, err
    }
    user.Id = req.ParamUint("id")
    return nil, c.userModel.UpdateByPK(&user)
}

func (c *UserController) Delete(req *web.Request) (any, error) {
    return nil, c.userModel.DeleteByPK(req.ParamUint("id"))
}

// ── Main ──
func main() {
    builder := wf.NewBuilder(config.LoadAutoConfig())
    builder.Model(&UserModel{})
    builder.Rest(&UserController{})
    builder.Build().Run(context.Background())
}
```

```bash
curl http://localhost:19009/users              # → [{"Id":1,"Name":"alice"}]
curl http://localhost:19009/users/1            # → {"Id":1,"Name":"alice"}
curl -X POST ... -d '{"Name":"bob"}'          # → {"Id":2,"Name":"bob"}
```

表自动创建，CRUD 全通。不需要写 SQL，不需要写 ORM 样板代码。

---

## 操作数据库

### Model 分两级，按需选择

| 类型 | 提供的方法 | 适用场景 |
|---|---|---|
| `Model[T]` | `Save`、`Query()`、`Update()`、`Delete()`、`CreateTable()`、`WithContext()` | 需要完全控制，用流式构建器 |
| `EntryModel[T, PK]` | Model 全部 + `FindByPK`、`FindAll`、`DeleteByPK`、`UpdateByPK`、`UpdateColumn`、`Page` | 实体有主键（最常见） |

`PK` 可以是 `uint`、`int`、`string` 等满足 `~uint | ~int | ~string` 约束的任意类型。

### 日常查询

```go
m := userModel.WithContext(req.Ctx())   // context 自动传播到所有 DB 调用

// 获取
user, err := m.FindByPK(1)                          // 按主键
users, err := m.FindAll()                            // 全表
user, err := m.Query().Where("email = ?", email).One()
users, err := m.Query().Where("status = ?", 1).Order("id desc").List(100)

// 分页
page := &web.Page{PageNo: 1, PageSize: 10}
users, total, err := m.Query().Where("age > ?", 18).Page(page)
pageAble, err := m.Query().Where("age > ?", 18).PageForWeb(page)  // 返回 PageAble[*User]

// 计数
count, err := m.Query().Where("status = ?", 1).Count()

// 聚合查询（SUM、AVG、GROUP BY、HAVING、DISTINCT）
var total float64
err := m.Aggregate().Select("SUM(amount)").Where("status = ?", 1).Aggregate(&total)

type CatStat struct { Category string; Total float64; Count int }
var stats []CatStat
err := m.Aggregate().
    Select("category, SUM(amount) as total, COUNT(*) as count").
    Group("category").
    Having("SUM(amount) > ?", 200).
    Aggregate(&stats)

// 关联查询（GORM Preload — 自动设置 Model() 子句）
users, err := m.Query().Preload("Orders").Preload("Profile").All()
user, err := m.Query().Where("id = ?", 1).Preload("Orders").One()

// Join（同样自动设置 Model() 子句，v1.0.14）
users, err := m.Query().Joins("JOIN orders ON orders.user_id = t_user.id").All()
```

> **注意**：`Query.One()` 在记录不存在时返回零值和 `nil` 错误——应检查返回值，而非 `gorm.ErrRecordNotFound`。

### 日常写入

```go
// 插入
err := m.Save(&User{Name: "alice"})

// 按主键更新
user.Name = "新名字"
err := m.UpdateByPK(user)

// 更新单列
err := m.UpdateColumn(1, "status", 0)

// 条件更新
err := m.Update().Where("status = ?", 0).UpdateForMap(map[string]any{"status": 1})

// 删除
err := m.DeleteByPK(1)
err := m.Delete().Where("status = ?", -1).Delete()
```

### 对比原生 GORM

```go
// Go Web Frame — 表名绑定一次，ctx 传播一次，分页内置
m := userModel.WithContext(req.Ctx())
user, _ := m.FindByPK(1)
users, total, _ := m.Query().Where("age > ?", 18).Page(&web.Page{PageNo: 1, PageSize: 10})

// 原生 GORM — 通过 GetGorm() 获取 *gorm.DB，直接使用 GORM 生态
var user User
db.GetGorm().WithContext(ctx).Table("t_user").Where("id = ?", 1).First(&user)
var users []User
db.GetGorm().WithContext(ctx).Table("t_user").Where("age > ?", 18).Offset(0).Limit(10).Find(&users)
var total int64
db.GetGorm().WithContext(ctx).Table("t_user").Where("age > ?", 18).Count(&total)
```

`EntryModel` 是辅助封装——减少样板代码，而非限制能力。内置方法不够用时（复杂 JOIN、子查询、窗口函数等），调 `db.GetGorm()` 直接使用 [GORM](https://gorm.io/zh_CN/docs/) 生态，两种方式可自由混用。

原生 GORM 每次都要重复表名、context 和类型。Go Web Frame 在构造时绑定一次，开发时差异巨大——有 20 个 Model 的时候最明显。

---

## 组织项目

### Builder：一个入口注册所有组件

```go
builder := wf.NewBuilder(config.LoadAutoConfig())

// 基础设施（最先初始化）
builder.Filter(&cors.Filter{})        // CORS 头
builder.Filter(&AuthFilter{})         // 认证检查

// 数据层
builder.Model(&UserModel{})
builder.Model(&OrderModel{})
builder.Model(&ProductModel{})

// 业务逻辑
builder.Service(&UserService{})       // 共享业务逻辑
builder.Service(&PaymentService{})

// HTTP 层
builder.Rest(&UserController{})
builder.Rest(&OrderController{})

// 后台任务
builder.Runner(&CleanupTask{})

// 启动
app := builder.Build()
app.Run(ctx)
```

注册顺序决定初始化顺序：Filter → Model → Service → Controller → Runner。依赖通过 `Init()` 自动注入。

### 运行时获取依赖

```go
// 在任何 Init() 或 handler 中，按类型获取：
userModel := wf.GetModel[*UserModel](ctx)
userService := wf.GetService[*UserService](ctx)
authFilter := wf.GetFilter[*AuthFilter](ctx)
cleanupTask := wf.GetRunner[*CleanupTask](ctx)
```

没有字符串 key，不做类型断言。泛型函数直接处理。

### Service 层：共享业务逻辑

多个 Controller 共用逻辑时，抽一个 Service：

```go
type UserService struct {
    core.IService
    userModel *UserModel
}

func (s *UserService) Init(ctx *core.Context) error {
    s.userModel = wf.GetModel[*UserModel](ctx)
    return nil
}

func (s *UserService) GetActiveUsers() ([]*User, error) {
    return s.userModel.Query().Where("status = ?", 1).All()
}

// 注册：
builder.Service(&UserService{})

// 在 Controller 里用：
userService := wf.GetService[*UserService](ctx)
users, _ := userService.GetActiveUsers()
```

### Model Group：多数据库

```go
// 默认数据库
builder.Model(&UserModel{}, &LogModel{})

// 分析系统用独立数据库
analyticsGroup := wf.NewModelGroupBuilder().
    Name("analytics").
    DB(analyticsDB).
    Model(&ReportModel{}).
    AutoCreateTable(true).
    Build()
builder.ModelGroup(analyticsGroup)
```

---

## 认证和中间件

### 路由级元数据（WithMeta）

在路由上声明，Filter 统一处理——不用在每个 handler 里写认证逻辑：

```go
func (c *ApiController) Init(ctx *core.Context) error {
    // 公开接口
    ctx.Get("/api/login", login).WithMeta(SkipAuth())

    // 需要登录
    ctx.Get("/api/profile", profile).WithMeta(RequireAuth())

    // 需要登录 + 特定权限
    ctx.Post("/api/admin/users", createUser).
        WithMeta(RequireAuth(), RequirePermission("admin:create_user"))
    return nil
}
```

```go
// 定义元数据工厂
func RequireAuth() web.MetaOption      { return web.WithValue("require_auth", true) }
func SkipAuth() web.MetaOption          { return web.WithValue("skip_auth", true) }
func RequirePermission(p string) web.MetaOption { return web.WithValue("require_permission", p) }
```

```go
// 一个 Filter 处理所有认证逻辑
func (f *AuthFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    if !req.HasMeta(RequireAuth()) || req.HasMeta(SkipAuth()) {
        return fc.Next()
    }

    token := req.Request().Header.Get("Authorization")
    if token == "" {
        return nil, errors.New("未登录")
    }

    // 验证 token，检查权限...
    return fc.Next()
}
```

### 全局 Filter

通过 `builder.Filter()` 注册的 Filter 对所有请求生效：

```go
// 日志
type LoggingFilter struct { core.IFilter }

func (f *LoggingFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    start := time.Now()
    result, err := fc.Next()
    log.Info("请求", zap.String("path", req.FullPath()), zap.Duration("耗时", time.Since(start)))
    return result, err
}

// CORS（内置，直接注册即可）
builder.Filter(&cors.Filter{})
```

### RestGroup：按路由分组应用 Filter

```go
apiGroup := core.NewRestGroupBuilder().
    ServerConfig(web.DefaultServerConfig()).
    ContextPath("/api/v1").
    Build()

apiGroup.AddFilter(&AuthFilter{})     // 只影响这个 Group 里的路由
apiGroup.AddRest(&UserController{})   // 这个 Controller 的所有路由都需要认证

builder.RestGroup(apiGroup)
```

---

## 常用场景

### 文件上传

```go
ctx.Post("/upload", func(req *web.Request) (any, error) {
    file, header, err := req.File("file")
    if err != nil {
        return nil, err
    }
    defer file.Close()

    dst := "./uploads/" + header.Filename
    if err := web.SaveUploadedFile(header, dst); err != nil {
        return nil, err
    }
    return map[string]string{"path": dst}, nil
})
```

### WebSocket

```go
ctx.WebSocket("/ws", func(ws *web.WebSocket) error {
    stream, err := ws.OpenStream()
    if err != nil {
        return err
    }
    defer stream.Close()
    for {
        typ, msg, err := stream.Read(stream.Context())
        if err != nil {
            return err
        }
        stream.Write(stream.Context(), typ, msg)  // echo
    }
})
```

WebSocket 连接采用延迟初始化——`OpenStream()` 支持通过 `AcceptOptions` 进行配置：

```go
stream, err := ws.OpenStream(
    web.WithOriginPatterns([]string{"example.com"}),
    web.WithCompressionMode(1),  // CompressionNoContextTakeover
)
```

### Server-Sent Events

```go
ctx.SSE("/events", func(stream *web.SSEStream) error {
    defer stream.Close()
    stream.SendRetry(3000)

    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
    for range ticker.C {
        stream.Send("update", fmt.Sprintf("time: %s", time.Now()))
    }
    return nil
})
```

### 静态文件 & SPA

```go
ctx.Static("/assets", "./public")
ctx.Static("/", "./frontend/dist")   // SPA 构建产物
```

```yaml
# 或通过配置——多目录查找，自动 404 回退：
web:
  server:
    locations:
      - view/dist
      - www
    page404: 404.html
```

### 反向代理

```go
ctx.ReverseProxy("/api/legacy", "http://127.0.0.1:8081")
```

### 后台任务

```go
type CleanupTask struct { core.IRunner }

func (t *CleanupTask) Init(ctx *core.Context) error { return nil }

func (t *CleanupTask) Run() error {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    for range ticker.C {
        // 清理过期会话...
    }
    return nil
}

builder.Runner(&CleanupTask{})
```

### 参数校验

```go
type CreateUserInput struct {
    Name  string `validate:"required,min=2,max=50"`
    Email string `validate:"required,email"`
}

func (c *Controller) Create(req *web.Request) (any, error) {
    var input CreateUserInput
    if err := req.BindJSON(&input); err != nil {
        return nil, err
    }

    validator := wf.GetService[*validator.Validator](req.Ctx())
    if err := validator.Validate(input); err != nil {
        return nil, web.NewValidationError().WithError(err)
    }

    // input 已校验，继续业务...
}
```

### 自定义 HTTP 响应

```go
// 返回普通 struct → 自动包装：{"code":200, "data":{...}, "msg":"ok"}
return &User{Id: 1, Name: "alice"}, nil

// 返回字符串 → 纯文本响应
return "ok", nil

// 指定状态码
return web.DataCode(http.StatusCreated, &user), nil

// 返回错误
return nil, errors.New("出错了")

// 返回业务错误码
return nil, web.NewValidationError().WithDetail("名字不能为空")

// 重定向
return web.Redirect("/new-url"), nil

// 文件下载
return &web.FileResponse{Path: "/path/to/report.pdf", FileName: "report.pdf"}, nil
```

---

## 配置文件

一个 YAML 覆盖所有配置。自动从 `./config/`、`~/.<appname>/`、`/etc/<appname>/` 加载。

```yaml
web:
  server:
    port: 8080
    context_path: /api            # 全局路由前缀
    ssl:                          # HTTPS：自动证书或本地证书
      enabled: true
      hosts:                      # Let's Encrypt 自动申请证书
        - example.com
      # certs:                    # 或使用本地证书文件（可选）
      #   - host: example.com
      #     cert-file: /path/to/fullchain.pem
      #     key-file: /path/to/privkey.pem

  db:
    type: mysql                   # mysql | postgres | sqlite
    host: localhost
    port: 3306
    user: root
    password: your_password
    database: mydb
    max_open_conns: 100
    max_idle_conns: 10
    conn_max_lifetime: 3600

  log:
    level: info                   # debug | info | warn | error
    path: ./logs/app.log
    max_size: 500                 # 单文件最大 MB，触发滚动
    max_backups: 3
    max_age: 30                   # 保留天数
    compress: true

  redis:
    addr: localhost:6379
    password: ""
    db: 0
```

PostgreSQL 和 SQLite：

```yaml
# PostgreSQL
db:
  type: postgres
  host: localhost
  port: 5432
  user: postgres
  password: ""
  database: mydb
  sslmode: disable

# SQLite
db:
  type: sqlite
  path: ./data.db
```

支持 JSON、YAML、TOML 格式。

### 自定义数据库驱动

注册自定义数据库驱动，用于框架未内置的数据库（如 SQL Server、ClickHouse 等）。框架底层使用 GORM 作为 ORM 引擎，因此自定义数据库必须有对应的 [GORM 驱动](https://gorm.io/docs/connecting_to_the_database.html)。

```go
// 1. 实现 db.IConfig 接口
type ClickHouseConfig struct {
    Host     string
    Port     int
    Database string
    User     string
    Password string
}

func (c *ClickHouseConfig) Connection() (*db.DB, error) {
    dsn := fmt.Sprintf("clickhouse://%s:%s@%s:%d/%s", c.User, c.Password, c.Host, c.Port, c.Database)
    gormDB, err := gorm.Open(clickhouse.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, err
    }
    return &db.DB{DB: gormDB}, nil
}

// 2. 在应用启动前注册
func main() {
    db.RegisterDB("clickhouse", &ClickHouseConfig{})
    builder := wf.NewBuilder(config.LoadAutoConfig())
    builder.Build().Start()
}
```

然后在配置文件中指定 `type` 为注册的类型名：

```yaml
web:
  db:
    type: clickhouse
    host: localhost
    port: 9000
    database: mydb
```

---

## 技术栈

预集成、生产验证过的组件：

| 层次 | 库 | 作用 |
|---|---|---|
| HTTP | Gin | 路由、中间件链、参数绑定 |
| ORM | GORM | SQL 驱动、迁移、join/preload |
| 配置 | Viper | 多格式、多路径自动加载 |
| 日志 | Zap | 结构化、分级、滚动 |
| JSON | Sonic | 高性能序列化 |
| 缓存 | Otter | 本地内存缓存 |
| 并发 | Conc | 结构化并发池，管理服务生命周期 |
| Redis | go-redis | 发布订阅、缓存 |
| SQLite | modernc/sqlite | 纯 Go，零 CGO |
| 校验 | go-playground/validator | struct tag 校验 |
| WebSocket | coder/websocket | 连接升级 + 读写 |
| 定时任务 | robfig/cron | 表达式调度器 |

---

## 项目结构

```
├── web_frame.go        # Builder、WebFrame 工厂
├── core/               # 接口（IService, IModel, IRest, IRunner, IFilter）、DI context
├── web/                # Request、Response、路由、Filter、SSE、WebSocket、静态文件
├── model/              # Model[T]、EntryModel[T, PK]、Query[T]、Update[T]、Delete[T]
├── db/                 # MySQL、PostgreSQL、SQLite 连接管理
├── redis/              # Redis 客户端封装
├── config/             # Viper 自动加载
├── log/                # Zap + lumberjack 滚动
├── component/          # 独立模块：auth、cache、captcha、cors、localcache、qrcode、ratelimit、schedule、validator
├── util/               # 加密、文件、字符串工具
└── example/            # 可运行的示例
    ├── helloworld/     # 最小应用
    ├── rest/           # REST Controller
    ├── model/          # 泛型 ORM 用法
    ├── filter/         # 认证 + 日志 Filter
    ├── withmeta/       # 路由元数据
    └── background/     # 后台任务
```

---

## 可选模块

重型依赖拆分为独立子模块，按需安装：

<!-- component-badges -->
[![auth](https://img.shields.io/github/v/tag/chuccp/go-web-frame?filter=component/auth/*&label=auth&color=blue)](https://pkg.go.dev/github.com/chuccp/go-web-frame/component/auth)
[![cache](https://img.shields.io/github/v/tag/chuccp/go-web-frame?filter=component/cache/*&label=cache&color=blue)](https://pkg.go.dev/github.com/chuccp/go-web-frame/component/cache)
[![captcha](https://img.shields.io/github/v/tag/chuccp/go-web-frame?filter=component/captcha/*&label=captcha&color=blue)](https://pkg.go.dev/github.com/chuccp/go-web-frame/component/captcha)
[![cors](https://img.shields.io/github/v/tag/chuccp/go-web-frame?filter=component/cors/*&label=cors&color=blue)](https://pkg.go.dev/github.com/chuccp/go-web-frame/component/cors)
[![localcache](https://img.shields.io/github/v/tag/chuccp/go-web-frame?filter=component/localcache/*&label=localcache&color=blue)](https://pkg.go.dev/github.com/chuccp/go-web-frame/component/localcache)
[![qrcode](https://img.shields.io/github/v/tag/chuccp/go-web-frame?filter=component/qrcode/*&label=qrcode&color=blue)](https://pkg.go.dev/github.com/chuccp/go-web-frame/component/qrcode)
[![ratelimit](https://img.shields.io/github/v/tag/chuccp/go-web-frame?filter=component/ratelimit/*&label=ratelimit&color=blue)](https://pkg.go.dev/github.com/chuccp/go-web-frame/component/ratelimit)
[![schedule](https://img.shields.io/github/v/tag/chuccp/go-web-frame?filter=component/schedule/*&label=schedule&color=blue)](https://pkg.go.dev/github.com/chuccp/go-web-frame/component/schedule)
[![validator](https://img.shields.io/github/v/tag/chuccp/go-web-frame?filter=component/validator/*&label=validator&color=blue)](https://pkg.go.dev/github.com/chuccp/go-web-frame/component/validator)
<!-- /component-badges -->

```bash
# 核心框架（组件按需安装）
go get github.com/chuccp/go-web-frame

# 按需安装
go get github.com/chuccp/go-web-frame/component/auth@v1.0.14
go get github.com/chuccp/go-web-frame/component/cache@v1.0.14
go get github.com/chuccp/go-web-frame/component/captcha@v1.0.14
go get github.com/chuccp/go-web-frame/component/cors@v1.0.14
go get github.com/chuccp/go-web-frame/component/localcache@v1.0.14
go get github.com/chuccp/go-web-frame/component/qrcode@v1.0.14
go get github.com/chuccp/go-web-frame/component/ratelimit@v1.0.14
go get github.com/chuccp/go-web-frame/component/schedule@v1.0.14
go get github.com/chuccp/go-web-frame/component/validator@v1.0.14
```

### 使用说明

#### Cache — 高性能内存缓存

```go
import "github.com/chuccp/go-web-frame/component/cache"

// 1. 注册
builder.Service(&cache.Cache{})

// 2. 在控制器或服务中使用
c := core.GetService[*cache.Cache](ctx)
c.Set("user:1", user)
val, ok := c.Get("user:1")
c.SetNX("lock:task", true, 30*time.Second) // 不存在则设置，带过期时间
```

#### Captcha — 滑动拼图验证码

```yaml
# application.yml
captcha:
  code_key: "your-32-character-key-here!!"  # 32 个字符，必填
  code_iv: "your-16-char-iv"                # 16 个字符，必填
```

```go
import "github.com/chuccp/go-web-frame/component/captcha"

// 1. 注册
builder.Service(&captcha.Captcha{})

// 2. 生成验证码（返回给前端）
c := core.GetService[*captcha.Captcha](ctx)
data, _ := c.GetCaptchaData()
// → 返回 SlideCaptchaData，包含 TileImage、MasterImage、ThumbCode

// 3. 校验用户滑动结果
result, ok := c.ValidateThumb(data.ThumbCode, userXOffset)
if ok {
    valid, _ := c.ValidateCode(result.CaptchaCode)
}
```

#### Schedule — Cron 定时任务

```go
import (
    "github.com/chuccp/go-web-frame/component/schedule"
    "github.com/robfig/cron/v3"
)

// 1. 注册为 Runner（注意不是 Service！）
sched := schedule.NewSchedule(cron.WithSeconds())
builder.Runner(sched)

// 2. 添加任务 — Build() 前后均可
sched.AddKeyFunc("cleanup", "0 0 * * *", func(ctx *core.Context) {
    // 每天午夜执行
})
sched.AddKeyFunc("report", "*/30 * * * * *", func(ctx *core.Context) {
    // 每 30 秒执行
})
```

#### QRCode — 二维码生成器

```go
import "github.com/chuccp/go-web-frame/component/qrcode"

// 生成到文件（条纹样式）
qrcode.GenerateStripeQRCode("https://example.com", "qr.png")

// 生成到内存，自定义样式
buf := qrcode.CreateBufferWriteCloser()
qrcode.GenerateQrcode("hello", buf, qrcode.WithCircleShape())
pngBytes := buf.Bytes()

// 自定义颜色
s := qrcode.NewStripeQRCode().WithModuleSize(10)
s.GenerateFile("https://example.com", "custom.png")
```

> QRCode 是独立工具包，无需注册，直接 import 调用即可。

#### RateLimit — 令牌桶限流器

```yaml
# application.yml（可选，以下为默认值）
rate_limit:
  limit: 10     # 每秒生成令牌数
  burst: 5      # 最大突发容量
```

```go
import "github.com/chuccp/go-web-frame/component/ratelimit"

// 1. 注册
builder.Service(&ratelimit.RateLimit{})

// 2. 使用 — 通常在 Filter 中
r := core.GetService[*ratelimit.RateLimit](ctx)
if r.Allow(req.ClientIP()) {
    return fc.Next()
}
return nil, errors.New("请求过于频繁")
```

#### Auth — Token 认证过滤器

泛型认证过滤器，提供 `SignIn`/`SignOut`/`User` 方法——为你的用户类型实现 `Authentication[U]` 接口：

```go
import auth "github.com/chuccp/go-web-frame/component/auth"

// 1. 为用户类型实现 Authentication[U] 接口
type MyUser struct { Id uint; Name string }
type MyAuth struct{}
func (a *MyAuth) Init(ctx *core.Context) error              { return nil }
func (a *MyAuth) SignIn(v any, r *web.Request) (any, error) { /* 设置 token/cookie */ }
func (a *MyAuth) SignOut(r *web.Request) (any, error)       { /* 清除 token */ }
func (a *MyAuth) User(r *web.Request) (*MyUser, error)      { /* 从 token 读取 */ }

// 2. 注册为 Filter
builder.Filter(auth.NewAuthenticationFilter[*MyUser](&MyAuth{}))

// 3. 标记需要登录的路由
ctx.Get("/api/profile", profile).WithMeta(auth.WithLogin())

// 4. 在 handler 中获取当前用户
authFilter := wf.GetFilter[*auth.AuthenticationFilter[*MyUser]](ctx)
user, err := authFilter.User(req)
```

#### CORS — 跨域资源共享过滤器

```go
import "github.com/chuccp/go-web-frame/component/cors"

// 注册 — 自动处理 OPTIONS 预检请求
builder.Filter(cors.NewCorsFilter())
```

默认策略：允许所有来源、支持凭证、允许 `Origin`/`Content-Length`/`Content-Type`/`Authorization` 请求头。

#### LocalCache — 文件本地缓存

将生成的文件（图片、PDF、报表）缓存到磁盘，支持懒生成：

```yaml
# application.yml
local_cache:
  open: true        # 启用磁盘缓存
  path: ./cache     # 缓存目录
```

```go
import "github.com/chuccp/go-web-frame/component/localcache"

builder.Service(&localcache.LocalCache{})

// 使用 — 命中缓存则直接返回，未命中则生成并缓存
lc := core.GetService[*localcache.LocalCache](ctx)
file, err := lc.GetFile(func(v ...any) ([]byte, error) {
    return generateReport(v...)  // 仅在缓存未命中时调用
}, "report", "2026-07")
```

#### Validator — 结构体验证（含自定义规则）

```go
import "github.com/chuccp/go-web-frame/component/validator"

builder.Service(&validator.Validator{})

type RegisterInput struct {
    Name     string `validate:"required,min=2"`
    Phone    string `validate:"mobile"`      // 中国手机号验证
    Password string `validate:"password"`    // 强密码（大写+小写+数字，最少 8 位）
}

v := wf.GetService[*validator.Validator](ctx)
if err := v.Validate(input); err != nil {
    // err 为 *validator.ValidationError，可用 errors.Is 判断具体错误码
    if errors.Is(err, validator.ErrMobileInvalid) { /* ... */ }
    return nil, err
}
```

### 发布子模块

每个子模块通过 tag 前缀独立版本管理：

```bash
# 1. 发布主模块
git tag v1.0.1
git push origin v1.0.1

# 2. 更新各子模块对主模块的依赖
for mod in auth cache captcha cors localcache qrcode ratelimit schedule validator; do
  cd component/$mod
  go get github.com/chuccp/go-web-frame@v1.0.1
  go mod edit -dropreplace github.com/chuccp/go-web-frame
  go mod tidy
  cd ../..
done

# 3. 为每个子模块打 tag（格式：component/<name>/vX.Y.Z）
git tag component/auth/v1.0.1
git tag component/cache/v1.0.1
git tag component/captcha/v1.0.1
git tag component/cors/v1.0.1
git tag component/localcache/v1.0.1
git tag component/qrcode/v1.0.1
git tag component/ratelimit/v1.0.1
git tag component/schedule/v1.0.1
git tag component/validator/v1.0.1

# 4. 推送所有 tag
git push origin --tags
```

Go proxy 会根据 `component/captcha/v1.0.1` tag 前缀自动找到对应目录的 `go.mod`。

---

## 获取帮助

- **[使用手册 (中文)](./docs-site/docs-zh/index.md)** — 完整使用文档
- **[使用手册 (英文)](./docs-site/docs-en/index.md)** — 英文文档
- **[架构设计](./ARCHITECTURE.md)** — 设计决策和内部实现
- **[CLAUDE.md](./CLAUDE.md)** — AI 编程助手指南
- **[Go Reference](https://pkg.go.dev/github.com/chuccp/go-web-frame)** — 包 API 文档

---

## 贡献

欢迎提交 PR。提交前请运行 `go test ./...`。

## License

MIT — 详见 [LICENSE](./LICENSE)

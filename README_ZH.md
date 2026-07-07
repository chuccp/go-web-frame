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
    file_path: ./data.db
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

// 关联查询（GORM Preload）
users, err := m.Query().Preload("Orders").Preload("Profile").All()
user, err := m.Query().Where("id = ?", 1).Preload("Orders").One()

// Join
users, err := m.Query().Joins("JOIN orders ON orders.user_id = t_user.id").All()
```

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

// 原生 GORM — 每次调用都要写表名、ctx、手动分页
var user User
db.WithContext(ctx).Table("t_user").Where("id = ?", 1).First(&user)
var users []User
db.WithContext(ctx).Table("t_user").Where("age > ?", 18).Offset(0).Limit(10).Find(&users)
var total int64
db.WithContext(ctx).Table("t_user").Where("age > ?", 18).Count(&total)
```

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
analyticsDB := db.NewDBFromConfig(analyticsConfig)
analyticsGroup := app.NewModelGroup(analyticsDB, "analytics")
analyticsGroup.AddModel(&ReportModel{})
analyticsGroup.AutoCreateTable(true)
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
    meta := req.HandlerMeta()

    if !meta.Has("require_auth") || meta.Has("skip_auth") {
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
    if err := req.SaveUploadedFile(header, dst); err != nil {
        return nil, err
    }
    return map[string]string{"path": dst}, nil
})
```

### WebSocket

```go
ctx.WebSocket("/ws", func(stream *web.WebSocketStream) error {
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

    validator := wf.GetService[*validator.Validator](req.Context())
    if err := validator.Validate(input); err != nil {
        return nil, web.NewValidationError(err.Error())
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
return nil, web.NewValidationError("名字不能为空")

// 重定向
return web.Redirect("/new-url"), nil

// 文件下载
return &web.File{Path: "/path/to/report.pdf", FileName: "report.pdf"}, nil
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
  file_path: ./data.db
```

支持 JSON、YAML、TOML 格式。

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
├── component/          # cors、cache、rate-limit、captcha、qrcode、cron、validator
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

```bash
# 核心框架（不含 captcha/qrcode/cron/otter）
go get github.com/chuccp/go-web-frame

# 按需安装
go get github.com/chuccp/go-web-frame/component/captcha@v1.0.1
go get github.com/chuccp/go-web-frame/component/schedule@v1.0.1
go get github.com/chuccp/go-web-frame/component/qrcode@v1.0.1
go get github.com/chuccp/go-web-frame/component/cache@v1.0.1
go get github.com/chuccp/go-web-frame/component/ratelimit@v1.0.1
```

### 发布子模块

每个子模块通过 tag 前缀独立版本管理：

```bash
# 1. 发布主模块
git tag v1.0.1
git push origin v1.0.1

# 2. 更新各子模块对主模块的依赖
for mod in captcha schedule qrcode cache ratelimit; do
  cd component/$mod
  go get github.com/chuccp/go-web-frame@v1.0.1
  go mod edit -dropreplace github.com/chuccp/go-web-frame
  go mod tidy
  cd ../..
done

# 3. 为每个子模块打 tag（格式：component/<name>/vX.Y.Z）
git tag component/captcha/v1.0.1
git tag component/schedule/v1.0.1
git tag component/qrcode/v1.0.1
git tag component/cache/v1.0.1
git tag component/ratelimit/v1.0.1

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

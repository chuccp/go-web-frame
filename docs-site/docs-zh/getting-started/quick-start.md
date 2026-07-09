# 快速开始

本指南将帮助你在几分钟内上手 Go Web Frame。

## 创建项目

```bash
mkdir myapp && cd myapp
go mod init myapp
go get github.com/chuccp/go-web-frame
```

## 30 秒：Hello World

创建 `main.go`：

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

    app := builder.Build()
    app.Run(context.Background())
}
```

```bash
go run main.go
# → http://localhost:19009
```

访问 `http://localhost:19009`，看到 `"Hello, World!"` 即表示成功。

## 5 分钟：REST + 数据库

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

// ── 模型（零 CRUD 样板） ──
type UserModel struct {
    *model.EntryModel[*User, uint]
}

func (m *UserModel) Init(database *db.DB, ctx *core.Context) error {
    m.EntryModel = model.NewEntryModel[*User, uint](database, "t_user")
    return m.CreateTable()
}

// ── 控制器 ──
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

// ── 入口 ──
func main() {
    cfg, _ := config.LoadSingleFileConfig("application.yml")
    builder := wf.NewBuilder(cfg)
    builder.Model(&UserModel{})
    builder.Rest(&UserController{})
    builder.Build().Run(context.Background())
}
```

```bash
curl http://localhost:19009/users              # → [{"Id":1,"Name":"alice"}]
curl http://localhost:19009/users/1            # → {"Id":1,"Name":"alice"}
curl -X POST http://localhost:19009/users \
  -H "Content-Type: application/json" \
  -d '{"Name":"bob"}'                         # → {"Id":2,"Name":"bob"}
```

表会自动创建，CRUD 直接可用，无需手写 SQL 和 ORM 装配代码。

## 核心概念

### Builder：一处注册所有组件

```go
cfg, _ := config.LoadSingleFileConfig("application.yml")
builder := wf.NewBuilder(cfg)

builder.Get("/", handler)
builder.Model(&UserModel{})
builder.Service(&UserService{})
builder.Filter(&AuthFilter{})
builder.Runner(&CleanupTask{})

app := builder.Build()
app.Run(ctx)
```

### 统一的 Handler 签名

```go
func handler(req *web.Request) (any, error)
```

- 返回 `any`：自动转换为 JSON 响应
- 返回 `error`：自动转换为错误响应

### 依赖注入

```go
func (s *UserService) Init(ctx *core.Context) error {
    s.userModel = wf.GetModel[*UserModel](ctx)
    return nil
}
```

## 项目结构建议

```
myapp/
├── main.go                 # 应用入口
├── go.mod
├── application.yml         # 配置文件
├── controller/             # REST 控制器
├── service/                # 业务逻辑层
├── model/                  # 数据访问层
├── entity/                 # 领域实体
├── filter/                 # HTTP 过滤器
└── runner/                 # 后台任务
```

## 下一步

- [Hello World](hello-world.md) - 更详细的入门示例
- [路由](../guide/routing.md) - 深入了解路由系统
- [控制器](../guide/controller.md) - 使用 REST 控制器组织代码
- [模型](../guide/model.md) - 类型安全 ORM 完整用法

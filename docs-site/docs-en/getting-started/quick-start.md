# Quick Start

This guide helps you create your first Go Web Frame application in minutes.

## Create Project

```bash
mkdir myapp && cd myapp
go mod init myapp
go get github.com/chuccp/go-web-frame
```

## 30 Seconds: Hello World

Create `application.yml`:

```yaml
web:
  server:
    port: 19009
    host: 0.0.0.0
```

Create `main.go`:

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

Visit `http://localhost:19009` and you should see `"Hello, World!"`.

## 5 Minutes: REST + Database

Add SQLite config to `application.yml`:

```yaml
web:
  server:
    port: 19009
    host: 0.0.0.0
  db:
    type: sqlite
    path: ./data.db
```

Create `main.go`:

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

// ──── Entity ────
type User struct {
    Id   uint   `gorm:"primaryKey;autoIncrement"`
    Name string `gorm:"size:255"`
}

// ──── Model (zero CRUD boilerplate) ────
type UserModel struct {
    *model.EntryModel[*User, uint]
}

func (m *UserModel) Init(database *db.DB, ctx *core.Context) error {
    m.EntryModel = model.NewEntryModel[*User, uint](database, "t_user")
    return m.CreateTable()
}

// ──── Controller ────
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

// ──── Entry ────
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

The table is auto-created and CRUD works out of the box — no manual SQL or ORM wiring.

## 10 Minutes: Request Parameters

### Path Parameters

```go
builder.Get("/users/:id", func(req *web.Request) (any, error) {
    id := req.Param("id")         // string "123"
    idInt := req.ParamInt("id")   // int 123
    idUint := req.ParamUint("id") // uint 123
    return map[string]any{"id": id, "idInt": idInt, "idUint": idUint}, nil
})
```

Visiting `/users/123`:

```json
{
  "id": "123",
  "idInt": 123,
  "idUint": 123
}
```

### Query Parameters

```go
builder.Get("/search", func(req *web.Request) (any, error) {
    keyword := req.Query("q")               // /search?q=go
    page := req.DefaultQuery("page", "1")   // with default
    return map[string]any{"keyword": keyword, "page": page}, nil
})
```

### JSON Request Body

```go
builder.Post("/users", func(req *web.Request) (any, error) {
    var user struct {
        Name  string `json:"name"`
        Email string `json:"email"`
    }
    if err := req.BindJSON(&user); err != nil {
        return nil, err
    }
    return map[string]any{
        "name":  user.Name,
        "email": user.Email,
    }, nil
})
```

Send a POST request:

```bash
curl -X POST http://localhost:19009/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com"}'
```

## Core Concepts

### Builder: Register Everything in One Place

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

### Unified Handler Signature

```go
func handler(req *web.Request) (any, error)
```

- Return `any`: automatically converted to JSON response
- Return `error`: automatically converted to error response

### Dependency Injection

```go
func (s *UserService) Init(ctx *core.Context) error {
    s.userModel = wf.GetModel[*UserModel](ctx)
    return nil
}
```

## Recommended Project Structure

```
myapp/
├── main.go                 # Application entry point
├── go.mod
├── application.yml         # Configuration file
├── controller/             # REST controllers
├── service/                # Business logic layer
├── model/                  # Data access layer
├── entity/                 # Domain entities
├── filter/                 # HTTP filters
└── runner/                 # Background tasks
```

## Next Steps

- [Routing](../guide/routing.md) - Learn about the routing system
- [Controller](../guide/controller.md) - Organize code with REST controllers
- [Model](../guide/model.md) - Type-safe ORM usage
- [Service](../guide/service.md) - Business logic and dependency injection

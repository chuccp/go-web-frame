# Controller

REST controllers in Go Web Frame.

## Define Controller

```go
type UserController struct {
    core.IService
}

func (u *UserController) Init(ctx *core.Context) error {
    ctx.Get("/users", u.List)
    ctx.Post("/users", u.Create)
    return nil
}
```

## Register Controller

```go
cfg, _ := config.LoadSingleFileConfig("application.yml")
builder := wf.NewBuilder(cfg)
builder.Rest(&UserController{})
```

## Request Handling

```go
func (u *UserController) List(req *web.Request) (any, error) {
    // Path params
    id := req.Param("id")
    
    // Query params
    page := req.Query("page")
    
    // JSON body
    var data struct {
        Name string `json:"name"`
    }
    if err := req.BindJSON(&data); err != nil {
        return nil, err
    }
    
    return data, nil
}
```

> **Note:** Get models and services in `Init()`, not in the handler. See [Dependency Injection](service.md).

## Request Methods

| Method | Description |
|--------|-------------|
| `Param(key)` | Path parameter (string) |
| `ParamInt(key)` | Path parameter (int) |
| `ParamUint(key)` | Path parameter (uint) |
| `Query(key)` | Query parameter (string) |
| `DefaultQuery(key, default)` | Query parameter with default |
| `BindJSON(v)` | Bind JSON body |
| `GetHeader(key)` | Get header |
| `Cookie()` | Get cookie helper |
| `Ctx()` | Get request `context.Context` |
| `ClientIP()` | Get client IP |
| `Page()` | Get pagination info |
| `GinContext()` | Get underlying `*gin.Context` |

## Response Types

```go
// JSON (default)
return map[string]any{"id": 1}, nil

// String
return "hello", nil

// File download
return &web.FileResponse{Path: "/path/file.pdf", FileName: "doc.pdf"}, nil

// Redirect
return web.Redirect("/new-url"), nil

// Error
return nil, errors.New("something wrong")
```

## File Upload

```go
func (u *UserController) Upload(c *web.Request) (any, error) {
    file, header, err := c.File("file")  // form field name
    if err != nil {
        return nil, err
    }
    defer file.Close()

    dst := "./uploads/" + header.Filename
    if err := web.SaveUploadedFile(header, dst); err != nil {
        return nil, err
    }
    return map[string]string{"path": dst}, nil
}
```

## Dependency Injection

Controllers can get other components via the `ctx` parameter in `Init`:

```go
type UserController struct {
    core.IService
    userService *UserService
}

func (c *UserController) Init(ctx *core.Context) error {
    // Get service
    c.userService = wf.GetService[*UserService](ctx)

    // Register routes
    ctx.Get("/users", c.List)
    ctx.Get("/users/:id", c.Get)

    return nil
}

func (c *UserController) List(req *web.Request) (any, error) {
    users, err := c.userService.GetAllUsers()
    if err != nil {
        return nil, err
    }
    return users, nil
}
```

## Next Steps

- [Model](model.md) - Type-safe ORM
- [Service](service.md) - Business logic layer
- [Filter/Middleware](filter.md) - Cross-cutting concerns
- [Runner](runner.md) - Background tasks
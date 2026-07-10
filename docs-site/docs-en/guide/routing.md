# Routing

Go Web Frame provides a powerful and flexible routing system built on Gin.

## Basic Routes

### HTTP Methods

```go
// Register routes using the Builder
builder.Get("/users", handler)        // GET
builder.Post("/users", handler)       // POST
builder.Put("/users/:id", handler)   // PUT
builder.Delete("/users/:id", handler) // DELETE
builder.Any("/api", handler)          // Match all HTTP methods
```

### Register Routes in a Controller

```go
type UserController struct {
    core.IService
}

func (c *UserController) Init(context *core.Context) error {
    context.Get("/users", c.List)
    context.Get("/users/:id", c.Get)
    context.Post("/users", c.Create)
    context.Put("/users/:id", c.Update)
    context.Delete("/users/:id", c.Delete)
    return nil
}
```

## Path Parameters

### Basic Parameters

```go
builder.Get("/users/:id", func(req *web.Request) (any, error) {
    id := req.Param("id")
    return "User ID: " + id, nil
})
```

### Multiple Parameters

```go
builder.Get("/users/:userId/posts/:postId", func(req *web.Request) (any, error) {
    userId := req.Param("userId")
    postId := req.Param("postId")
    return map[string]any{
        "userId": userId,
        "postId": postId,
    }, nil
})
```

## Parameter Type Conversion

`web.Request` provides type-safe methods:

```go
idStr := req.Param("id")         // "123" (string)
idInt := req.ParamInt("id")      // 123 (int)
idUint := req.ParamUint("id")    // 123 (uint)
```

## Query Parameters

```go
builder.Get("/search", func(req *web.Request) (any, error) {
    keyword := req.Query("q")
    page := req.DefaultQuery("page", "1")
    return map[string]any{
        "keyword": keyword,
        "page":    page,
    }, nil
})
```

## Route Groups

Use `wf.NewRestGroupBuilder()` to create route groups with independent port, prefix, filters, and converters:

### Basic Route Group

```go
restGroup := wf.NewRestGroupBuilder().
    Rest(&UserController{}).
    Port(8081).
    Build()
builder.RestGroup(restGroup)
```

### Route Prefix (ContextPath)

```go
restGroup := wf.NewRestGroupBuilder().
    ContextPath("/api/v1").
    Rest(&UserController{}).
    Port(8081).
    Build()
```

### Route Group Filters

```go
import auth2 "github.com/chuccp/go-web-frame/component/auth"

restGroup := wf.NewRestGroupBuilder().
    Rest(&UserController{}).
    Filter(auth2.NewAuthenticationFilter(&MyAuthenticator{})).
    Port(8081).
    Build()
```

### Route Group Converter

```go
restGroup := wf.NewRestGroupBuilder().
    Rest(&UserController{}).
    Converter(&APIConverter{}).
    Port(8081).
    Build()
```

### RestGroupBuilder API

| Method | Description |
|------|------|
| `Rest(rest ...IRest)` | Register REST controllers |
| `Port(port int)` | Set listen port |
| `ContextPath(path string)` | Set route prefix |
| `Filter(filters ...IFilter)` | Set filters |
| `Converter(converter IConverter)` | Set response converter |
| `ServerConfig(config *web.ServerConfig)` | Set server config (SSL, timeout, etc.) |
| `Build()` | Build the route group |

## Route Metadata (WithMeta)

Tag routes declaratively with `.WithMeta()`, handled in a global filter:

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

See the [Filter](filter.md) documentation for route metadata details.

## Next Steps

- [Controller](controller.md) - Organize code with REST controllers
- [Filter](filter.md) - Add cross-cutting concerns
- [WebSocket & SSE](../advanced/websocket-sse.md) - Real-time communication
- [Static Files & Proxy](../advanced/static-proxy.md) - Static files and reverse proxy

# Routing

HTTP routing system in Go Web Frame.

## Basic Routes

```go
builder.Get("/users", handler)
builder.Post("/users", handler)
builder.Put("/users/:id", handler)
builder.Delete("/users/:id", handler)
```

## Path Parameters

```go
builder.Get("/users/:id", func(c *web.Request) (any, error) {
    id := c.Param("id")        // string
    idInt := c.ParamInt("id")  // int
    return map[string]any{"id": id}, nil
})
```

## Query Parameters

```go
builder.Get("/search", func(c *web.Request) (any, error) {
    q := c.Query("q")
    page := c.QueryInt("page")
    return map[string]any{"query": q, "page": page}, nil
})
```

## REST Controller

```go
type UserController struct {
    core.IService
}

func (u *UserController) Init(ctx *core.Context) error {
    ctx.Get("/users", u.List)
    ctx.Get("/users/:id", u.Get)
    ctx.Post("/users", u.Create)
    ctx.Put("/users/:id", u.Update)
    ctx.Delete("/users/:id", u.Delete)
    return nil
}

builder.Rest(&UserController{})
```

## Route Metadata (WithMeta)

```go
builder.Get("/admin", handler).WithMeta(RequireAuth(), RequirePermission("admin"))
```

## Static Files

```go
ctx.Static("/assets", "./public")
```

## Reverse Proxy

```go
ctx.ReverseProxy("/api", "http://backend:8081")
```

## WebSocket

```go
ctx.WebSocket("/ws", func(stream *web.WebSocketStream) error {
    defer stream.Close()
    for {
        typ, data, err := stream.Read(stream.Context())
        if err != nil {
            return nil
        }
        stream.Write(stream.Context(), typ, data) // echo
    }
})
```

## SSE (Server-Sent Events)

```go
ctx.SSE("/events", func(stream *web.SSEStream) error {
    defer stream.Close()
    stream.Send("update", "data")
    return nil
})
```
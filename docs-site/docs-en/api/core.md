# Core API

Core abstractions and DI container.

## WebFrame

Main application instance.

```go
// Create
builder := wf.NewBuilder(config)
app := builder.Build()

// Run
app.Run(ctx)
```

## Builder

Explicit registration pattern.

```go
builder := wf.NewBuilder(config)

builder.Get("/path", handler)
builder.Post("/path", handler)
builder.Rest(&Controller{})
builder.Model(&Model{})
builder.Service(&Service{})
builder.Filter(&Filter{})
builder.Runner(&Runner{})

app := builder.Build()
```

## Context

DI container.

```go
// Get services
userService := wf.GetService[*UserService](ctx)
userModel := wf.GetModel[*UserModel](ctx)

// Get config
port := ctx.Config().GetInt("web.server.port")

// Register routes
ctx.Get("/users", handler)
ctx.Post("/users", handler)
```

## Interfaces

```go
type IService interface {
    Init(ctx *Context) error
}

type IModel interface {
    IService
    Init(db *DB, ctx *Context) error
}

type IRest interface {
    IService
}

type IFilter interface {
    IService
    Handle(fc FilterChain, req *Request) (any, error)
}

type IRunner interface {
    IService
    Run() error
}
```
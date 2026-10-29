# Best Practices

Recommended patterns and practices.

## Project Structure

```
myapp/
├── main.go
├── config/
│   └── config.yaml
├── controller/
│   ├── user_controller.go
│   └── order_controller.go
├── service/
│   ├── user_service.go
│   └── order_service.go
├── model/
│   ├── user.go
│   ├── user_model.go
│   ├── order.go
│   └── order_model.go
├── filter/
│   └── auth_filter.go
└── runner/
    └── cleanup_runner.go
```

## Layer Separation

| Layer | Responsibility |
|-------|----------------|
| Controller | HTTP handling, request/response |
| Service | Business logic, orchestration |
| Model | Data access, CRUD |
| Filter | Cross-cutting concerns |

## Dependency Direction

```
Controller → Service → Model
```

- Controller calls Service
- Service calls Model
- Model should not call Controller or Service

## Error Handling

```go
// Service layer
func (s *UserService) GetUser(id uint) (*User, error) {
    user, err := s.userModel.FindByPK(id)
    if err != nil {
        return nil, fmt.Errorf("user not found: %w", err)
    }
    return user, nil
}

// Controller layer
func (c *UserController) Get(req *web.Request) (any, error) {
    user, err := userService.GetUser(id)
    if err != nil {
        return nil, err  // framework handles error response
    }
    return user, nil
}
```

## Authentication Pattern

Use WithMeta for declarative auth:

```go
// Define
func RequireAuth() web.MetaOption {
    return web.WithValue("require_auth", true)
}

// Declare on routes
builder.Get("/profile", handler).WithMeta(RequireAuth())

// Handle in filter
func (f *AuthFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    if req.HasMeta(RequireAuth()) {
        // verify token
    }
    return fc.Next()
}
```

## Templates and Static Assets: Hot Reload in Development, Embedded in Release

Keep templates and static files on disk while developing and embed them in release
binaries. A build tag picks the declarations; `Reload(true)` enables hot reload and
`Reload(false)` serves the embedded copies:

```go
// assets_dev.go
//go:build !release

package webui

import "github.com/chuccp/go-web-frame/web"

// Development: reread the working tree on every request.
func templateFiles() *web.EmbedFileSystem {
    return web.EmbedFs("webui/templates").Reload(true)
}

func staticFiles() *web.EmbedFileSystem {
    return web.EmbedFs("webui/static").Reload(true)
}
```

```go
// assets_release.go
//go:build release

package webui

import (
    "embed"

    "github.com/chuccp/go-web-frame/web"
)

//go:embed templates/*.html
var templates embed.FS

//go:embed static
var static embed.FS

// Release: serve the embedded copies and never touch the disk.
func templateFiles() *web.EmbedFileSystem {
    return web.EmbedFs("templates", templates).Reload(false)
}

func staticFiles() *web.EmbedFileSystem {
    return web.EmbedFs("static", static).Reload(false)
}
```

```go
templateOption, err := web.LoadTemplateOption(
    templateFiles(),
    web.WithLayouts("layout", "bare"),
    web.WithPartials("partials"),
)
ctx.AddEmbedFs("/static", staticFiles())
```

```bash
go run main.go              # hot reload from disk
go build -tags release .    # templates and static assets embedded
```

If the development directory differs from the working directory, pass it with
`ReloadPath` — for example
`web.EmbedFs("templates", templates).Reload(true).ReloadPath("webui/templates")`.

## Testing

```go
func TestUserService(t *testing.T) {
    // Setup
    cfg, _ := config.LoadSingleFileConfig("test.yml")
    builder := wf.NewBuilder(cfg)
    builder.Model(&UserModel{})
    builder.Service(&UserService{})
    app := builder.Build()
    
    // Use Test() to initialize without starting HTTP server
    err := app.Test(func(ctx *core.Context) error {
        userService := wf.GetService[*UserService](ctx)
        
        // Test
        user, err := userService.GetUser(1)
        assert.NoError(t, err)
        assert.Equal(t, "alice", user.Name)
        return nil
    })
    assert.NoError(t, err)
}
```
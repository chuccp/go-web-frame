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
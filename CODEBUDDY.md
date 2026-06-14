# CODEBUDDY.md

This file provides guidance to CodeBuddy Code when working with code in this repository.

## Common Development Commands

### Build and Run
```bash
# Run example applications
go run example/helloworld/helloworld.go
go run example/rest/rest.go
go run example/model/model.go

# Build the framework (library only, no main package)
go build ./...
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests in a specific package
go test ./core
go test ./web
go test ./model

# Run tests with verbose output
go test -v ./...

# Run a specific test case
go test -v ./core -run TestSpecificFunction

# Run tests with coverage
go test -cover ./...
```

### Code Formatting
```bash
# Format all code
gofmt -w ./

# Or use go fmt
go fmt ./...
```

### Dependency Management
```bash
# Add a new dependency
go get github.com/example/package

# Update dependencies
go get -u ./...

# Tidy up go.mod and go.sum
go mod tidy
```

## High-Level Architecture

This is a Go web framework built on Gin, providing structured web application development with dependency injection and type-safe ORM.

### Core Architecture Flow

```
HTTP Request → Gin Engine → Filter Chain → Handler → Converter → HTTP Response
                           ↓
                    Dependency Injection (Context)
                           ↓
              Models | Services | Components | Runners
```

### Key Directories

| Directory | Purpose |
|-----------|---------|
| `core/` | Core abstractions and interfaces (IService, IModel, IRest, IFilter, IRunner) |
| `web/` | HTTP layer built on Gin - routing, request handling, response conversion |
| `db/` | Database abstraction using GORM (MySQL, SQLite, PostgreSQL) |
| `model/` | Type-safe generic base models with zero-boilerplate CRUD |
| `config/` | Configuration management with Viper (JSON, YAML, TOML) |
| `log/` | Structured logging with Zap |
| `component/` | Reusable components (cache, rate limiting, captcha, cron, validation) |
| `example/` | Example applications demonstrating framework usage |

### Component Initialization Order

Components initialize in a specific order to ensure dependencies are available:

1. `IModel.Init(db, ctx)` - Data access layer
2. `IService.Init(ctx)` - Business logic services (including components)
4. `IFilter.Init(ctx)` - HTTP middleware filters
5. `IRunner.Run(ctx)` - Background tasks

### Dependency Injection

The framework uses `core.Context` as a DI container. Components are registered during initialization and retrieved using type-safe generic getters:

```go
// Registration happens automatically via Init()
// Retrieval uses generic getters:
userService := wf.GetService[*UserService](ctx)
userModel := wf.GetModel[*UserModel](ctx)
```

### Request Handling Pattern

```
1. Route Registration: ctx.Get("/users/:id", handler)
2. Request Receives: *web.Request (abstraction over Gin context)
3. Handler Returns: (any, error)
4. Converter Transforms: result → HTTP response (JSON by default)
```

### Configuration Auto-Loading

Configuration auto-loads from multiple locations (later overrides earlier):

1. `./config/` - Development (project-local)
2. `~/.<appname>/` - User-specific settings
3. `/etc/<appname>/` - System-wide (production)

Supported formats: JSON, YAML, TOML

## Key Entry Points

- `web_frame.go` - Main package with factory methods (`NewWithAutoConfig()`, `New()`)
- `core/context.go` - Dependency injection context
- `core/server.go` - Server managing REST groups and runners
- `web/handles.go` - Route registration and handler storage
- `web/request.go` - Request abstraction with helper methods

## Important Patterns

### Model Definition

Use `model.EntryModel[T, PK]` for entities with a primary key to get built-in CRUD methods. `PK` is the primary key type (`uint`, `int`, `string`, etc.):

```go
type UserModel struct {
    *model.EntryModel[*User, uint]
}
```

### REST Controller

Implement `core.IRest` interface and embed `core.IService` for structured REST APIs:

```go
type UserController struct {
    core.IService
}
```

### Context Propagation

Use `WithContext(ctx)` to propagate request cancellation, timeouts, and tracing to database operations:

```go
func (c *UserController) GetUsers(req *web.Request) (any, error) {
    // Inject request context once — all subsequent operations carry it
    return c.userModel.WithContext(req.Ctx()).FindAll()
}

// Custom timeout
ctx, cancel := context.WithTimeout(req.Ctx(), 5*time.Second)
defer cancel()
users, err := userModel.WithContext(ctx).FindAll()
```

`req.Ctx()` returns the per-request `context.Context`, auto-cancelled when the request completes.

### Service Layer

Get dependencies via context in `Init()` method:

```go
func (s *UserService) Init(ctx *core.Context) error {
    s.userModel = wf.GetModel[*UserModel](ctx)
    return nil
}
```

## Additional Documentation

For detailed information, refer to:

- **CLAUDE.md** - Comprehensive development guide with examples
- **ARCHITECTURE.md** - Detailed architecture design and component interactions
- **BEST_PRACTICES.md** - Recommended patterns and practices
- **README.md** - Project overview and quick start guide

## Tech Stack

- **Gin** - High-performance HTTP web framework
- **GORM** - ORM library with multi-database support
- **Viper** - Configuration management
- **Zap** - Structured logging
- **go-redis** - Redis client
- **modernc.org/sqlite** - Pure Go SQLite (no CGO required)

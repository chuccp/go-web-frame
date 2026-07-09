# Architecture Design

This document describes the internal architecture and design decisions of Go Web Frame.

## Design Philosophy

### 1. Convention over Configuration
- Auto-loading configuration from standard locations
- Sensible defaults for all components
- Zero-setup for common use cases

### 2. Separation of Concerns
- **Registration**: Components register themselves (Context layer)
- **Storage**: Unified route registry (RouteTree layer)
- **Execution**: Actual HTTP handling (HttpServer layer)

### 3. Dependency Injection
- All components receive dependencies through Context
- No global state or singletons
- Easy testing with mock dependencies

## Core Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Application                              │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                    WebFrame (Facade)                      │   │
│  │  NewBuilder() | Rest() | Model() | Build() | Run()       │   │
│  └──────────────────────────────────────────────────────────┘   │
│                              │                                   │
│                              ▼                                   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                    Core Context                           │   │
│  │  - Manages DI container                                  │   │
│  │  - Provides config access                                │   │
│  │  - Route registration (Context.Get/Post/etc)             │   │
│  └──────────────────────────────────────────────────────────┘   │
│                              │                                   │
│                              ▼                                   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                    HandlerConfig                          │   │
│  │  - Routes (Route storage)                                │   │
│  │  - Filters (middleware chain)                            │   │
│  │  - Converter (response transformation)                   │   │
│  └──────────────────────────────────────────────────────────┘   │
│                              │                                   │
│                              ▼                                   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                    HttpServer                             │   │
│  │  - Process Routes → Gin routes                           │   │
│  │  - Handle static files, reverse proxy                    │   │
│  │  - Graceful shutdown                                     │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

## Request Flow

```
HTTP Request
     │
     ▼
┌─────────┐
│  Gin    │ ← Engine handles raw HTTP
└────┬────┘
     │
     ▼
┌─────────────────────────────────────────────────────────┐
│                    Filter Chain                          │
│  ┌─────────┐   ┌─────────┐   ┌─────────┐               │
│  │ Filter1 │ → │ Filter2 │ → │ FilterN │ → Handler     │
│  └─────────┘   └─────────┘   └─────────┘               │
└─────────────────────────────────────────────────────────┘
     │
     ▼
┌─────────────────┐
│    Handler      │ ← Returns (any, error)
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   Converter     │ ← Transforms result to HTTP response
└────────┬────────┘
         │
         ▼
   HTTP Response
```

## Component Model

### Interface Hierarchy

```
IService (base)
    ├── IModel    (data access)
    ├── IRest     (REST controllers)
    ├── IRunner   (background tasks)
    └── IFilter   (HTTP middleware)
```

### Initialization Order

```
1. IModel.Init(db, ctx)          ← First, DB ready
2. IService.Init(ctx)            ← Second, services (including components) can use models
3. IFilter.Init(ctx)             ← Third, filters can use services
4. IRest.Init(ctx)               ← Fourth, controllers register routes
5. IRunner.Run()                 ← Last, background tasks start
```

## Route Registry Design

### Unified Route Storage

All routes (API, static files, reverse proxy) are stored in the `HandlerConfig`:

```go
type Route struct {
    path     string
    method   string
    handlers []HandlerFunc
    meta     map[string]any  // route metadata for filters
}

type HandlerMeta struct {
    keys   []string
    values map[string]any
}
```

### Route Registration

```go
// In Context (used by controllers)
ctx.Get("/users", handler)
ctx.Post("/users", handler).WithMeta(RequireAuth())

// In Builder (top-level routes)
builder.Get("/health", handler)
builder.Post("/api/login", handler)
```

### Processing in HttpServer

```go
func (s *HttpServer) Handle(config *HandlerConfig) {
    for _, route := range config.Routes() {
        s.engine.Handle(route.method, route.path, route.handlers...)
    }
}
```

**Benefits:**
- Single source of truth for all routes
- Consistent registration API
- Route metadata for filter-based auth/permissions

## Dependency Injection

### Context as DI Container

```go
type Context struct {
    config     IConfig
    modelMap   map[string]IModel
    serviceMap map[string]IService
    runnerMap  map[string]IRunner
    // ...
}
```

### Type-Safe Retrieval

```go
// Generic getter with type assertion
func GetService[T IService](c *Context) T {
    return c.GetService(func(s IService) bool {
        _, ok := s.(T)
        return ok
    }).(T)
}

// Usage
userService := wf.GetService[*UserService](ctx)
userModel := wf.GetModel[*UserModel](ctx)
```

## Error Handling Strategy

### Unified Error Conversion

```go
// In core/interface.go
type IConverter interface {
    Init(ctx *Context) error
}

// In web/converter.go
type Converter interface {
    Request(fc FilterChain, req *Request)
}

// Default: error → JSON response with error message
type DefaultConverter struct{}

func (c *DefaultConverter) Request(fc FilterChain, req *Request) {
    result, err := fc.Next()
    if err != nil {
        req.Response().AbortWithError(err)
        return
    }
    req.Response().AbortWithStatusJSON(200, result)
}
```

### Benefits
- Handlers focus on business logic
- Consistent error responses
- Custom converters for different API styles

## Configuration Strategy

### Multi-Location Auto-Discovery

```
./config/           ← Development, project-local
~/.<appname>/       ← User-specific settings
/etc/<appname>/     ← System-wide, production
```

### Priority: Later locations override earlier

```go
// Auto-config loads from standard locations
cfg := config.LoadAutoConfig()
builder := wf.NewBuilder(cfg)

// Or explicit single file
cfg, _ := config.LoadSingleFileConfig("application.yml")
builder := wf.NewBuilder(cfg)
```

## Context Propagation

### Request Context → Database

The framework supports propagating HTTP request context to database operations for cancellation, timeouts, and tracing:

```
HTTP Request (context.Context)
     │
     ▼
web.Request.Ctx()                      ← auto-created per request
     │
     ▼
model.WithContext(req.Ctx())           ← shallow copy, concurrent-safe
     │
     ▼
db.DB.WithContext(ctx)                 ← gorm.DB.WithContext(ctx)
     │
     ▼
All GORM operations carry context      ← cancellation, timeout, trace propagation
```

Key design decisions:
- `WithContext(ctx)` returns a shallow copy — the original model is never mutated, safe for concurrent use across requests
- `Model[T]` and `EntryModel[T, PK]` return concrete types, preserving type-safe chaining
- `Query[T]`, `Update[T]`, `Delete[T]` support chainable `WithContext(ctx)` for builder-level injection
- `model` package depends only on `context.Context` (stdlib), not on `web.Request` — clean separation

## Database Abstraction

### Connection Pool Configuration

```yaml
db:
  type: mysql
  host: localhost
  port: 3306
  max_open_conns: 100      # Connection pool limits
  max_idle_conns: 10
  conn_max_lifetime: 3600  # Seconds
```

### Transaction Support

```go
tx := ctx.GetTransaction()
err := tx.Exec(func(tx *db.DB) error {
    // All operations within transaction
    userModel := wf.GetReNewModel[*UserModel](tx, ctx)
    if err := userModel.Save(user); err != nil {
        return err  // Will rollback
    }
    return nil  // Will commit
})
```

## Static Files & SPA Support

### Multi-Location Lookup

```yaml
server:
  locations:
    - view/dist    # Frontend build output
    - www          # Static assets
  page404: 404.html  # SPA fallback
```

### Request Flow

```
Request /dashboard
    │
    ▼
Check file exists in view/dist/dashboard?
    │ No
    ▼
Check file exists in www/dashboard?
    │ No
    ▼
Accept: text/html?
    │ Yes
    ▼
Return 404.html (SPA entry point)
```

## HTTPS & Auto Certificate

### Let's Encrypt Integration

```yaml
server:
  port: 443
  ssl:
    enabled: true
    hosts:
      - example.com
      - api.example.com
```

### Automatic Flow

```
1. Server starts on port 443
2. First request to example.com
3. Auto-request certificate from Let's Encrypt
4. Cache certificate in ./certs/
5. Serve HTTPS with cached certificate
```

## Extension Points

### Adding New Route Type

1. Add field to `HandlerInfo`
2. Add `IsXxx()` method
3. Add `AddXxx()` to `Handles`
4. Add `handleXxx()` to `HttpServer`

### Adding New Component

1. Implement `IService`
2. Register with `builder.Service()` or `ctx.AddService()`
3. Retrieve via `wf.GetService[T](ctx)`

### Custom Converter

```go
type GraphQLConverter struct {
    core.IConverter
}

func (c *GraphQLConverter) Init(ctx *core.Context) error {
    return nil
}

func (c *GraphQLConverter) Request(fc web.FilterChain, req *web.Request) {
    // Custom GraphQL response formatting
}
```

## Performance Considerations

### Route Matching
- Gin's radix tree for O(k) lookup where k = path length
- No reflection in hot path

### Memory
- HandlerInfo pooling (future optimization)
- Minimal allocations per request

### Concurrency
- All components are goroutine-safe
- Context uses RWMutex for concurrent access

## Testing Strategy

### Unit Tests
- Each package independently testable
- Mock interfaces for external dependencies

### Integration Tests
- Use `httptest.Server` for HTTP tests
- Test full request/response cycle

```go
builder := wf.NewBuilder(config.LoadAutoConfig())
builder.Rest(&UserController{})
builder.Model(&UserModel{})
app := builder.Build()

// Use httptest for testing
ts := httptest.NewServer(app.Engine())
defer ts.Close()
resp, _ := http.Get(ts.URL + "/api/users")
```
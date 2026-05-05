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
│  │  NewWithAutoConfig() | AddRest() | AddModel() | Run()    │   │
│  └──────────────────────────────────────────────────────────┘   │
│                              │                                   │
│                              ▼                                   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                    Core Context                           │   │
│  │  - Manages DI container                                  │   │
│  │  - Provides config access                                │   │
│  │  - Route registration (Context.Get/Post/etc)            │   │
│  └──────────────────────────────────────────────────────────┘   │
│                              │                                   │
│                              ▼                                   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                    HandlerConfig                          │   │
│  │  - Handles (RouteTree storage)                           │   │
│  │  - Filters (middleware chain)                            │   │
│  │  - Converter (response transformation)                   │   │
│  └──────────────────────────────────────────────────────────┘   │
│                              │                                   │
│                              ▼                                   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                    HttpServer                             │   │
│  │  - Process RouteTree → Gin routes                        │   │
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
4. IFilter.Init(ctx)             ← Fourth, filters can use services
5. IRunner.Run(ctx)              ← Last, background tasks start
```

## Route Registry Design

### Unified RouteTree

All routes (API, static files, reverse proxy) are stored in a single `RouteTree`:

```go
type RouteTree map[string]RouteInfo  // method → handlers

type HandlerInfo struct {
    path        string
    handlers    []HandlerFunc
    fs          http.FileSystem  // for StaticFs
    targetUrl   string           // for ReverseProxy
}
```

### Type Detection

```go
func (h *HandlerInfo) IsStaticFs() bool     { return h.fs != nil }
func (h *HandlerInfo) IsReverseProxy() bool { return h.targetUrl != "" }
```

### Processing in HttpServer

```go
func (s *HttpServer) Handle(config *HandlerConfig) {
    for method, routes := range config.RouteTree() {
        for _, info := range routes {
            switch {
            case info.IsReverseProxy():
                s.handleReverseProxy(method, info)
            case info.IsStaticFs():
                s.handleStaticFs(info)
            default:
                s.handleAPI(method, info)
            }
        }
    }
}
```

**Benefits:**
- Single source of truth for all routes
- Consistent registration API
- Easy to extend with new route types

## Dependency Injection

### Context as DI Container

```go
type Context struct {
    config       IConfig
    modelMap     map[string]IModel
    serviceMap   map[string]IService
    componentMap map[string]IService  // removed, merged into serviceMap
    runnerMap    map[string]IRunner
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
```

## Error Handling Strategy

### Unified Error Conversion

```go
type Converter interface {
    Request(fc FilterChain, req *Request)
}

// Default: error → JSON response with error message
func (c *Converter) Request(fc FilterChain, req *Request) {
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
func NewWithAutoConfig() *WebFrame {
    cfg := config.NewAutoConfig("appname")
    cfg.AddSearchPath("./config")
    cfg.AddSearchPath(filepath.Join(os.UserHomeDir(), ".appname"))
    cfg.AddSearchPath("/etc/appname")
    // ...
}
```

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
err := tx.Execute(func(db *gorm.DB) error {
    // All operations within transaction
    if err := userRepo.Create(db, user); err != nil {
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
2. Register with `AddService()`
3. Retrieve via `GetService[T](ctx)`

### Custom Converter

```go
type GraphQLConverter struct{}

func (c *GraphQLConverter) Request(fc FilterChain, req *Request) {
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
server := web.NewHttpServer(config, certManager)
server.Handle(handlerConfig)

ts := httptest.NewServer(server.Engine())
resp, _ := http.Get(ts.URL + "/api/users")
```
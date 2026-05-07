# Filter/Middleware

HTTP request filtering for cross-cutting concerns.

## Define Filter

```go
type AuthFilter struct {
    core.IFilter
}

func (f *AuthFilter) Init(ctx *core.Context) error {
    return nil
}

func (f *AuthFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    // Pre-processing
    token := req.GetHeader("Authorization")
    if token == "" {
        return nil, errors.New("unauthorized")
    }
    
    // Call next handler
    result, err := fc.Next()
    
    // Post-processing (optional)
    return result, err
}
```

## Register Filter

```go
builder.Filter(&AuthFilter{})
builder.Filter(&cors.Filter{})
```

## Route Metadata (WithMeta)

Declare metadata per route:

```go
func RequireAuth() web.MetaOption {
    return web.WithValue("require_auth", true)
}

builder.Get("/profile", handler).WithMeta(RequireAuth())
```

Read metadata in filter:

```go
func (f *AuthFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    meta := req.HandlerMeta()
    
    if meta.GetBool("require_auth") {
        // check auth
    }
    
    return fc.Next()
}
```

## Built-in Filters

| Filter | Description |
|--------|-------------|
| `cors.Filter` | CORS support |
| `rate_limit.Filter` | Rate limiting |

## Gin Middleware Compatibility

```go
type GzipFilter struct {
    handler gin.HandlerFunc
}

func (f *GzipFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    f.handler(req.GinContext())
    return fc.Next()
}
```
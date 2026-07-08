# Web API

HTTP handling and routing.

## Request

Request wrapper with helper methods.

```go
// Path params
id := c.Param("id")
idInt := c.ParamInt("id")
idUint := c.ParamUint("id")

// Query params
q := c.Query("q")
qInt := c.QueryInt("q")

// JSON body
var data MyStruct
c.BindJSON(&data)

// Get JSON values
name, _ := c.GetJsonStringValue("name")

// Headers
auth := c.GetHeader("Authorization")

// Pagination
page, _ := c.Page()

// Cookie
cookie := c.Cookie()
val := cookie.Get("session")

// Context
ctx := c.Context()

// Gin context
ginCtx := c.GinContext()

// Client IP
ip := c.ClientIP()
```

## Response Types

```go
// JSON (default)
return map[string]any{"key": "value"}, nil

// String
return "text", nil

// File download
return &web.File{Path: "/path/file.pdf", FileName: "doc.pdf"}, nil

// Redirect
return web.Redirect("/new-url"), nil

// Error
return nil, errors.New("error")
```

## HasMeta

Query route metadata (set via `WithMeta`) in filters or handlers:

```go
// Check if route has a metadata key
if req.HasMeta(RequireAuth()) {
    // route requires authentication
}
```

## MetaOption

```go
// Define
func RequireAuth() web.MetaOption {
    return web.WithValue("require_auth", true)
}

// Use on route
builder.Get("/profile", handler).WithMeta(RequireAuth())

// Query in filter — unified with WithMeta
func (f *AuthFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    if req.HasMeta(RequireAuth()) {
        // verify token
    }
    return fc.Next()
}
```

Built-in constructors:
- `web.WithKey(keys ...string)` — match if **any** key exists
- `web.WithValue(key string, value any)` — match if key exists and value equals (via `reflect.DeepEqual`)
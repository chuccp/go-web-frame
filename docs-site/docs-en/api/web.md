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

## web.Message

`web/message.go` provides the standard response format used by the framework:

```go
type Message struct {
    Code int    `json:"code"`
    Data any    `json:"data"`
    Msg  string `json:"msg"`
    Type string `json:"type"`
}
```

Helper functions:

```go
// Success responses
return web.Ok(), nil
return web.Ok("success"), nil
return web.Data(users), nil
return web.DataCode(http.StatusCreated, user), nil
return web.DataType("table", data), nil

// Check status
msg := web.Data(users)
if msg.IsOK() { /* code == 200 */ }
```

Usage in controllers:

```go
func (c *UserController) List(req *web.Request) (any, error) {
    users, err := c.userService.GetAll()
    if err != nil {
        return nil, err
    }
    return web.Data(users), nil
}

func (c *UserController) Create(req *web.Request) (any, error) {
    if err := c.userService.Create(req); err != nil {
        return nil, err
    }
    return web.Ok("created"), nil
}
```

## ErrorCode

`web/error_code.go` defines structured error codes. Returning an `*ErrorCode` from a handler lets the converter map it to the correct HTTP status.

### Error code constants

| Constant | Value | Meaning |
|---|---|---|
| `CodeOK` | 200 | Success |
| `CodeBadRequest` | 400 | Bad request |
| `CodeUnauthorized` | 401 | Unauthorized |
| `CodeForbidden` | 403 | Forbidden |
| `CodeNotFound` | 404 | Not found |
| `CodeMethodNotAllowed` | 405 | Method not allowed |
| `CodeTooManyRequests` | 429 | Too many requests |
| `CodeInternalError` | 500 | Internal server error |
| `CodeServiceUnavailable` | 503 | Service unavailable |
| `CodeValidationFailed` | 1001 | Validation failed |
| `CodeDuplicateEntry` | 1002 | Duplicate entry |
| `CodeTokenExpired` | 1003 | Token expired |
| `CodeTokenInvalid` | 1004 | Token invalid |

### Constructing errors

```go
return nil, web.NewBadRequest().WithDetail("invalid params")
return nil, web.NewUnauthorized().WithError(err)
return nil, web.NewForbidden().WithDetail("access denied")
return nil, web.NewNotFound().WithDetail("user not found")
return nil, web.NewTooManyRequests().WithDetail("rate limited")
return nil, web.NewInternalError().WithError(err)
return nil, web.NewServiceUnavailable().WithDetail("try later")

// Business errors
return nil, web.NewValidationError().WithError(err)
return nil, web.NewDuplicateEntry().WithDetail("email exists")
return nil, web.NewTokenExpired().WithDetail("token expired")
return nil, web.NewTokenInvalid().WithDetail("token invalid")

// Custom code
return nil, web.NewErrorCode(2001, "custom error").WithDetail("...")
```

### ErrorCode structure

```go
type ErrorCode struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Detail  string `json:"detail,omitempty"`
    Err     error  `json:"-"`
}
```

- `Code` is serialized to the response JSON.
- `Message` is a short summary.
- `Detail` is optional extra information.
- `Err` holds the wrapped error; it is not serialized but supports `errors.Is` / `errors.As`.

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
# Error Code

`web/error_code.go` provides a structured error type `ErrorCode` for returning HTTP error responses with business codes from handlers.

## ErrorCode Structure

```go
type ErrorCode struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Detail  string `json:"detail,omitempty"`
    Err     error  `json:"-"`
}
```

| Field | Description |
|---|---|
| `Code` | Business/HTTP error code, serialized to the response JSON |
| `Message` | Short error summary |
| `Detail` | Optional detailed description |
| `Err` | Wrapped original error, not serialized, supports `errors.Is` / `errors.As` |

## Error Code Constants

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

## Constructing Errors

### Pre-defined constructors

Each pre-defined error has a constructor that returns a mutable copy:

```go
return nil, web.NewBadRequest().WithDetail("invalid params")     // 400
return nil, web.NewUnauthorized().WithError(err)                 // 401
return nil, web.NewForbidden().WithDetail("access denied")       // 403
return nil, web.NewNotFound().WithDetail("user not found")       // 404
return nil, web.NewMethodNotAllowed().WithDetail("not allowed")  // 405
return nil, web.NewTooManyRequests().WithDetail("rate limited")  // 429
return nil, web.NewInternalError().WithError(err)                // 500
return nil, web.NewServiceUnavailable().WithDetail("try later")  // 503

// Business errors
return nil, web.NewValidationError().WithError(err)              // 1001
return nil, web.NewDuplicateEntry().WithDetail("email exists")   // 1002
return nil, web.NewTokenExpired().WithDetail("token expired")    // 1003
return nil, web.NewTokenInvalid().WithDetail("token invalid")    // 1004
```

### Custom error code

```go
return nil, web.NewErrorCode(2001, "custom error").WithDetail("custom detail")
```

## Chaining Methods

```go
err := web.NewNotFound().
    WithDetail("user #123 not found").
    WithError(sqlErr)
```

- `WithDetail(detail string) *ErrorCode` — attach detailed information
- `WithError(err error) *ErrorCode` — wrap the original error

## Usage in Handlers

```go
func (c *UserController) Get(req *web.Request) (any, error) {
    user, err := c.userService.GetById(req.ParamUint("id"))
    if err != nil {
        return nil, web.NewInternalError().WithError(err)
    }
    if user == nil {
        return nil, web.NewNotFound().WithDetail("user not found")
    }
    return user, nil
}
```

## Error Mapping

`web.ClassifyError(value, err)` maps an error to an HTTP status code and message. `DefaultConverter` uses this internally:

1. If the error chain contains `*ErrorCode`, use its `Code`.
2. If the return value is `*web.Message` with `Code != 200`, use `Message.Code`.
3. `os.ErrNotExist` → 404; `os.ErrPermission` → 403.
4. Any other error → 500.

You can also call it directly in your own converter or filter:

```go
code, msg := web.ClassifyError(nil, err)
req.Response().AbortWithStatusJSON(code, &web.Message{Code: code, Msg: msg})
```

## Handle Errors Uniformly in a Filter

```go
func (f *ErrorHandlerFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    result, err := fc.Next()
    if err == nil {
        return result, nil
    }

    var ec *web.ErrorCode
    if errors.As(err, &ec) {
        req.Response().AbortWithStatusJSON(ec.Code, ec)
        return nil, nil
    }

    req.Response().AbortWithStatusJSON(500, web.NewInternalError().WithError(err))
    return nil, nil
}
```

## Full Example

```go
package main

import (
    "context"
    "errors"
    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/config"
    "github.com/chuccp/go-web-frame/core"
    "github.com/chuccp/go-web-frame/web"
)

type UserController struct{ core.IService }

func (c *UserController) Init(ctx *core.Context) error {
    ctx.Get("/users/:id", c.Get)
    return nil
}

func (c *UserController) Get(req *web.Request) (any, error) {
    if req.ParamUint("id") == 0 {
        return nil, web.NewBadRequest().WithDetail("id is required")
    }
    return map[string]any{"id": req.ParamUint("id"), "name": "alice"}, nil
}

type AuthFilter struct{ core.IFilter }

func (f *AuthFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    if req.GetHeader("Authorization") == "" {
        return nil, web.NewUnauthorized().WithDetail("missing token")
    }
    return fc.Next()
}

func main() {
    cfg, _ := config.LoadSingleFileConfig("application.yml")
    builder := wf.NewBuilder(cfg)
    builder.Filter(&AuthFilter{})
    builder.Rest(&UserController{})
    builder.Build().Run(context.Background())
}
```

## Next Steps

- [Response Converter](converter.md) - How converters handle ErrorCode
- [Filter/Middleware](filter.md) - Handle errors uniformly in filters
- [Web API Reference](../api/web.md) - web.Message and error responses

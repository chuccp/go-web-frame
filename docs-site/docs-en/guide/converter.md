# Response Converter

The response converter turns handler return values into HTTP responses. The `web.Converter` interface has a single method:

```go
type Converter interface {
    Request(filterChain FilterChain, request *Request)
}
```

The framework uses `web.DefaultConverter` by default. It already handles `*web.Message`, `*web.ErrorCode`, file downloads, redirects, WebSocket, SSE, and more. When writing a custom converter, **embed `*web.DefaultConverter`** and only override what you need, so you don't accidentally fail to parse messages, files, or errors.

> Note: `RestGroupBuilder.Converter()` accepts `core.IConverter` (which is `IService` + `web.Converter`), so a converter registered with a REST group must also implement `Init(ctx *core.Context) error`.

## Default Conversion Rules

`web.DefaultConverter` handles these return types:

| Handler return value | Response behavior |
|---|---|
| `*web.Message` | JSON response using `Message.Code` as HTTP status; redirects when `Code == 301` |
| `*web.ErrorCode` | JSON error response via `ClassifyError` |
| `*web.FileResponse` | File download |
| `*web.FileSystemResponse` | Serve file from `http.FileSystem` |
| `*web.SSEResponse` | Open Server-Sent Events stream |
| `*web.WSResponse` | Upgrade to WebSocket |
| `*web.ReverseProxyResponse` | Reverse proxy |
| `*os.File` | Download as attachment |
| `string` | Plain text |
| `error` | Error handling flow |
| Any other value | Wrap as `web.Data(value)` and return JSON |

### Error mapping

`web.ClassifyError(value, err)` resolves errors in this priority:

1. `*web.ErrorCode` — uses its `Code` and `Error()` text
2. `*web.Message` with `Code != 200` — uses `Message.Code`
3. `os.ErrNotExist` → 404; `os.ErrPermission` → 403
4. Any other error → 500

## Simplest Custom Converter

To add logging, metrics, or other cross-cutting logic, just delegate to the default converter:

```go
package converter

import (
    "time"
    "github.com/chuccp/go-web-frame/core"
    "github.com/chuccp/go-web-frame/web"
    "go.uber.org/zap"
)

type LoggingConverter struct {
    core.IConverter
    *web.DefaultConverter
}

func (c *LoggingConverter) Init(ctx *core.Context) error {
    return nil
}

func (c *LoggingConverter) Request(fc web.FilterChain, req *web.Request) {
    start := time.Now()
    c.DefaultConverter.Request(fc, req)
    zap.L().Info("request converted",
        zap.String("path", req.FullPath()),
        zap.Duration("cost", time.Since(start)),
    )
}
```

All messages, error codes, files, and redirects are still handled by the default logic.

## Custom Unified Response Format

If you want to wrap plain success values into `{code, data, msg}` while still letting the default converter handle errors, messages, files, etc.:

```go
type APIConverter struct {
    core.IConverter
    *web.DefaultConverter
}

func (c *APIConverter) Init(ctx *core.Context) error { return nil }

func (c *APIConverter) Request(fc web.FilterChain, req *web.Request) {
    result, err := fc.Next()
    if err != nil {
        // Let the default converter handle ErrorCode / Message / os errors
        c.DefaultConverter.Error(req, result, err)
        return
    }

    switch v := result.(type) {
    case nil:
        req.Response().JSON(200, web.Ok())
    case *web.Message:
        // Already a standard Message, pass through
        c.DefaultConverter.Message(req, v)
    case *web.ErrorCode:
        c.DefaultConverter.Error(req, v, v)
    case string:
        c.DefaultConverter.String(req, v)
    case *web.FileResponse, *web.FileSystemResponse, *web.SSEResponse,
         *web.WSResponse, *web.ReverseProxyResponse, *os.File:
        // Files, WebSocket, SSE, etc. go to the default converter
        c.DefaultConverter.Request(fc, req)
    default:
        // Custom wrapping
        req.Response().JSON(200, web.Data(v))
    }
}
```

> Note: except for the first pattern where you fully delegate to `DefaultConverter.Request`, once you call `fc.Next()` yourself, do **not** call `c.DefaultConverter.Request(fc, req)` again — that would run the filter chain twice. For specific types, call the exported helper methods on `DefaultConverter` directly (`Message`, `Error`, `FileResponse`, etc.).

## Register the Converter

Register a converter on a `RestGroupBuilder`:

```go
func main() {
    cfg, _ := config.LoadSingleFileConfig("application.yml")
    builder := wf.NewBuilder(cfg)

    restGroup := wf.NewRestGroupBuilder().
        Rest(&UserController{}).
        Converter(&APIConverter{DefaultConverter: &web.DefaultConverter{}}).
        Port(8081).
        Build()
    builder.RestGroup(restGroup)

    builder.Build().Run(context.Background())
}
```

## Converter vs Filter

| Feature | Filter | Converter |
|---|---|---|
| Position | Before/after handler | Last step of the filter chain |
| Responsibility | Auth, logging, rate limiting, etc. | Convert handler result to HTTP response |
| Chaining | Multiple filters run in sequence | One converter handles the request |

## Full Example

```go
package main

import (
    "context"
    "net/http"
    "os"
    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/config"
    "github.com/chuccp/go-web-frame/core"
    "github.com/chuccp/go-web-frame/web"
)

type UserController struct{ core.IService }

func (c *UserController) Init(ctx *core.Context) error {
    ctx.Get("/users", c.List)
    ctx.Get("/error", c.Error)
    return nil
}

func (c *UserController) List(req *web.Request) (any, error) {
    return []map[string]any{{"id": 1, "name": "alice"}}, nil
}

func (c *UserController) Error(req *web.Request) (any, error) {
    return nil, web.NewBadRequest().WithDetail("invalid request")
}

type APIConverter struct {
    core.IConverter
    *web.DefaultConverter
}

func (c *APIConverter) Init(ctx *core.Context) error { return nil }

func (c *APIConverter) Request(fc web.FilterChain, req *web.Request) {
    result, err := fc.Next()
    if err != nil {
        c.DefaultConverter.Error(req, result, err)
        return
    }

    switch v := result.(type) {
    case nil:
        req.Response().JSON(200, web.Ok())
    case *web.Message, *web.ErrorCode, string, *web.FileResponse,
         *web.FileSystemResponse, *web.SSEResponse, *web.WSResponse,
         *web.ReverseProxyResponse, *os.File:
        c.DefaultConverter.Request(fc, req)
    default:
        req.Response().JSON(200, web.Data(v))
    }
}

func main() {
    cfg, _ := config.LoadSingleFileConfig("application.yml")
    builder := wf.NewBuilder(cfg)

    restGroup := wf.NewRestGroupBuilder().
        Rest(&UserController{}).
        Converter(&APIConverter{DefaultConverter: &web.DefaultConverter{}}).
        Port(8081).
        Build()
    builder.RestGroup(restGroup)

    builder.Build().Run(context.Background())
}
```

## Next Steps

- [Filter/Middleware](filter.md) - Filter chain overview
- [Controller](controller.md) - Handler return values and converters
- [Web API Reference](../api/web.md) - Response types and error codes

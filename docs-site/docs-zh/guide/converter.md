# 响应转换器

响应转换器（Converter）负责把 Handler 的返回值转换成 HTTP 响应。`web.Converter` 接口本身只有一个方法：

```go
type Converter interface {
    Request(filterChain FilterChain, request *Request)
}
```

框架默认使用 `web.DefaultConverter`，它已经处理了 `*web.Message`、`*web.ErrorCode`、文件下载、重定向、WebSocket、SSE 等常见返回值。自定义 Converter 时，**推荐嵌入 `*web.DefaultConverter`**，只覆盖需要定制的行为，避免遗漏 Message、文件等类型的解析。

> 注意：`RestGroupBuilder.Converter()` 接收的是 `core.IConverter`（即 `IService` + `web.Converter`），因此注册到 REST 分组的自定义 Converter 还需要实现 `Init(ctx *core.Context) error`。

## 默认转换规则

`web.DefaultConverter` 对 Handler 返回值的转换规则如下：

| Handler 返回值 | 响应行为 |
|---|---|
| `*web.Message` | 以 `Message.Code` 为 HTTP 状态码返回 JSON；`Code == 301` 时重定向 |
| `*web.ErrorCode` | 通过 `ClassifyError` 映射为 JSON 错误响应 |
| `*web.FileResponse` | 文件下载 |
| `*web.FileSystemResponse` | 从 `http.FileSystem` 读取并返回文件 |
| `*web.SSEResponse` | 建立 Server-Sent Events 流 |
| `*web.WSResponse` | 升级 WebSocket 连接 |
| `*web.ReverseProxyResponse` | 反向代理 |
| `*os.File` | 作为附件下载 |
| `string` | 直接写入纯文本 |
| `error` | 走错误处理流程 |
| 其他任意值 | 包装为 `web.Data(value)` 返回 JSON |

### 错误映射

`web.ClassifyError(value, err)` 按以下优先级处理错误：

1. `*web.ErrorCode`：使用其 `Code` 和 `Error()` 文本
2. `*web.Message` 且 `Code != 200`：使用 `Message.Code`
3. `os.ErrNotExist` → 404；`os.ErrPermission` → 403
4. 其他 error → 500

## 最简单的自定义 Converter

如果只是想加日志、指标等横切逻辑，可以直接调用默认 Converter：

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

这样所有 Message、ErrorCode、文件、重定向等仍然走默认逻辑，你只需要关心额外逻辑。

## 定制统一返回格式

如果需要把成功响应统一包装成 `{code, data, msg}`，同时继续让默认 Converter 处理错误、Message、文件等，可以这样做：

```go
type APIConverter struct {
    core.IConverter
    *web.DefaultConverter
}

func (c *APIConverter) Init(ctx *core.Context) error { return nil }

func (c *APIConverter) Request(fc web.FilterChain, req *web.Request) {
    result, err := fc.Next()
    if err != nil {
        // 错误交给默认 Converter，自动处理 ErrorCode / Message / os 错误
        c.DefaultConverter.Error(req, result, err)
        return
    }

    switch v := result.(type) {
    case nil:
        req.Response().JSON(200, web.Ok())
    case *web.Message:
        // 已经返回标准 Message，直接透传
        c.DefaultConverter.Message(req, v)
    case *web.ErrorCode:
        c.DefaultConverter.Error(req, v, v)
    case string:
        c.DefaultConverter.String(req, v)
    case *web.FileResponse, *web.FileSystemResponse, *web.SSEResponse,
         *web.WSResponse, *web.ReverseProxyResponse, *os.File:
        // 文件、WebSocket、SSE 等交给默认 Converter
        c.DefaultConverter.Request(fc, req)
    default:
        // 自定义统一包装
        req.Response().JSON(200, web.Data(v))
    }
}
```

> 注意：除了第一种“完全委托给默认 Converter”的场景外，一旦你手动调用了 `fc.Next()`，就不要再调用 `c.DefaultConverter.Request(fc, req)`，否则 Filter 链会被执行两次。需要处理特定类型时，直接调用 `DefaultConverter` 上对应的导出方法（`Message`、`Error`、`FileResponse` 等）。

## 注册 Converter

Converter 通过 `RestGroupBuilder.Converter()` 注册到 REST 分组：

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

## Converter 与 Filter 的区别

| 特性 | Filter | Converter |
|---|---|---|
| 位置 | Handler 之前/之后 | Filter 链的最后一步 |
| 职责 | 认证、日志、限流等横切逻辑 | 把 Handler 返回值写成 HTTP 响应 |
| 链式执行 | 多个 Filter 依次执行 | 一个请求最终只被一个 Converter 处理 |

## 完整示例

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

## 下一步

- [过滤器/中间件](filter.md) - 了解 Filter 链
- [控制器](controller.md) - 控制器返回值与 Converter 配合
- [Web API 参考](../api/web.md) - 响应类型与错误码

# Web API 参考

本文档介绍 Go Web Frame 的 Web 层 API。

## HandlerFunc

处理器函数签名：

```go
func(req *web.Request) (any, error)
```

- 返回 `any`：自动转换为 JSON 响应
- 返回 `error`：自动转换为错误响应

## 路由注册

### 基本路由

路由注册在 `Builder` 上进行：

```go
builder := wf.NewBuilder(cfg)

// HTTP 方法
builder.Get("/users", handler)
builder.Post("/users", handler)
builder.Put("/users/:id", handler)
builder.Delete("/users/:id", handler)

// 任意方法
builder.Any("/api", handler)

app := builder.Build()
```

### 路由前缀

使用 `RestGroupBuilder` 的 `ContextPath` 为路由添加前缀：

```go
group := wf.NewRestGroupBuilder()
group.ContextPath("/api/v1")
group.Rest(&UserController{})
builder.RestGroup(group.Build())
```

### 路由元数据

```go
// 定义元数据选项
func Public() web.MetaOption {
    return web.WithValue("public", true)
}

func RequireAuth() web.MetaOption {
    return web.WithValue("auth", true)
}

// 注册带元数据的路由
ctx.Get("/login", c.Login).WithMeta(Public())
ctx.Get("/users", c.List).WithMeta(RequireAuth())
```

## 响应类型

### JSON 响应（默认）

```go
func handler(req *web.Request) (any, error) {
    return map[string]any{
        "message": "Hello",
        "status":  "success",
    }, nil
}
```

### 字符串响应

```go
func handler(req *web.Request) (any, error) {
    return "plain text response", nil
}
```

### 文件下载

```go
func handler(req *web.Request) (any, error) {
    return &web.File{
        Path:     "/path/to/file.pdf",
        FileName: "document.pdf",
    }, nil
}
```

### 重定向

```go
func handler(req *web.Request) (any, error) {
    return web.Redirect("/new-url"), nil
}
```

### 自定义响应

```go
func handler(req *web.Request) (any, error) {
    req.Response().Header().Set("Content-Type", "application/xml")
    req.Response().Write([]byte("<xml>...</xml>"))
    return nil, nil
}
```

## web.Message 标准响应

框架提供 `web.Message` 类型作为标准响应格式：

```go
type Message struct {
    Code int    `json:"code"`
    Data any    `json:"data"`
    Msg  string `json:"msg"`
    Type string `json:"type"`
}
```

### 辅助函数

```go
// 成功响应（code=200）
return web.Ok(), nil
return web.Ok("操作成功"), nil

// 带数据的成功响应
return web.Data(users), nil

// 带自定义状态码和数据的响应
return web.DataCode(201, createdUser), nil

// 带类型的响应
return web.DataType("table", data), nil

// 错误响应 — 返回 *ErrorCode，converter 自动识别状态码
return nil, errors.New("something went wrong")         // 500
return nil, web.NewInternalError().WithDetail("server error")
return nil, web.NewValidationError().WithError(err)     // 1001
return nil, web.NewNotFound().WithDetail("user #123")   // 404

// 未授权响应（code=401）
return nil, web.NewUnauthorized().WithError(err)

// 重定向响应
return web.Redirect("/login"), nil
```

### 在控制器中使用

```go
func (c *UserController) List(req *web.Request) (any, error) {
    users, err := c.userService.GetAllUsers()
    if err != nil {
        return nil, err
    }
    return web.Data(users), nil
}

func (c *UserController) Create(req *web.Request) (any, error) {
    var input UserInput
    if err := req.BindJSON(&input); err != nil {
        return nil, err
    }
    if err := c.userService.Create(&input); err != nil {
        return nil, err
    }
    return web.Ok("创建成功"), nil
}
```

## 静态文件

### 静态文件目录

在控制器的 `Init` 方法中注册：

```go
func (c *MyController) Init(ctx *core.Context) error {
    ctx.Static("/static", "./www")
    return nil
}
```

访问 `/static/style.css` 会返回 `./www/style.css` 文件。

### 静态文件系统

```go
func (c *MyController) Init(ctx *core.Context) error {
    ctx.StaticFs("/assets", http.Dir("./dist"))
    return nil
}
```

## 反向代理

```go
func (c *MyController) Init(ctx *core.Context) error {
    ctx.ReverseProxy("/api", "http://backend:8080")
    return nil
}
```

所有 `/api/*` 请求都会被代理到 `http://backend:8080/api/*`。

### WebSocket 支持

```go
func (c *MyController) Init(ctx *core.Context) error {
    ctx.WebSocket("/ws", func(stream *web.WebSocketStream) error {
        defer stream.Close()
        for {
            typ, data, err := stream.Read(stream.Context())
            if err != nil {
                return nil
            }
            stream.Write(stream.Context(), typ, data)
        }
    })
    return nil
}
```

### SSE 支持

```go
func (c *MyController) Init(ctx *core.Context) error {
    ctx.SSE("/events", func(stream *web.SSEStream) error {
        defer stream.Close()
        stream.Send("message", "Hello")
        return nil
    })
    return nil
}
```

## 错误码（ErrorCode）

`web/error_code.go` 定义了统一的错误码结构体和预定义错误。Handler 返回 `*ErrorCode` 时，框架会自动映射为对应的 HTTP 状态码和业务错误码。

### 错误码常量

| 常量 | 值 | 说明 |
|---|---|---|
| `CodeOK` | 200 | 成功 |
| `CodeBadRequest` | 400 | 请求参数错误 |
| `CodeUnauthorized` | 401 | 未授权 |
| `CodeForbidden` | 403 | 禁止访问 |
| `CodeNotFound` | 404 | 资源不存在 |
| `CodeMethodNotAllowed` | 405 | 方法不允许 |
| `CodeTooManyRequests` | 429 | 请求过于频繁 |
| `CodeInternalError` | 500 | 服务器内部错误 |
| `CodeServiceUnavailable` | 503 | 服务不可用 |
| `CodeValidationFailed` | 1001 | 校验失败 |
| `CodeDuplicateEntry` | 1002 | 重复记录 |
| `CodeTokenExpired` | 1003 | Token 已过期 |
| `CodeTokenInvalid` | 1004 | Token 无效 |

### 构造错误响应

```go
// 预定义错误构造函数（均返回可修改的副本）
return nil, web.NewBadRequest().WithDetail("参数错误")          // 400
return nil, web.NewUnauthorized().WithError(err)              // 401
return nil, web.NewForbidden().WithDetail("无权限")            // 403
return nil, web.NewNotFound().WithDetail("user not found")    // 404
return nil, web.NewTooManyRequests().WithDetail("请稍后再试")   // 429
return nil, web.NewInternalError().WithError(err)             // 500
return nil, web.NewServiceUnavailable().WithDetail("服务暂不可用") // 503

// 业务错误
return nil, web.NewValidationError().WithError(err)           // 1001
return nil, web.NewDuplicateEntry().WithDetail("邮箱已存在")    // 1002
return nil, web.NewTokenExpired().WithDetail("token 过期")     // 1003
return nil, web.NewTokenInvalid().WithDetail("token 无效")     // 1004

// 自定义错误码
return nil, web.NewErrorCode(2001, "custom error").WithDetail("自定义错误详情")
```

### ErrorCode 结构

```go
type ErrorCode struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Detail  string `json:"detail,omitempty"`
    Err     error  `json:"-"`
}
```

- `Code`：业务错误码，会放入响应 JSON
- `Message`：错误摘要
- `Detail`：详细说明（可选）
- `Err`：原始错误（不会序列化，支持 `errors.Is` / `errors.As` 链）

### 在 Filter 中统一处理错误

```go
func (f *ErrorHandlerFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    result, err := fc.Next()
    if err == nil {
        return result, nil
    }

    // 如果是 *ErrorCode，可直接拿到 code
    var ec *web.ErrorCode
    if errors.As(err, &ec) {
        req.Response().AbortWithStatusJSON(ec.Code, ec)
        return nil, nil
    }

    // 普通 error 统一返回 500
    req.Response().AbortWithStatusJSON(500, web.NewInternalError().WithError(err))
    return nil, nil
}
```

## 下一步

- [模型 API](model.md) - 了解模型层 API
- [核心 API](../api/core.md) - 了解核心 API

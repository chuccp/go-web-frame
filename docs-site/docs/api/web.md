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
builder.Patch("/users/:id", handler)

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
    ctx.WebSocket("/ws", func(conn *websocket.Conn, req *web.Request) {
        // WebSocket 处理逻辑
    })
    return nil
}
```

### SSE 支持

```go
func (c *MyController) Init(ctx *core.Context) error {
    ctx.SSE("/events", func(writer web.SSEWriter, req *web.Request) {
        // SSE 事件处理逻辑
    })
    return nil
}
```

## 下一步

- [模型 API](model.md) - 了解模型层 API
- [核心 API](../api/core.md) - 了解核心 API

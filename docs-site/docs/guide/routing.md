# 路由

Go Web Frame 提供了强大而灵活的路由系统，基于 Gin。

## 基本路由

### HTTP 方法

```go
// 使用 Builder 注册路由
builder.Get("/users", handler)        // GET 请求
builder.Post("/users", handler)       // POST 请求
builder.Put("/users/:id", handler)   // PUT 请求
builder.Delete("/users/:id", handler) // DELETE 请求
builder.Any("/api", handler)          // 匹配所有 HTTP 方法
```

### 在控制器中注册路由

```go
type UserController struct {
    core.IService
}

func (c *UserController) Init(context *core.Context) error {
    // 直接注册路由
    context.Get("/users", c.List)
    context.Get("/users/:id", c.Get)
    context.Post("/users", c.Create)
    context.Put("/users/:id", c.Update)
    context.Delete("/users/:id", c.Delete)
    return nil
}
```

## 路径参数

### 基本参数

```go
builder.Get("/users/:id", func(req *web.Request) (any, error) {
    id := req.Param("id") // 获取路径参数
    return "User ID: " + id, nil
})
```

访问 `/users/123`，返回 `User ID: 123`

### 多个参数

```go
builder.Get("/users/:userId/posts/:postId", func(req *web.Request) (any, error) {
    userId := req.Param("userId")
    postId := req.Param("postId")
    return map[string]any{
        "userId": userId,
        "postId": postId,
    }, nil
})
```

访问 `/users/1/posts/2`，返回：

```json
{
  "userId": "1",
  "postId": "2"
}
```

## 参数类型转换

`web.Request` 提供了类型安全的方法：

```go
// 字符串（默认）
idStr := req.Param("id")         // "123"

// 整数
idInt := req.ParamInt("id")      // 123 (int)

// 无符号整数
idUint := req.ParamUint("id")    // 123 (uint)
```

## 查询参数

```go
builder.Get("/search", func(req *web.Request) (any, error) {
    keyword := req.Query("q")       // /users?q=go
    page := req.DefaultQuery("page", "1") // 带默认值
    return map[string]any{
        "keyword": keyword,
        "page":    page,
    }, nil
})
```

访问 `/search?q=go&page=2`，返回：

```json
{
  "keyword": "go",
  "page": "2"
}
```

## REST 控制器

### 创建 REST 控制器

```go
package main

import (
    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/core"
    "github.com/chuccp/go-web-frame/web"
    "github.com/chuccp/go-web-frame/config"
    "go.uber.org/zap"
)

type UserController struct {
    core.IService
    userService *UserService
}

func (c *UserController) Init(context *core.Context) error {
    // 获取依赖的服务
    c.userService = wf.GetService[*UserService](context)
    
    // 注册路由
    context.Get("/users", c.List)
    context.Get("/users/:id", c.Get)
    context.Post("/users", c.Create)
    context.Put("/users/:id", c.Update)
    context.Delete("/users/:id", c.Delete)
    
    return nil
}

func (c *UserController) List(req *web.Request) (any, error) {
    users, err := c.userService.GetAllUsers()
    if err != nil {
        return nil, err
    }
    return users, nil
}

func (c *UserController) Get(req *web.Request) (any, error) {
    id := req.ParamUint("id")
    user, err := c.userService.GetUserById(id)
    if err != nil {
        return nil, err
    }
    return user, nil
}

func main() {
    // 加载配置
    fileConfig, err := config.LoadSingleFileConfig("application.yml")
    if err != nil {
        zap.L().Fatal("加载配置失败", zap.Error(err))
    }

    // 创建 Builder
    builder := wf.NewBuilder(fileConfig)
    
    // 注册 REST 控制器
    restGroupBuilder := wf.NewRestGroupBuilder()
    restGroupBuilder.Rest(&UserController{})
    restGroupBuilder.Port(8081)
    restGroup := restGroupBuilder.Build()
    builder.RestGroup(restGroup)
    
    // 构建应用
    app := builder.Build()
    
    // 启动应用
    err = app.Start()
    if err != nil {
        zap.L().Fatal("启动应用失败", zap.Error(err))
    }
}
```

## 路由组

使用 `wf.NewRestGroupBuilder()` 创建路由组，路由组可以配置独立的端口、前缀、过滤器和响应转换器：

### 基本路由组

```go
func main() {
    builder := wf.NewBuilder(fileConfig)

    // 创建路由组
    restGroup := wf.NewRestGroupBuilder().
        Rest(&UserController{}).
        Port(8081).
        Build()

    builder.RestGroup(restGroup)

    app := builder.Build()
    app.Start()
}
```

### 路由前缀（ContextPath）

使用 `ContextPath` 为路由组添加统一前缀：

```go
restGroup := wf.NewRestGroupBuilder().
    ContextPath("/api/v1").  // 所有路由前缀 /api/v1
    Rest(&UserController{}).
    Port(8081).
    Build()
```

生成的路由：

- `GET /api/v1/users/`
- `GET /api/v1/users/:id`
- `POST /api/v1/users/`
- `PUT /api/v1/users/:id`
- `DELETE /api/v1/users/:id`

### 路由组过滤器

路由组可以配置独立的过滤器：

```go
import auth2 "github.com/chuccp/go-web-frame/component/auth"

restGroup := wf.NewRestGroupBuilder().
    Rest(&UserController{}).
    Filter(auth2.NewAuthenticationFilter(&MyAuthenticator{})).
    Port(8081).
    Build()
```

### 路由组响应转换器

使用 `Converter` 为路由组配置自定义响应格式：

```go
restGroup := wf.NewRestGroupBuilder().
    Rest(&UserController{}).
    Converter(&APIConverter{}).
    Port(8081).
    Build()
```

### RestGroupBuilder 完整 API

| 方法 | 说明 |
|------|------|
| `Rest(rest ...IRest)` | 注册 REST 控制器 |
| `Port(port int)` | 设置监听端口 |
| `ContextPath(path string)` | 设置路由前缀 |
| `Filter(filters ...IFilter)` | 设置过滤器 |
| `Converter(converter IConverter)` | 设置响应转换器 |
| `ServerConfig(config *web.ServerConfig)` | 设置服务器配置（SSL、超时等） |
| `Build()` | 构建路由组 |

## 静态文件

### 静态文件目录

在控制器的 `Init` 方法中注册：

```go
func (c *MyController) Init(ctx *core.Context) error {
    // 注册静态文件目录，访问 /static/* 返回 ./www/*
    ctx.Static("/static", "./www")
    return nil
}
```

访问 `/static/style.css` 会返回 `./www/style.css` 文件。

### 静态文件系统

使用 `http.FileSystem` 接口注册静态文件，支持嵌入文件系统：

```go
import "net/http"

func (c *MyController) Init(ctx *core.Context) error {
    // 使用 http.Dir 注册静态文件系统
    ctx.StaticFs("/assets", http.Dir("./dist"))
    return nil
}
```

## 反向代理

在控制器中注册反向代理，将请求转发到后端服务：

```go
func (c *MyController) Init(ctx *core.Context) error {
    // 所有 /api/* 请求会被代理到 http://backend:8080/api/*
    ctx.ReverseProxy("/api", "http://backend:8080")
    return nil
}
```

## WebSocket

注册 WebSocket 端点：

```go
import "github.com/gorilla/websocket"

func (c *MyController) Init(ctx *core.Context) error {
    ctx.WebSocket("/ws", func(conn *websocket.Conn, req *web.Request) {
        // WebSocket 处理逻辑
        defer conn.Close()
        for {
            _, message, err := conn.ReadMessage()
            if err != nil {
                break
            }
            conn.WriteMessage(websocket.TextMessage, message)
        }
    })
    return nil
}
```

## SSE（Server-Sent Events）

注册 SSE 端点，用于服务器推送事件：

```go
func (c *MyController) Init(ctx *core.Context) error {
    ctx.SSE("/events", func(stream *web.SSEStream) error {
        // 设置 SSE 响应头
        stream.SetHeaders()

        // 启动心跳保活（每 30 秒）
        stopHeartbeat := stream.StartHeartbeat(30 * time.Second)
        defer stopHeartbeat()

        // 发送命名事件
        stream.Send("message", "Hello World")

        // 发送默认消息（无事件名）
        stream.SendMessage("ping")

        // 发送带 ID 的事件（客户端可断线重连）
        stream.SendWithID("1", "update", `{"status":"ok"}`)

        // 设置重连时间（毫秒）
        stream.SendRetry(3000)

        // 发送心跳
        stream.Heartbeat()

        // 关闭流
        stream.Close()

        return nil
    })
    return nil
}
```

### SSEStream API

| 方法 | 说明 |
|------|------|
| `SetHeaders()` | 设置 SSE 响应头（Content-Type、Cache-Control 等） |
| `Send(event, data)` | 发送命名事件 |
| `SendMessage(data)` | 发送默认消息（无事件名） |
| `SendWithID(id, event, data)` | 发送带 ID 的事件，客户端断线重连时可使用 Last-Event-ID |
| `SendRetry(retryMs)` | 设置客户端重连时间（毫秒） |
| `Heartbeat()` | 发送心跳注释，保持连接活跃 |
| `StartHeartbeat(interval)` | 启动定时心跳 goroutine，返回停止函数 |
| `Close()` | 关闭 SSE 流 |
| `Done()` | 返回关闭通知 channel |

## 完整示例

```go
package main

import (
    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/core"
    "github.com/chuccp/go-web-frame/web"
    "github.com/chuccp/go-web-frame/config"
    "go.uber.org/zap"
)

func main() {
    // 加载配置
    fileConfig, err := config.LoadSingleFileConfig("application.yml")
    if err != nil {
        zap.L().Fatal("加载配置失败", zap.Error(err))
    }

    // 创建 Builder
    builder := wf.NewBuilder(fileConfig)
    
    // 基本路由
    builder.Get("/", func(req *web.Request) (any, error) {
        return "Welcome!", nil
    })
    
    // 路径参数
    builder.Get("/users/:id", func(req *web.Request) (any, error) {
        id := req.Param("id")
        return map[string]any{"id": id}, nil
    })
    
    // 查询参数
    builder.Get("/search", func(req *web.Request) (any, error) {
        q := req.Query("q")
        return map[string]any{"keyword": q}, nil
    })
    
    // JSON 请求体
    builder.Post("/users", func(req *web.Request) (any, error) {
        var user struct {
            Name string `json:"name"`
        }
        if err := req.BindJSON(&user); err != nil {
            return nil, err
        }
        return map[string]any{"name": user.Name}, nil
    })
    
    // 构建应用
    app := builder.Build()
    
    // 启动应用
    err = app.Start()
    if err != nil {
        zap.L().Fatal("启动应用失败", zap.Error(err))
    }
}
```

## 下一步

- [控制器](controller.md) - 使用 REST 控制器组织代码
- [过滤器/中间件](filter.md) - 添加横切关注点

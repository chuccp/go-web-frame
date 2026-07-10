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
    keyword := req.Query("q")       // /search?q=go
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

## 路由元数据（WithMeta）

通过 `.WithMeta()` 给路由绑定元数据，在全局 Filter 中集中处理：

```go
func RequireAuth() web.MetaOption      { return web.WithValue("require_auth", true) }
func SkipAuth() web.MetaOption          { return web.WithValue("skip_auth", true) }
func RequirePermission(p string) web.MetaOption { return web.WithValue("require_permission", p) }

func (c *ApiController) Init(ctx *core.Context) error {
    ctx.Get("/api/login", c.Login).WithMeta(SkipAuth())
    ctx.Get("/api/profile", c.Profile).WithMeta(RequireAuth())
    ctx.Post("/api/admin/users", c.CreateUser).
        WithMeta(RequireAuth(), RequirePermission("admin:create_user"))
    return nil
}
```

详见 [过滤器](filter.md) 文档中的路由元数据章节。

## 下一步

- [控制器](controller.md) - 使用 REST 控制器组织代码
- [过滤器](filter.md) - 添加横切关注点
- [WebSocket 与 SSE](../advanced/websocket-sse.md) - 实时通信
- [静态文件与代理](../advanced/static-proxy.md) - 静态文件和反向代理

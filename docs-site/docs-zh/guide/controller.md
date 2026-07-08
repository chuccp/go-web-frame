# 控制器

控制器负责处理 HTTP 请求并返回响应。

## REST 控制器

嵌入 `core.IService` 接口来创建结构化的 REST 控制器。

### 基本结构

```go
package main

import (
    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/core"
    "github.com/chuccp/go-web-frame/web"
    auth2 "github.com/chuccp/go-web-frame/component/auth"
)

type UserController struct {
    core.IService  // 嵌入 IService 接口
    userService *UserService
}

// Init 在控制器初始化时注册路由
func (c *UserController) Init(context *core.Context) error {
    // 获取依赖的服务
    c.userService = wf.GetService[*UserService](context)
    
    // 直接注册路由（不带元数据）
    context.Get("/users", c.List)
    context.Get("/users/:id", c.Get)
    context.Post("/users", c.Create)
    
    // 带元数据注册路由
    context.Put("/users/:id", c.Update).WithMeta(auth2.WithLogin())
    context.Delete("/users/:id", c.Delete).WithMeta(auth2.WithLogin())
    
    return nil
}

func (c *UserController) List(req *web.Request) (any, error) {
    // 获取用户列表
    return []map[string]any{
        {"id": float64(1), "name": "Alice"},
        {"id": float64(2), "name": "Bob"},
    }, nil
}

func (c *UserController) Get(req *web.Request) (any, error) {
    id := req.ParamInt("id")
    return map[string]any{"id": float64(id), "name": "Alice"}, nil
}

func (c *UserController) Create(req *web.Request) (any, error) {
    var input struct {
        Name string `json:"name"`
    }
    if err := req.BindJSON(&input); err != nil {
        return nil, err
    }
    return map[string]any{"id": float64(3), "name": input.Name}, nil
}

func (c *UserController) Update(req *web.Request) (any, error) {
    id := req.ParamInt("id")
    return map[string]any{"id": float64(id), "updated": true}, nil
}

func (c *UserController) Delete(req *web.Request) (any, error) {
    id := req.ParamInt("id")
    return map[string]any{"id": float64(id), "deleted": true}, nil
}
```

## 请求处理

### 获取请求数据

```go
func (c *UserController) Handle(req *web.Request) (any, error) {
    // 路径参数
    id := req.Param("id")           // 字符串
    idInt := req.ParamInt("id")     // 整数
    idUint := req.ParamUint("id")   // 无符号整数

    // 查询参数
    keyword := req.Query("q")       // /users?q=go

    // JSON 请求体
    var data map[string]any
    if err := req.BindJSON(&data); err != nil {
        return nil, err
    }

    // 直接获取 JSON 值
    jsonObject, err := req.Json()
    if err != nil {
        return nil, err
    }
    name := jsonObject.GetString("name")
    age := jsonObject.GetInt("age")

    // 便捷方法：直接从请求体获取 JSON 值（无需先解析 Json()）
    name, _ := req.GetJsonStringValue("name")
    name, _ := req.GetJsonStringValueOrDefault("name", "default")
    age, _ := req.GetJsonIntValue("age")
    age, _ := req.GetJsonIntValueOrDefault("age", 0)

    return "ok", nil
}
```

### 请求头

```go
func (c *UserController) Handle(req *web.Request) (any, error) {
    // 获取请求头
    auth := req.GetHeader("Authorization")
    userAgent := req.GetHeader("User-Agent")

    // 客户端 IP
    ip := req.ClientIP()

    // HTTP 方法
    method := req.Request().Method

    return map[string]any{
        "auth":      auth,
        "userAgent": userAgent,
        "ip":        ip,
        "method":    method,
    }, nil
}
```

### Cookie 操作

```go
func (c *UserController) Handle(req *web.Request) (any, error) {
    cookie := req.Cookie()

    // 获取 Cookie
    sessionId := cookie.Get("session_id")

    // 设置 Cookie
    cookie.Set("token", "xxx", 3600)  // name, value, maxAge

    return "ok", nil
}
```

### 分页

```go
func (c *UserController) List(req *web.Request) (any, error) {
    // 自动从 GET/POST 参数中检测分页参数
    page, err := req.Page()
    if err != nil {
        return nil, err
    }

    // page.PageNo  - 页码（从 1 开始）
    // page.PageSize - 每页数量
    // page.LastId  - 最后一条记录的 ID（用于游标分页）

    return map[string]any{
        "pageNo":   page.PageNo,
        "pageSize": page.PageSize,
    }, nil
}
```

## 响应类型

### JSON 响应（默认）

```go
func (c *UserController) GetJSON(req *web.Request) (any, error) {
    return map[string]any{"key": "value"}, nil
}
```

### 字符串响应

```go
func (c *UserController) GetString(req *web.Request) (any, error) {
    return "plain text response", nil
}
```

### 文件下载

```go
func (c *UserController) Download(req *web.Request) (any, error) {
    return &web.File{
        Path:     "/path/to/file.pdf",
        FileName: "document.pdf",
    }, nil
}
```

### 重定向

```go
func (c *UserController) Redirect(req *web.Request) (any, error) {
    return web.Redirect("/new-url"), nil
}
```

### 错误响应

```go
func (c *UserController) Error(req *web.Request) (any, error) {
    return nil, errors.New("something went wrong")
}
```

### 统一响应格式（web.Message）

框架提供 `web.Message` 类型，用于构建统一格式的 API 响应：

```go
type Message struct {
    Code int    `json:"code"`
    Data any    `json:"data"`
    Msg  string `json:"msg"`
    Type string `json:"type"`
}
```

#### 辅助函数

```go
// 成功响应
return web.Ok(), nil                    // {"code":200,"data":null,"msg":"","type":""}
return web.Ok("操作成功"), nil           // {"code":200,"data":null,"msg":"操作成功","type":""}

// 带数据的成功响应
return web.Data(users), nil             // {"code":200,"data":[...],"msg":"","type":""}

// 带类型标签的响应
return web.DataType("table", data), nil // {"code":200,"data":...,"type":"table"}

// 带自定义状态码的响应
return web.DataCode(201, user), nil     // {"code":201,"data":{...},"msg":"","type":""}

// 错误响应 — 返回 *ErrorCode，converter 自动识别状态码
return nil, web.NewInternalError().WithDetail("服务器错误")
return nil, web.NewInternalError().WithError(err)
return nil, web.NewUnauthorized().WithError(err)

// 检查响应是否成功
msg := web.Data(users)
if msg.IsOK() { /* code == 200 */ }
```

#### 在控制器中使用

```go
func (c *UserController) List(req *web.Request) (any, error) {
    users, err := c.userService.GetAllUsers()
    if err != nil {
        return nil, err
    }
    return web.Data(users), nil
}

func (c *UserController) Create(req *web.Request) (any, error) {
    var input CreateUserInput
    if err := req.BindJSON(&input); err != nil {
        return nil, err
    }
    if err := c.userService.Create(&input); err != nil {
        return nil, err
    }
    return web.Ok("创建成功"), nil
}
```

### 自定义响应转换器（IConverter）

`DefaultConverter` 是框架默认的响应转换器，处理逻辑如下：

| handler 返回值 | 响应行为 |
|----------------|----------|
| `error` | 返回 HTTP 500，`*ErrorCode` 自动识别状态码 |
| `*web.Message` | 如果是 Redirect 类型则 301 跳转，否则返回 JSON |
| `string` | 直接写入纯文本 |
| `*web.File` / `*os.File` | 文件下载 |
| 其他值 | 包装为 `web.Data(result)` 返回 JSON |

如果需要统一所有 API 的响应格式（例如强制所有响应都走 `web.Message`），可以实现 `core.IConverter` 接口：

```go
type APIConverter struct {
    core.IConverter
}

func (c *APIConverter) Init(ctx *core.Context) error {
    return nil
}

func (c *APIConverter) Request(fc web.FilterChain, req *web.Request) {
    result, err := fc.Next()
    if err != nil {
        // 所有错误统一返回 Message 格式
        req.Response().AbortWithStatusJSON(500, &web.Message{Code: 500, Msg: err.Error()})
        return
    }
    // 如果已经是 Message 类型，直接返回
    if msg, ok := result.(*web.Message); ok {
        req.Response().AbortWithStatusJSON(200, msg)
        return
    }
    // 其他类型包装为 Message
    req.Response().AbortWithStatusJSON(200, web.Data(result))
}
```

注册转换器到 RestGroup：

```go
restGroup := wf.NewRestGroupBuilder().
    Rest(&UserController{}).
    Converter(&APIConverter{}).
    Port(8081).
    Build()
builder.RestGroup(restGroup)
```

## 依赖注入

控制器可以通过 `Init` 方法的 `ctx` 参数获取其他组件：

```go
type UserController struct {
    core.IService
    userService *UserService
}

func (c *UserController) Init(ctx *core.Context) error {
    // 获取服务
    c.userService = wf.GetService[*UserService](ctx)

    // 注册路由
    ctx.Get("/users", c.List)
    ctx.Get("/users/:id", c.Get)
    
    return nil
}

func (c *UserController) List(req *web.Request) (any, error) {
    // 使用服务
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
```

## 完整示例

```go
package main

import (
    "errors"
    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/config"
    "github.com/chuccp/go-web-frame/core"
    "github.com/chuccp/go-web-frame/web"
)

type UserController struct {
    core.IService
    userService *UserService
}

func (c *UserController) Init(ctx *core.Context) error {
    // 获取服务
    c.userService = wf.GetService[*UserService](ctx)

    // 注册路由
    ctx.Get("/users", c.List)
    ctx.Get("/users/:id", c.Get)
    ctx.Post("/users", c.Create)
    ctx.Put("/users/:id", c.Update)
    ctx.Delete("/users/:id", c.Delete)
    
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
    if user == nil {
        return nil, errors.New("user not found")
    }
    return user, nil
}

func (c *UserController) Create(req *web.Request) (any, error) {
    var input struct {
        Name  string `json:"name" validate:"required"`
        Email string `json:"email" validate:"required,email"`
    }
    if err := req.BindJSON(&input); err != nil {
        return nil, err
    }
    user, err := c.userService.CreateUser(input.Name, input.Email)
    if err != nil {
        return nil, err
    }
    return user, nil
}

func main() {
    // 方式一：使用 Builder（推荐）
    cfg, _ := config.LoadConfig("application.yml")
    builder := wf.NewBuilder(cfg)
    builder.Service(&UserService{})
    builder.Rest(&UserController{})
    app := builder.Build()
    app.Start()
}
```

## 下一步

- [模型](model.md) - 使用类型安全的 ORM
- [服务](service.md) - 业务逻辑层
- [过滤器/中间件](filter.md) - 添加横切关注点
- [后台任务](runner.md) - Runner 和定时任务

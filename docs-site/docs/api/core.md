# 核心 API 参考

本文档介绍 Go Web Frame 的核心 API。

## WebFrame

`WebFrame` 是应用的主入口点。

### 工厂方法

框架使用 `Builder` 模式创建应用：

```go
// 加载配置文件
cfg, err := config.LoadSingleFileConfig("application.yml")
if err != nil {
    // 处理错误
}

// 使用 Builder 和自定义配置创建应用
builder := wf.NewBuilder(cfg)
app := builder.Build()
```

### 注册方法

注册方法在 `Builder` 上调用，然后通过 `Build()` 构建 `WebFrame`：

```go
builder := wf.NewBuilder(config)

// 注册 REST 控制器
builder.Rest(rest ...core.IRest)

// 注册服务
builder.Service(services ...core.IService)

// 注册过滤器
builder.Filter(filters ...core.IFilter)

// 注册模型
builder.Model(models ...core.IModel)

// 注册模型组
builder.ModelGroup(modelGroups ...core.IModelGroup)

// 注册 REST 组
builder.RestGroup(restGroups ...*core.RestGroup)

// 注册后台任务
builder.Runner(runners ...core.IRunner)

// 注册组件
builder.Component(components ...core.IComponent)

// 构建应用
app := builder.Build()
```

### 启动方法

```go
// 启动应用（阻塞直到关闭）
err := app.Start()

// 使用 context 启动应用（支持优雅关闭）
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
err := app.Run(ctx)
```

`Run(ctx)` 支持通过 context 取消来优雅关闭 HTTP 服务器和后台 Runner。

### 测试方法

`Test()` 用于在不启动 HTTP 服务器的情况下初始化应用，适合编写单元测试。它会执行完整的初始化流程（加载配置、初始化模型、服务、组件等），并返回一个可用的 `*core.Context` 供测试使用。

```go
// 基本用法
err := app.Test(func(ctx *core.Context) error {
    // 在初始化的上下文中运行测试逻辑
    service := wf.GetService[*UserService](ctx)
    user, err := service.GetUserById(1)
    if err != nil {
        return err
    }
    // 断言...
    return nil
})
```

#### 完整测试示例

```go
package main

import (
    "testing"

    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/config"
    "github.com/stretchr/testify/assert"
    "myapp/model"
    "myapp/service"
)

func TestUserService(t *testing.T) {
    // 加载测试配置
    cfg, err := config.LoadSingleFileConfig("test.yml")
    if err != nil {
        t.Fatal(err)
    }

    // 构建应用
    builder := wf.NewBuilder(cfg)
    builder.Model(&model.UserModel{})
    builder.Service(&service.UserService{})
    app := builder.Build()

    // 使用 Test 方法初始化上下文（不启动 HTTP 服务器）
    err = app.Test(func(ctx *core.Context) error {
        // 从上下文中获取服务
        userService := wf.GetService[*service.UserService](ctx)

        // 测试业务逻辑
        user, err := userService.GetUserById(1)
        assert.NoError(t, err)
        assert.NotNil(t, user)
        assert.Equal(t, "test_user", user.Name)

        return nil
    })
    assert.NoError(t, err)
}
```

#### 注意事项

- `Test()` 不会启动 HTTP 服务器和后台 Runner，仅完成依赖注入和组件初始化
- 在测试函数内部可以正常访问所有服务、模型和组件
- 如果需要数据库操作，确保测试配置指向测试数据库
- 多个测试用例应使用独立的测试配置，避免数据互相干扰

## Context

`core.Context` 是依赖注入容器。

### 获取组件

```go
// 获取服务
service := wf.GetService[*UserService](ctx)

// 获取模型
model := wf.GetModel[*UserModel](ctx)

// 获取组件
component := wf.GetComponent[*CacheComponent](ctx)

// 获取任务
runner := wf.GetRunner[*EmailRunner](ctx)

// 获取过滤器
filter := wf.GetFilter[*AuthFilter](ctx)
```

### 注册路由

```go
// 在控制器中注册路由
func (c *UserController) Init(context *core.Context) error {
    context.Get("/users", c.List)
    context.Get("/users/:id", c.Get)
    context.Post("/users", c.Create)
    context.Put("/users/:id", c.Update)
    context.Delete("/users/:id", c.Delete)
    return nil
}
```

### 获取配置

```go
// 获取配置值
value := ctx.GetConfig().GetString("key")
value := ctx.GetConfig().GetInt("key")
value := ctx.GetConfig().GetBool("key")
```

### 获取事务

```go
// 获取默认模型组的事务
tx := ctx.GetTransaction()

// 获取命名模型组的事务
tx := ctx.GetTransactionByName("user_group")
```

## Request

`web.Request` 提供请求处理方法。

### 路径参数

```go
// 获取路径参数（字符串）
id := req.Param("id")

// 获取路径参数（整数）
idInt := req.ParamInt("id")

// 获取路径参数（无符号整数）
idUint := req.ParamUint("id")
```

### 查询参数

```go
// 获取查询参数
keyword := req.Query("q")

// 带默认值的查询参数
page := req.DefaultQuery("page", "1")

// 获取 JSON 请求体
jsonObject, err := req.Json()

// 绑定 JSON 到结构体
var data MyStruct
err := req.BindJSON(&data)
```

### JsonObject

`req.Json()` 返回 `JsonObject` 类型，提供便捷的 JSON 值获取方法：

```go
jsonObject, err := req.Json()
if err != nil {
    return nil, err
}

// 获取字符串
name := jsonObject.GetString("name")

// 获取整数
age := jsonObject.GetInt("age")

// 获取浮点数
price := jsonObject.GetFloat("price")

// 获取布尔值
active := jsonObject.GetBool("active")

// 获取数组
tags := jsonObject.GetSlice("tags")

// 获取嵌套对象
subObj := jsonObject.GetObject("address")
city := subObj.GetString("city")
```

### 请求头

```go
// 获取请求头
token := req.GetHeader("Authorization")

// 获取客户端 IP
ip := req.ClientIP()
```

### Cookie

```go
// 获取 Cookie 助手
cookie := req.Cookie()

// 获取 Cookie 值
value := cookie.Get("session_id")

// 设置 Cookie
cookie.Set("token", "xxx", 3600)  // name, value, maxAge
```

### 分页

```go
// 获取分页参数
page, err := req.Page()

// page.PageNo  - 页码（从 1 开始）
// page.PageSize - 每页数量
// page.LastId  - 最后一条记录的 ID（用于游标分页）
```

## Response

`web.Response` 提供响应处理方法。

### 设置响应

```go
// 设置状态码
req.Response().WriteHeader(200)

// 设置响应头
req.Response().Header().Set("Content-Type", "application/json")

// 写入响应体
req.Response().Write([]byte("Hello"))
```

### 快捷方法

```go
// 返回 JSON 响应
req.Response().AbortWithStatusJSON(200, data)

// 返回错误响应
req.Response().AbortWithError(err)
```

## 下一步

- [Web API](web.md) - 了解 Web 层 API
- [模型 API](model.md) - 了解模型层 API

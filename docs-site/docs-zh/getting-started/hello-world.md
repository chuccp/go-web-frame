# Hello World

本指南将带你创建第一个 Go Web Frame 应用。

## 最简单的应用

创建 `main.go`：

```go
package main

import (
    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/web"
    "github.com/chuccp/go-web-frame/config"
    "go.uber.org/zap"
)

func main() {
    // 加载配置
    cfg, err := config.LoadSingleFileConfig("application.yml")
    if err != nil {
        zap.L().Fatal("加载配置失败", zap.Error(err))
    }

    // 创建应用
    builder := wf.NewBuilder(cfg)
    
    // 注册路由
    builder.Get("/", func(req *web.Request) (any, error) {
        return "Hello, World!", nil
    })
    
    app := builder.Build()
    
    // 启动应用
    err = app.Start()
    if err != nil {
        zap.L().Fatal("启动失败", zap.Error(err))
    }
}
```

运行：

```bash
go run main.go
```

访问 `http://localhost:8081`，你会看到：

```json
"Hello, World!"
```

## 返回 JSON 响应

```go
builder.Get("/api/hello", func(req *web.Request) (any, error) {
    return map[string]any{
        "message": "Hello, World!",
        "status":  "success",
    }, nil
})
```

响应：

```json
{
  "message": "Hello, World!",
  "status": "success"
}
```

## 获取请求参数

### 路径参数

```go
builder.Get("/users/:id", func(req *web.Request) (any, error) {
    id := req.Param("id")         // 字符串
    idInt := req.ParamInt("id")   // 整数
    return map[string]any{"id": id, "idInt": idInt}, nil
})
```

访问 `/users/123`，返回：

```json
{
  "id": "123",
  "idInt": 123
}
```

### 查询参数

```go
builder.Get("/search", func(req *web.Request) (any, error) {
    keyword := req.Query("q")
    return map[string]any{
        "keyword": keyword,
    }, nil
})
```

访问 `/search?q=go`，返回：

```json
{
  "keyword": "go"
}
```

### JSON 请求体

```go
builder.Post("/users", func(req *web.Request) (any, error) {
    var user struct {
        Name  string `json:"name"`
        Email string `json:"email"`
    }
    if err := req.BindJSON(&user); err != nil {
        return nil, err
    }
    return map[string]any{
        "name":  user.Name,
        "email": user.Email,
    }, nil
})
```

发送 POST 请求：

```bash
curl -X POST http://localhost:8081/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com"}'
```

## 完整示例

```go
package main

import (
    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/config"
    "github.com/chuccp/go-web-frame/web"
    "go.uber.org/zap"
)

var users = []map[string]any{
    {"id": float64(1), "name": "Alice"},
    {"id": float64(2), "name": "Bob"},
}

func main() {
    cfg, err := config.LoadSingleFileConfig("application.yml")
    if err != nil {
        zap.L().Fatal("加载配置失败", zap.Error(err))
    }

    builder := wf.NewBuilder(cfg)

    // 获取所有用户
    builder.Get("/users", func(req *web.Request) (any, error) {
        return users, nil
    })

    // 获取单个用户
    builder.Get("/users/:id", func(req *web.Request) (any, error) {
        id := req.ParamInt("id")
        if id <= 0 || id > len(users) {
            return nil, nil
        }
        return users[id-1], nil
    })

    // 创建用户
    builder.Post("/users", func(req *web.Request) (any, error) {
        var user map[string]any
        if err := req.BindJSON(&user); err != nil {
            return nil, err
        }
        user["id"] = float64(len(users) + 1)
        users = append(users, user)
        return user, nil
    })

    app := builder.Build()
    app.Start()
}
```

## 下一步

- [路由](../guide/routing.md) - 深入了解路由系统
- [控制器](../guide/controller.md) - 使用 REST 控制器组织代码

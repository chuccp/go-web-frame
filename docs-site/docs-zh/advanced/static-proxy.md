# 静态文件与代理

本文档介绍 Go Web Frame 中静态文件服务和反向代理的使用。

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

### 通过配置注册静态文件

在配置文件中指定静态文件目录：

```yaml
web:
  server:
    locations:              # 静态文件目录
      - ./view/dist         # Vue/React 构建产物
      - ./www               # 静态资源
    page404: index.html     # SPA 模式：未匹配路由返回 index.html
```

当 `page404` 设置为 `index.html` 时，所有未匹配的路由都会返回该文件，实现 SPA 前端路由支持。

## 反向代理

在控制器中注册反向代理，将请求转发到后端服务：

```go
func (c *MyController) Init(ctx *core.Context) error {
    // 所有 /api/* 请求会被代理到 http://backend:8080/api/*
    ctx.ReverseProxy("/api", "http://backend:8080")
    return nil
}
```

所有 `/api/*` 请求都会被代理到 `http://backend:8080/api/*`。

## 下一步

- [路由](../guide/routing.md) - 路由系统
- [WebSocket 与 SSE](websocket-sse.md) - 实时通信
- [部署](deployment.md) - 生产环境部署（含 HTTPS 配置）

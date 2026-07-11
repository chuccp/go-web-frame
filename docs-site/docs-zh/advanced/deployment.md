# 部署

本文档介绍 Go Web Frame 应用的部署相关配置，包括 HTTPS、优雅关闭等。

## 优雅关闭

使用 `Run(ctx)` 配合信号实现优雅关闭：

```go
package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"

    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/config"
    "github.com/chuccp/go-web-frame/web"
)

func main() {
    cfg, _ := config.LoadSingleFileConfig("application.yml")
    builder := wf.NewBuilder(cfg)
    builder.Get("/", func(req *web.Request) (any, error) {
        return "Hello", nil
    })
    app := builder.Build()

    // 创建带取消的 context
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // 监听系统信号
    go func() {
        sigCh := make(chan os.Signal, 1)
        signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
        <-sigCh
        // 收到信号，取消 context 触发优雅关闭
        cancel()
    }()

    // Run 会在 context 取消后优雅关闭 HTTP 服务器和后台 Runner
    err := app.Run(ctx)
    if err != nil {
        panic(err)
    }
}
```

`Run(ctx)` 与 `Start()` 的区别：

| 方法 | 行为 |
|------|------|
| `Start()` | 启动应用并阻塞，不支持优雅关闭 |
| `Run(ctx)` | 启动应用，当 context 取消时优雅关闭 HTTP 服务器和后台 Runner |

## HTTPS / SSL

### 配置文件方式

```yaml
web:
  server:
    port: 443
  ssl:
    enabled: true
    hosts:
      - example.com
      - www.example.com
```

| 配置项 | 说明 |
|--------|------|
| `web.ssl.enabled` | 是否启用 HTTPS |
| `web.ssl.hosts` | 证书域名列表（用于 Let's Encrypt 自动证书） |
| `web.ssl.certs` | 本地证书配置列表（用于已有证书文件） |

### Let's Encrypt 自动证书

框架内置 Let's Encrypt 自动证书管理。启用 SSL 后，`CertManager` 会自动为配置的域名申请和续期证书：

```yaml
web:
  server:
    port: 443
  ssl:
    enabled: true
    hosts:
      - yourdomain.com
```

启动后框架自动完成：
1. 向 Let's Encrypt 申请域名证书
2. 自动续期（证书快过期时自动更新）
3. HTTP 请求自动重定向到 HTTPS

> **注意**：使用 Let's Encrypt 自动证书需要服务器 80 和 443 端口对外可访问，且域名已解析到服务器 IP。

### 本地证书文件

如果已有证书文件（如从 CA 机构获取或自签名证书），可以通过 `certs` 直接配置：

```yaml
web:
  server:
    port: 443
    ssl:
      enabled: true
      certs:
        - host: example.com
          cert-file: /etc/ssl/example.com/fullchain.pem
          key-file: /etc/ssl/example.com/privkey.pem
        - host: api.example.com
          cert-file: /etc/ssl/api.example.com/fullchain.pem
          key-file: /etc/ssl/api.example.com/privkey.pem
```

每个 `certs` 条目将一个域名映射到其证书和私钥文件。框架根据请求的 TLS SNI（服务器名称指示）自动选择正确的证书。

### 混合模式：本地证书 + 自动证书兆底

可以组合使用两种方式——部分域名用本地证书，其他域名用 Let's Encrypt：

```yaml
web:
  server:
    port: 443
    ssl:
      enabled: true
      hosts:                      # 这些域名用 Let's Encrypt 自动申请
        - auto.example.com
      certs:                      # 这些域名用本地证书
        - host: example.com
          cert-file: /etc/ssl/example.com/fullchain.pem
          key-file: /etc/ssl/example.com/privkey.pem
```

证书匹配优先级：本地证书精确匹配 > 通配符匹配 > autocert 自动证书 > 自签证书兆底。

### 自签证书兆底

当没有证书匹配某个域名时（无本地证书、无通配符、无 autocert），框架会自动生成内存中的 ECDSA P-256 自签证书（有效期 1 年）并缓存。开发环境下无需配置外部证书。

- 支持 IPv6 地址，包括 HTTP Host header 中的方括号格式（`[::1]`、`[::1]:8443`）
- 生成的证书按主机名缓存，后续请求复用

### 证书加载容错

如果某个证书文件加载失败（如文件不存在），框架会记录错误并继续运行——其他证书仍正常加载，HTTP 服务器在无 TLS 模式下启动。单个证书路径错误不会阻塞整个应用。

### 手动管理证书

如果使用自己的证书，可以通过 `ServerConfig` 编程配置：

```go
import "github.com/chuccp/go-web-frame/web"

serverConfig := web.DefaultServerConfig()
serverConfig.Port = 443
serverConfig.SSL = &web.SSLConfig{
    Enabled: true,
    Hosts:   []string{"example.com"},
}

restGroup := wf.NewRestGroupBuilder().
    ServerConfig(serverConfig).
    Rest(&UserController{}).
    Build()
builder.RestGroup(restGroup)
```

### 同时运行 HTTP 和 HTTPS

可以注册两个 RestGroup，分别监听不同端口：

```go
// HTTPS（443）
httpsGroup := wf.NewRestGroupBuilder().
    Port(443).
    ServerConfig(sslServerConfig).
    Rest(&UserController{}).
    Build()

// HTTP（80）重定向到 HTTPS
httpGroup := wf.NewRestGroupBuilder().
    Port(80).
    Rest(&UserController{}).
    Build()

builder.RestGroup(httpsGroup, httpGroup)
```

## 服务器配置

### 完整配置项

```yaml
web:
  server:
    port: 8081              # 监听端口
    context_path: /api      # 路由前缀（类似 Tomcat 的 context-path）
    locations:              # 静态文件目录
      - ./view/dist
      - www
    page404: 404.html       # 404 页面
  ssl:
    enabled: false          # 是否启用 HTTPS
    hosts: []               # Let's Encrypt 域名列表
    certs: []               # 本地证书配置列表
```

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `web.server.port` | 监听端口 | 8081 |
| `web.server.context_path` | 路由前缀 | 无 |
| `web.server.locations` | 静态文件目录 | 无 |
| `web.server.page404` | 404 页面 | 无 |
| `web.ssl.enabled` | 是否启用 HTTPS | false |
| `web.ssl.hosts` | 证书域名列表 | 无 |
| `web.ssl.certs` | 本地证书配置列表 | 无 |

### Context Path

`context_path` 为所有路由添加统一前缀，类似 Java Tomcat 的 context-path：

```yaml
web:
  server:
    context_path: /myapp
```

此时 `/users` 路由的实际访问路径为 `/myapp/users`。

也可以在 RestGroup 级别设置：

```go
restGroup := wf.NewRestGroupBuilder().
    ContextPath("/api/v1").
    Rest(&UserController{}).
    Port(8081).
    Build()
```

### 静态文件目录（SPA 支持）

`locations` 配置静态文件目录，支持 SPA 应用：

```yaml
web:
  server:
    locations:
      - ./view/dist         # Vue/React 构建产物
      - ./www               # 静态资源
    page404: index.html     # SPA 模式：未匹配路由返回 index.html
```

当 `page404` 设置为 `index.html` 时，所有未匹配的路由都会返回该文件，实现 SPA 前端路由支持。

## 生产环境建议

1. **使用 `Run(ctx)` 配合信号** — 确保应用可以优雅关闭
2. **启用 HTTPS** — 生产环境必须使用 HTTPS
3. **配置日志级别为 info** — 避免大量 debug 日志影响性能
4. **配置数据库连接池** — 根据并发量调整 `max_open_conns` 和 `max_idle_conns`
5. **使用反向代理** — 如果需要 Nginx 等前置代理，框架内置 `ReverseProxy` 支持

## 下一步

- [数据库高级用法](database.md) - 事务、模型组等
- [最佳实践](../best-practices.md) - 推荐的使用模式

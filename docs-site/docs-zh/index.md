---
hide:
  - navigation
  - toc
---

# Go Web Frame

> 集成式 Go 后端开发工具包 — 零样板 CRUD、声明式路由元数据、显式依赖注入，并无缝复用 Gin 生态。

[5 分钟上手](getting-started/quick-start.md){ .md-button .md-button--primary }
[安装](getting-started/installation.md){ .md-button }
[源码 :simple-github:](https://github.com/chuccp/go-web-frame){ .md-button }

---

## 特性

- **:material-database-check: 零样板 CRUD** — `Model[T]` / `EntryModel[T, PK]` 消除 CRUD 样板，从数据库到 Handler 全链路类型安全，无需代码生成。
- **:material-shield-key: 声明式路由元数据** — 通过 `.WithMeta()` 给路由打标记，一个全局 Filter 集中处理认证、权限、限流等横切逻辑。
- **:material-link-variant: 显式依赖注入** — 用 Builder 模式显式注册组件，初始化顺序完全透明、可控，告别隐式扫描的"黑箱操作"。
- **:material-transit-connection-variant: Gin 生态兼容** — 封装自己的请求/响应抽象，同时暴露 `req.GinContext()`，CORS、Gzip 等中间件可直接复用。
- **:material-database-settings: GORM 生态兼容** — `db.GetGorm()` 暴露原生 `*gorm.DB`，任意 GORM 驱动、插件、回调、标签零成本接入。
- **:material-package-variant: 内置组件开箱即用** — 限流、缓存、验证码、定时任务、认证、CORS 等常用组件已预集成，一行代码即可启用。
- **:material-lock-check: 生产就绪** — 内置 Let's Encrypt 自动 HTTPS、优雅关停、连接池管理、结构化日志，直接部署到生产环境。

---

## 30 秒 Hello World

```bash
go get github.com/chuccp/go-web-frame
```

```go
package main

import (
    "context"
    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/config"
    "github.com/chuccp/go-web-frame/web"
)

func main() {
    cfg, _ := config.LoadSingleFileConfig("application.yml")
    builder := wf.NewBuilder(cfg)
    builder.Get("/", func(c *web.Request) (any, error) {
        return "Hello, World!", nil
    })
    builder.Build().Run(context.Background())
}
```

```bash
go run main.go
# → http://localhost:19009
```

## :material-compass: 快速导航

| | | |
|---|---|---|
| [:material-download: 安装](getting-started/installation.md) | [:material-rocket-launch: 5 分钟上手](getting-started/quick-start.md) | [:material-routes: 路由](guide/routing.md) |
| [:material-format-list-bulleted: 控制器](guide/controller.md) | [:material-database: 模型](guide/model.md) | [:material-server: 服务](guide/service.md) |
| [:material-shield: 过滤器](guide/filter.md) | [:material-swap-horizontal: 响应转换器](guide/converter.md) | [:material-alert-circle: 错误码](guide/error-code.md) |
| [:material-cog: 配置](guide/configuration.md) | [:material-file-document: 日志](guide/logging.md) | [:material-run: 后台任务](guide/runner.md) |
| [:material-package-variant: 内置组件](guide/components.md) | [:material-web: WebSocket & SSE](advanced/websocket-sse.md) | [:material-database-cog: 数据库高级](advanced/database.md) |
| [:material-rocket: 部署](advanced/deployment.md) | | |

## :material-layers: 技术栈

| 层级 | 库 | 作用 |
|---|---|---|
| HTTP | Gin | 路由、中间件链、参数绑定 |
| ORM | GORM | 底层 SQL 驱动、迁移、Join/Preload |
| 配置 | Viper | 多格式、多路径加载 |
| 日志 | Zap | 结构化、分级、文件轮转 |
| JSON | Sonic | 高性能序列化 |
| 缓存 | Otter | 本地内存缓存 |
| Redis | go-redis | 发布订阅、缓存 |
| SQLite | modernc/sqlite | 纯 Go、零 CGO |
| 校验 | go-playground/validator | 结构体标签校验 |
| WebSocket | coder/websocket | 连接升级与读写 |
| 定时任务 | robfig/cron | 表达式定时调度 |

## :simple-github: 社区

- [GitHub 仓库](https://github.com/chuccp/go-web-frame)
- [问题反馈](https://github.com/chuccp/go-web-frame/issues)
- [最佳实践](best-practices.md)
- [更新日志](changelog.md)

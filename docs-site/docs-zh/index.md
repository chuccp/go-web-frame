---
hide:
  - navigation
  - toc
---

# Go Web Frame

<div class="gw-hero" markdown>

# Go Web Frame

<p class="gw-subtitle">集成式 Go 后端开发工具包 — 零样板 CRUD、声明式路由元数据、显式依赖注入，并无缝复用 Gin 生态。</p>

<div class="gw-cta" markdown>
<a href="getting-started/quick-start.md" class="gw-cta-primary">:material-rocket-launch: 5 分钟上手</a>
<a href="getting-started/installation.md" class="gw-cta-secondary">:material-download: 安装</a>
<a href="https://github.com/chuccp/go-web-frame" class="gw-cta-secondary">:simple-github: 源码</a>
</div>

</div>

<div class="gw-features" markdown>

<a href="getting-started/quick-start.md" class="gw-feature-card" markdown>
<span class="gw-feature-icon">:material-database-check:</span>
#### 零样板 CRUD
用 `Model[T]` / `EntryModel[T, PK]` 消除 CRUD 样板，从数据库到 Handler 全链路类型安全，无需代码生成。
</a>

<a href="guide/filter.md" class="gw-feature-card" markdown>
<span class="gw-feature-icon">:material-shield-key:</span>
#### 声明式路由元数据
通过 `.WithMeta()` 给路由打标记，一个全局 Filter 集中处理认证、权限、限流等横切逻辑。
</a>

<a href="guide/service.md" class="gw-feature-card" markdown>
<span class="gw-feature-icon">:material-link-variant:</span>
#### 显式依赖注入
用 Builder 模式显式注册组件，初始化顺序完全透明、可控，告别隐式扫描的"黑箱操作"。
</a>

<a href="guide/routing.md" class="gw-feature-card" markdown>
<span class="gw-feature-icon">:material-transit-connection-variant:</span>
#### Gin 生态兼容
封装自己的请求/响应抽象，同时暴露 `req.GinContext()`，CORS、Gzip 等中间件可直接复用。
</a>

<a href="guide/components.md" class="gw-feature-card" markdown>
<span class="gw-feature-icon">:material-package-variant:</span>
#### 内置组件开箱即用
限流、缓存、验证码、定时任务、认证、CORS 等常用组件已预集成，一行代码即可启用。
</a>

<a href="advanced/deployment.md" class="gw-feature-card" markdown>
<span class="gw-feature-icon">:material-lock-check:</span>
#### 生产就绪
内置 Let's Encrypt 自动 HTTPS、优雅关停、连接池管理、结构化日志，直接部署到生产环境。
</a>

</div>

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

<div class="gw-quicknav" markdown>

<a href="getting-started/installation.md"><span class="gw-quicknav-icon">:material-download:</span> 安装</a>
<a href="getting-started/quick-start.md"><span class="gw-quicknav-icon">:material-rocket-launch:</span> 5 分钟上手</a>
<a href="guide/routing.md"><span class="gw-quicknav-icon">:material-routes:</span> 路由</a>
<a href="guide/controller.md"><span class="gw-quicknav-icon">:material-format-list-bulleted:</span> 控制器</a>
<a href="guide/model.md"><span class="gw-quicknav-icon">:material-database:</span> 模型</a>
<a href="guide/service.md"><span class="gw-quicknav-icon">:material-server:</span> 服务</a>
<a href="guide/filter.md"><span class="gw-quicknav-icon">:material-shield:</span> 过滤器</a>
<a href="guide/converter.md"><span class="gw-quicknav-icon">:material-swap-horizontal:</span> 响应转换器</a>
<a href="guide/error-code.md"><span class="gw-quicknav-icon">:material-alert-circle:</span> 错误码</a>
<a href="guide/configuration.md"><span class="gw-quicknav-icon">:material-cog:</span> 配置</a>
<a href="guide/logging.md"><span class="gw-quicknav-icon">:material-file-document:</span> 日志</a>
<a href="guide/runner.md"><span class="gw-quicknav-icon">:material-run:</span> 后台任务</a>
<a href="guide/components.md"><span class="gw-quicknav-icon">:material-package-variant:</span> 内置组件</a>
<a href="advanced/websocket-sse.md"><span class="gw-quicknav-icon">:material-web:</span> WebSocket 与 SSE</a>
<a href="advanced/database.md"><span class="gw-quicknav-icon">:material-database-cog:</span> 数据库高级</a>
<a href="advanced/deployment.md"><span class="gw-quicknav-icon">:material-rocket:</span> 部署</a>

</div>

## :material-layers: 技术栈

<div class="gw-techstack" markdown>
<span class="gw-badge">:simple-go: Gin</span>
<span class="gw-badge">:material-database: GORM</span>
<span class="gw-badge">:material-file-cog: Viper</span>
<span class="gw-badge">:material-lightning-bolt: Zap</span>
<span class="gw-badge">:material-code-json: Sonic</span>
<span class="gw-badge">:material-database-sync: go-redis</span>
<span class="gw-badge">:simple-sqlite: modernc/sqlite</span>
<span class="gw-badge">:material-check-circle: validator</span>
<span class="gw-badge">:material-web: coder/websocket</span>
<span class="gw-badge">:material-clock-time-four: robfig/cron</span>
</div>

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

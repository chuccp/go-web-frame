---
hide:
  - navigation
  - toc
---

# Go Web Frame

<div class="gw-hero" markdown>

# Go Web Frame

<p class="gw-subtitle">An integrated Go backend toolkit — zero-boilerplate CRUD, declarative route metadata, explicit dependency injection, and seamless reuse of the Gin ecosystem.</p>

<div class="gw-cta" markdown>
<a href="getting-started/quick-start.md" class="gw-cta-primary">:material-rocket-launch: Quick Start</a>
<a href="getting-started/installation.md" class="gw-cta-secondary">:material-download: Installation</a>
<a href="https://github.com/chuccp/go-web-frame" class="gw-cta-secondary">:simple-github: Source</a>
</div>

</div>

<div class="gw-features" markdown>

<a href="getting-started/quick-start.md" class="gw-feature-card" markdown>
<span class="gw-feature-icon">:material-database-check:</span>
#### Zero-Boilerplate CRUD
`Model[T]` / `EntryModel[T, PK]` eliminate CRUD boilerplate with full type safety from database to handler — no code generation.
</a>

<a href="guide/filter.md" class="gw-feature-card" markdown>
<span class="gw-feature-icon">:material-shield-key:</span>
#### Declarative Route Metadata
Tag routes with `.WithMeta()` and handle auth, permissions, and rate limiting in one global filter.
</a>

<a href="guide/service.md" class="gw-feature-card" markdown>
<span class="gw-feature-icon">:material-link-variant:</span>
#### Explicit Dependency Injection
Register components through a fluent Builder — initialization order is fully transparent and controllable.
</a>

<a href="guide/routing.md" class="gw-feature-card" markdown>
<span class="gw-feature-icon">:material-transit-connection-variant:</span>
#### Gin Ecosystem Compatible
Wraps HTTP requests/responses while exposing `req.GinContext()` — CORS, Gzip, and other middlewares work out of the box.
</a>

<a href="guide/components.md" class="gw-feature-card" markdown>
<span class="gw-feature-icon">:material-package-variant:</span>
#### Built-in Components
Rate limiting, caching, captcha, cron scheduling, authentication, CORS — all pre-integrated and ready to use.
</a>

<a href="advanced/deployment.md" class="gw-feature-card" markdown>
<span class="gw-feature-icon">:material-lock-check:</span>
#### Production Ready
Let's Encrypt auto HTTPS, graceful shutdown, connection pooling, structured logging — deploy to production with confidence.
</a>

</div>

## 30-Second Hello World

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

## :material-compass: Quick Navigation

<div class="gw-quicknav" markdown>

<a href="getting-started/installation.md"><span class="gw-quicknav-icon">:material-download:</span> Installation</a>
<a href="getting-started/quick-start.md"><span class="gw-quicknav-icon">:material-rocket-launch:</span> Quick Start</a>
<a href="guide/routing.md"><span class="gw-quicknav-icon">:material-routes:</span> Routing</a>
<a href="guide/controller.md"><span class="gw-quicknav-icon">:material-format-list-bulleted:</span> Controller</a>
<a href="guide/model.md"><span class="gw-quicknav-icon">:material-database:</span> Model</a>
<a href="guide/service.md"><span class="gw-quicknav-icon">:material-server:</span> Service</a>
<a href="guide/filter.md"><span class="gw-quicknav-icon">:material-shield:</span> Filter</a>
<a href="guide/converter.md"><span class="gw-quicknav-icon">:material-swap-horizontal:</span> Converter</a>
<a href="guide/error-code.md"><span class="gw-quicknav-icon">:material-alert-circle:</span> Error Code</a>
<a href="guide/configuration.md"><span class="gw-quicknav-icon">:material-cog:</span> Configuration</a>
<a href="guide/logging.md"><span class="gw-quicknav-icon">:material-file-document:</span> Logging</a>
<a href="guide/runner.md"><span class="gw-quicknav-icon">:material-run:</span> Runner</a>
<a href="guide/components.md"><span class="gw-quicknav-icon">:material-package-variant:</span> Components</a>
<a href="advanced/websocket-sse.md"><span class="gw-quicknav-icon">:material-web:</span> WebSocket & SSE</a>
<a href="advanced/database.md"><span class="gw-quicknav-icon">:material-database-cog:</span> Database</a>
<a href="advanced/deployment.md"><span class="gw-quicknav-icon">:material-rocket:</span> Deployment</a>

</div>

## :material-layers: Tech Stack

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

| Layer | Library | Role |
|---|---|---|
| HTTP | Gin | Router, middleware chain, parameter binding |
| ORM | GORM | Underlying SQL driver, migrations, joins/preload |
| Config | Viper | Multi-format, multi-path loading |
| Logging | Zap | Structured, leveled, rotated logging |
| JSON | Sonic | High-performance marshal/unmarshal |
| Cache | Otter | Local in-memory cache |
| Redis | go-redis | Pub/sub, caching |
| SQLite | modernc/sqlite | Pure Go, zero CGO |
| Validation | go-playground/validator | Struct tag validation |
| WebSocket | coder/websocket | Upgrade, read/write |
| Cron | robfig/cron | Expression-based scheduling |

## :simple-github: Community

- [GitHub Repository](https://github.com/chuccp/go-web-frame)
- [Issue Tracker](https://github.com/chuccp/go-web-frame/issues)
- [Best Practices](best-practices.md)
- [Changelog](changelog.md)

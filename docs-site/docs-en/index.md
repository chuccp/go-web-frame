---
hide:
  - navigation
  - toc
---

# Go Web Frame

> An integrated Go backend toolkit — zero-boilerplate CRUD, declarative route metadata, explicit dependency injection, and seamless reuse of the Gin ecosystem.

[Quick Start](getting-started/quick-start.md){ .md-button .md-button--primary }
[Installation](getting-started/installation.md){ .md-button }
[Source :simple-github:](https://github.com/chuccp/go-web-frame){ .md-button }

---

## Features

- **:material-database-check: Zero-Boilerplate CRUD** — `Model[T]` / `EntryModel[T, PK]` eliminate CRUD boilerplate with full type safety from database to handler — no code generation.
- **:material-shield-key: Declarative Route Metadata** — Tag routes with `.WithMeta()` and handle auth, permissions, and rate limiting in one global filter.
- **:material-link-variant: Explicit Dependency Injection** — Register components through a fluent Builder — initialization order is fully transparent and controllable.
- **:material-transit-connection-variant: Gin Ecosystem Compatible** — Wraps HTTP requests/responses while exposing `req.GinContext()` — CORS, Gzip, and other middlewares work out of the box.
- **:material-package-variant: Built-in Components** — Rate limiting, caching, captcha, cron scheduling, authentication, CORS — all pre-integrated and ready to use.
- **:material-lock-check: Production Ready** — Let's Encrypt auto HTTPS, graceful shutdown, connection pooling, structured logging — deploy to production with confidence.

---

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

| | | |
|---|---|---|
| [:material-download: Installation](getting-started/installation.md) | [:material-rocket-launch: Quick Start](getting-started/quick-start.md) | [:material-routes: Routing](guide/routing.md) |
| [:material-format-list-bulleted: Controller](guide/controller.md) | [:material-database: Model](guide/model.md) | [:material-server: Service](guide/service.md) |
| [:material-shield: Filter](guide/filter.md) | [:material-swap-horizontal: Converter](guide/converter.md) | [:material-alert-circle: Error Code](guide/error-code.md) |
| [:material-cog: Configuration](guide/configuration.md) | [:material-file-document: Logging](guide/logging.md) | [:material-run: Runner](guide/runner.md) |
| [:material-package-variant: Components](guide/components.md) | [:material-web: WebSocket & SSE](advanced/websocket-sse.md) | [:material-database-cog: Database](advanced/database.md) |
| [:material-rocket: Deployment](advanced/deployment.md) | | |

## :material-layers: Tech Stack

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

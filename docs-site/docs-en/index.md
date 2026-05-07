# Go Web Frame User Guide

Welcome to Go Web Frame — a modern, feature-rich Go web framework.

## What is Go Web Frame?

Go Web Frame combines the best open-source components from the Go ecosystem: Gin (HTTP), GORM (ORM), Viper (config), Zap (logging), Sonic (JSON), Otter (cache), and more. It provides declarative route metadata, type-safe generic ORM, and dependency injection out of the box.

**Key Features:**
- **WithMeta**: Declare metadata per route (auth, permissions, rate limit flags) and handle uniformly in filters
- **Builder Pattern**: Explicit registration, controllable initialization order, no implicit scanning
- **Generic ORM**: Zero-boilerplate CRUD with `Model[T]`, no code generation
- **Transparent Context**: All dependencies injectable via `GetService[T]`, debug-friendly

## Tech Stack

### Core Framework
| Component | Description |
|------|------|
| [Gin](https://github.com/gin-gonic/gin) | High-performance HTTP web framework |
| [GORM](https://gorm.io/) | Powerful ORM library with multi-database support |
| [Viper](https://github.com/spf13/viper) | Complete configuration solution |
| [Zap](https://go.uber.org/zap) | Uber's high-performance structured logging library |

### Data Storage
| Component | Description |
|------|------|
| [gorm/driver/mysql](https://gorm.io/driver/mysql/) | GORM MySQL driver |
| [gorm/driver/postgres](https://gorm.io/driver/postgres/) | GORM PostgreSQL driver |
| [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) | Pure Go SQLite implementation, no CGO dependency |
| [go-redis](https://github.com/redis/go-redis) | Redis client recommended by Redis |

## Quick Links

### Getting Started

- [Installation](getting-started/installation.md) - Requirements and installation
- [Quick Start](getting-started/quick-start.md) - Create your first application
- [Hello World](getting-started/hello-world.md) - The simplest example

### User Guide

- [Routing](guide/routing.md) - HTTP routing system (path params, query params, REST controllers, static files, reverse proxy, WebSocket, SSE)
- [Controller](guide/controller.md) - REST controllers and request handling
- [Service](guide/service.md) - Business logic layer and dependency injection
- [Model](guide/model.md) - Type-safe ORM (Model, EntryModel, query builder, transactions)
- [Filter/Middleware](guide/filter.md) - HTTP request filtering (auth, logging, CORS, rate limiting, route metadata)
- [Configuration](guide/configuration.md) - Configuration management (YAML/JSON/TOML, environment variables, multi-environment)
- [Logging](guide/logging.md) - Structured logging (Zap, file rotation, log levels)
- [Runner](guide/runner.md) - Runner and scheduled task management
- [Components](guide/components.md) - Built-in framework components (rate limiting, auth, cron, cache, captcha, etc.)

### Advanced Topics

- [Database](advanced/database.md) - Transactions, model groups, migrations, raw SQL
- [Deployment](advanced/deployment.md) - HTTPS, SSL certificates, graceful shutdown, production configuration

### API Reference

- [Core API](api/core.md) - WebFrame, Builder, Context, Request, Response
- [Web API](api/web.md) - HandlerFunc, route registration, response types, web.Message
- [Model API](api/model.md) - Model, EntryModel, Query, Transaction
- [Util API](api/util.md) - String, crypto, file, network, time utilities

### Best Practices

- [Best Practices](best-practices.md) - Recommended project structure, layer separation, error handling, auth, testing

## Main Features

- **Dependency Injection** - Type-safe DI container based on Context
- **Type-safe ORM** - Zero-boilerplate generic Model
- **Flexible Configuration** - Multi-location, multi-format config files (YAML/JSON/TOML)
- **Component System** - Reusable standalone components (rate limiting, auth, cron, etc.)
- **HTTPS Auto Certificate** - Integrated Let's Encrypt

## Community

- [GitHub](https://github.com/chuccp/go-web-frame)
- [Issue Tracker](https://github.com/chuccp/go-web-frame/issues)
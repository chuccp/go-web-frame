# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Development Commands

### Build and Run
```bash
# Run the hello world example
go run example/helloworld/helloworld.go

# Run the rest example
go run example/rest/rest.go

# Run the ORM model example
go run example/model/model.go

# Build the framework (library only)
go build

# Run application as a daemon/service
# (Requires implementing AppService interface)
go run your_app.go
# Stop the daemon
go run your_app.go -stop
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests in a specific package
go test ./core
go test ./web

# Run tests with verbose output
go test -v ./core

# Run a specific test case
go test -v ./core -run TestSpecificFunction
```

### Formatting
```bash
# Format all code with gofmt
gofmt -w ./...

# Alternative formatting with gofumpt (if installed)
gofumpt -w ./...
```

### Linting
```bash
# Install linter (if not already installed)
go install golang.org/x/lint/golint@latest

# Run linter
golint ./...
```

### Dependency Management
```bash
# Add a new dependency
go get github.com/example/package

# Update dependencies
go get -u ./...

# Tidy up go.mod and go.sum
go mod tidy
```

## High-Level Architecture

### Overview
This is a Go web framework built on top of Gin, providing a structured approach to building web applications with:
- Dependency injection
- MVC-like architecture
- Type-safe generic ORM with zero boilerplate
- Database integration (SQLite, MySQL, Redis)
- Component-based system
- Daemon/service mode for production deployment

### Core Components

#### 1. Core Abstractions (./core)
The framework is built around several key interfaces that define the component model:
- `IService`: Base interface for services that need initialization
- `IModel`: Data access layer interface with CRUD and table management
- `IRest`: REST controller interface (extends `IService`)
- `IComponent`: Independent components that initialize with config
- `IRunner`: Background task runners (extends `IService` and `IRun`)
- `IFilter`: HTTP request filters (extends `IService` and `web.Filter`)

#### 2. Web Layer (./web)
Built on Gin, provides:
- Request/Response handling
- Routing with HTTP method support (GET, POST, PUT, DELETE, etc.)
- Filter/middleware support
- Conversion between service responses and HTTP responses

#### 3. Data Access
- `./db`: Database abstraction layer supporting multiple databases (MySQL, SQLite) using GORM
- `./model`: Type-safe generic base model implementation with zero-boilerplate CRUD operations
- `./sqlite`: SQLite-specific configuration and initialization
- `./redis`: Redis integration for caching and messaging

#### 4. Other Key Packages
- `./config`: Configuration management with Viper, supports JSON, YAML, TOML
- `./log`: Structured logging based on Zap with rotation support
- `./component`: Reusable components: cache, local cache, rate limiting, captcha, QR code, cron scheduled tasks, input validation
- `./util`: Comprehensive utility functions for strings, time, crypto, networking, and more

### Application Structure
A typical application using this framework will:
1. Create a `WebFrame` instance with `NewWithAutoConfig()` or `New(config)`
2. Register routes directly or add REST controllers
3. Add models, services, components, and runners
4. Start the application with `Run(ctx)` or `Start()`

### Key Entry Points
- `web_frame.go`: Main package entry point with factory methods (`NewWithAutoConfig()`, `New()`) and registration methods
- `core/context.go`: Core context for dependency injection - all components initialize through this context
- `core/server.go`: Server implementation that manages REST groups and background runners
- `web/handles.go`: Request routing and handler registration
- `web/request.go`: Request abstraction with helper methods for binding, params, query
- `daemon.go`: Daemon/service wrapper for running applications as system services

### Dependency Injection
The framework uses a context-based DI container:
- All components implement the `IService` interface with `Init(ctx *Context) error`
- Services can be retrieved from context using generic getters: `wf.GetService[T](ctx)`, `wf.GetModel[T](ctx)`
- Context provides access to configuration through `ctx.Config()`

### Configuration
- Uses auto-loading config by default (supports JSON, YAML, TOML)
- Auto-loads from common locations: `./config/`, `~/.<appname>/`, `/etc/<appname>/`
- Can be customized by providing a config to `New()`
- Database configuration under `db` key
- Server configuration under `server` key
- Logging configuration under `log` key

### Daemon/Service Mode
The framework supports running applications as system services using:
- `daemon.go`: Provides service wrappers for Windows (Service Control Manager), Linux (systemd), and macOS (launchd)
- Implement the `AppService` interface with `Start()` and `Close()` methods to use daemon mode
- Use `-stop` flag to stop a running service
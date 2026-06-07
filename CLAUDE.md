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
- `IComponent`: *Removed* — components now implement `IService` directly, registered via `Service()`, retrieved via `GetService[T]`
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

## Framework Usage Guide

### 1. Quick Start - Hello World

The simplest way to create a web application:

```go
package main

import (
    "context"
    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/web"
)

func main() {
    // Create application with auto-loading config
    app := wf.NewWithAutoConfig()

    // Register a simple route
    app.Get("/", func(c *web.Request) (any, error) {
        return "hello world", nil
    })

    // Run the application
    app.Start()
}
```

### 2. REST Controller

Create structured REST APIs by implementing `IRest`:

```go
package main

import (
    "context"
    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/core"
    "github.com/chuccp/go-web-frame/web"
)

type UserController struct {
    core.IService  // Embed IService interface
}

// Init registers routes during initialization
func (u *UserController) Init(ctx *core.Context) error {
    ctx.Get("/users", u.GetUsers)
    ctx.Get("/users/:id", u.GetUser)
    ctx.Post("/users", u.CreateUser)
    ctx.Put("/users/:id", u.UpdateUser)
    ctx.Delete("/users/:id", u.DeleteUser)
    return nil
}

func (u *UserController) GetUsers(c *web.Request) (any, error) {
    return map[string]any{"users": []string{"alice", "bob"}}, nil
}

func (u *UserController) GetUser(c *web.Request) (any, error) {
    id := c.Param("id")        // Get path parameter
    return map[string]any{"id": id, "name": "alice"}, nil
}

func (u *UserController) CreateUser(c *web.Request) (any, error) {
    var user struct {
        Name string `json:"name"`
    }
    if err := c.BindJSON(&user); err != nil {
        return nil, err
    }
    return map[string]any{"id": 1, "name": user.Name}, nil
}

func (u *UserController) UpdateUser(c *web.Request) (any, error) {
    id := c.ParamInt("id")     // Get path parameter as int
    // Update logic here
    return map[string]any{"id": id, "updated": true}, nil
}

func (u *UserController) DeleteUser(c *web.Request) (any, error) {
    id := c.ParamInt("id")
    // Delete logic here
    return map[string]any{"id": id, "deleted": true}, nil
}

func main() {
    app := wf.NewWithAutoConfig()
    app.AddRest(&UserController{})
    app.Start()
}
```

### 3. Model Layer - Type-Safe Generic ORM

Define models with zero boilerplate CRUD operations:

```go
package main

import (
    "time"
    "github.com/chuccp/go-web-frame/core"
    "github.com/chuccp/go-web-frame/db"
    "github.com/chuccp/go-web-frame/model"
)

// Define entity struct
type User struct {
    Id         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    Name       string    `gorm:"size:255;not null" json:"name"`
    Email      string    `gorm:"size:255;unique" json:"email"`
    Status     int       `gorm:"default:1" json:"status"`
    CreateTime time.Time `json:"createTime"`
    UpdateTime time.Time `json:"updateTime"`
}

// Define model struct with embedded generic Model
type UserModel struct {
    *model.Model[*User]
    db *db.DB
    c  *core.Context
}

// Init initializes the model
func (m *UserModel) Init(db *db.DB, c *core.Context) error {
    m.db = db
    m.c = c
    m.Model = model.NewModel[*User](db, "t_user")
    return m.CreateTable()  // Auto-create table
}

// Usage examples:
// - m.Save(&User{Name: "alice"})           // Save record
// - m.Query().Where("status = ?", 1).All() // Query all active users
// - m.Query().Where("id = ?", 1).One()     // Query single record
// - m.Update().Where("id = ?", 1).UpdateForMap(map[string]any{"name": "bob"})
// - m.Delete().Where("id = ?", 1).Delete()
```

### 4. EntryModel - Enhanced Model with Built-in Methods

For entities with a primary key and optional `CreateTime`/`UpdateTime` fields.
`EntryModel` accepts two type parameters: `T` (entity type) and `PK` (primary key type, must satisfy `PKConstraint`: `uint`, `int`, `string`, etc.).

```go
package main

import (
    "github.com/chuccp/go-web-frame/core"
    "github.com/chuccp/go-web-frame/db"
    "github.com/chuccp/go-web-frame/model"
)

type User struct {
    Id   uint   `gorm:"primaryKey;autoIncrement" json:"id"`
    Name string `gorm:"size:255" json:"name"`
}

type UserModel struct {
    *model.EntryModel[*User, uint]
}

func (m *UserModel) Init(db *db.DB, c *core.Context) error {
    m.EntryModel = model.NewEntryModel[*User, uint](db, "t_user")
    return m.CreateTable()
}

// EntryModel provides additional methods:
// - FindByPK(id)                    // Find by primary key
// - FindAll()                       // Find all records
// - DeleteByPK(id)                  // Delete by primary key
// - UpdateByPK(entity)              // Update by primary key
// - Page(page)                      // Pagination query
// - UpdateColumn(id, column, val)   // Update single column
```

### 5. Query Operations

Fluent query builder:

```go
// Basic queries
users, err := userModel.Query().All()                                    // All records
user, err := userModel.Query().Where("id = ?", 1).One()                  // Single record
count, err := userModel.Query().Where("status = ?", 1).Count()           // Count

// Pagination
page := &web.Page{PageNo: 1, PageSize: 10}
users, total, err := userModel.Query().Where("status = ?", 1).Page(page)

// Order and limit
users, err := userModel.Query().Order("id desc").List(100)

// Web pagination (returns PageAble struct)
pageAble, err := userModel.Query().Where("status = ?", 1).PageForWeb(page)

// Update operations
err := userModel.Update().Where("id = ?", 1).UpdateForMap(map[string]any{"name": "new_name"})
err := userModel.Update().Where("id = ?", 1).UpdateColumn("status", 0)

// Delete operations
err := userModel.Delete().Where("id = ?", 1).Delete()

// Preload associations (GORM eager loading)
users, err := userModel.Query().Preload("Profile").Preload("Role").All()
user, err := userModel.Query().Where("id = ?", 1).Preload("Profile").One()
```

### 6. Request Handling

The `web.Request` provides rich request handling:

```go
func (u *UserController) HandleRequest(c *web.Request) (any, error) {
    // Path parameters
    id := c.Param("id")           // string
    idInt := c.ParamInt("id")     // int
    idUint := c.ParamUint("id")   // uint

    // Query parameters
    page := c.Query("page")       // /users?page=1

    // JSON body binding
    var data MyStruct
    if err := c.BindJSON(&data); err != nil {
        return nil, err
    }

    // Get JSON values directly
    name, _ := c.GetJsonStringValue("name")
    age, _ := c.GetJsonIntValue("age")

    // Pagination
    page, err := c.Page()         // Auto-detect from GET/POST

    // Headers
    auth := c.GetHeader("Authorization")

    // Client info
    ip := c.ClientIP()
    method := c.Request().Method

    // Cookie operations
    cookie := c.Cookie()
    value := cookie.Get("session_id")
    cookie.Set("token", "xxx", 3600)  // name, value, maxAge

    return map[string]any{"result": "ok"}, nil
}
```

### 7. Service Layer - Business Logic

Create reusable services with dependency injection:

```go
package main

import (
    "github.com/chuccp/go-web-frame/core"
    wf "github.com/chuccp/go-web-frame"
)

type UserService struct {
    core.IService
    userModel *UserModel
}

func (s *UserService) Init(ctx *core.Context) error {
    // Get model from context
    s.userModel = wf.GetModel[*UserModel](ctx)
    return nil
}

func (s *UserService) GetUserById(id uint) (*User, error) {
    return s.userModel.FindByPK(id)
}

// Register service in main:
// app.AddService(&UserService{})
// Then get it: userService := wf.GetService[*UserService](ctx)
```

### 8. Filter/Middleware

Create filters for cross-cutting concerns:

```go
package main

import (
    "errors"
    "github.com/chuccp/go-web-frame/core"
    "github.com/chuccp/go-web-frame/web"
)

type AuthFilter struct {
    core.IFilter
}

func (f *AuthFilter) Init(ctx *core.Context) error {
    return nil
}

func (f *AuthFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
    // Pre-processing
    token := req.GetHeader("Authorization")
    if token == "" {
        return nil, errors.New("unauthorized")
    }

    // Call next handler
    result, err := fc.Next()

    // Post-processing (optional)
    return result, err
}

// Register filter:
// app.AddFilter(&AuthFilter{})
```

### 9. Background Runner

Create background tasks:

```go
package main

import (
    "context"
    "time"
    "github.com/chuccp/go-web-frame/core"
)

type CleanupTask struct {
    core.IRunner
}

func (t *CleanupTask) Init(ctx *core.Context) error {
    return nil
}

func (t *CleanupTask) Run() error {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            // Execute cleanup task
        }
    }
}

// Register runner:
// app.AddRunner(&CleanupTask{})
```

### 10. Configuration

Configuration is auto-loaded from multiple locations:
- `./config/` (current directory)
- `~/.<appname>/` (user home)
- `/etc/<appname>/` (system config)

Supported formats: JSON, YAML, TOML

Example config.json:
```json
{
    "server": {
        "port": 8080,
        "host": "0.0.0.0"
    },
    "db": {
        "type": "sqlite",
        "path": "./data.db"
    },
    "log": {
        "level": "info",
        "path": "./logs/app.log"
    }
}
```

MySQL configuration:
```json
{
    "db": {
        "type": "mysql",
        "host": "localhost",
        "port": 3306,
        "user": "root",
        "password": "password",
        "database": "mydb"
    }
}
```

### 11. Component

Components are services registered with `Service()` and retrieved with `GetService[T]`:

```go
package main

import (
    "github.com/chuccp/go-web-frame/core"
)

type CacheComponent struct {
    core.IService
    // component fields
}

func (c *CacheComponent) Init(ctx *core.Context) error {
    // Initialize component with config
    // ctx.GetConfig().UnmarshalKey("cache", &cacheConfig)
    return nil
}

// Register component (before dependent services for correct init order):
// builder.Service(&CacheComponent{})
// Retrieve: wf.GetService[*CacheComponent](ctx)
```

### 12. Model Groups

Group models for shared database connection and transactions:

```go
func main() {
    app := wf.NewWithAutoConfig()

    // Create model group
    group := app.NewModelGroup(db, "user_group")
    group.AddModel(&UserModel{}, &ProfileModel{})
    group.AutoCreateTable(true)

    // Or use default model group
    app.AddModel(&UserModel{}, &OrderModel{})
    app.SetDefaultDB(db)

    app.Start()
}
```

### 13. Response Types

Return different response types from handlers:

```go
// JSON response (default)
func (c *Controller) GetJSON(req *web.Request) (any, error) {
    return map[string]any{"key": "value"}, nil
}

// String response
func (c *Controller) GetString(req *web.Request) (any, error) {
    return "plain text response", nil
}

// File download
func (c *Controller) Download(req *web.Request) (any, error) {
    return &web.File{Path: "/path/to/file.pdf", FileName: "document.pdf"}, nil
}

// Redirect
func (c *Controller) Redirect(req *web.Request) (any, error) {
    return web.Redirect("/new-url"), nil
}

// Error response
func (c *Controller) Error(req *web.Request) (any, error) {
    return nil, errors.New("something went wrong")
}
```
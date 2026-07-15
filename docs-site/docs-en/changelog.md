# Changelog

All notable changes to this project are documented here.

## [Unreleased]

## [1.0.12] - 2026-07-15

### Added
- **`MemFileSystem.ExistsFile()`**: Checks whether a path exists as a file (not a directory) with a single `Stat` call, used in SPA 404 fallback to avoid matching directories.
- **`Table.Model()`**: New `Model(value)` method on `db.Table` wrapping GORM's Model, required for Preload to resolve associations correctly.
- **`Server.ServerConfig()`**: Exposes the Server's configuration object for external access.
- **`Server.Listen()` / `Server.ListenTLS()`**: HTTP/TLS listen logic moved from Runner to Server for independent usage.
- **`Server.AddFilters()`**: Convenience method for adding multiple filters at once.
- **Access logging**: Debug-level request logging in `optionsMiddleware` printing HTTP method and path.
- **`util.DefaultPage()`**: Centralized pagination defaults (PageNo=1, PageSize=10), replacing duplicate validation across query methods.

### Refactored
- **`core/server.go`**: Removed `Init()` phase — servers are now created on demand during `Run()`; added `AddIRunner`/`AddRestGroup`.
- **`core/context.go`**: `NewContext` no longer takes `*web.Server`; removed server field from root Context.
- **`core/rest.go`**: `Handles` flows through `RestGroup` to Server; `Build()` constructs the struct directly.
- **`web/runner.go`**: Listen logic delegated to `Server.Listen`/`ListenTLS`; removed duplicate implementations and dead `logTLSListen`.
- **`web_frame.go`**: No longer pre-creates Server; handles passed through RestGroup; simplified init flow.
- **`component/cors/cors.go`**: Extracted `crosHandlerFunc`, added nil guard in `Handle`.
- **`model/clause.go`**: `ListPage`/`ExecPage`/`Page` all use `DefaultPage`, eliminating duplicate validation.

### Fixed
- SPA 404 fallback now uses `ExistsFile` instead of `Exists` to avoid matching directories.
- `initServer` properly propagates server creation errors with `errors.WithStackIf` instead of swallowing nil.
- Runner `Init` moved from `Run()` to `web_frame.init()` to avoid runners depending on uninitialized REST routes.
- `optionsMiddleware` registered before `justInitRoute` for correct middleware ordering.

## [1.0.11] - 2026-07-11

### Added
- **Self-Signed Certificate Generation**: When no wildcard certificate matches a host, the framework now automatically generates an in-memory ECDSA P-256 self-signed certificate (valid for 1 year) and caches it. Eliminates the need for external certs in development.
- **IPv6 TLS Support**: Self-signed certificates correctly handle IPv6 addresses, including bracketed format (`[::1]`, `[::1]:8443`) from HTTP Host headers.
- **`Servers.GetHandler()`**: Returns an `http.Handler` that dispatches requests to the correct `Server` engine based on the port in the Host header. Each server's routes, filters, and ContextPath remain fully independent.
- **`WebFrame.GetHandler()`**: Initializes the full application (DB, services, routes) and returns an `http.Handler` for use with `httptest.NewServer()`.
- **`core.Server.GetHandler()`**: Exposes the underlying `Servers.GetHandler()` for test and embedding scenarios.
- Multi-domain server listening — all matched domains now log their own URL on startup.

### Fixed
- `certStore.init()` now collects all certificate errors instead of failing on the first one, ensuring as many certs as possible are loaded.
- Certificate loading failures no longer block HTTP server startup — errors are logged and the server continues without TLS.
- Various MkDocs 2.0 compatibility fixes (removed deprecated plugins, empty overrides, invalid icons).

### Changed
- `core.NewServer()` now accepts an external `*web.Servers` parameter, ensuring builder-registered routes are available through `GetHandler()`.
- Docs rewritten with standard markdown for MkDocs 2.0 compatibility.

## [1.0.0] - 2026-04-07

### Added
- Initial release of Go Web Frame
- **Core Features**
  - Dependency injection with Context-based DI container
  - MVC-like architecture with clear separation of concerns
  - Type-safe generic ORM with zero boilerplate CRUD operations
  - Built-in support for SQLite, MySQL, PostgreSQL, Redis

- **Web Layer**
  - REST controller support with automatic route registration
  - Filter/middleware chain for cross-cutting concerns
  - Route metadata support via `.WithMeta()`
  - Static file serving with SPA fallback support
  - Reverse proxy support
  - WebSocket and Server-Sent Events (SSE)
  - Automatic HTTPS with Let's Encrypt certificates

- **Data Access**
  - Generic `Model[T]` for type-safe database operations
  - `EntryModel[T, PK]` for entities with primary key support
  - Fluent query builder with chain API
  - Transaction support
  - Connection pool configuration

- **Components**
  - Cache component (Redis-based)
  - Local cache component (Otter-based)
  - Rate limiting component
  - Captcha generation
  - QR code generation
  - Cron scheduled tasks
  - Input validation
  - Authentication filter
  - CORS filter

- **Infrastructure**
  - Auto-loading configuration (JSON/YAML/TOML/INI)
  - Structured logging with Zap + log rotation
  - Builder pattern for application construction

- **Documentation**
  - Multi-language README (English, Chinese, Traditional Chinese, Japanese)
  - MkDocs documentation site (English + Chinese)
  - Example applications for common use cases

## Version History

| Version | Date | Highlights |
|---------|------|------------|
| 1.0.12 | 2026-07-15 | ExistsFile, Model/Preload associations, core refactor, DefaultPage |
| 1.0.11 | 2026-07-11 | Self-signed certs, GetHandler for testing, TLS error resilience |
| 1.0.0 | 2026-04-07 | Initial release with core features |

## Next Steps

- [Home](index.md) — Back to home
- [Quick Start](getting-started/installation.md) — Installation guide

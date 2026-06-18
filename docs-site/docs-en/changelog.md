# Changelog

All notable changes to this project are documented here.

## [Unreleased]

### Added
- **SSL Local Certificates**: `SSLCert` struct and `SSLConfig.Certs` field for configuring local certificate files. Supports multi-domain local certificates with SNI-based selection and autocert fallback.
- **Context Propagation**: `WithContext(ctx)` on `Model[T]`, `EntryModel[T, PK]`, `Query[T]`, `Update[T]`, `Delete[T]` for propagating request-scoped cancellation, timeouts, and tracing to database operations.
- `Request.Ctx()` method on `web.Request` to expose the per-request `context.Context`.
- `DB.WithContext(ctx)` and `Table.WithContext(ctx)` in the `db` package as the foundation layer.

### Fixed
- `Builder.Any()` now correctly registers handlers for all HTTP methods (was only registering GET).
- `Config.GetBoolOrDefault()` now returns the default value when the key is not set in config.
- Renamed internal `getBb()` to `getDB()` in `model.Model[T]` for clarity.

### Changed
- ORM comparison docs updated — context propagation is now a key differentiator vs GORM's built-in generics.

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
| 1.0.0 | 2026-04-07 | Initial release with core features |

## Next Steps

- [Home](index.md) — Back to home
- [Quick Start](getting-started/installation.md) — Installation guide

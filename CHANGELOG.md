# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

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
  - Automatic HTTPS with Let's Encrypt certificates

- **Data Access**
  - Generic `Model[T]` for type-safe database operations
  - `EntryModel[T, PK]` for entities with primary key support (uint, int, string, etc.)
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

- **Infrastructure**
  - Auto-loading configuration (JSON/YAML/TOML)
  - Structured logging with Zap
  - Log rotation support
  - Daemon/service mode for Windows/Linux/macOS

- **Documentation**
  - Multi-language README (English, Chinese, Traditional Chinese, Japanese)
  - Comprehensive CLAUDE.md for AI-assisted development
  - Example applications for common use cases

### Architecture
- Unified RouteTree for all route types (API, static files, reverse proxy)
- Clean separation: Registration → Storage → Execution
- Interface-based component model (IService, IModel, IRest, IRunner, IFilter, IComponent)

---

## Version History Summary

| Version | Date | Highlights |
|---------|------|------------|
| 1.0.0 | 2026-04-07 | Initial release with core features |
# Changelog

## v0.8.4

- Remove IComponent interface, merge into IService
- Simplify IComponent to embed IService
- Replace encoding/json with bytedance/sonic

## v0.8.3

- Add logging documentation
- Add deployment documentation
- Add util API documentation
- Add Redis documentation
- Add WebSocket documentation
- Add IConverter documentation
- Fix documentation inaccuracies

## v0.8.2

- Add route metadata (WithMeta) support
- Add Builder pattern initialization
- Add generic Model[T] with zero-boilerplate CRUD

## v0.8.0

- Initial release
- Gin + GORM + Viper + Zap integration
- Dependency injection via Context
- REST controller support
- Multiple database support (MySQL, PostgreSQL, SQLite)
- Redis integration
- Built-in CORS filter
- Rate limiting component
- Captcha component
- QR code component
- Cron scheduled tasks
- WebSocket support
- SSE support
- HTTPS with Let's Encrypt
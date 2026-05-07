# Configuration

Configuration management with Viper.

## Config Locations

Auto-loaded from (in order):
- `./config/`
- `~/.<appname>/`
- `/etc/<appname>/`

## Config Formats

- JSON
- YAML
- TOML

## Example config.yaml

```yaml
web:
  server:
    port: 8080
    context_path: /api
  db:
    type: mysql
    host: localhost
    port: 3306
    user: root
    password: secret
    database: mydb
  log:
    level: info
    path: ./logs/app.log
  redis:
    addr: localhost:6379
```

## Load Config

```go
// Auto load
builder := wf.NewBuilder(config.LoadAutoConfig())

// Custom config
cfg := config.NewConfig()
cfg.Set("web.server.port", 9000)
builder := wf.NewBuilder(cfg)
```

## Read Config

```go
func (s *MyService) Init(ctx *core.Context) error {
    port := ctx.Config().GetInt("web.server.port")
    dbType := ctx.Config().GetString("web.db.type")
    return nil
}
```

## Database Configuration

| Database | Config |
|----------|--------|
| MySQL | `type: mysql`, `host`, `port`, `user`, `password`, `database` |
| PostgreSQL | `type: postgres`, `host`, `port`, `user`, `password`, `database` |
| SQLite | `type: sqlite`, `file_path` |
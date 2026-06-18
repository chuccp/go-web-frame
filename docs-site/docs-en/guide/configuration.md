# Configuration

Go Web Frame provides flexible configuration management with multi-location and multi-format support, powered by Viper.

## Loading Configuration

### Single File (Recommended)

Use `config.LoadSingleFileConfig()` to load a single YAML file:

```go
cfg, err := config.LoadSingleFileConfig("application.yml")
if err != nil {
    log.PanicErrors("Failed to load config", err)
}

builder := wf.NewBuilder(cfg)
app := builder.Build()
```

### Multiple Files

Use `config.LoadConfig()` to load multiple files in order. Later files override earlier ones:

```go
cfg, err := config.LoadConfig(
    "/etc/myapp/application.yml",  // system-level config
    "./application.yml",            // project-level config
)
if err != nil {
    log.PanicErrors("Failed to load config", err)
}

builder := wf.NewBuilder(cfg)
app := builder.Build()
```

### Auto-Discovery

Use `config.LoadAutoConfig()` for zero-config setup:

```go
builder := wf.NewBuilder(config.LoadAutoConfig())
```

### From Bytes

Use `config.NewFromBytes()` to load config from raw bytes:

```go
cfg, err := config.NewFromBytes(yamlBytes, "yaml")
```

## Config Formats

Supported formats:

- **JSON**
- **YAML** (recommended)
- **TOML**
- **INI**

### YAML Example (application.yml)

```yaml
web:
  server:
    port: 8081
  log:
    path: ./logs/app.log
    level: debug
  db:
    type: mysql
    host: localhost
    port: 3306
    database: mydb
    user: root
    password: your_password

redis:
  addr: localhost:6379
  password: ""
  db: 0

log:
  level: info
  path: ./logs/app.log
```

### JSON Example (config.json)

```json
{
  "web": {
    "server": {
      "port": 8081
    }
  },
  "db": {
    "type": "sqlite",
    "path": "./data.db"
  }
}
```

### TOML Example (config.toml)

```toml
[web.server]
port = 8081

[web.db]
type = "sqlite"
path = "./data.db"
```

## Configuration Reference

### Server

```yaml
web:
  server:
    port: 8081
    context_path: /api
    locations:
      - ./view/dist
      - www
    page404: 404.html
    ssl:
      enabled: false
      hosts: []
      certs: []
```

| Key | Description | Default |
|-----|-------------|---------|
| `web.server.port` | Listen port | 19009 |
| `web.server.context_path` | Route prefix | none |
| `web.server.locations` | Static file directories | none |
| `web.server.page404` | 404 fallback page | none |
| `web.ssl.enabled` | Enable HTTPS | false |
| `web.ssl.hosts` | Domains for Let's Encrypt | none |
| `web.ssl.certs` | Local certificate entries | none |

### Database

#### SQLite

```yaml
web:
  db:
    type: sqlite
    file_path: ./data.db
```

#### MySQL

```yaml
web:
  db:
    type: mysql
    host: localhost
    port: 3306
    database: mydb
    user: root
    password: your_password
    max_open_conns: 100
    max_idle_conns: 10
    conn_max_lifetime: 3600
```

#### PostgreSQL

```yaml
web:
  db:
    type: postgres
    host: localhost
    port: 5432
    database: mydb
    user: postgres
    password: your_password
    sslmode: disable
    max_open_conns: 100
    max_idle_conns: 10
    conn_max_lifetime: 3600
```

### Redis

```yaml
redis:
  addr: localhost:6379
  password: ""
  db: 0
  pool_size: 10
  min_idle_conns: 5
```

### Logging

```yaml
log:
  level: info              # debug | info | warn | error
  path: ./logs/app.log
  max_size: 100            # MB per file before rotation
  max_backups: 7           # max rotated files to keep
  max_age: 30              # days to keep old logs
  compress: true           # compress rotated files
```

| Key | Description | Default |
|-----|-------------|---------|
| `log.level` | Log level | `info` |
| `log.path` | Log file path | `./logs/app.log` |
| `log.max_size` | Max file size (MB) | 100 |
| `log.max_backups` | Max backup files | 7 |
| `log.max_age` | Max retention days | 30 |
| `log.compress` | Compress backups | `true` |

## Environment Variables

Config values support environment variable substitution:

```yaml
web:
  db:
    host: ${DB_HOST:localhost}
    port: ${DB_PORT:3306}
    user: ${DB_USER:root}
    password: ${DB_PASSWORD:}
    database: ${DB_NAME:mydb}
```

Format: `${ENV_VAR:default_value}`

You can also read environment variables directly in code:

```go
port := util.GetEnvIntOrDefault("APP_PORT", 8081)
host := util.GetEnvOrDefault("APP_HOST", "localhost")
```

## Accessing Configuration

### In a Service

```go
type UserService struct {
    core.IService
    maxRetries int
}

func (s *UserService) Init(ctx *core.Context) error {
    s.maxRetries = ctx.GetConfig().GetInt("service.max_retries")
    return nil
}
```

### Available Methods

```go
cfg := ctx.GetConfig()

cfg.GetString("key")              // string
cfg.GetInt("key")                 // int
cfg.GetBoolOrDefault("key", false) // bool with default
cfg.GetIntOrDefault("key", 0)     // int with default
cfg.GetStringOrDefault("key", "")  // string with default
cfg.HasKey("key")                 // check if key exists

// Deserialize into struct
var myConfig MyConfig
cfg.UnmarshalKey("service", &myConfig)

// Unmarshal entire config
cfg.Unmarshal(&myConfig)
```

## Multi-Environment Configuration

### Approach 1: File Naming Convention

```
config/
├── application.yml          # base config
├── application-dev.yml      # development overrides
├── application-prod.yml     # production overrides
└── application-test.yml     # test overrides
```

Load multiple files in order:

```go
cfg, err := config.LoadConfig("application.yml", "application-dev.yml")
builder := wf.NewBuilder(cfg)
```

### Approach 2: Multi-Location Override

Load from multiple locations; later locations override earlier ones:

```go
cfg, err := config.LoadConfig(
    "/etc/myapp/application.yml",   // system-level
    "~/.myapp/application.yml",     // user-level
    "./application.yml",            // project-level
)
```

## Complete Example

```yaml
web:
  server:
    port: 8081
    context_path: /api
  db:
    type: mysql
    host: ${DB_HOST:localhost}
    port: ${DB_PORT:3306}
    database: ${DB_NAME:mydb}
    user: ${DB_USER:root}
    password: ${DB_PASSWORD:}
    max_open_conns: 100
    max_idle_conns: 10
  log:
    path: ./logs/app.log
    level: info

redis:
  addr: ${REDIS_ADDR:localhost:6379}
  password: ${REDIS_PASSWORD:}
  db: 0

log:
  level: info
  path: ./logs/app.log
  max_size: 100
  max_backups: 7
  max_age: 30
  compress: true
```

## Next Steps

- [Database](../advanced/database.md) — Transactions, model groups
- [Deployment](../advanced/deployment.md) — HTTPS, graceful shutdown
- [Best Practices](../best-practices.md) — Recommended patterns

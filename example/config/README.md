# Configuration Example

This example demonstrates how to use the configuration system in Go Web Frame.

## Features Demonstrated

1. **Loading configuration from YAML files**
2. **Unmarshaling config into structs**
3. **Reading config values programmatically**
4. **Creating custom configuration services**
5. **Exposing config via REST API**

## Running the Example

### Build

```bash
cd example/config
go build config.go
```

### Run

```bash
./config
```

The application will start on port 8080 (or as configured in `application.example.yml`).

## API Endpoints

### Health Check
```bash
curl http://localhost:8080/health
```

### Get Full Configuration
```bash
curl http://localhost:8080/config
```

### Get Server Configuration
```bash
curl http://localhost:8080/config/server
```

### Get Database Configuration
```bash
curl http://localhost:8080/config/database
```

### Get Redis Configuration
```bash
curl http://localhost:8080/config/redis
```

## Configuration File

The example uses `application.example.yml` as the configuration file. You can customize it for your needs:

```yaml
server:
  port: 8080
  mode: debug

web:
  db:
    type: mysql
    host: localhost
    port: 3306
    username: root
    password: your_password
    database: your_database

log:
  level: info
  path: ./logs/app.log

redis:
  addr: localhost:6379
  password: ""
  db: 0
```

## Usage Examples

### Load Config from a Single File

```go
cfg, err := config.LoadSingleFileConfig("./application.example.yml")
if err != nil {
    log.Error("Failed to load config", zap.Error(err))
    return
}

// Read specific values
port := cfg.GetInt("server.port")
mode := cfg.GetString("server.mode")
```

### Load Config from Multiple Files

```go
cfg, err := config.LoadConfig("./config/base.yml", "./config/override.yml")
```

### Load Config from Bytes

```go
yamlData := []byte(`
server:
  port: 9090
  mode: debug
`)

cfg, err := config.NewFromBytes(yamlData, "yaml")
```

### Unmarshal into Struct

```go
type AppConfig struct {
    Server ServerConfig `mapstructure:"server"`
    DB     DBConfig     `mapstructure:"db"`
}

cfg := ctx.GetConfig()
var appConfig AppConfig
err := cfg.Unmarshal(&appConfig)
```

### Check if Key Exists

```go
if cfg.HasKey("server.ssl.enabled") {
    sslEnabled := cfg.GetBoolOrDefault("server.ssl.enabled", false)
}
```

### Set Custom Configuration Values

```go
cfg := ctx.GetConfig()
cfg.Put("app.name", "My App")
cfg.Put("app.version", "1.0.0")
```

## Best Practices

1. **Always use environment variables for sensitive data** (passwords, API keys)
2. **Use struct tags** for type-safe config access
3. **Provide default values** for optional configuration
4. **Validate configuration** at application startup
5. **Never log sensitive configuration values**

## Configuration Keys Reference

| Key | Type | Description |
|-----|------|-------------|
| `server.port` | int | Server port number |
| `server.mode` | string | Server mode (debug/release) |
| `web.db.type` | string | Database type (mysql/postgres/sqlite) |
| `web.db.host` | string | Database host |
| `web.db.port` | int | Database port |
| `web.db.username` | string | Database username |
| `web.db.password` | string | Database password |
| `web.db.database` | string | Database name |
| `log.level` | string | Log level (debug/info/warn/error) |
| `log.path` | string | Log file path |
| `redis.addr` | string | Redis address |
| `redis.password` | string | Redis password |
| `redis.db` | int | Redis database number |
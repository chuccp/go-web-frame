# 配置

Go Web Frame 提供灵活的配置管理，支持多位置和多种格式。

## 自动配置

### 使用自动配置（推荐）

```go
app := wf.NewWithAutoConfig()
```

框架会自动从以下位置加载配置（后面的覆盖前面的）：

1. `./config/` - 开发环境（项目本地）
2. `~/.<appname>/` - 用户特定设置
3. `/etc/<appname>/` - 系统级（生产环境）

### 手动配置

```go
// 使用配置文件（INI 格式）
fileConfig, err := config.LoadSingleFileConfig("config.ini")
if err != nil {
	log.Panic("加载配置失败", zap.Error(err))
}

builder := wf.NewBuilder(fileConfig)
app := builder.Build()
```

## 配置文件格式

框架使用 Viper 库，支持以下格式：

- **INI**（默认，推荐用于简单配置）
- **JSON**
- **YAML**（推荐用于复杂配置）
- **TOML**

### YAML 示例（application.yml）

```yaml
# 服务器配置
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

# Redis 配置
redis:
  addr: localhost:6379
  password: ""
  db: 0

# 日志配置
log:
  level: info
  path: ./logs/app.log
```

### JSON 示例（config.json）

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

### INI 示例（config.ini）

```ini
[web.server]
port = 8081

[web.db]
type = sqlite
path = ./data.db
```

## 配置项

### Server 配置

```yaml
web:
  server:
    port: 8081          # 端口
    host: 0.0.0.0       # 主机
    read_timeout: 30    # 读取超时（秒）
    write_timeout: 30   # 写入超时（秒）
  ssl:
    enabled: false       # 是否启用 HTTPS
  locations:            # 静态文件目录
    - ./view/dist
    - www
  page404: 404.html     # 404 页面
```

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `web.server.port` | 端口 | 8081 |
| `web.server.host` | 主机 | `0.0.0.0` |
| `web.server.read_timeout` | 读取超时（秒） | 30 |
| `web.server.write_timeout` | 写入超时（秒） | 30 |
| `web.ssl.enabled` | 是否启用 HTTPS | false |

### 数据库配置

#### SQLite

```yaml
web:
  db:
    type: sqlite
    path: ./data.db
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
    max_open_conns: 100      # 最大打开连接数
    max_idle_conns: 10        # 最大空闲连接数
    conn_max_lifetime: 3600   # 连接最大生命周期（秒）
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
```

### Redis 配置

```yaml
redis:
  addr: localhost:6379
  password: ""
  db: 0
  pool_size: 10           # 连接池大小
  min_idle_conns: 5        # 最小空闲连接数
```

### 日志配置

```yaml
log:
  level: info              # 日志级别（debug/info/warn/error）
  path: ./logs/app.log    # 日志文件路径
  max_size: 100           # 单文件最大大小（MB）
  max_backups: 7           # 最大备份文件数
  max_age: 30              # 最大保留天数
  compress: true           # 是否压缩备份文件
```

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `log.level` | 日志级别（debug/info/warn/error） | `info` |
| `log.path` | 日志文件路径 | `./logs/app.log` |
| `log.max_size` | 单文件最大大小（MB） | 100 |
| `log.max_backups` | 最大备份文件数 | 7 |
| `log.max_age` | 最大保留天数 | 30 |
| `log.compress` | 是否压缩备份文件 | `true` |

## 环境变量

配置支持环境变量读取：

```go
// 在代码中使用环境变量
port := util.GetEnvIntOrDefault("APP_PORT", 8081)
host := util.GetEnvOrDefault("APP_HOST", "localhost")
```

也可以在配置文件中引用环境变量：

```yaml
web:
  db:
    host: ${DB_HOST:localhost}
    port: ${DB_PORT:3306}
    user: ${DB_USER:root}
    password: ${DB_PASSWORD:}
    database: ${DB_NAME:mydb}
```

格式：`${ENV_VAR:default_value}`

## 访问配置

### 在服务中访问

```go
type UserService struct {
    core.IService
    maxRetries int
    timeout    time.Duration
}

func (s *UserService) Init(context *core.Context) error {
    // 通过 context 获取配置并读取
    s.maxRetries = context.GetConfig().GetInt("service.max_retries")
    s.timeout = context.GetConfig().GetDuration("service.timeout")
    return nil
}
```

### 可用方法

```go
// 获取字符串
value := ctx.GetConfig().GetString("key")

// 获取整数
value := ctx.GetConfig().GetInt("key")

// 获取布尔值
value := ctx.GetConfig().GetBool("key")

// 获取浮点数
value := ctx.GetConfig().GetFloat64("key")

// 获取持续时间
value := ctx.GetConfig().GetDuration("key")

// 获取字符串数组
value := ctx.GetConfig().GetStringSlice("key")

// 获取字符串映射
value := ctx.GetConfig().GetStringMap("key")

// 反序列化到结构体
var config MyConfig
ctx.GetConfig().UnmarshalKey("key", &config)
```

## 多环境配置

### 方式一：配置文件命名

```
config/
├── application.yml          # 基础配置
├── application-dev.yml      # 开发环境
├── application-prod.yml     # 生产环境
└── application-test.yml     # 测试环境
```

通过 `config.LoadConfig` 加载多个配置文件：

```go
cfg, err := config.LoadConfig("application.yml", "application-dev.yml")
if err != nil {
    log.Panic("加载配置失败", zap.Error(err))
}
builder := wf.NewBuilder(cfg)
```

### 方式二：多位置覆盖

框架支持按顺序加载多个配置文件，后面的配置会覆盖前面的：

```go
cfg, err := config.LoadConfig(
    "/etc/myapp/application.yml",   // 系统级配置
    "~/.myapp/application.yml",     // 用户级配置
    "./application.yml",            // 项目级配置
)
```

## 完整示例

```yaml
# 服务器配置
web:
  server:
    port: 8081
    host: 0.0.0.0
  log:
    path: ./logs/app.log
    level: debug
  db:
    type: mysql
    host: ${DB_HOST:localhost}
    port: ${DB_PORT:3306}
    database: ${DB_NAME:mydb}
    user: ${DB_USER:root}
    password: ${DB_PASSWORD:}

# Redis 配置
redis:
  addr: ${REDIS_ADDR:localhost:6379}
  password: ${REDIS_PASSWORD:}
  db: 0

# 日志配置
log:
  level: info
  path: ./logs/app.log
```

## 下一步

- [数据库高级用法](../advanced/database.md) - 事务、迁移等
- [最佳实践](../best-practices.md) - 推荐的使用模式

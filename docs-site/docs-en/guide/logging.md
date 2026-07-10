# Logging

The framework provides structured logging built on [Zap](https://go.uber.org/zap), with file rotation support.

## Basic Usage

```go
import (
    "github.com/chuccp/go-web-frame/log"
    "go.uber.org/zap"
)

// Different log levels
log.Debug("debug message", zap.String("key", "value"))
log.Info("server started", zap.Int("port", 8081))
log.Warn("slow query detected", zap.Duration("elapsed", time.Since(start)))
log.Error("failed to save user", zap.Error(err))
log.Fatal("unrecoverable error", zap.Error(err))  // calls os.Exit(1) after writing
log.Panic("critical failure", zap.Error(err))     // panics after writing

// Multiple errors
log.Errors("multiple errors occurred", err1, err2)
log.PanicErrors("critical errors", err1, err2)
```

## Usage in Services

```go
type UserService struct {
    core.IService
    userModel *UserModel
}

func (s *UserService) CreateUser(input *CreateUserInput) (*User, error) {
    log.Info("creating user", zap.String("email", input.Email))

    user, err := s.doCreateUser(input)
    if err != nil {
        log.Error("failed to create user",
            zap.String("email", input.Email),
            zap.Error(err),
        )
        return nil, err
    }

    log.Info("user created", zap.Uint("id", user.Id))
    return user, nil
}
```

## Usage in Controllers

```go
func (c *UserController) Get(req *web.Request) (any, error) {
    id := req.ParamUint("id")
    log.Debug("get user request", zap.Uint("id", id))

    user, err := c.userService.GetUserById(id)
    if err != nil {
        log.Error("get user failed", zap.Uint("id", id), zap.Error(err))
        return nil, err
    }
    return user, nil
}
```

## Common Zap Fields

```go
zap.String("key", "value")       // string
zap.Int("count", 42)             // int
zap.Uint("id", uint(1))          // uint
zap.Float64("rate", 3.14)        // float64
zap.Bool("active", true)         // bool
zap.Duration("elapsed", took)    // duration
zap.Error(err)                   // error (auto key "error")
zap.Any("data", someValue)       // any value
```

## Configuration

```yaml
log:
  level: info              # debug/info/warn/error
  path: ./logs/app.log     # log file path
  write: true              # whether to write to file
  max_size: 100            # max file size (MB)
  max_backups: 7           # max backup files
  max_age: 30              # max retention days
  compress: true           # compress rotated files
```

| Key | Description | Default |
|-----|-------------|---------|
| `log.level` | Log level | `info` |
| `log.path` | Log file path | `./logs/app.log` |
| `log.write` | Write to file | `true` |
| `log.max_size` | Max file size (MB) | 100 |
| `log.max_backups` | Max backup files | 7 |
| `log.max_age` | Max retention days | 30 |
| `log.compress` | Compress backups | `true` |

## Log Levels

From lowest to highest:

| Level | Method | Description |
|-------|--------|-------------|
| Debug | `log.Debug()` | Debug info, usually disabled in production |
| Info | `log.Info()` | Normal runtime info |
| Warn | `log.Warn()` | Warning, doesn't affect operation but needs attention |
| Error | `log.Error()` | Error, affects functionality but program continues |
| Fatal | `log.Fatal()` | Fatal error, exits program after logging |
| Panic | `log.Panic()` | Critical error, panics after logging |

## Log API Reference

| Function | Signature | Description |
|----------|-----------|-------------|
| `Debug` | `Debug(msg string, fields ...zap.Field)` | Debug level log |
| `Info` | `Info(msg string, fields ...zap.Field)` | Info level log |
| `Warn` | `Warn(msg string, fields ...zap.Field)` | Warn level log |
| `Error` | `Error(msg string, fields ...zap.Field)` | Error level log |
| `Errors` | `Errors(msg string, errs ...error)` | Multi-error log |
| `Fatal` | `Fatal(msg string, fields ...zap.Field)` | Fatal level (exits program) |
| `Panic` | `Panic(msg string, fields ...zap.Field)` | Panic level |
| `PanicErrors` | `PanicErrors(msg string, errs ...error)` | Multi-error Panic |
| `PrintPanic` | `PrintPanic(errs ...error)` | Print panic errors |
| `Sync` | `Sync() error` | Flush log buffer |

## Next Steps

- [Configuration](configuration.md) - Configuration management
- [Deployment](../advanced/deployment.md) - Production recommendations

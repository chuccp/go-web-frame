# Logging

Structured logging with Zap.

## Configuration

```yaml
web:
  log:
    level: info
    path: ./logs/app.log
    max_size: 100      # MB
    max_backups: 5     # files
    max_age: 7         # days
    compress: true
```

## Log Levels

- `debug`
- `info`
- `warn`
- `error`

## Usage

```go
import "github.com/chuccp/go-web-frame/log"

log.Info("message")
log.Errorf("error: %v", err)
log.Warnf("warning: %s", msg)
log.Debugf("debug: %d", num)
```

## Structured Logging

```go
log.InfoKV("user login", 
    "user_id", 123,
    "ip", "192.168.1.1",
)
```

## Panic Recovery

```go
log.PrintPanic(err)
```
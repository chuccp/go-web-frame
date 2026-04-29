# 日志

框架基于 [Zap](https://go.uber.org/zap) 提供结构化日志，支持文件轮转。

## 基本用法

```go
import (
    "github.com/chuccp/go-web-frame/log"
    "go.uber.org/zap"
)

// 不同日志级别
log.Debug("debug message", zap.String("key", "value"))
log.Info("server started", zap.Int("port", 8081))
log.Warn("slow query detected", zap.Duration("elapsed", time.Since(start)))
log.Error("failed to save user", zap.Error(err))
log.Fatal("unrecoverable error", zap.Error(err))  // 写入日志后调用 os.Exit(1)
log.Panic("critical failure", zap.Error(err))     // 写入日志后 panic

// 多个错误
log.Errors("multiple errors occurred", err1, err2)
log.PanicErrors("critical errors", err1, err2)
```

## 在服务中使用

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

## 在控制器中使用

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

## 在 Runner 中使用

```go
func (r *CleanupRunner) Run() error {
    ticker := time.NewTicker(30 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-r.ctx.Done():
            log.Info("cleanup runner stopping")
            return nil
        case <-ticker.C:
            if err := r.cleanup(); err != nil {
                log.Error("cleanup failed", zap.Error(err))
            }
        }
    }
}
```

## 常用 Zap 字段

```go
zap.String("key", "value")       // 字符串
zap.Int("count", 42)             // 整数
zap.Uint("id", uint(1))          // 无符号整数
zap.Float64("rate", 3.14)        // 浮点数
zap.Bool("active", true)         // 布尔值
zap.Duration("elapsed", took)    // 时间间隔
zap.Error(err)                   // 错误（自动用 "error" 作为 key）
zap.Any("data", someValue)       // 任意值
```

## 配置

```yaml
log:
  level: info              # 日志级别（debug/info/warn/error）
  path: ./logs/app.log    # 日志文件路径
  write: true             # 是否写入文件
  max_size: 100           # 单文件最大大小（MB）
  max_backups: 7          # 最大备份文件数
  max_age: 30             # 最大保留天数
  compress: true           # 是否压缩备份文件
```

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `log.level` | 日志级别 | `info` |
| `log.path` | 日志文件路径 | `./logs/app.log` |
| `log.write` | 是否写入文件 | `true` |
| `log.max_size` | 单文件最大大小（MB） | 100 |
| `log.max_backups` | 最大备份文件数 | 7 |
| `log.max_age` | 最大保留天数 | 30 |
| `log.compress` | 是否压缩备份文件 | `true` |

## 日志级别

从低到高：

| 级别 | 方法 | 说明 |
|------|------|------|
| Debug | `log.Debug()` | 调试信息，生产环境通常关闭 |
| Info | `log.Info()` | 常规运行信息 |
| Warn | `log.Warn()` | 警告，不影响运行但需关注 |
| Error | `log.Error()` | 错误，影响功能但程序可继续 |
| Fatal | `log.Fatal()` | 致命错误，写入日志后退出程序 |
| Panic | `log.Panic()` | 严重错误，写入日志后 panic |

## 日志 API 参考

| 函数 | 签名 | 说明 |
|------|------|------|
| `Debug` | `Debug(msg string, fields ...zap.Field)` | Debug 级别日志 |
| `Info` | `Info(msg string, fields ...zap.Field)` | Info 级别日志 |
| `Warn` | `Warn(msg string, fields ...zap.Field)` | Warn 级别日志 |
| `Error` | `Error(msg string, fields ...zap.Field)` | Error 级别日志 |
| `Errors` | `Errors(msg string, errs ...error)` | 多错误日志 |
| `Fatal` | `Fatal(msg string, fields ...zap.Field)` | Fatal 级别（退出程序） |
| `Panic` | `Panic(msg string, fields ...zap.Field)` | Panic 级别 |
| `PanicErrors` | `PanicErrors(msg string, errs ...error)` | 多错误 Panic |
| `PrintPanic` | `PrintPanic(errs ...error)` | 打印 Panic 错误 |
| `Sync` | `Sync() error` | 刷新日志缓冲区 |

## 下一步

- [配置](configuration.md) - 了解配置管理
- [部署](../advanced/deployment.md) - 生产环境建议

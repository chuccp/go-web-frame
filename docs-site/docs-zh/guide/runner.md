# 后台任务（Runner）

Runner 用于在后台执行周期性或长时间运行的任务，如定时清理、数据同步、邮件发送等。

## 创建 Runner

嵌入 `core.IRunner` 接口：

```go
package runner

import (
    "context"
    "time"
    "github.com/chuccp/go-web-frame/core"
)

type CleanupRunner struct {
    core.IRunner
    ctx *core.Context
}

func (r *CleanupRunner) Init(c *core.Context) error {
    r.ctx = c
    return nil
}

// Run 方法无参数，返回 error
func (r *CleanupRunner) Run() error {
    ticker := time.NewTicker(30 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-r.ctx.Done():
            // 优雅退出：处理剩余任务
            return nil
        case <-ticker.C:
            // 执行清理任务
            r.cleanup()
        }
    }
}

func (r *CleanupRunner) cleanup() {
    // 清理逻辑
}
```

## 注册 Runner

在 `main.go` 中注册：

```go
builder := wf.NewBuilder(cfg)
builder.Runner(&runner.CleanupRunner{})
app := builder.Build()
app.Start()
```

## 使用定时任务调度

框架内置 `component/schedule` 组件，支持 cron 表达式调度：

```go
package runner

import (
    "github.com/chuccp/go-web-frame/component/schedule"
    "github.com/chuccp/go-web-frame/core"
    "github.com/chuccp/go-web-frame/log"
    "go.uber.org/zap"
)

type ScheduledRunner struct {
    schedule *schedule.Schedule
    ctx      *core.Context
}

func (r *ScheduledRunner) Init(c *core.Context) error {
    r.ctx = c
    r.schedule = core.GetRunner[*schedule.Schedule](c)
    return nil
}

func (r *ScheduledRunner) Run() error {
    // 添加定时任务（cron 表达式）
    _, err := r.schedule.AddFunc("0 */5 * * * ?", func(c *core.Context) {
        log.Info("running scheduled task", zap.String("name", "data-sync"))
        r.syncData()
    })
    return err
}

func (r *ScheduledRunner) syncData() {
    // 数据同步逻辑
}
```

注册方式：

```go
builder := wf.NewBuilder(cfg)
// 先注册调度组件，再注册自定义 Runner
builder.Runner(schedule.NewScheduleWithSeconds(), &runner.ScheduledRunner{})
app := builder.Build()
app.Start()
```

## 在服务中获取组件

Runner 的 `Init` 方法中可以获取其他服务、模型和组件：

```go
type EmailRunner struct {
    core.IRunner
    emailService *service.EmailService
    ctx          *core.Context
}

func (r *EmailRunner) Init(c *core.Context) error {
    r.ctx = c
    r.emailService = core.GetService[*service.EmailService](c)
    return nil
}

func (r *EmailRunner) Run() error {
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-r.ctx.Done():
            return nil
        case <-ticker.C:
            r.emailService.SendPendingEmails()
        }
    }
}
```

## 完整示例

```go
package main

import (
    "sync"
    "time"

    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/config"
    "github.com/chuccp/go-web-frame/core"
    "github.com/chuccp/go-web-frame/log"
    atomic2 "go.uber.org/atomic"
    "go.uber.org/zap"
)

type DataSyncRunner struct {
    core.IRunner
    ctx     *core.Context
    mu      sync.Mutex
    isRunning *atomic2.Bool
}

func (r *DataSyncRunner) Init(c *core.Context) error {
    r.ctx = c
    r.isRunning = atomic2.NewBool(false)
    return nil
}

func (r *DataSyncRunner) Run() error {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    r.syncData() // 立即执行一次

    for {
        select {
        case <-r.ctx.Done():
            log.Info("data sync runner stopping")
            return nil
        case <-ticker.C:
            r.syncData()
        }
    }
}

func (r *DataSyncRunner) syncData() {
    // 防止并发执行
    if !r.isRunning.CompareAndSwap(false, true) {
        return
    }
    defer r.isRunning.Swap(false)

    r.mu.Lock()
    defer r.mu.Unlock()

    log.Info("data sync started")
    // 执行数据同步
    log.Info("data sync completed")
}

func main() {
    cfg, err := config.LoadSingleFileConfig("application.yml")
    if err != nil {
        log.Panic("加载配置失败", zap.Error(err))
    }

    builder := wf.NewBuilder(cfg)
    builder.Runner(&DataSyncRunner{})
    app := builder.Build()
    app.Start()
}
```

## 下一步

- [组件](components.md) - 了解框架内置组件
- [服务](service.md) - 业务逻辑层
- [模型](model.md) - 数据访问层

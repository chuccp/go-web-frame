# Runner (Background Tasks)

Runners execute periodic or long-running tasks in the background, such as scheduled cleanup, data sync, email sending, etc.

## Create a Runner

Embed the `core.IRunner` interface:

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

// Run takes no arguments and returns error
func (r *CleanupRunner) Run() error {
    ticker := time.NewTicker(30 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-r.ctx.Done():
            // Graceful shutdown: process remaining tasks
            return nil
        case <-ticker.C:
            // Execute cleanup task
            r.cleanup()
        }
    }
}

func (r *CleanupRunner) cleanup() {
    // Cleanup logic
}
```

## Register a Runner

Register in `main.go`:

```go
builder := wf.NewBuilder(cfg)
builder.Runner(&runner.CleanupRunner{})
app := builder.Build()
app.Start()
```

## Scheduled Task Dispatch

The framework includes a `component/schedule` component that supports cron expression scheduling:

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
    // Add a scheduled task (cron expression)
    _, err := r.schedule.AddFunc("0 */5 * * * ?", func(c *core.Context) {
        log.Info("running scheduled task", zap.String("name", "data-sync"))
        r.syncData()
    })
    return err
}

func (r *ScheduledRunner) syncData() {
    // Data sync logic
}
```

Registration:

```go
builder := wf.NewBuilder(cfg)
// Register the schedule component first, then custom Runners
builder.Runner(schedule.NewScheduleWithSeconds(), &runner.ScheduledRunner{})
app := builder.Build()
app.Start()
```

## Get Components in a Runner

The `Init` method can retrieve other services, models, and components:

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

## Full Example

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
    ctx        *core.Context
    mu         sync.Mutex
    isRunning  *atomic2.Bool
}

func (r *DataSyncRunner) Init(c *core.Context) error {
    r.ctx = c
    r.isRunning = atomic2.NewBool(false)
    return nil
}

func (r *DataSyncRunner) Run() error {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    r.syncData() // run immediately

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
    // Prevent concurrent execution
    if !r.isRunning.CompareAndSwap(false, true) {
        return
    }
    defer r.isRunning.Swap(false)

    r.mu.Lock()
    defer r.mu.Unlock()

    log.Info("data sync started")
    // Execute data sync
    log.Info("data sync completed")
}

func main() {
    cfg, err := config.LoadSingleFileConfig("application.yml")
    if err != nil {
        log.Panic("failed to load config", zap.Error(err))
    }

    builder := wf.NewBuilder(cfg)
    builder.Runner(&DataSyncRunner{})
    app := builder.Build()
    app.Start()
}
```

## Next Steps

- [Components](components.md) - Built-in framework components
- [Service](service.md) - Business logic layer
- [Model](model.md) - Data access layer

# Runner

Background task runners.

## Define Runner

```go
type CleanupRunner struct {
    core.IRunner
}

func (r *CleanupRunner) Init(ctx *core.Context) error {
    return nil
}

func (r *CleanupRunner) Run() error {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            // Execute cleanup
        }
    }
}
```

## Register Runner

```go
builder.Runner(&CleanupRunner{})
```

## Cron Tasks

Use the cron component:

```go
import "github.com/chuccp/go-web-frame/component"

cron := component.NewCron()
cron.AddFunc("0 0 * * *", func() {
    // Daily task
})
cron.Start()
```
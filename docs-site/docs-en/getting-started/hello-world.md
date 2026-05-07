# Hello World

The simplest Go Web Frame application.

## Code

```go
package main

import (
    "context"
    "time"
    wf "github.com/chuccp/go-web-frame"
    "github.com/chuccp/go-web-frame/config"
    "github.com/chuccp/go-web-frame/log"
    "github.com/chuccp/go-web-frame/web"
)

func main() {
    builder := wf.NewBuilder(config.LoadAutoConfig())
    
    builder.Get("/", func(c *web.Request) (any, error) {
        return "Hello, World!", nil
    })
    
    app := builder.Build()
    
    // Graceful shutdown
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    // Auto shutdown after 10 seconds (demo)
    go func() {
        time.Sleep(time.Second * 10)
        cancel()
    }()
    
    if err := app.Run(ctx); err != nil {
        log.PrintPanic(err)
    }
}
```

## Key Points

1. **Builder Pattern**: `NewBuilder()` creates the builder
2. **Route Registration**: `builder.Get()` registers GET route
3. **Build**: `builder.Build()` creates the WebFrame instance
4. **Run**: `app.Run(ctx)` starts the server with graceful shutdown support

## Request Handler

Handler signature:

```go
func(c *web.Request) (any, error)
```

- `web.Request`: Request wrapper with helper methods
- Return `any` for response, `error` for errors

## Next Steps

- [Routing](../guide/routing.md) - Full routing guide
- [Controller](../guide/controller.md) - REST controllers
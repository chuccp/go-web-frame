# WebSocket & SSE

This document covers WebSocket and Server-Sent Events (SSE) usage in Go Web Frame.

## WebSocket

### Basic Usage

Register a WebSocket endpoint using `*web.WebSocketStream` (based on `coder/websocket`):

```go
func (c *MyController) Init(ctx *core.Context) error {
    ctx.WebSocket("/ws", func(stream *web.WebSocketStream) error {
        defer stream.Close()
        for {
            typ, message, err := stream.Read(stream.Context())
            if err != nil {
                break
            }
            stream.Write(stream.Context(), typ, message) // echo
        }
    })
    return nil
}
```

### Full Example (Chat Room)

```go
type ChatController struct {
    core.IService
    clients map[*web.WebSocketStream]bool
    mu      sync.Mutex
}

func (c *ChatController) Init(ctx *core.Context) error {
    c.clients = make(map[*web.WebSocketStream]bool)
    ctx.WebSocket("/ws/chat", c.HandleChat)
    return nil
}

func (c *ChatController) HandleChat(stream *web.WebSocketStream) error {
    c.mu.Lock()
    c.clients[stream] = true
    c.mu.Unlock()

    defer func() {
        c.mu.Lock()
        delete(c.clients, stream)
        c.mu.Unlock()
        stream.Close()
    }()

    for {
        _, message, err := stream.Read(stream.Context())
        if err != nil {
            break
        }
        c.mu.Lock()
        for client := range c.clients {
            err := client.WriteString(stream.Context(), string(message))
            if err != nil {
                client.Close()
                delete(c.clients, client)
            }
        }
        c.mu.Unlock()
    }
}
```

### WebSocketStream API

| Method | Description |
|------|------|
| `Read(ctx) (MessageType, []byte, error)` | Read message (any type) |
| `Write(ctx, typ, data)` | Write message |
| `WriteText(ctx, data)` | Write text message |
| `WriteString(ctx, s)` | Write string message |
| `WriteBinary(ctx, data)` | Write binary message |
| `ReadText(ctx) ([]byte, error)` | Read text message, error if non-text |
| `Ping(ctx)` | Send Ping frame |
| `Close()` | Close connection (graceful) |
| `Done() <-chan struct{}` | Returns close notification channel |
| `Context() context.Context` | Returns stream context |
| `Request() *web.Request` | Returns original HTTP request |

## SSE (Server-Sent Events)

Register an SSE endpoint for server push. `SetHeaders()` is called automatically by the framework:

```go
func (c *MyController) Init(ctx *core.Context) error {
    ctx.SSE("/events", func(stream *web.SSEStream) error {
        defer stream.Close()

        // Start heartbeat (every 30s, stops automatically on Close)
        stream.StartHeartbeat(30 * time.Second)

        // Send named event
        stream.Send("message", "Hello World")

        // Send default message (no event name)
        stream.SendMessage("ping")

        // Send event with ID (for client reconnection)
        stream.SendWithID("1", "update", `{"status":"ok"}`)

        // Set retry interval (milliseconds)
        stream.SendRetry(3000)

        // Send heartbeat
        stream.Heartbeat()

        return nil
    })
    return nil
}
```

### SSEStream API

| Method | Description |
|------|------|
| `Send(event, data)` | Send named event |
| `SendMessage(data)` | Send default message (no event name) |
| `SendWithID(id, event, data)` | Send event with ID for client reconnection via Last-Event-ID |
| `SendRetry(retryMs)` | Set client retry interval (milliseconds) |
| `Heartbeat()` | Send heartbeat comment to keep connection alive |
| `StartHeartbeat(interval)` | Start periodic heartbeat goroutine, auto-stops on Close |
| `Close()` | Close SSE stream (waits for background goroutine to exit) |
| `Done() <-chan struct{}` | Returns close notification channel |
| `SetHeader(key, value)` | Set custom response header |

> **Note:** `SetHeaders()` (Content-Type, Cache-Control, etc.) is called automatically by the converter. No manual call needed in handlers.

## Next Steps

- [Static Files & Proxy](static-proxy.md) - Static files and reverse proxy
- [Routing](../guide/routing.md) - Routing system
- [Controller](../guide/controller.md) - REST controllers
- [Deployment](deployment.md) - Production deployment

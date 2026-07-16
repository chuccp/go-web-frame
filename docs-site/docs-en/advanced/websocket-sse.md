# WebSocket & SSE

This document covers WebSocket and Server-Sent Events (SSE) usage in Go Web Frame.

## WebSocket

### Basic Usage

Register a WebSocket endpoint. The handler receives a `*web.WebSocket` — call `OpenStream()` to get a readable/writable stream (based on `coder/websocket`):

```go
func (c *MyController) Init(ctx *core.Context) error {
    ctx.WebSocket("/ws", func(ws *web.WebSocket) error {
        stream, err := ws.OpenStream()
        if err != nil {
            return err
        }
        defer stream.Close()
        for {
            typ, message, err := stream.Read(stream.Context())
            if err != nil {
                break
            }
            stream.Write(stream.Context(), typ, message) // echo
        }
        return nil
    })
    return nil
}
```

> **Changed in v1.0.14**: Handler signature changed from `func(stream *WebSocketStream)` to `func(ws *WebSocket)`. The connection is lazily initialized — the WebSocket handshake only happens on the first `OpenStream()` or `Read`/`Write` call.

### Configuring AcceptOptions

`OpenStream()` accepts variadic `AcceptOptions` to configure WebSocket accept behavior:

```go
stream, err := ws.OpenStream(
    web.WithOriginPatterns([]string{"example.com", "*.example.com"}),
    web.WithSubprotocols("chat"),
    web.WithCompressionMode(1),              // CompressionNoContextTakeover
    web.WithCompressionThreshold(128),       // minimum size before compression (bytes)
    web.WithInsecureSkipVerify(false),       // disable origin verification (not recommended)
    web.WithOnPingReceived(func(ctx context.Context, payload []byte) bool {
        return true  // return false to suppress automatic pong response
    }),
    web.WithOnPongReceived(func(ctx context.Context, payload []byte) {
        // callback when a pong frame is received
    }),
)
```

| Option Function | Description |
|----------------|-------------|
| `WithOriginPatterns(patterns []string)` | List of host patterns for authorized origins |
| `WithSubprotocols(protocols ...string)` | Subprotocols to negotiate during accept |
| `WithInsecureSkipVerify(skip bool)` | Disable origin verification (use OriginPatterns instead) |
| `WithCompressionMode(mode int)` | Compression mode: 0=Disabled, 1=NoContextTakeover, 2=ContextTakeover |
| `WithCompressionThreshold(threshold int)` | Minimum message size before compression is applied (bytes) |
| `WithOnPingReceived(fn)` | Ping frame callback, return false to suppress pong |
| `WithOnPongReceived(fn)` | Pong frame callback |

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

func (c *ChatController) HandleChat(ws *web.WebSocket) error {
    stream, err := ws.OpenStream()
    if err != nil {
        return err
    }

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
    return nil
}
```

### WebSocket API

| Method | Description |
|------|------|
| `OpenStream(opts ...AcceptOptions) (*WebSocketStream, error)` | Accepts the WebSocket connection and returns the stream (lazy init — handshake on first call) |
| `Request() *web.Request` | Returns the original HTTP request that initiated the WebSocket upgrade |
| `Close()` | Closes the WebSocket connection |

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

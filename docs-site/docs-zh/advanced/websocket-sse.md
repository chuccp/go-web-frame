# WebSocket 与 SSE

本文档介绍 Go Web Frame 中 WebSocket 和 Server-Sent Events (SSE) 的使用。

## WebSocket

### 基本用法

注册 WebSocket 端点，使用 `*web.WebSocketStream`（基于 `coder/websocket`）：

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

### 完整示例（聊天室）

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
    // 新客户端加入
    c.mu.Lock()
    c.clients[stream] = true
    c.mu.Unlock()

    defer func() {
        // 客户端离开
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
        // 广播消息给所有客户端
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

| 方法 | 说明 |
|------|------|
| `Read(ctx) (MessageType, []byte, error)` | 读取消息（任意类型） |
| `Write(ctx, typ, data)` | 写入消息 |
| `WriteText(ctx, data)` | 写入文本消息 |
| `WriteString(ctx, s)` | 写入字符串消息 |
| `WriteBinary(ctx, data)` | 写入二进制消息 |
| `ReadText(ctx) ([]byte, error)` | 读取文本消息，非文本则报错 |
| `Ping(ctx)` | 发送 Ping 帧 |
| `Close()` | 关闭连接（正常关闭） |
| `Done() <-chan struct{}` | 返回关闭通知 channel |
| `Context() context.Context` | 返回流的 context |
| `Request() *web.Request` | 返回原始 HTTP 请求 |

## SSE（Server-Sent Events）

注册 SSE 端点，用于服务器推送事件。`SetHeaders()` 由框架自动调用，无需手动设置：

```go
func (c *MyController) Init(ctx *core.Context) error {
    ctx.SSE("/events", func(stream *web.SSEStream) error {
        defer stream.Close()

        // 启动心跳保活（每 30 秒，Close 时自动停止）
        stream.StartHeartbeat(30 * time.Second)

        // 发送命名事件
        stream.Send("message", "Hello World")

        // 发送默认消息（无事件名）
        stream.SendMessage("ping")

        // 发送带 ID 的事件（客户端可断线重连）
        stream.SendWithID("1", "update", `{"status":"ok"}`)

        // 设置重连时间（毫秒）
        stream.SendRetry(3000)

        // 发送心跳
        stream.Heartbeat()

        return nil
    })
    return nil
}
```

### SSEStream API

| 方法 | 说明 |
|------|------|
| `Send(event, data)` | 发送命名事件 |
| `SendMessage(data)` | 发送默认消息（无事件名） |
| `SendWithID(id, event, data)` | 发送带 ID 的事件，客户端断线重连时可使用 Last-Event-ID |
| `SendRetry(retryMs)` | 设置客户端重连时间（毫秒） |
| `Heartbeat()` | 发送心跳注释，保持连接活跃 |
| `StartHeartbeat(interval)` | 启动定时心跳 goroutine，Close 时自动等待退出 |
| `Close()` | 关闭 SSE 流（等待后台 goroutine 退出） |
| `Done() <-chan struct{}` | 返回关闭通知 channel |
| `SetHeader(key, value)` | 设置自定义响应头 |

> **注意：** `SetHeaders()`（Content-Type、Cache-Control 等）由 converter 自动调用，无需在 handler 中手动调用。

## 下一步

- [静态文件与代理](static-proxy.md) - 静态文件和反向代理
- [路由](../guide/routing.md) - 路由系统
- [控制器](../guide/controller.md) - REST 控制器
- [部署](deployment.md) - 生产环境部署

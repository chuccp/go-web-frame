package main

import (
	"context"
	"fmt"
	"time"

	wf "github.com/chuccp/go-web-frame"
	config2 "github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/web"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type WebSocketController struct {
	core.IService
}

func (c *WebSocketController) Init(ctx *core.Context) error {
	// WebSocket endpoint - echo server
	ctx.WebSocket("/ws", func(conn *websocket.Conn) error {
		log.Info("WebSocket client connected")
		defer log.Info("WebSocket client disconnected")

		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				return err
			}

			log.Info("WebSocket received", zap.String("message", string(message)))

			// Echo back
			err = conn.WriteMessage(messageType, message)
			if err != nil {
				return err
			}
		}
	})

	// WebSocket endpoint with custom upgrader
	ctx.WebSocket("/ws/chat", func(conn *websocket.Conn) error {
		// Set read limit
		conn.SetReadLimit(512)

		// Set read deadline
		_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		// Set pong handler to reset read deadline
		conn.SetPongHandler(func(string) error {
			_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		done := make(chan struct{})

		// Read messages
		go func() {
			defer close(done)
			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					return
				}
				// Broadcast message logic would go here
				log.Info("Chat message", zap.String("message", string(message)))
			}
		}()

		// Ping and read
		for {
			select {
			case <-done:
				return nil
			case <-ticker.C:
				_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return err
				}
			}
		}
	})

	return nil
}

type SSEController struct {
	core.IService
}

func (c *SSEController) Init(ctx *core.Context) error {
	// SSE endpoint - time events
	ctx.SSE("/events/time", func(stream *web.SSEStream) error {
		log.Info("SSE client connected to time stream")
		defer log.Info("SSE client disconnected from time stream")

		// Start heartbeat to keep connection alive
		stopHeartbeat := stream.StartHeartbeat(15 * time.Second)
		defer close(stopHeartbeat)

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		count := 0
		for {
			select {
			case <-stream.Done():
				return nil
			case <-ticker.C:
				count++
				err := stream.Send("time", fmt.Sprintf(`{"count": %d, "time": "%s"}`, count, time.Now().Format(time.RFC3339)))
				if err != nil {
					return err
				}

				// Stop after 100 events (demo purpose)
				if count >= 100 {
					return nil
				}
			}
		}
	})

	// SSE endpoint - counter with custom events
	ctx.SSE("/events/counter", func(stream *web.SSEStream) error {
		log.Info("SSE client connected to counter stream")

		// Send initial retry setting (reconnect after 3 seconds if disconnected)
		_ = stream.SendRetry(3000)

		for i := 0; i < 10; i++ {
			select {
			case <-stream.Done():
				return nil
			default:
			}

			// Send different event types
			if i%3 == 0 {
				_ = stream.SendWithID(fmt.Sprintf("%d", i), "milestone", fmt.Sprintf("Milestone reached: %d", i))
			} else {
				_ = stream.Send("update", fmt.Sprintf("Counter: %d", i))
			}

			time.Sleep(time.Second)
		}

		// Send final message
		_ = stream.SendMessage("Counter finished!")
		return nil
	})

	return nil
}

func main() {
	builder := wf.NewBuilder(config2.LoadAutoConfig())

	// Add WebSocket and SSE controllers
	builder.Rest(&WebSocketController{})
	builder.Rest(&SSEController{})

	// Add a simple index page with JavaScript clients
	builder.Get("/", func(c *web.Request) (any, error) {
		return `<!DOCTYPE html>
<html>
<head>
    <title>WebSocket & SSE Demo</title>
</head>
<body>
    <h1>WebSocket & SSE Demo</h1>

    <h2>WebSocket Echo</h2>
    <div id="ws-messages"></div>
    <input type="text" id="ws-input" placeholder="Type message...">
    <button onclick="sendWS()">Send</button>

    <h2>SSE Time Stream</h2>
    <div id="sse-time"></div>

    <h2>SSE Counter</h2>
    <div id="sse-counter"></div>

    <script>
        // WebSocket
        const ws = new WebSocket('ws://' + location.host + '/ws');
        ws.onmessage = (e) => {
            document.getElementById('ws-messages').innerHTML += '<p>Received: ' + e.data + '</p>';
        };
        function sendWS() {
            const input = document.getElementById('ws-input');
            ws.send(input.value);
            input.value = '';
        }

        // SSE Time
        const sseTime = new EventSource('/events/time');
        sseTime.addEventListener('time', (e) => {
            document.getElementById('sse-time').innerHTML += '<p>' + e.data + '</p>';
        });

        // SSE Counter
        const sseCounter = new EventSource('/events/counter');
        sseCounter.addEventListener('update', (e) => {
            document.getElementById('sse-counter').innerHTML += '<p>' + e.data + '</p>';
        });
        sseCounter.addEventListener('milestone', (e) => {
            document.getElementById('sse-counter').innerHTML += '<p><strong>' + e.data + '</strong></p>';
        });
    </script>
</body>
</html>`, nil
	})

	app := builder.Build()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Info("Server starting on :19009")
	log.Info("Open http://localhost:19009 in your browser")

	if err := app.Run(ctx); err != nil {
		log.PrintPanic(err)
	}
}

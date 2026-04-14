package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestContext_WebSocketRegistersInRouteTree(t *testing.T) {
	handles := NewHandles()

	info := handles.AddWebSocket("/ws", func(conn *websocket.Conn) error {
		return nil
	}, nil)

	assert.Equal(t, "/ws", info.RelativePath())
	assert.True(t, info.IsWebSocket())
	assert.NotNil(t, info.Upgrader())
	assert.True(t, handles.HasHandler(http.MethodGet, "/ws"))
}

func TestContext_SSERegistersInRouteTree(t *testing.T) {
	handles := NewHandles()

	info := handles.AddSSE("/events", func(stream *SSEStream) error {
		return nil
	})

	assert.Equal(t, "/events", info.RelativePath())
	assert.True(t, info.IsSSE())
	assert.True(t, handles.HasHandler(http.MethodGet, "/events"))
}

func TestWebSocket_EchoHandler(t *testing.T) {
	handles := NewHandles()
	handles.AddWebSocket("/ws", func(conn *websocket.Conn) error {
		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				return err
			}
			err = conn.WriteMessage(messageType, message)
			if err != nil {
				return err
			}
		}
	}, nil)

	server := NewHttpServer(DefaultServerConfig(), NewCertManager())
	server.Handle(NewHandlerConfig(nil, handles, DefaultServerConfig()))

	ts := httptest.NewServer(server.Engine())
	defer ts.Close()

	// Convert http URL to ws URL
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")

	// Connect as WebSocket client
	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial(wsURL+"/ws", nil)
	assert.NoError(t, err)
	defer conn.Close()

	// Send message
	testMessage := "Hello WebSocket!"
	err = conn.WriteMessage(websocket.TextMessage, []byte(testMessage))
	assert.NoError(t, err)

	// Read echoed message
	messageType, message, err := conn.ReadMessage()
	assert.NoError(t, err)
	assert.Equal(t, websocket.TextMessage, messageType)
	assert.Equal(t, testMessage, string(message))
}

func TestSSE_EventHandler(t *testing.T) {
	handles := NewHandles()
	handles.AddSSE("/events", func(stream *SSEStream) error {
		stream.Send("test", "hello")
		return nil
	})

	server := NewHttpServer(DefaultServerConfig(), NewCertManager())
	server.Handle(NewHandlerConfig(nil, handles, DefaultServerConfig()))

	ts := httptest.NewServer(server.Engine())
	defer ts.Close()

	// Connect as SSE client
	resp, err := http.Get(ts.URL + "/events")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))
}

func TestSSEStream_SendAndHeartbeat(t *testing.T) {
	w := httptest.NewRecorder()
	stream := NewSSEStream(w)
	assert.NotNil(t, stream)

	stream.SetHeaders()

	// Test Send
	err := stream.Send("message", "test data")
	assert.NoError(t, err)

	// Test SendMessage
	err = stream.SendMessage("plain message")
	assert.NoError(t, err)

	// Test Heartbeat
	err = stream.Heartbeat()
	assert.NoError(t, err)

	// Test SendWithID
	err = stream.SendWithID("123", "event", "data with id")
	assert.NoError(t, err)

	// Test SendRetry
	err = stream.SendRetry(5000)
	assert.NoError(t, err)

	stream.Close()

	// Verify closed stream returns error
	err = stream.Send("test", "should fail")
	assert.Error(t, err)
}

func TestDefaultWebSocketUpgrader(t *testing.T) {
	upgrader := DefaultWebSocketUpgrader()
	assert.NotNil(t, upgrader)
	assert.Equal(t, 1024, upgrader.ReadBufferSize)
	assert.Equal(t, 1024, upgrader.WriteBufferSize)
	assert.True(t, upgrader.CheckOrigin(nil))
}

func TestWebSocketUpgraderWithOptions(t *testing.T) {
	checkOrigin := func(r *http.Request) bool {
		if r == nil {
			return false
		}
		return r.Header.Get("Origin") == "https://example.com"
	}
	upgrader := WebSocketUpgraderWithOptions(2048, 4096, checkOrigin)
	assert.NotNil(t, upgrader)
	assert.Equal(t, 2048, upgrader.ReadBufferSize)
	assert.Equal(t, 4096, upgrader.WriteBufferSize)

	// Test with a request
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	req.Header.Set("Origin", "https://example.com")
	assert.True(t, upgrader.CheckOrigin(req))

	req.Header.Set("Origin", "https://other.com")
	assert.False(t, upgrader.CheckOrigin(req))
}

func TestSSEStream_HeartbeatGoroutine(t *testing.T) {
	w := httptest.NewRecorder()
	stream := NewSSEStream(w)
	stream.SetHeaders()

	// Start heartbeat
	stop := stream.StartHeartbeat(50 * time.Millisecond)

	// Wait for a few heartbeats
	time.Sleep(150 * time.Millisecond)

	// Stop heartbeat
	close(stop)
	stream.Close()

	// Verify heartbeat was sent (body should contain heartbeat comments)
	body := w.Body.String()
	assert.Contains(t, body, ": heartbeat")
}
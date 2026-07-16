// Package web: integration tests for SSE, WebSocket, and ReverseProxy.
package web

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
)

// newTestServer creates an httptest.Server from a configured Server,
// registering all routes on the gin engine without binding a real port.
func newTestServer(t *testing.T, setup func(server *Server)) *httptest.Server {
	t.Helper()
	servers := NewServers()
	server, err := servers.CreateServer(&ServerConfig{Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	setup(server)
	server.initRoute()
	ts := httptest.NewServer(server.engine)
	t.Cleanup(ts.Close)
	return ts
}

// ---------------------------------------------------------------------------
// SSE Tests
// ---------------------------------------------------------------------------

func TestSSE_Send(t *testing.T) {
	ts := newTestServer(t, func(server *Server) {
		server.AddSSE("/events", func(stream *SSEStream) error {
			defer stream.Close()
			return stream.Send("greeting", "hello")
		})
	})

	resp, err := http.Get(ts.URL + "/events")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("Content-Type = %q, want text/event-stream", ct)
	}

	body, _ := io.ReadAll(resp.Body)
	want := "event: greeting\ndata: hello\n\n"
	if string(body) != want {
		t.Errorf("body = %q, want %q", string(body), want)
	}
}

func TestSSE_SendMessage(t *testing.T) {
	ts := newTestServer(t, func(server *Server) {
		server.AddSSE("/events", func(stream *SSEStream) error {
			defer stream.Close()
			return stream.SendMessage("no-event-data")
		})
	})

	resp, err := http.Get(ts.URL + "/events")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	want := "data: no-event-data\n\n"
	if string(body) != want {
		t.Errorf("body = %q, want %q", string(body), want)
	}
}

func TestSSE_SendWithID(t *testing.T) {
	ts := newTestServer(t, func(server *Server) {
		server.AddSSE("/events", func(stream *SSEStream) error {
			defer stream.Close()
			return stream.SendWithID("42", "update", "payload")
		})
	})

	resp, err := http.Get(ts.URL + "/events")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	want := "id: 42\nevent: update\ndata: payload\n\n"
	if string(body) != want {
		t.Errorf("body = %q, want %q", string(body), want)
	}
}

func TestSSE_SendRetry(t *testing.T) {
	ts := newTestServer(t, func(server *Server) {
		server.AddSSE("/events", func(stream *SSEStream) error {
			defer stream.Close()
			if err := stream.SendRetry(5000); err != nil {
				return err
			}
			return stream.SendMessage("after-retry")
		})
	})

	resp, err := http.Get(ts.URL + "/events")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	want := "retry: 5000\n\ndata: after-retry\n\n"
	if string(body) != want {
		t.Errorf("body = %q, want %q", string(body), want)
	}
}

func TestSSE_Heartbeat(t *testing.T) {
	ts := newTestServer(t, func(server *Server) {
		server.AddSSE("/events", func(stream *SSEStream) error {
			defer stream.Close()
			return stream.Heartbeat()
		})
	})

	resp, err := http.Get(ts.URL + "/events")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	want := ": heartbeat\n\n"
	if string(body) != want {
		t.Errorf("body = %q, want %q", string(body), want)
	}
}

func TestSSE_MultipleEvents(t *testing.T) {
	ts := newTestServer(t, func(server *Server) {
		server.AddSSE("/events", func(stream *SSEStream) error {
			defer stream.Close()
			for i := 0; i < 3; i++ {
				if err := stream.Send("tick", fmt.Sprintf("%d", i)); err != nil {
					return err
				}
			}
			return nil
		})
	})

	resp, err := http.Get(ts.URL + "/events")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	want := "event: tick\ndata: 0\n\nevent: tick\ndata: 1\n\nevent: tick\ndata: 2\n\n"
	if string(body) != want {
		t.Errorf("body = %q, want %q", string(body), want)
	}
}

func TestSSE_StreamContext(t *testing.T) {
	ts := newTestServer(t, func(server *Server) {
		server.AddSSE("/events", func(stream *SSEStream) error {
			defer stream.Close()
			select {
			case <-stream.Done():
				t.Error("stream.Done() should not be closed before Close()")
			default:
			}
			stream.Close()
			select {
			case <-stream.Done():
				// expected
			case <-time.After(time.Second):
				t.Error("stream.Done() should be closed after Close()")
			}
			return nil
		})
	})

	resp, err := http.Get(ts.URL + "/events")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	io.ReadAll(resp.Body)
}

func TestSSE_StartHeartbeat(t *testing.T) {
	ts := newTestServer(t, func(server *Server) {
		server.AddSSE("/events", func(stream *SSEStream) error {
			stream.StartHeartbeat(50 * time.Millisecond)
			time.Sleep(150 * time.Millisecond)
			stream.Close()
			return nil
		})
	})

	resp, err := http.Get(ts.URL + "/events")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	count := strings.Count(string(body), ": heartbeat\n\n")
	if count < 2 {
		t.Errorf("expected at least 2 heartbeats, got %d", count)
	}
}

func TestSSE_SetHeader(t *testing.T) {
	ts := newTestServer(t, func(server *Server) {
		server.AddSSE("/events", func(stream *SSEStream) error {
			defer stream.Close()
			stream.SetHeader("X-Custom", "test-value")
			return stream.Send("test", "data")
		})
	})

	resp, err := http.Get(ts.URL + "/events")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if v := resp.Header.Get("X-Custom"); v != "test-value" {
		t.Errorf("X-Custom = %q, want %q", v, "test-value")
	}
}

// ---------------------------------------------------------------------------
// WebSocket Tests
// ---------------------------------------------------------------------------

func TestWebSocket_Echo(t *testing.T) {
	ts := newTestServer(t, func(server *Server) {
		server.AddWebSocket("/ws", func(webSocket *WebSocket) error {
			stream, err := webSocket.OpenStream()
			if err != nil {
				return err
			}
			for {
				typ, data, err := stream.Read(stream.Context())
				if err != nil {
					return nil
				}
				if err := stream.Write(stream.Context(), typ, data); err != nil {
					return err
				}
			}
		})
	})

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	messages := []string{"hello", "world", "foo"}
	for _, msg := range messages {
		if err := conn.Write(ctx, websocket.MessageText, []byte(msg)); err != nil {
			t.Fatal(err)
		}
		_, data, err := conn.Read(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != msg {
			t.Errorf("echo = %q, want %q", string(data), msg)
		}
	}
}

func TestWebSocket_Binary(t *testing.T) {
	ts := newTestServer(t, func(server *Server) {
		server.AddWebSocket("/ws", func(webSocket *WebSocket) error {
			stream, err := webSocket.OpenStream()
			if err != nil {
				return err
			}
			for {
				typ, data, err := stream.Read(stream.Context())
				if err != nil {
					return nil
				}
				if err := stream.Write(stream.Context(), typ, data); err != nil {
					return err
				}
			}
		})
	})

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	binData := []byte{0x00, 0x01, 0x02, 0xFF}
	if err := conn.Write(ctx, websocket.MessageBinary, binData); err != nil {
		t.Fatal(err)
	}
	_, data, err := conn.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != len(binData) {
		t.Fatalf("binary length = %d, want %d", len(data), len(binData))
	}
	for i, b := range data {
		if b != binData[i] {
			t.Errorf("byte[%d] = %02x, want %02x", i, b, binData[i])
		}
	}
}

func TestWebSocket_Ping(t *testing.T) {
	ts := newTestServer(t, func(server *Server) {
		server.AddWebSocket("/ws", func(webSocket *WebSocket) error {
			stream, err := webSocket.OpenStream()
			if err != nil {
				return err
			}
			defer stream.Close()
			return stream.Ping(stream.Context())
		})
	})

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	// Server sends ping then closes; read to unblock.
	conn.Read(ctx)
}

func TestWebSocket_WriteString(t *testing.T) {
	ts := newTestServer(t, func(server *Server) {
		server.AddWebSocket("/ws", func(webSocket *WebSocket) error {
			stream, err := webSocket.OpenStream()
			if err != nil {
				return err
			}
			defer stream.Close()
			return stream.WriteString(stream.Context(), "string-reply")
		})
	})

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	if err := conn.Write(ctx, websocket.MessageText, []byte("trigger")); err != nil {
		t.Fatal(err)
	}

	typ, data, err := conn.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if typ != websocket.MessageText {
		t.Errorf("type = %v, want MessageText", typ)
	}
	if string(data) != "string-reply" {
		t.Errorf("data = %q, want %q", string(data), "string-reply")
	}
}

func TestWebSocket_ContextDone(t *testing.T) {
	handlerDone := make(chan struct{})
	ts := newTestServer(t, func(server *Server) {
		server.AddWebSocket("/ws", func(webSocket *WebSocket) error {
			stream, err := webSocket.OpenStream()
			if err != nil {
				return err
			}
			defer close(handlerDone)
			for {
				_, _, err := stream.Read(stream.Context())
				if err != nil {
					return nil
				}
			}
		})
	})

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}

	conn.Write(ctx, websocket.MessageText, []byte("hi"))
	conn.Read(ctx)

	conn.Close(websocket.StatusNormalClosure, "bye")

	select {
	case <-handlerDone:
	case <-time.After(2 * time.Second):
		t.Error("handler did not exit after client disconnect")
	}
}

func TestWebSocket_ReadText(t *testing.T) {
	ts := newTestServer(t, func(server *Server) {
		server.AddWebSocket("/ws", func(webSocket *WebSocket) error {
			stream, err := webSocket.OpenStream()
			if err != nil {
				return err
			}
			defer stream.Close()
			data, err := stream.ReadText(stream.Context())
			if err != nil {
				return err
			}
			return stream.WriteString(stream.Context(), "got:"+string(data))
		})
	})

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	if err := conn.Write(ctx, websocket.MessageText, []byte("ping")); err != nil {
		t.Fatal(err)
	}

	_, data, err := conn.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "got:ping" {
		t.Errorf("data = %q, want %q", string(data), "got:ping")
	}
}

func TestWebSocket_TextAndBinary(t *testing.T) {
	ts := newTestServer(t, func(server *Server) {
		server.AddWebSocket("/ws", func(webSocket *WebSocket) error {
			stream, err := webSocket.OpenStream()
			if err != nil {
				return err
			}
			defer stream.Close()
			for {
				typ, data, err := stream.Read(stream.Context())
				if err != nil {
					return nil
				}
				switch typ {
				case websocket.MessageText:
					if err := stream.WriteString(stream.Context(), "text:"+string(data)); err != nil {
						return err
					}
				case websocket.MessageBinary:
					if err := stream.WriteBinary(stream.Context(), append([]byte("bin:"), data...)); err != nil {
						return err
					}
				}
			}
		})
	})

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	// Text round-trip
	if err := conn.Write(ctx, websocket.MessageText, []byte("hello")); err != nil {
		t.Fatal(err)
	}
	_, data, err := conn.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "text:hello" {
		t.Errorf("text echo = %q, want %q", string(data), "text:hello")
	}

	// Binary round-trip
	if err := conn.Write(ctx, websocket.MessageBinary, []byte{0xDE, 0xAD}); err != nil {
		t.Fatal(err)
	}
	_, data, err = conn.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 6 || string(data[:4]) != "bin:" {
		t.Errorf("binary echo = %v, want [bin: 0xDE 0xAD]", data)
	}
}

// ---------------------------------------------------------------------------
// ReverseProxy Tests
// ---------------------------------------------------------------------------

func TestReverseProxy_Get(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Backend", "true")
		fmt.Fprintf(w, "backend:%s", r.URL.Path)
	}))
	defer backend.Close()

	ts := newTestServer(t, func(server *Server) {
		server.AddReverseProxy("/api", backend.URL)
	})

	resp, err := http.Get(ts.URL + "/api/hello")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "backend:/hello" {
		t.Errorf("body = %q, want %q", string(body), "backend:/hello")
	}
	if resp.Header.Get("X-Backend") != "true" {
		t.Error("X-Backend header not forwarded")
	}
}

func TestReverseProxy_Post(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("backend method = %s, want POST", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"echo":"%s"}`, string(body))
	}))
	defer backend.Close()

	ts := newTestServer(t, func(server *Server) {
		server.AddReverseProxy("/api", backend.URL)
	})

	resp, err := http.Post(ts.URL+"/api/data", "application/json", strings.NewReader("hello"))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	want := `{"echo":"hello"}`
	if string(body) != want {
		t.Errorf("body = %q, want %q", string(body), want)
	}
}

func TestReverseProxy_PathRewrite(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "path:%s", r.URL.Path)
	}))
	defer backend.Close()

	ts := newTestServer(t, func(server *Server) {
		server.AddReverseProxy("/api/v1", backend.URL+"/backend")
	})

	resp, err := http.Get(ts.URL + "/api/v1/users/123")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "path:/backend/users/123" {
		t.Errorf("body = %q, want %q", string(body), "path:/backend/users/123")
	}
}

func TestReverseProxy_XForwardedHeaders(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "for:%s|proto:%s|host:%s",
			r.Header.Get("X-Forwarded-For"),
			r.Header.Get("X-Forwarded-Proto"),
			r.Header.Get("X-Forwarded-Host"))
	}))
	defer backend.Close()

	ts := newTestServer(t, func(server *Server) {
		server.AddReverseProxy("/api", backend.URL)
	})

	resp, err := http.Get(ts.URL + "/api/test")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "for:") {
		t.Errorf("X-Forwarded-For not set: %s", string(body))
	}
}

func TestReverseProxy_AllMethods(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "method:%s", r.Method)
	}))
	defer backend.Close()

	ts := newTestServer(t, func(server *Server) {
		server.AddReverseProxy("/api", backend.URL)
	})

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
	for _, method := range methods {
		req, _ := http.NewRequest(method, ts.URL+"/api/test", nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("%s: %v", method, err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		want := "method:" + method
		if string(body) != want {
			t.Errorf("%s: body = %q, want %q", method, string(body), want)
		}
	}
}

func TestReverseProxy_BackendDown(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	backendURL := backend.URL
	backend.Close()

	ts := newTestServer(t, func(server *Server) {
		server.AddReverseProxy("/api", backendURL)
	})

	resp, err := http.Get(ts.URL + "/api/test")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadGateway && resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status = %d, want 502 or 500", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// Combined: SSE + WebSocket on same server
// ---------------------------------------------------------------------------

func TestCombined_SSEAndWS(t *testing.T) {
	ts := newTestServer(t, func(server *Server) {
		server.AddSSE("/sse", func(stream *SSEStream) error {
			defer stream.Close()
			return stream.Send("sse", "from-sse")
		})
		server.AddWebSocket("/ws", func(webSocket *WebSocket) error {
			stream, err := webSocket.OpenStream()
			if err != nil {
				return err
			}
			defer stream.Close()
			return stream.WriteString(stream.Context(), "from-ws")
		})
	})

	// Test SSE
	resp, err := http.Get(ts.URL + "/sse")
	if err != nil {
		t.Fatal(err)
	}
	sseBody, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(sseBody), "from-sse") {
		t.Errorf("SSE body = %q, want contains 'from-sse'", string(sseBody))
	}

	// Test WS
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	_, data, err := conn.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "from-ws" {
		t.Errorf("WS data = %q, want %q", string(data), "from-ws")
	}
}

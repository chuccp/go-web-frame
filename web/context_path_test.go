package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestHttpServer_JoinContextPath(t *testing.T) {
	tests := []struct {
		name         string
		contextPath  string
		relativePath string
		expected     string
	}{
		{
			name:         "empty context path",
			contextPath:  "",
			relativePath: "/users",
			expected:     "/users",
		},
		{
			name:         "context path without leading slash",
			contextPath:  "api",
			relativePath: "/users",
			expected:     "/api/users",
		},
		{
			name:         "context path with leading slash",
			contextPath:  "/api",
			relativePath: "/users",
			expected:     "/api/users",
		},
		{
			name:         "context path with trailing slash",
			contextPath:  "/api/",
			relativePath: "/users",
			expected:     "/api/users",
		},
		{
			name:         "root path",
			contextPath:  "/api",
			relativePath: "/",
			expected:     "/api/",
		},
		{
			name:         "relative path without leading slash",
			contextPath:  "/api",
			relativePath: "users",
			expected:     "/api/users",
		},
		{
			name:         "both empty",
			contextPath:  "",
			relativePath: "/",
			expected:     "/",
		},
		{
			name:         "nested path",
			contextPath:  "/api/v1",
			relativePath: "/users/:id",
			expected:     "/api/v1/users/:id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := joinContextPath(tt.contextPath, tt.relativePath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHttpServer_HandleWithContextPath(t *testing.T) {
	// Create server config with context path
	serverConfig := &ServerConfig{
		Port:        19009,
		ContextPath: "/api",
	}

	handles := NewHandles()
	handles.Handle(http.MethodGet, "/hello", func(req *Request) (any, error) {
		return "Hello, World!", nil
	})

	server := NewHttpServer(serverConfig, NewCertManager())
	server.AddHandle(NewHandlerConfig(nil, handles, serverConfig))
	server.Handle()

	ts := httptest.NewServer(server.Engine())
	defer ts.Close()

	// Test without context path - should return 404
	resp, err := http.Get(ts.URL + "/hello")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()

	// Test with context path - should return 200
	resp, err = http.Get(ts.URL + "/api/hello")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

func TestHttpServer_WebSocketWithContextPath(t *testing.T) {
	serverConfig := &ServerConfig{
		Port:        19009,
		ContextPath: "/ws",
	}

	handles := NewHandles()
	handles.AddWebSocket("/echo", func(conn *websocket.Conn) error {
		return nil
	}, nil)

	server := NewHttpServer(serverConfig, NewCertManager())
	server.AddHandle(NewHandlerConfig(nil, handles, serverConfig))
	server.Handle()

	// Verify route is registered with context path
	routes := server.Engine().Routes()
	var found bool
	for _, route := range routes {
		if route.Path == "/ws/echo" && route.Method == http.MethodGet {
			found = true
			break
		}
	}
	assert.True(t, found, "WebSocket route should be registered with context path")
}

func TestHttpServer_SSEWithContextPath(t *testing.T) {
	serverConfig := &ServerConfig{
		Port:        19009,
		ContextPath: "/events",
	}

	handles := NewHandles()
	handles.AddSSE("/stream", func(stream *SSEStream) error {
		return nil
	})

	server := NewHttpServer(serverConfig, NewCertManager())
	server.AddHandle(NewHandlerConfig(nil, handles, serverConfig))
	server.Handle()

	// Verify route is registered with context path
	routes := server.Engine().Routes()
	var found bool
	for _, route := range routes {
		if route.Path == "/events/stream" && route.Method == http.MethodGet {
			found = true
			break
		}
	}
	assert.True(t, found, "SSE route should be registered with context path")
}

func TestHttpServer_StaticFsWithContextPath(t *testing.T) {
	serverConfig := &ServerConfig{
		Port:        19009,
		ContextPath: "/static",
	}

	handles := NewHandles()
	handles.AddStaticFs("/assets", http.Dir("."))

	server := NewHttpServer(serverConfig, NewCertManager())
	server.AddHandle(NewHandlerConfig(nil, handles, serverConfig))
	server.Handle()

	// Verify routes are registered with context path
	routes := server.Engine().Routes()
	var getFound, headFound bool
	for _, route := range routes {
		if route.Path == "/static/assets/*filepath" {
			if route.Method == http.MethodGet {
				getFound = true
			}
			if route.Method == http.MethodHead {
				headFound = true
			}
		}
	}
	assert.True(t, getFound, "Static GET route should be registered with context path")
	assert.True(t, headFound, "Static HEAD route should be registered with context path")
}

func TestHandlerInfo_FullPath(t *testing.T) {
	handles := NewHandles()
	info := handles.Handle(http.MethodGet, "/users", func(req *Request) (any, error) {
		return "ok", nil
	})

	// Before Handle is called on server, fullPath should be empty
	assert.Empty(t, info.FullPath())
	assert.Equal(t, "/users", info.RelativePath())

	// After Handle is called on server, fullPath should be set
	serverConfig := DefaultServerConfig()
	server := NewHttpServer(serverConfig, NewCertManager())
	server.AddHandle(NewHandlerConfig(nil, handles, serverConfig))
	server.Handle()

	assert.Equal(t, "/users", info.FullPath())
}

func TestHandlerInfo_FullPathWithContextPath(t *testing.T) {
	serverConfig := &ServerConfig{
		Port:        19009,
		ContextPath: "/api/v1",
	}

	handles := NewHandles()
	info := handles.Handle(http.MethodGet, "/users/:id", func(req *Request) (any, error) {
		return "ok", nil
	})

	server := NewHttpServer(serverConfig, NewCertManager())
	server.AddHandle(NewHandlerConfig(nil, handles, serverConfig))
	server.Handle()

	assert.Equal(t, "/users/:id", info.RelativePath())
	assert.Equal(t, "/api/v1/users/:id", info.FullPath())
}

func TestRequest_ContextPath(t *testing.T) {
	serverConfig := &ServerConfig{
		Port:        19009,
		ContextPath: "/api",
	}

	handles := NewHandles()
	var capturedContextPath string
	handles.Handle(http.MethodGet, "/test", func(req *Request) (any, error) {
		capturedContextPath = req.ContextPath()
		return "ok", nil
	})

	server := NewHttpServer(serverConfig, NewCertManager())
	server.AddHandle(NewHandlerConfig(nil, handles, serverConfig))
	server.Handle()

	ts := httptest.NewServer(server.Engine())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/test")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "/api", capturedContextPath)
	resp.Body.Close()
}

func TestRequest_ContextPath_Empty(t *testing.T) {
	serverConfig := DefaultServerConfig()

	handles := NewHandles()
	var capturedContextPath string
	handles.Handle(http.MethodGet, "/test", func(req *Request) (any, error) {
		capturedContextPath = req.ContextPath()
		return "ok", nil
	})

	server := NewHttpServer(serverConfig, NewCertManager())
	server.AddHandle(NewHandlerConfig(nil, handles, serverConfig))
	server.Handle()

	ts := httptest.NewServer(server.Engine())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/test")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Empty(t, capturedContextPath)
	resp.Body.Close()
}

func TestHandles_AddReverseProxy(t *testing.T) {
	handles := NewHandles()
	info := handles.AddReverseProxy("/proxy", "http://localhost:8080")

	assert.Equal(t, "/proxy", info.RelativePath())
	assert.True(t, info.IsReverseProxy())
	assert.Equal(t, "http://localhost:8080", info.TargetUrl())

	// Reverse proxy should register all HTTP methods
	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete} {
		assert.True(t, handles.HasHandler(method, "/proxy"), "method %s should be registered", method)
	}
}
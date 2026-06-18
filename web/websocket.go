package web

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// DefaultWebSocketUpgrader returns a websocket.Upgrader with sensible defaults:
// 1024-byte read/write buffers and permissive origin checking.
func DefaultWebSocketUpgrader() *websocket.Upgrader {
	return &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// Allow all origins by default (configure in production)
			return true
		},
	}
}

// WebSocketUpgraderWithOptions returns a websocket.Upgrader with custom buffer sizes
// and origin checking. If checkOrigin is nil, all origins are allowed.
func WebSocketUpgraderWithOptions(readBufferSize, writeBufferSize int, checkOrigin func(r *http.Request) bool) *websocket.Upgrader {
	if checkOrigin == nil {
		checkOrigin = func(r *http.Request) bool { return true }
	}
	return &websocket.Upgrader{
		ReadBufferSize:  readBufferSize,
		WriteBufferSize: writeBufferSize,
		CheckOrigin:     checkOrigin,
	}
}
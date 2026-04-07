package web

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// DefaultWebSocketUpgrader returns a websocket.Upgrader with sensible defaults
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

// WebSocketUpgraderWithOptions returns a websocket.Upgrader with custom options
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
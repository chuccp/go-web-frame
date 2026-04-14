// Package web provides the HTTP routing, request handling, and response layer.
package web

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// WebSocketHandler handles WebSocket connections.
type WebSocketHandler func(conn *websocket.Conn) error

// SSEHandler handles Server-Sent Events streams.
type SSEHandler func(stream *SSEStream) error

// SSEStream represents a Server-Sent Events stream.
type SSEStream struct {
	writer http.ResponseWriter
	flush  http.Flusher
	done   chan struct{}
}

// HandlerInfo contains metadata about a registered route handler.
type HandlerInfo struct {
	path        string
	fullPath    string
	handlerMeta *HandlerMeta
	handlers    []HandlerFunc
	fs          http.FileSystem     // 静态文件系统，用于 StaticFs
	targetUrl   string              // 反向代理目标 URL，用于 ReverseProxy
	wsHandler   WebSocketHandler    // WebSocket handler
	sseHandler  SSEHandler          // SSE handler
	wsUpgrader  *websocket.Upgrader // WebSocket upgrader config
}

// RelativePath returns the route path as registered (without context path prefix).
func (hi *HandlerInfo) RelativePath() string {
	return hi.path
}

// FullPath returns the complete route path including context path prefix.
func (hi *HandlerInfo) FullPath() string {
	return hi.fullPath
}

// HandlerMeta returns the metadata attached to this handler.
func (hi *HandlerInfo) HandlerMeta() *HandlerMeta {
	return hi.handlerMeta
}

// HandlerFunc returns the handler functions for this route.
func (hi *HandlerInfo) HandlerFunc() []HandlerFunc {
	return hi.handlers
}

// FileSystem returns the http.FileSystem for static file routes.
func (hi *HandlerInfo) FileSystem() http.FileSystem {
	return hi.fs
}

// TargetUrl returns the target URL for reverse proxy routes.
func (hi *HandlerInfo) TargetUrl() string {
	return hi.targetUrl
}

// WebSocketHandler returns the WebSocket handler function.
func (hi *HandlerInfo) WebSocketHandler() WebSocketHandler {
	return hi.wsHandler
}

// SSEHandler returns the SSE handler function.
func (hi *HandlerInfo) SSEHandler() SSEHandler {
	return hi.sseHandler
}

// Upgrader returns the WebSocket upgrader configuration.
func (hi *HandlerInfo) Upgrader() *websocket.Upgrader {
	return hi.wsUpgrader
}

// IsStaticFs reports whether this route serves static files.
func (hi *HandlerInfo) IsStaticFs() bool {
	return hi.fs != nil
}

// IsReverseProxy reports whether this route is a reverse proxy.
func (hi *HandlerInfo) IsReverseProxy() bool {
	return hi.targetUrl != ""
}

// IsWebSocket reports whether this route handles WebSocket connections.
func (hi *HandlerInfo) IsWebSocket() bool {
	return hi.wsHandler != nil
}

// IsSSE reports whether this route handles Server-Sent Events.
func (hi *HandlerInfo) IsSSE() bool {
	return hi.sseHandler != nil
}

// RouteInfo is a collection of HandlerInfo for a route.
type RouteInfo []*HandlerInfo

// NewHandlerInfo creates a new HandlerInfo for a standard HTTP handler.
func NewHandlerInfo(path string, handlers ...HandlerFunc) *HandlerInfo {
	return &HandlerInfo{handlerMeta: NewHandlerMeta(), path: path, handlers: handlers}
}

// NewStaticFsHandlerInfo creates a new HandlerInfo for static file serving.
func NewStaticFsHandlerInfo(relativePath string, fs http.FileSystem) *HandlerInfo {
	return &HandlerInfo{handlerMeta: NewHandlerMeta(), path: relativePath, fs: fs}
}

// NewReverseProxyHandlerInfo creates a new HandlerInfo for reverse proxy.
func NewReverseProxyHandlerInfo(relativePath string, targetUrl string) *HandlerInfo {
	return &HandlerInfo{handlerMeta: NewHandlerMeta(), path: relativePath, targetUrl: targetUrl}
}

// NewWebSocketHandlerInfo creates a new HandlerInfo for WebSocket.
func NewWebSocketHandlerInfo(relativePath string, handler WebSocketHandler, upgrader *websocket.Upgrader) *HandlerInfo {
	return &HandlerInfo{handlerMeta: NewHandlerMeta(), path: relativePath, wsHandler: handler, wsUpgrader: upgrader}
}

// NewSSEHandlerInfo creates a new HandlerInfo for Server-Sent Events.
func NewSSEHandlerInfo(relativePath string, handler SSEHandler) *HandlerInfo {
	return &HandlerInfo{handlerMeta: NewHandlerMeta(), path: relativePath, sseHandler: handler}
}

// RouteTree maps HTTP methods to their registered route information.
type RouteTree map[string]RouteInfo

// Set adds a handler to the route tree for the given method.
func (rt RouteTree) Set(method string, handlerInfo *HandlerInfo) {
	rt[method] = append(rt[method], handlerInfo)
}

// Has reports whether a route is registered for the given method and path.
func (rt RouteTree) Has(method, path string) bool {
	if rt[method] != nil {
		for _, info := range rt[method] {
			if info.path == path {
				return true
			}
		}
	}
	return false
}

// GetHandlerMeta returns the HandlerMeta for a route by method and full path.
func (rt RouteTree) GetHandlerMeta(method, path string) *HandlerMeta {
	if rt[method] != nil {
		for _, info := range rt[method] {
			if info.fullPath == path {
				return info.handlerMeta
			}
		}
	}
	return NewHandlerMeta()
}

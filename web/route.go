package web

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// WebSocketHandler handles WebSocket connections
type WebSocketHandler func(conn *websocket.Conn) error

// SSEHandler handles Server-Sent Events streams
type SSEHandler func(stream *SSEStream) error

// SSEStream represents a Server-Sent Events stream
type SSEStream struct {
	writer http.ResponseWriter
	flush  http.Flusher
	done   chan struct{}
}

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

func (hi *HandlerInfo) RelativePath() string {
	return hi.path
}
func (hi *HandlerInfo) FullPath() string {
	return hi.fullPath
}
func (hi *HandlerInfo) HandlerMeta() *HandlerMeta {
	return hi.handlerMeta
}
func (hi *HandlerInfo) HandlerFunc() []HandlerFunc {
	return hi.handlers
}
func (hi *HandlerInfo) FileSystem() http.FileSystem {
	return hi.fs
}
func (hi *HandlerInfo) TargetUrl() string {
	return hi.targetUrl
}
func (hi *HandlerInfo) WebSocketHandler() WebSocketHandler {
	return hi.wsHandler
}
func (hi *HandlerInfo) SSEHandler() SSEHandler {
	return hi.sseHandler
}
func (hi *HandlerInfo) Upgrader() *websocket.Upgrader {
	return hi.wsUpgrader
}
func (hi *HandlerInfo) IsStaticFs() bool {
	return hi.fs != nil
}
func (hi *HandlerInfo) IsReverseProxy() bool {
	return hi.targetUrl != ""
}
func (hi *HandlerInfo) IsWebSocket() bool {
	return hi.wsHandler != nil
}
func (hi *HandlerInfo) IsSSE() bool {
	return hi.sseHandler != nil
}

type RouteInfo []*HandlerInfo

func NewHandlerInfo(path string, handlers ...HandlerFunc) *HandlerInfo {
	return &HandlerInfo{handlerMeta: NewHandlerMeta(), path: path, handlers: handlers}
}

func NewStaticFsHandlerInfo(relativePath string, fs http.FileSystem) *HandlerInfo {
	return &HandlerInfo{handlerMeta: NewHandlerMeta(), path: relativePath, fs: fs}
}

func NewReverseProxyHandlerInfo(relativePath string, targetUrl string) *HandlerInfo {
	return &HandlerInfo{handlerMeta: NewHandlerMeta(), path: relativePath, targetUrl: targetUrl}
}

func NewWebSocketHandlerInfo(relativePath string, handler WebSocketHandler, upgrader *websocket.Upgrader) *HandlerInfo {
	return &HandlerInfo{handlerMeta: NewHandlerMeta(), path: relativePath, wsHandler: handler, wsUpgrader: upgrader}
}

func NewSSEHandlerInfo(relativePath string, handler SSEHandler) *HandlerInfo {
	return &HandlerInfo{handlerMeta: NewHandlerMeta(), path: relativePath, sseHandler: handler}
}

type RouteTree map[string]RouteInfo

func (rt RouteTree) Set(method string, handlerInfo *HandlerInfo) {
	rt[method] = append(rt[method], handlerInfo)
}

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

// Package web: WebSocket stream support.
package web

import (
	"context"
	"fmt"
	"sync"

	"github.com/chuccp/go-web-frame/log"
	"github.com/coder/websocket"
	"go.uber.org/zap"
)

// WebSocketHandler is the function signature for WebSocket stream handlers.
type WebSocketHandler func(stream *WebSocketStream) error

// WSResponse is a handler return value that signals the converter
// to accept the WebSocket upgrade and invoke the handler.
type WSResponse struct {
	Handler WebSocketHandler
}

type AcceptOptions struct {
	OriginPatterns []string
}

// WebSocketStream wraps a coder/websocket connection with web2 Request and context.
type WebSocketStream struct {
	cancel        context.CancelFunc
	request       *Request
	ctx           context.Context
	conn          *websocket.Conn
	AcceptOptions *AcceptOptions
	mu            sync.Mutex
}

func newWebSocketStream(r *Request) *WebSocketStream {
	ctx, cancel := context.WithCancel(r.Ctx())
	return &WebSocketStream{
		request: r,
		ctx:     ctx,
		cancel:  cancel,
		AcceptOptions: &AcceptOptions{
			OriginPatterns: []string{"*"},
		},
	}
}

func (ws *WebSocketStream) initConnection() error {
	if ws.conn != nil {
		return nil
	}
	ws.mu.Lock()
	defer ws.mu.Unlock()
	if ws.conn != nil {
		return nil
	}
	acceptOptions := &websocket.AcceptOptions{
		OriginPatterns: ws.AcceptOptions.OriginPatterns,
	}
	conn, err := websocket.Accept(ws.request.response, ws.request.Request(), acceptOptions)
	if err != nil {
		log.Debug("converter: WebSocket accept error", zap.Error(err))
		if abortErr := ws.request.response.AbortWithError(err); abortErr != nil {
			log.Debug("converter: WebSocket abort error", zap.Error(abortErr))
		}
		return err
	}
	ws.conn = conn
	return nil
}

// Request returns the underlying web2 Request.
func (ws *WebSocketStream) Request() *Request {
	return ws.request
}

// Conn returns the underlying WebSocket connection.
func (ws *WebSocketStream) Conn() *websocket.Conn {
	conn, err := ws.getConn()
	if err != nil {
		log.Error("converter: WebSocket initConnection error", zap.Error(err))
	}
	return conn
}

func (ws *WebSocketStream) getConn() (*websocket.Conn, error) {
	if ws.conn != nil {
		return ws.conn, nil
	}
	err := ws.initConnection()
	if err != nil {
		return nil, err
	}
	return ws.conn, nil
}

// Context returns the stream context, cancelled on Close or request disconnect.
func (ws *WebSocketStream) Context() context.Context {
	return ws.ctx
}

// Done returns a channel that is closed when the stream is closed or the request disconnects.
func (ws *WebSocketStream) Done() <-chan struct{} {
	return ws.ctx.Done()
}

// Read reads any message type from the WebSocket connection.
func (ws *WebSocketStream) Read(ctx context.Context) (websocket.MessageType, []byte, error) {
	conn, err := ws.getConn()
	if err != nil {
		return 0, nil, err
	}
	return conn.Read(ctx)
}

// Write writes a message to the WebSocket connection.
func (ws *WebSocketStream) Write(ctx context.Context, typ websocket.MessageType, data []byte) error {
	conn, err := ws.getConn()
	if err != nil {
		return err
	}
	return conn.Write(ctx, typ, data)
}

// WriteText writes a text message.
func (ws *WebSocketStream) WriteText(ctx context.Context, data []byte) error {
	return ws.Write(ctx, websocket.MessageText, data)
}

// WriteString writes a text message from a string.
func (ws *WebSocketStream) WriteString(ctx context.Context, s string) error {
	return ws.Write(ctx, websocket.MessageText, []byte(s))
}

// WriteBinary writes a binary message.
func (ws *WebSocketStream) WriteBinary(ctx context.Context, data []byte) error {
	return ws.Write(ctx, websocket.MessageBinary, data)
}

// ReadText reads a text message, returning an error if the message is not text.
func (ws *WebSocketStream) ReadText(ctx context.Context) ([]byte, error) {
	typ, data, err := ws.Read(ctx)
	if err != nil {
		return nil, err
	}
	if typ != websocket.MessageText {
		return nil, fmt.Errorf("expected text message, got %v", typ)
	}
	return data, nil
}

// Ping sends a ping frame to the client.
func (ws *WebSocketStream) Ping(ctx context.Context) error {
	conn, err := ws.getConn()
	if err != nil {
		return err
	}
	return conn.Ping(ctx)
}

// Close closes the WebSocket connection and cancels the stream context.
func (ws *WebSocketStream) Close() {
	defer ws.cancel()
	conn, err := ws.getConn()
	if err != nil {
		return
	}
	if closeErr := conn.Close(websocket.StatusNormalClosure, ""); closeErr != nil {
		log.Debug("websocket close error", zap.Error(closeErr))
	}
}

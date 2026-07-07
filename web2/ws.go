package web2

import (
	"context"
	"fmt"

	"github.com/coder/websocket"
)

// WebSocketHandler is the function signature for WebSocket stream handlers.
type WebSocketHandler func(stream *WebSocketStream) error

// WSResponse is a handler return value that signals the converter
// to accept the WebSocket upgrade and invoke the handler.
type WSResponse struct {
	Handler WebSocketHandler
}

// WebSocketStream wraps a coder/websocket connection with web2 Request and context.
type WebSocketStream struct {
	cancel  context.CancelFunc
	request *Request
	ctx     context.Context
	conn    *websocket.Conn
}

func newWebSocketStream(r *Request, conn *websocket.Conn) *WebSocketStream {
	ctx, cancel := context.WithCancel(r.Ctx())
	return &WebSocketStream{
		request: r,
		ctx:     ctx,
		cancel:  cancel,
		conn:    conn,
	}
}

// Request returns the underlying web2 Request.
func (ws *WebSocketStream) Request() *Request {
	return ws.request
}

// Conn returns the underlying WebSocket connection.
func (ws *WebSocketStream) Conn() *websocket.Conn {
	return ws.conn
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
	return ws.conn.Read(ctx)
}

// Write writes a message to the WebSocket connection.
func (ws *WebSocketStream) Write(ctx context.Context, typ websocket.MessageType, data []byte) error {
	return ws.conn.Write(ctx, typ, data)
}

// WriteText writes a text message.
func (ws *WebSocketStream) WriteText(ctx context.Context, data []byte) error {
	return ws.conn.Write(ctx, websocket.MessageText, data)
}

// WriteString writes a text message from a string.
func (ws *WebSocketStream) WriteString(ctx context.Context, s string) error {
	return ws.conn.Write(ctx, websocket.MessageText, []byte(s))
}

// WriteBinary writes a binary message.
func (ws *WebSocketStream) WriteBinary(ctx context.Context, data []byte) error {
	return ws.conn.Write(ctx, websocket.MessageBinary, data)
}

// ReadText reads a text message, returning an error if the message is not text.
func (ws *WebSocketStream) ReadText(ctx context.Context) ([]byte, error) {
	typ, data, err := ws.conn.Read(ctx)
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
	return ws.conn.Ping(ctx)
}

// Close closes the WebSocket connection and cancels the stream context.
func (ws *WebSocketStream) Close() {
	ws.cancel()
	err := ws.conn.Close(websocket.StatusNormalClosure, "")
	if err != nil {
		return
	}
}

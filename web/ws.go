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
	// Subprotocols lists the WebSocket subprotocols that Accept will negotiate
	// with the client. The empty subprotocol will always be negotiated as per
	// RFC 6455. If you would like to reject it, close the connection when
	// c.Subprotocol() == "".
	Subprotocols []string

	// InsecureSkipVerify is used to disable Accept's origin verification behaviour.
	// You probably want to use OriginPatterns instead.
	InsecureSkipVerify bool

	// OriginPatterns lists the host patterns for authorized origins.
	// The request host is always authorized. Use this to enable cross origin WebSockets.
	OriginPatterns []string

	// CompressionMode controls the compression mode.
	// Defaults to CompressionDisabled.
	CompressionMode websocket.CompressionMode

	// CompressionThreshold controls the minimum size of a message before
	// compression is applied. Defaults to 512 bytes for CompressionNoContextTakeover
	// and 128 bytes for CompressionContextTakeover.
	CompressionThreshold int

	// OnPingReceived is an optional callback invoked synchronously when a ping
	// frame is received. If the callback returns false, the subsequent pong
	// frame will not be sent.
	OnPingReceived func(ctx context.Context, payload []byte) bool

	// OnPongReceived is an optional callback invoked synchronously when a pong
	// frame is received.
	OnPongReceived func(ctx context.Context, payload []byte)
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
		Subprotocols:         ws.AcceptOptions.Subprotocols,
		InsecureSkipVerify:   ws.AcceptOptions.InsecureSkipVerify,
		OriginPatterns:       ws.AcceptOptions.OriginPatterns,
		CompressionMode:      ws.AcceptOptions.CompressionMode,
		CompressionThreshold: ws.AcceptOptions.CompressionThreshold,
		OnPingReceived:       ws.AcceptOptions.OnPingReceived,
		OnPongReceived:       ws.AcceptOptions.OnPongReceived,
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

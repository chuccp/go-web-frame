// Package web: Server-Sent Events stream support.
package web

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sourcegraph/conc/panics"
)

// SSEHandler is the function signature for SSE stream handlers.
type SSEHandler func(stream *SSEStream) error

// SSEResponse is a handler return value that signals the converter
// to set up an SSE stream and invoke the handler.
type SSEResponse struct {
	Handler SSEHandler
}

// SSEStream represents a Server-Sent Events stream backed by a web Request.
type SSEStream struct {
	mu        sync.Mutex
	cancel    context.CancelFunc
	request   *Request
	ctx       context.Context
	bgWorkers sync.WaitGroup
}

// NewSSEStream creates a new SSE stream from a web Request.
func NewSSEStream(r *Request) *SSEStream {
	ctx, cancel := context.WithCancel(r.Ctx())
	return &SSEStream{
		request: r,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Request returns the underlying web Request.
func (s *SSEStream) Request() *Request {
	return s.request
}

// Context returns the stream's context, cancelled when the stream closes.
func (s *SSEStream) Context() context.Context {
	return s.ctx
}

// send writes an already-formatted SSE frame to the response writer.
// Caller must hold s.mu.
func (s *SSEStream) send(frame string) error {
	w := s.request.response
	if _, err := fmt.Fprint(w, frame); err != nil {
		return err
	}
	w.Flush()
	return nil
}

// Send sends an SSE event with the given event name and data.
func (s *SSEStream) Send(event string, data string) error {
	select {
	case <-s.ctx.Done():
		return fmt.Errorf("stream closed")
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	return s.send(fmt.Sprintf("event: %s\ndata: %s\n\n", event, data))
}

// SendMessage sends a default message event without a named event type.
func (s *SSEStream) SendMessage(data string) error {
	select {
	case <-s.ctx.Done():
		return fmt.Errorf("stream closed")
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	return s.send(fmt.Sprintf("data: %s\n\n", data))
}

// SendWithID sends an SSE event with an ID, event name, and data.
func (s *SSEStream) SendWithID(id string, event string, data string) error {
	select {
	case <-s.ctx.Done():
		return fmt.Errorf("stream closed")
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	return s.send(fmt.Sprintf("id: %s\nevent: %s\ndata: %s\n\n", id, event, data))
}

// SendRetry sends a reconnection interval hint (in milliseconds) to the client.
func (s *SSEStream) SendRetry(retryMs int) error {
	select {
	case <-s.ctx.Done():
		return fmt.Errorf("stream closed")
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	return s.send(fmt.Sprintf("retry: %d\n\n", retryMs))
}

// SetHeaders writes the standard SSE headers on the response.
func (s *SSEStream) SetHeaders() {
	s.mu.Lock()
	defer s.mu.Unlock()

	w := s.request.response
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
}

// SetHeader sets a custom header on the SSE response.
func (s *SSEStream) SetHeader(key string, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	w := s.request.response
	w.Header().Set(key, value)
}

// Close closes the SSE stream and waits for all background goroutines to exit.
func (s *SSEStream) Close() {
	s.cancel()
	s.bgWorkers.Wait()
}

// Done returns a channel that is closed when the stream is closed.
func (s *SSEStream) Done() <-chan struct{} {
	return s.ctx.Done()
}

// Heartbeat sends a comment line to keep the connection alive.
func (s *SSEStream) Heartbeat() error {
	select {
	case <-s.ctx.Done():
		return fmt.Errorf("stream closed")
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	return s.send(": heartbeat\n\n")
}

// StartHeartbeat starts a periodic heartbeat goroutine that sends SSE comment
// lines at the given interval. The goroutine exits automatically when the stream
// is closed or the request disconnects. Close() blocks until the heartbeat
// goroutine has fully exited.
func (s *SSEStream) StartHeartbeat(interval time.Duration) {
	s.StartHeartbeatWithContext(s.ctx, interval)
}

// StartHeartbeatWithContext starts a periodic heartbeat goroutine with an
// external context. The goroutine exits when ctx is cancelled, the stream is
// closed, or the request disconnects. Close() blocks until the goroutine exits.
func (s *SSEStream) StartHeartbeatWithContext(ctx context.Context, interval time.Duration) {
	s.bgWorkers.Add(1)
	go panics.Try(func() {
		defer s.bgWorkers.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-s.ctx.Done():
				return
			case <-ticker.C:
				if err := s.Heartbeat(); err != nil {
					return
				}
			}
		}
	})
}

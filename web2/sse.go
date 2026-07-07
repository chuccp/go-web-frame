package web2

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// SSEHandler is the function signature for SSE stream handlers.
type SSEHandler func(stream *SSEStream) error

// SSEStream represents a Server-Sent Events stream backed by a web2 Request.
type SSEStream struct {
	cancel  context.CancelFunc
	request *Request
	ctx     context.Context
}

// NewSSEStream creates a new SSE stream from a web2 Request.
func NewSSEStream(r *Request) *SSEStream {
	ctx, cancel := context.WithCancel(r.Ctx())
	return &SSEStream{
		request: r,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Send sends an SSE event with the given event name and data.
func (s *SSEStream) Send(event string, data string) error {
	select {
	case <-s.ctx.Done():
		return fmt.Errorf("stream closed")
	default:
	}

	w := s.request.response
	fmt.Fprintf(w, "event: %s\n", event)
	fmt.Fprintf(w, "data: %s\n\n", data)
	w.Flush()
	return nil
}

// SendMessage sends a default message event without a named event type.
func (s *SSEStream) SendMessage(data string) error {
	select {
	case <-s.ctx.Done():
		return fmt.Errorf("stream closed")
	default:
	}

	w := s.request.response
	fmt.Fprintf(w, "data: %s\n\n", data)
	w.Flush()
	return nil
}

// SendWithID sends an SSE event with an ID, event name, and data.
func (s *SSEStream) SendWithID(id string, event string, data string) error {
	select {
	case <-s.ctx.Done():
		return fmt.Errorf("stream closed")
	default:
	}

	w := s.request.response
	fmt.Fprintf(w, "id: %s\n", id)
	fmt.Fprintf(w, "event: %s\n", event)
	fmt.Fprintf(w, "data: %s\n\n", data)
	w.Flush()
	return nil
}

// SendRetry sends a reconnection interval hint (in milliseconds) to the client.
func (s *SSEStream) SendRetry(retryMs int) error {
	select {
	case <-s.ctx.Done():
		return fmt.Errorf("stream closed")
	default:
	}

	w := s.request.response
	fmt.Fprintf(w, "retry: %d\n\n", retryMs)
	w.Flush()
	return nil
}

// SetHeaders writes the standard SSE headers on the response.
func (s *SSEStream) setHeaders() {
	w := s.request.response
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
}
func (s *SSEStream) SetHeader(key string, value string) {
	w := s.request.response
	w.Header().Set(key, value)
}

// Close closes the SSE stream, signaling all goroutines to stop.
func (s *SSEStream) Close() {
	s.cancel()
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

	w := s.request.response
	fmt.Fprintf(w, ": heartbeat\n\n")
	w.Flush()
	return nil
}

// StartHeartbeat starts a periodic heartbeat goroutine.
// Returns a stop function that blocks until the goroutine has exited.
func (s *SSEStream) StartHeartbeat(interval time.Duration) func() {
	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-s.ctx.Done():
				return
			case <-stop:
				return
			case <-ticker.C:
				if err := s.Heartbeat(); err != nil {
					return
				}
			}
		}
	}()
	return func() {
		close(stop)
		wg.Wait()
	}
}

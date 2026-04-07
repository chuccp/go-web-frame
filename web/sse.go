package web

import (
	"fmt"
	"net/http"
	"time"
)

// NewSSEStream creates a new SSE stream from http.ResponseWriter
func NewSSEStream(w http.ResponseWriter) *SSEStream {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil
	}
	return &SSEStream{
		writer: w,
		flush:  flusher,
		done:   make(chan struct{}),
	}
}

// Send sends an event with the given event name and data
func (s *SSEStream) Send(event string, data string) error {
	select {
	case <-s.done:
		return fmt.Errorf("stream closed")
	default:
	}

	fmt.Fprintf(s.writer, "event: %s\n", event)
	fmt.Fprintf(s.writer, "data: %s\n\n", data)
	s.flush.Flush()
	return nil
}

// SendMessage sends a message without event name (default message type)
func (s *SSEStream) SendMessage(data string) error {
	select {
	case <-s.done:
		return fmt.Errorf("stream closed")
	default:
	}

	fmt.Fprintf(s.writer, "data: %s\n\n", data)
	s.flush.Flush()
	return nil
}

// SendWithID sends an event with id, event name and data
func (s *SSEStream) SendWithID(id string, event string, data string) error {
	select {
	case <-s.done:
		return fmt.Errorf("stream closed")
	default:
	}

	fmt.Fprintf(s.writer, "id: %s\n", id)
	fmt.Fprintf(s.writer, "event: %s\n", event)
	fmt.Fprintf(s.writer, "data: %s\n\n", data)
	s.flush.Flush()
	return nil
}

// SendRetry sets the reconnection time in milliseconds
func (s *SSEStream) SendRetry(retryMs int) error {
	select {
	case <-s.done:
		return fmt.Errorf("stream closed")
	default:
	}

	fmt.Fprintf(s.writer, "retry: %d\n\n", retryMs)
	s.flush.Flush()
	return nil
}

// SetHeaders sets the SSE headers on the response writer
func (s *SSEStream) SetHeaders() {
	s.writer.Header().Set("Content-Type", "text/event-stream")
	s.writer.Header().Set("Cache-Control", "no-cache")
	s.writer.Header().Set("Connection", "keep-alive")
	s.writer.Header().Set("Access-Control-Allow-Origin", "*")
}

// Close closes the stream
func (s *SSEStream) Close() {
	close(s.done)
}

// Done returns a channel that's closed when the stream is closed
func (s *SSEStream) Done() <-chan struct{} {
	return s.done
}

// Heartbeat sends a comment as heartbeat to keep the connection alive
func (s *SSEStream) Heartbeat() error {
	select {
	case <-s.done:
		return fmt.Errorf("stream closed")
	default:
	}

	fmt.Fprintf(s.writer, ": heartbeat\n\n")
	s.flush.Flush()
	return nil
}

// StartHeartbeat starts a periodic heartbeat goroutine
func (s *SSEStream) StartHeartbeat(interval time.Duration) chan struct{} {
	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-s.done:
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
	return stop
}
package core

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/web"
)

// contextIgnoringRunner never returns from Run, mimicking a runner that does
// not watch the application context (e.g. the background-task example in
// CLAUDE.md that loops on a ticker).
type contextIgnoringRunner struct{}

func (r *contextIgnoringRunner) Init(ctx *Context) error { return nil }
func (r *contextIgnoringRunner) Run() error              { select {} }

// TestServerRun_ContextIgnoringRunner verifies that a runner which ignores the
// context cannot block shutdown: Run returns once the context is cancelled and
// the grace period expires.
func TestServerRun_ContextIgnoringRunner(t *testing.T) {
	oldGrace := runnerShutdownGrace
	runnerShutdownGrace = 50 * time.Millisecond
	defer func() { runnerShutdownGrace = oldGrace }()

	ctx, cancel := context.WithCancel(context.Background())
	coreCtx := NewContext(config.NewConfig(), ctx)

	server := NewServer(coreCtx)
	server.AddIRunner(&contextIgnoringRunner{})
	if _, err := server.servers.CreateServerWithContext(&web.ServerConfig{Port: 0}, coreCtx); err != nil {
		t.Fatalf("create server: %v", err)
	}

	done := make(chan error, 1)
	go func() { done <- server.Run() }()

	time.Sleep(50 * time.Millisecond) // let the listener come up
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Run() = %v, want context.Canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run() did not return: a runner that ignores the context blocked shutdown")
	}
}

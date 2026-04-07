package main

import (
	"context"
	"time"

	wf "github.com/chuccp/go-web-frame"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/web"
)

// BackgroundRunner is an example background task that runs periodically
type BackgroundRunner struct {
	// Embed the core.IRunner interface
	core.IRunner
}

// Init initializes the background runner
func (r *BackgroundRunner) Init(ctx *core.Context) error {
	// No additional initialization needed
	return nil
}

// Run is called by the framework to start the background task
// It runs until context is canceled
func (r *BackgroundRunner) Run(ctx context.Context) error {
	log.Info("Background runner starting...")

	// Run a periodic task every 5 seconds
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("Background runner stopping...")
			return nil
		case <-ticker.C:
			// Do your background work here
			log.Debug("Background runner is working...")
		}
	}
}

// Another background task example that runs once at startup
type StartupTask struct {
	core.IRunner
}

func (t *StartupTask) Init(ctx *core.Context) error {
	return nil
}

func (t *StartupTask) Run(ctx context.Context) error {
	log.Info("Running startup initialization task...")
	// Do one-time initialization here
	// For example: warm up cache, load configuration, etc.
	log.Info("Startup task completed")
	return nil
}

func main() {
	// Create application with auto config
	app := wf.NewWithAutoConfig()

	// Add a simple route
	app.Get("/", func(c *web.Request) (any, error) {
		return map[string]string{
			"status": "running",
			"message": "Background tasks are running",
		}, nil
	})

	// Register background runners
	// The framework will manage their lifecycle automatically
	app.AddRunner(&BackgroundRunner{})
	app.AddRunner(&StartupTask{})

	// Run with context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Auto shutdown after 30 seconds for demonstration
	go func() {
		time.Sleep(30 * time.Second)
		log.Info("Auto shutdown triggered...")
		cancel()
	}()

	if err := app.Run(ctx); err != nil {
		log.PrintPanic(err)
	}

	log.Info("Application exited gracefully")
}

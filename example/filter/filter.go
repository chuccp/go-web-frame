package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	wf "github.com/chuccp/go-web-frame"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/web"
	"go.uber.org/zap"
)

// RequestIDFilter adds a unique request ID to each request for tracing
type RequestIDFilter struct {
	core.IFilter
}

func (f *RequestIDFilter) Init(ctx *core.Context) error {
	// Initialization if needed
	return nil
}

// Handle processes the filter chain
func (f *RequestIDFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
	// Generate request ID (in real app you might use a library)
	requestID := fmt.Sprintf("%d", time.Now().UnixNano())

	// Add request ID to request context for use in handler
	req.GinContext().Set("request_id", requestID)

	// Continue processing the request
	return fc.Next()
}

// LoggingFilter logs the request processing time
type LoggingFilter struct {
	core.IFilter
}

func (f *LoggingFilter) Init(ctx *core.Context) error {
	return nil
}

func (f *LoggingFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
	start := time.Now()

	// Log incoming request
	log.Info("Incoming request",
		zap.String("method", req.Request().Method),
		zap.String("path", req.Request().URL.Path),
		zap.String("remote", req.RemoteAddr()),
	)

	// Call next handler
	result, err := fc.Next()

	// Log processing time
	elapsed := time.Since(start)
	log.Info("Request completed",
		zap.String("method", req.Request().Method),
		zap.String("path", req.Request().URL.Path),
		zap.Duration("elapsed", elapsed),
	)

	return result, err
}

// AuthFilter checks authentication before processing
type AuthFilter struct {
	core.IFilter
}

func (f *AuthFilter) Init(ctx *core.Context) error {
	return nil
}

func (f *AuthFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
	// Get authorization header
	auth := req.Request().Header.Get("Authorization")

	if auth == "" {
		// Return error will abort processing and send error response
		return nil, errors.New("authorization required")
	}

	// Verify token here...
	// If invalid: return nil, errors.New("invalid token")

	// If valid, continue processing
	return fc.Next()
}

func main() {
	app := wf.NewWithAutoConfig()

	// Add global filters that apply to all routes
	app.AddFilter(&LoggingFilter{})
	app.AddFilter(&RequestIDFilter{})

	// Public route doesn't need auth
	app.Get("/api/public", func(c *web.Request) (any, error) {
		return map[string]string{
			"message": "public resource",
		}, nil
	})

	// Create a separate rest group for protected routes that requires auth
	// This way all routes in this group get the AuthFilter
	protectedGroup := app.NewRestGroup(web.DefaultServerConfig())
	protectedGroup.AddFilter(&AuthFilter{})

	// Add a REST controller for protected routes
	protectedGroup.AddRest(&ProtectedController{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := app.Run(ctx); err != nil {
		log.PrintPanic(err)
	}
}

// ProtectedController handles protected routes that require authentication
type ProtectedController struct {
	core.IRest
}

func (c *ProtectedController) Init(ctx *core.Context) error {
	ctx.Get("/api/protected", func(c *web.Request) (any, error) {
		// Get request ID from context set by filter
		requestID, _ := c.GinContext().Get("request_id")

		return map[string]any{
			"message":    "protected resource accessed",
			"request_id": requestID,
		}, nil
	})
	return nil
}

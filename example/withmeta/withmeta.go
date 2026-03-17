package main

import (
	"context"
	"errors"

	wf "github.com/chuccp/go-web-frame"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/web"
)

// AuthMetaKey is a metadata key for authentication requirement
const AuthMetaKey = "require_auth"

// LoginAuthFilter is an authentication filter that checks route metadata
type LoginAuthFilter struct {
	core.IFilter
}

func (f *LoginAuthFilter) Init(ctx *core.Context) error {
	return nil
}

func (f *LoginAuthFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
	// Get handler metadata
	meta := req.HandlerMeta()

	// Check if this route requires authentication
	requireAuth := meta.Has(AuthMetaKey)
	if !requireAuth {
		// No auth required, skip checking and continue
		return fc.Next()
	}

	// Check for skip_auth meta - useful for public routes inside authenticated groups
	if meta.Has("skip_auth") {
		return fc.Next()
	}

	// Get token from request
	token := req.Request().Header.Get("Authorization")
	if token == "" {
		return nil, errors.New("missing authorization token")
	}

	// Verify token here (e.g., JWT verification)
	// For demonstration, we just check that it's not empty
	// In real app, you would verify and set user info to context:
	// user := verifyToken(token)
	// req.GinContext().Set("user", user)

	return fc.Next()
}

// RequireAuth creates a meta option that marks a route as requiring authentication
// Usage: app.Get("/api/profile", handler, web.WithMeta(RequireAuth()))
func RequireAuth() web.MetaOption {
	return web.WithValue(AuthMetaKey, true)
}

// RequirePermission creates a meta option that checks for specific permission
func RequirePermission(permission string) web.MetaOption {
	return web.WithValue("require_permission", permission)
}

// SkipAuth creates a meta option that skips authentication even for routes inside an authenticated group
func SkipAuth() web.MetaOption {
	return web.WithValue("skip_auth", true)
}

// ApiController is an example REST controller with protected and public routes
type ApiController struct {
	core.IRest
}

func (c *ApiController) Init(ctx *core.Context) error {
	// Public routes - no auth needed
	ctx.Get("/api/login", func(c *web.Request) (any, error) {
		// Handle login, return JWT token
		return map[string]string{
			"status": "ok",
			"token": "example-jwt-token",
		}, nil
	}).WithMeta(SkipAuth()) // Skip auth since login doesn't need to be authenticated

	ctx.Get("/api/public/info", func(c *web.Request) (any, error) {
		return map[string]string{
			"message": "this is public information",
		}, nil
	})

	// Protected route - requires auth via .WithMeta(RequireAuth())
	ctx.Get("/api/profile", func(c *web.Request) (any, error) {
		// Get authenticated user from context set by filter
		user, exists := c.GinContext().Get("user")
		if !exists {
			user = "current-user"
		}
		return map[string]any{
			"user": user,
			"id":   "123",
		}, nil
	}).WithMeta(RequireAuth())

	// Protected route with permission check - multiple meta options supported
	ctx.Post("/api/admin/users", func(c *web.Request) (any, error) {
		// Create user logic here
		return map[string]string{
			"status": "created",
		}, nil
	}).WithMeta(RequireAuth(), RequirePermission("admin:create_user"))

	return nil
}

func main() {
	app := wf.NewWithAutoConfig()

	// Add global authentication filter - checks all routes that have require_auth meta
	app.AddFilter(&LoginAuthFilter{})

	// Create a rest group for API
	apiGroup := app.NewRestGroup(web.DefaultServerConfig())

	// Add our controller to the group - all routes are registered in controller.Init()
	apiGroup.AddRest(&ApiController{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := app.Run(ctx); err != nil {
		log.PrintPanic(err)
	}
}

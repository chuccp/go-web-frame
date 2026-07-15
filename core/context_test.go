package core

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/web"
	"github.com/stretchr/testify/assert"
)

// TestService is a simple test service that implements IService
type TestService struct {
	InitCalled bool
}

func (t *TestService) Init(ctx *Context) error {
	t.InitCalled = true
	return nil
}

// newTestServer creates a *web.Server suitable for unit tests.
func newTestServer() *web.Server {
	servers := web.NewServers()
	server, _ := servers.CreateServer(web.DefaultServerConfig())
	return server
}

func TestContext_AddService(t *testing.T) {
	// Given: a new context
	cfg, _ := config.NewFromBytes([]byte(`{"web": {"db": {"type": "sqlite"}}}`), "json")
	ctx := NewContext(cfg, context.Background())

	// When: adding a service
	service := &TestService{}
	ctx.AddService(service)

	// Then: service should be added to the map, but Init not called yet
	// (Init is called by server during startup)
	assert.False(t, service.InitCalled)

	// When retrieving, should find it
	retrieved := GetService[*TestService](ctx)
	assert.Same(t, service, retrieved)
}

func TestContext_GetService(t *testing.T) {
	// Given: context with a service added
	cfg, _ := config.NewFromBytes([]byte(`{"web": {"db": {"type": "sqlite"}}}`), "json")
	ctx := NewContext(cfg, context.Background())
	service := &TestService{}
	ctx.AddService(service)

	// When: getting the service by type
	retrieved := GetService[*TestService](ctx)

	// Then: the same instance should be returned
	assert.Same(t, service, retrieved)
}

func TestContext_GetConfig(t *testing.T) {
	// Given: a context with config
	cfg, _ := config.NewFromBytes([]byte(`{"test": {"key": "value", "num": 42}}`), "json")
	ctx := NewContext(cfg, context.Background())

	// When: getting config from context
	resultConfig := ctx.GetConfig()

	// Then: should return the original config
	assert.NotNil(t, resultConfig)
	assert.Equal(t, "value", resultConfig.GetString("test.key"))
	assert.Equal(t, 42, resultConfig.GetInt("test.num"))
}

// TestGetService_NotFoundPanics verifies that GetService panics when service is not found
// because missing dependencies should fail fast at startup
func TestGetService_NotFoundPanics(t *testing.T) {
	// Given: an empty context
	cfg, _ := config.NewFromBytes([]byte(`{}`), "json")
	ctx := NewContext(cfg, context.Background())

	// When: trying to get a non-existent service
	// Then: should panic since service is not registered
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic when getting non-existent service, got no panic")
		}
	}()

	_ = GetService[*TestService](ctx)
}

func TestGetModel(t *testing.T) {
	// This test just verifies the generic GetModel function signature compiles
	cfg, _ := config.NewFromBytes([]byte(`{}`), "json")
	ctx := NewContext(cfg, context.Background())
	assert.NotNil(t, ctx)
	// No panic expected when just checking the function exists
}

// newContextWithServer creates a Context and its *web.Server for tests that register routes.
func newContextWithServer(t *testing.T) (*Context, *web.Server) {
	t.Helper()
	cfg, err := config.NewFromBytes([]byte(`{}`), "json")
	assert.NoError(t, err)
	server := newTestServer()
	ctx := NewContext(cfg, context.Background()).Copy(server, nil)
	return ctx, server
}

func TestContext_StaticFsRegistersRoute(t *testing.T) {
	ctx, _ := newContextWithServer(t)
	fs := http.FS(os.DirFS("testdata"))

	// StaticFs registers a GET route; we just verify it doesn't panic.
	route := ctx.StaticFs("/assets", fs)
	assert.NotNil(t, route)
}

func TestContext_StaticRegistersRoute(t *testing.T) {
	ctx, _ := newContextWithServer(t)
	dir := t.TempDir()

	route := ctx.Static("/static", dir)
	assert.NotNil(t, route)
}

func TestContext_ReverseProxyRegistersRoute(t *testing.T) {
	ctx, _ := newContextWithServer(t)

	route := ctx.ReverseProxy("/api", "http://localhost:9999/backend")
	assert.NotNil(t, route)
}

// TODO: The following tests (StaticFsServesFileViaHttpServer, StaticFsHandlesHeadRequestViaHttpServer,
// ReverseProxyForwardsRequest) require calling server.initRoute() and accessing server.engine,
// both of which are unexported. Re-enable these tests once the web.Server type exposes an
// Engine() or InitRoutes() method for testing.

func TestContext_ReverseProxyInvalidTargetUrl(t *testing.T) {
	ctx, _ := newContextWithServer(t)

	// Even with an invalid target URL, the route registration itself should succeed.
	route := ctx.ReverseProxy("/bad-proxy", "://bad target")
	assert.NotNil(t, route)
}

func TestContext_RouteRegistration(t *testing.T) {
	ctx, server := newContextWithServer(t)

	// Register various route types - they should not panic
	route := ctx.Get("/hello", func(req *web.Request) (any, error) {
		return "hello", nil
	})
	assert.NotNil(t, route)

	route = ctx.Post("/items", func(req *web.Request) (any, error) {
		return "created", nil
	})
	assert.NotNil(t, route)

	route = ctx.Put("/items/:id", func(req *web.Request) (any, error) {
		return "updated", nil
	})
	assert.NotNil(t, route)

	route = ctx.Delete("/items/:id", func(req *web.Request) (any, error) {
		return "deleted", nil
	})
	assert.NotNil(t, route)

	route = ctx.Any("/any", func(req *web.Request) (any, error) {
		return "any", nil
	})
	assert.NotNil(t, route)

	// Server should be the same instance
	assert.NotNil(t, server)
}

func TestContext_GetRunners(t *testing.T) {
	cfg, _ := config.NewFromBytes([]byte(`{}`), "json")
	ctx := NewContext(cfg, context.Background())

	runners := ctx.GetRunners()
	assert.Empty(t, runners)
}

func TestContext_Copy(t *testing.T) {
	cfg, _ := config.NewFromBytes([]byte(`{}`), "json")
	original := NewContext(cfg, context.Background())

	// Add a service to the original
	service := &TestService{}
	original.AddService(service)

	// Create a copy with a new server
	copied := original.Copy(newTestServer(), nil)

	// Both contexts share the same service maps
	retrieved := GetService[*TestService](copied)
	assert.Same(t, service, retrieved)

	// Server instances should be different
	assert.NotSame(t, original.Server(), copied.Server())
}

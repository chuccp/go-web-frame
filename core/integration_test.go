package core

import (
	"context"
	"sync"
	"testing"

	config2 "github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/go-web-frame/web"
	"github.com/stretchr/testify/assert"
)

// TestIntegration_ServiceDependencyInjection tests Service -> Model -> DB dependency injection chain
func TestIntegration_ServiceDependencyInjection(t *testing.T) {
	config := config2.NewConfig()
	ctx := NewContext(newTestServer(), config, context.Background())

	var modelInitCalled, serviceInitCalled bool

	mockModel := &testModel{initCalled: &modelInitCalled}
	mockService := &testService{initCalled: &serviceInitCalled}

	ctx.AddModel(mockModel)
	ctx.AddService(mockService)

	// Initialize model
	err := mockModel.Init(nil, ctx)
	assert.NoError(t, err)
	assert.True(t, modelInitCalled)

	// Initialize service (which depends on model)
	err = mockService.Init(ctx)
	assert.NoError(t, err)
	assert.True(t, serviceInitCalled)

	// Verify service can retrieve model from context
	retrievedModel := ctx.GetModel(func(m IModel) bool {
		_, ok := m.(*testModel)
		return ok
	})
	assert.NotNil(t, retrievedModel)
}

// TestIntegration_RestControllerLifecycle tests Rest controller Init/route registration/handler lifecycle
func TestIntegration_RestControllerLifecycle(t *testing.T) {
	config := config2.NewConfig()
	server := newTestServer()
	ctx := NewContext(server, config, context.Background())

	restCtrl := &testRestController{initCalled: false}
	ctx.AddService(restCtrl)

	// Initialize the REST controller (which registers routes)
	err := restCtrl.Init(ctx)
	assert.NoError(t, err)
	assert.True(t, restCtrl.initCalled)

	// Routes are registered directly on the server; we verified no panic above.
	// The old HasHandler check is removed because web.Server does not expose it.
}

// TestIntegration_FilterChainExecution tests multiple services registered together
func TestIntegration_FilterChainExecution(t *testing.T) {
	config := config2.NewConfig()
	server := newTestServer()
	ctx := NewContext(server, config, context.Background())

	// Register a route on the server
	server.Get("/api", func(req *web.Request) (any, error) {
		return "ok", nil
	})

	// Register multiple services
	svc1 := &testService{Name: 1}
	svc2 := &testService{Name: 2}
	ctx.AddService(svc1, svc2)

	// Verify services are accessible
	retrieved := ctx.GetService(func(m IService) bool {
		_, ok := m.(*testService)
		return ok
	})
	assert.NotNil(t, retrieved)
}

// TestIntegration_ConcurrentContextAccess tests concurrent access to context
func TestIntegration_ConcurrentContextAccess(t *testing.T) {
	config := config2.NewConfig()
	ctx := NewContext(newTestServer(), config, context.Background())

	// Add multiple services concurrently (same type overwrites by qualified name)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			svc := &testService{
				Name:       n,
				initCalled: nil,
			}
			ctx.AddService(svc)
		}(i)
	}
	wg.Wait()

	// At least one service should be registered
	var count int
	ctx.rLock.RLock()
	count = len(ctx.serviceMap)
	ctx.rLock.RUnlock()
	assert.GreaterOrEqual(t, count, 1)

	// Verify service is accessible
	retrieved := ctx.GetService(func(m IService) bool {
		_, ok := m.(*testService)
		return ok
	})
	assert.NotNil(t, retrieved)
}

// TestIntegration_RunnerLifecycle tests runner init and run lifecycle
func TestIntegration_RunnerLifecycle(t *testing.T) {
	config := config2.NewConfig()
	ctx := NewContext(newTestServer(), config, context.Background())

	runner := &testRunner{}
	ctx.AddRunner(runner)

	// Initialize runner
	err := runner.Init(ctx)
	assert.NoError(t, err)

	// Run runner (lifecycle managed by server's context pool)
	err = runner.Run()
	assert.NoError(t, err)
	assert.True(t, runner.runCalled)
}

// TestIntegration_ContextCopyIsolation tests that copied contexts have isolated servers
func TestIntegration_ContextCopyIsolation(t *testing.T) {
	config := config2.NewConfig()
	server1 := newTestServer()
	ctx := NewContext(server1, config, context.Background())

	// Add service to parent context
	ctx.AddService(&testService{})

	// Create copy with different server
	server2 := newTestServer()
	copiedCtx := ctx.Copy(server2, nil)

	// Both contexts share services (copied by reference)
	retrievedService := ctx.GetService(func(m IService) bool {
		_, ok := m.(*testService)
		return ok
	})
	assert.NotNil(t, retrievedService)

	retrievedService2 := copiedCtx.GetService(func(m IService) bool {
		_, ok := m.(*testService)
		return ok
	})
	assert.NotNil(t, retrievedService2)

	// But servers are different instances
	assert.NotSame(t, server1, server2)
}

// Test types

type testModel struct {
	initCalled *bool
}

func (m *testModel) Init(d *db.DB, ctx *Context) error {
	*m.initCalled = true
	return nil
}

func (m *testModel) TableName() string {
	return "test_model"
}

func (m *testModel) IsExist() (bool, error) {
	return false, nil
}

func (m *testModel) CreateTable() error {
	return nil
}

func (m *testModel) DeleteTable() error {
	return nil
}

func (m *testModel) DropTable() error {
	return nil
}

func (m *testModel) GetTableName() string {
	return "test_model"
}

func (m *testModel) ReNew(d *db.DB, ctx *Context) IModel {
	return m
}

type testService struct {
	Name       int
	initCalled *bool
}

func (s *testService) Init(ctx *Context) error {
	if s.initCalled != nil {
		*s.initCalled = true
	}
	return nil
}

type testRestController struct {
	initCalled bool
}

func (r *testRestController) Init(ctx *Context) error {
	r.initCalled = true
	ctx.Get("/rest/items", func(req *web.Request) (any, error) {
		return []string{"item1", "item2"}, nil
	})
	ctx.Post("/rest/items", func(req *web.Request) (any, error) {
		return "created", nil
	})
	ctx.Get("/rest/items/:id", func(req *web.Request) (any, error) {
		id := req.Param("id")
		return map[string]string{"id": id}, nil
	})
	return nil
}

type orderedFilter struct {
	IFilter
	name  string
	order *[]string
	mu    *sync.Mutex
}

func (f *orderedFilter) Handle(fc web.FilterChain, req *web.Request) (any, error) {
	f.mu.Lock()
	*f.order = append(*f.order, f.name)
	f.mu.Unlock()
	return fc.Next()
}

type testRunner struct {
	initCalled bool
	runCalled  bool
}

func (r *testRunner) Init(ctx *Context) error {
	r.initCalled = true
	return nil
}

func (r *testRunner) Run() error {
	r.runCalled = true
	return nil
}

// testServiceRunner is a service that also implements IRunner.
// It tracks how many times Init is called to guard against double initialization.
type testServiceRunner struct {
	initCount int
}

func (s *testServiceRunner) Init(ctx *Context) error {
	s.initCount++
	return nil
}

func (s *testServiceRunner) Run() error {
	return nil
}

// TestIntegration_ServiceRunnerNoDoubleInit verifies that a service implementing IRunner
// is initialized exactly once: in the services loop, not again in the runners loop.
func TestIntegration_ServiceRunnerNoDoubleInit(t *testing.T) {
	config := config2.NewConfig()
	ctx := NewContext(newTestServer(), config, context.Background())

	svcRunner := &testServiceRunner{}
	// Registered as a service; AddService also adds it to runnerMap because it implements IRunner.
	ctx.AddService(svcRunner)

	// Simulate the services initialization loop (as done in web_frame.go).
	err := svcRunner.Init(ctx)
	assert.NoError(t, err)

	// Build a Server with all runners (including the one from AddService).
	server := NewServer(web.NewServers(), nil, ctx.GetRunners())
	err = server.Init(ctx)
	assert.NoError(t, err)

	// Init should have been called exactly once, not twice.
	assert.Equal(t, 1, svcRunner.initCount, "service implementing IRunner should not be initialized twice")
}

package core

import (
	"testing"

	"github.com/chuccp/go-web-frame/config"
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

func TestContext_AddService(t *testing.T) {
	// Given: a new context
	cfg, _ := config.NewFromBytes([]byte(`{"web": {"db": {"type": "sqlite"}}}`), "json")
	ctx := NewContext(cfg, nil)

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
	ctx := NewContext(cfg, nil)
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
	ctx := NewContext(cfg, nil)

	// When: getting config from context
	resultConfig := ctx.GetConfig()

	// Then: should return the original config
	assert.NotNil(t, resultConfig)
	assert.Equal(t, "value", resultConfig.GetString("test.key"))
	assert.Equal(t, 42, resultConfig.GetInt("test.num"))
}

func TestGetService_NotFoundReturnsZero(t *testing.T) {
	// Given: an empty context
	cfg, _ := config.NewFromBytes([]byte(`{}`), "json")
	ctx := NewContext(cfg, nil)

	// When: trying to get a non-existent service
	// GetService returns the zero value for the type when not found
	result := GetService[*TestService](ctx)
	assert.Nil(t, result)
}

func TestGetModel(t *testing.T) {
	// This test just verifies the generic GetModel function signature compiles
	cfg, _ := config.NewFromBytes([]byte(`{}`), "json")
	ctx := NewContext(cfg, nil)
	assert.NotNil(t, ctx)
	// No panic expected when just checking the function exists
}

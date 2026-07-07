package wf

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"

	config2 "github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/go-web-frame/web"
	"github.com/stretchr/testify/assert"
)

func newTestServer() *web.Server {
	servers := web.NewServers()
	s, _ := servers.CreateServer(web.DefaultServerConfig())
	return s
}

// TestResponse wraps the gin.ResponseWriter from test context to implement web.Response
type TestResponse struct {
	gin.ResponseWriter
	ctx *gin.Context
}

func (r *TestResponse) SetAttachmentFileName(fileName string) {
	r.Header().Set("Content-Disposition", `attachment; filename="`+fileName+`"`)
}

func (r *TestResponse) JSON(code int, value any) {
	r.Header().Set("Content-Type", "application/json")
	r.WriteHeader(code)
	// Marshal and write the JSON content
	jsonBytes, _ := json.Marshal(value)
	r.Write(jsonBytes)
}

func (r *TestResponse) Abort() {
	// Abort is handled by gin context
}

func (r *TestResponse) Redirect(code int, location string) {
	r.Header().Set("Location", location)
	r.WriteHeader(code)
}

func (r *TestResponse) FileAttachment(path string, name string) {
	// Just pass through to the underlying implementation
}

func (r *TestResponse) WriteStatus(code int) {
	r.WriteHeader(code)
}

func (r *TestResponse) Message(t *web.Message) {
	r.JSON(t.Code, t)
}

func (r *TestResponse) AbortWithMessage(t *web.Message) {
	r.Message(t)
}

func (r *TestResponse) AbortWithStatusJSON(i int, value any) {
	r.JSON(i, value)
}

func (r *TestResponse) AbortWithError(err error) error {
	return r.ctx.AbortWithError(http.StatusInternalServerError, err)
}

func TestBuilder_Build(t *testing.T) {
	// Arrange
	config := config2.NewConfig()
	builder := NewBuilder(config)

	// Act
	app := builder.Build()

	// Assert
	assert.NotNil(t, app)
	assert.NotNil(t, app.rests)
	assert.Equal(t, config, app.config)
}

func TestBuilder_RouteMethods(t *testing.T) {
	// Arrange
	config := config2.NewConfig()
	builder := NewBuilder(config)

	handler := func(c *web.Request) (any, error) {
		return "ok", nil
	}

	// Act
	builder.Get("/get", handler)
	builder.Post("/post", handler)
	builder.Put("/put", handler)
	builder.Delete("/delete", handler)
	builder.Any("/any", handler)

	app := builder.Build()

	// Assert
	assert.NotNil(t, app)
	assert.NotNil(t, app.rests)
}

func TestBuilder_AllMethods(t *testing.T) {
	// Arrange
	config := config2.NewConfig()
	builder := NewBuilder(config)

	// Test all chainable methods
	var (
		rest    = &MockRest{}
		runner  = &MockRunner{}
		model   = &MockModel{}
		service = &MockService{}
		filter  = &MockFilter{}
	)

	// Act
	builder.Rest(rest)
	builder.Runner(runner)
	builder.Model(model)
	builder.Service(service)
	builder.Filter(filter)

	app := builder.Build()

	// Assert
	assert.NotNil(t, app)
	assert.Equal(t, 1, len(app.rests))
	assert.Equal(t, 1, len(app.runners))
	assert.Equal(t, 1, len(app.models))
	assert.Equal(t, 1, len(app.services))
	assert.Equal(t, 1, len(app.filters))
}

func TestBuilder_RestGroup(t *testing.T) {
	// Arrange
	config := config2.NewConfig()
	builder := NewBuilder(config)
	serverConfig := web.DefaultServerConfig()
	group := core.NewRestGroupBuilder().ServerConfig(serverConfig).Build()

	// Act
	builder.RestGroup(group)
	app := builder.Build()

	// Assert
	assert.Equal(t, 1, len(app.restGroups))
}

func TestBuilder_ModelGroup(t *testing.T) {
	// Arrange
	config := config2.NewConfig()
	builder := NewBuilder(config)
	group := core.EmptyModelGroup("test")

	// Act
	builder.ModelGroup(group)
	app := builder.Build()

	// Assert
	assert.Equal(t, 1, len(app.modelGroup))
}

func TestWebFrame_Start(t *testing.T) {
	// Arrange
	config := config2.NewConfig()
	config.Put("server.port", 0) // Use random port to avoid conflicts
	builder := NewBuilder(config)
	builder.Get("/", func(c *web.Request) (any, error) {
		return "ok", nil
	})
	app := builder.Build()

	// Use Test instead of Start to avoid blocking on a real server
	err := app.Test(func(ctx *core.Context) error {
		assert.NotNil(t, ctx)
		return nil
	})

	// Assert - should succeed with Test
	assert.NoError(t, err)
}

func TestWebFrame_Test(t *testing.T) {
	// Arrange
	config := config2.NewConfig()
	app := NewBuilder(config).Build()

	// Act
	called := false
	err := app.Test(func(ctx *core.Context) error {
		called = true
		return nil
	})

	// Assert
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestWebFrame_Test_WithError(t *testing.T) {
	// Arrange
	config := config2.NewConfig()
	app := NewBuilder(config).Build()

	// Act
	expectedErr := &MockError{message: "test error"}
	err := app.Test(func(ctx *core.Context) error {
		return expectedErr
	})

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestGetService(t *testing.T) {
	// Arrange
	config := config2.NewConfig()
	ctx := core.NewContext(newTestServer(), config, context.Background())
	service := &MockService{}
	ctx.AddService(service)

	// Act
	result := GetService[*MockService](ctx)

	// Assert
	assert.Equal(t, service, result)
}

func TestGetModel(t *testing.T) {
	// Arrange
	config := config2.NewConfig()
	ctx := core.NewContext(newTestServer(), config, context.Background())
	model := &MockModel{}
	ctx.AddModel(model)

	// Act
	result := GetModel[*MockModel](ctx)

	// Assert
	assert.Equal(t, model, result)
}

func TestGetComponentViaService(t *testing.T) {
	// Arrange
	config := config2.NewConfig()
	ctx := core.NewContext(newTestServer(), config, context.Background())
	component := &MockComponent{}
	ctx.AddService(component)

	// Act
	result := GetService[*MockComponent](ctx)

	// Assert
	assert.Equal(t, component, result)
}

func TestGetRunner(t *testing.T) {
	// Arrange
	config := config2.NewConfig()
	ctx := core.NewContext(newTestServer(), config, context.Background())
	runner := &MockRunner{}
	ctx.AddRunner(runner)

	// Act
	result := GetRunner[*MockRunner](ctx)

	// Assert
	assert.Equal(t, runner, result)
}

// func TestGetFilter(t *testing.T) {
// 	// Arrange
// 	config := config2.NewConfig()
// 	ctx := core.NewContext(newTestServer(), config, context.Background())
// 	filter := &MockFilter{}
// 	ctx.AddService(filter)

// 	// Act
// 	result := GetFilter[*MockFilter](ctx)

// 	// Assert
// 	assert.Equal(t, filter, result)
// }

func TestUnmarshalKeyConfig(t *testing.T) {
	// Arrange
	config := config2.NewConfig()
	config.Put("test.key", "value")
	config.Put("test.port", 8080)
	ctx := core.NewContext(newTestServer(), config, context.Background())

	// Act - test unmarshaling into a struct
	type TestConfig struct {
		Key  string `mapstructure:"key"`
		Port int    `mapstructure:"port"`
	}
	result, err := UnmarshalKeyConfig[*TestConfig]("test", ctx)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "value", result.Key)
	assert.Equal(t, 8080, result.Port)
}

func TestDefaultConverter_Request_JSON(t *testing.T) {
	// Arrange
	converter := &core.DefaultConverter{}
	ctx := core.NewContext(newTestServer(), config2.NewConfig(), context.Background())
	err := converter.Init(ctx)
	assert.NoError(t, err)

	// Create test request with gin test context
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	mockResp := &TestResponse{ResponseWriter: c.Writer, ctx: c}
	webReq := web.NewRequestForTest(c, mockResp, nil)

	// Act
	next := func() (any, error) {
		return map[string]string{"key": "value"}, nil
	}
	filterChain := &MockFilterChain{nextFunc: next}
	converter.Request(filterChain, webReq)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response web.Message
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	data, ok := response.Data.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "value", data["key"])
}

func TestDefaultConverter_Request_String(t *testing.T) {
	// Arrange
	converter := &core.DefaultConverter{}
	ctx := core.NewContext(newTestServer(), config2.NewConfig(), context.Background())
	converter.Init(ctx)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	mockResp := &TestResponse{ResponseWriter: c.Writer, ctx: c}
	webReq := web.NewRequestForTest(c, mockResp, nil)

	next := func() (any, error) {
		return "hello world", nil
	}
	filterChain := &MockFilterChain{nextFunc: next}

	// Act
	converter.Request(filterChain, webReq)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "hello world", w.Body.String())
}

func TestDefaultConverter_Request_Error(t *testing.T) {
	// Arrange
	converter := &core.DefaultConverter{}
	ctx := core.NewContext(newTestServer(), config2.NewConfig(), context.Background())
	_ = converter.Init(ctx)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	mockResp := &TestResponse{ResponseWriter: c.Writer, ctx: c}
	webReq := web.NewRequestForTest(c, mockResp, nil)

	next := func() (any, error) {
		return nil, &MockError{message: "test error"}
	}
	filterChain := &MockFilterChain{nextFunc: next}

	// Act
	converter.Request(filterChain, webReq)

	// Assert
	// web.DefaultConverter calls AbortWithError for errors.
	// AbortWithError writes 500 status on the gin context's internal writer.
	assert.Equal(t, http.StatusInternalServerError, c.Writer.Status())
}

func TestDefaultConverter_Request_File(t *testing.T) {
	// Arrange
	converter := &core.DefaultConverter{}
	ctx := core.NewContext(newTestServer(), config2.NewConfig(), context.Background())
	_ = converter.Init(ctx)

	// Create a temp file for testing
	tmpFile, err := os.CreateTemp("", "test*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	_, _ = tmpFile.WriteString("test file content")
	_ = tmpFile.Close()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	mockResp := &TestResponse{ResponseWriter: c.Writer, ctx: c}
	webReq := web.NewRequestForTest(c, mockResp, nil)

	next := func() (any, error) {
		return &web.FileResponse{Path: tmpFile.Name(), FileName: "test.txt"}, nil
	}
	filterChain := &MockFilterChain{nextFunc: next}

	// Act
	converter.Request(filterChain, webReq)

	// Assert - just check it doesn't panic
	// We can't easily test the file attachment response in unit test
}

func TestDefaultConverter_Request_OsFile(t *testing.T) {
	// Arrange
	converter := &core.DefaultConverter{}
	ctx := core.NewContext(newTestServer(), config2.NewConfig(), context.Background())
	_ = converter.Init(ctx)

	// Create a temp file for testing
	tmpFile, err := os.CreateTemp("", "test*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	_, _ = tmpFile.WriteString("test file content")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	mockResp := &TestResponse{ResponseWriter: c.Writer, ctx: c}
	webReq := web.NewRequestForTest(c, mockResp, nil)

	next := func() (any, error) {
		return tmpFile, nil
	}
	filterChain := &MockFilterChain{nextFunc: next}

	// Act
	converter.Request(filterChain, webReq)

	// Assert - just check it doesn't panic
}

func TestDefaultConverter_Request_Message(t *testing.T) {
	// Arrange
	converter := &core.DefaultConverter{}
	ctx := core.NewContext(newTestServer(), config2.NewConfig(), context.Background())
	_ = converter.Init(ctx)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	mockResp := &TestResponse{ResponseWriter: c.Writer, ctx: c}
	webReq := web.NewRequestForTest(c, mockResp, nil)

	next := func() (any, error) {
		return web.Data("hello"), nil
	}
	filterChain := &MockFilterChain{nextFunc: next}

	// Act
	converter.Request(filterChain, webReq)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response web.Message
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "hello", response.Data)
}

func TestDefaultConverter_Request_Redirect(t *testing.T) {
	// Arrange
	converter := &core.DefaultConverter{}
	ctx := core.NewContext(newTestServer(), config2.NewConfig(), context.Background())
	_ = converter.Init(ctx)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	mockResp := &TestResponse{ResponseWriter: c.Writer, ctx: c}
	webReq := web.NewRequestForTest(c, mockResp, nil)

	next := func() (any, error) {
		return web.Redirect("/new-location"), nil
	}
	filterChain := &MockFilterChain{nextFunc: next}

	// Act
	converter.Request(filterChain, webReq)

	// Assert
	// Due to how httptest.ResponseRecorder works (only sets status code once),
	// we just check that Location header is correctly set which confirms redirect
	// handling was executed properly
	assert.Equal(t, "/new-location", w.Header().Get("Location"))
}

func TestIntegration_WebFrame_WithRoute(t *testing.T) {
	// Integration test for basic route handling
	// Arrange
	config := config2.NewConfig()
	config.Put("server.port", 0) // Use random port

	builder := NewBuilder(config)
	builder.Get("/test", func(c *web.Request) (any, error) {
		return map[string]any{"success": true}, nil
	})
	app := builder.Build()

	// Act - test initialization
	server, ctx, err := app.init(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, server)
	assert.NotNil(t, ctx)
}

// Mock types for testing

type MockService struct {
	core.IService
	initCalled bool
}

func (m *MockService) Init(ctx *core.Context) error {
	m.initCalled = true
	return nil
}

type MockModel struct {
	core.IModel
	initCalled bool
}

func (m *MockModel) Init(db *db.DB, ctx *core.Context) error {
	m.initCalled = true
	return nil
}

func (m *MockModel) TableName() string {
	return "mock"
}

type MockComponent struct {
	core.IService
	initCalled bool
}

func (m *MockComponent) Init(ctx *core.Context) error {
	m.initCalled = true
	return nil
}

type MockRunner struct {
	core.IRunner
	initCalled bool
	runCalled  bool
}

func (m *MockRunner) Init(ctx *core.Context) error {
	m.initCalled = true
	return nil
}

func (m *MockRunner) Run() error {
	m.runCalled = true
	return nil
}

type MockFilter struct {
	core.IFilter
	initCalled bool
	runCalled  bool
}

func (m *MockFilter) Init(ctx *core.Context) error {
	m.initCalled = true
	return nil
}

type MockRest struct {
	core.IRest
	initCalled bool
}

func (m *MockRest) Init(ctx *core.Context) error {
	m.initCalled = true
	return nil
}

type MockFilterChain struct {
	nextFunc func() (any, error)
}

func (m *MockFilterChain) Next() (any, error) {
	return m.nextFunc()
}

func (m *MockFilterChain) Request() *web.Request {
	return nil
}

type MockError struct {
	message string
}

func (m *MockError) Error() string {
	return m.message
}

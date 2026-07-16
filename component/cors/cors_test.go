package cors

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	config2 "github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func newTestCORSServer(t *testing.T, filter *Filter) *httptest.Server {
	t.Helper()
	servers := web.NewServers()
	server, _ := servers.CreateServer(web.DefaultServerConfig())
	server.AddFilter(filter)
	server.Get("/hello", func(r *web.Request) (any, error) {
		return "hello", nil
	})
	ts := httptest.NewServer(server.GetHandler())
	t.Cleanup(ts.Close)
	return ts
}

func newTestContext() *core.Context {
	cfg := config2.NewConfig()
	return core.NewContext(cfg, context.Background())
}

func TestCORS_Init(t *testing.T) {
	f := &Filter{}
	err := f.Init(newTestContext())
	assert.NoError(t, err)
	assert.NotNil(t, f.handlerFunc)
}

func TestCORS_NewCorsFilter(t *testing.T) {
	f := NewCorsFilter()
	assert.NotNil(t, f)
}

func TestCORS_PreflightRequest(t *testing.T) {
	f := &Filter{}
	err := f.Init(newTestContext())
	assert.NoError(t, err)

	// Simulate an OPTIONS preflight request through the filter's Handle method
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodOptions, "/test", nil)
	c.Request.Header.Set("Origin", "http://example.com")
	c.Request.Header.Set("Access-Control-Request-Method", "GET")

	mockResp := &mockResponse{ResponseWriter: c.Writer, ctx: c}
	req := web.NewRequestForTest(c, mockResp, web.NewHandlerMeta())

	called := false
	fc := &mockFilterChain{nextFunc: func() (any, error) {
		called = true
		return nil, nil
	}}

	result, err := f.Handle(fc, req)
	assert.NoError(t, err)
	assert.Nil(t, result)
	assert.False(t, called, "filter chain should not be called for OPTIONS")
}

func TestCORS_GetRequest_AllowOrigin(t *testing.T) {
	f := &Filter{}
	err := f.Init(newTestContext())
	assert.NoError(t, err)

	ts := newTestCORSServer(t, f)

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/hello", nil)
	req.Header.Set("Origin", "http://example.com")
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "http://example.com", resp.Header.Get("Access-Control-Allow-Origin"))
}

func TestCORS_Handle_NonOptions(t *testing.T) {
	f := &Filter{}
	_ = f.Init(newTestContext())

	// Simulate a non-OPTIONS request through the filter chain
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	mockResp := &mockResponse{ResponseWriter: c.Writer, ctx: c}
	req := web.NewRequestForTest(c, mockResp, web.NewHandlerMeta())

	called := false
	fc := &mockFilterChain{nextFunc: func() (any, error) {
		called = true
		return "result", nil
	}}

	result, err := f.Handle(fc, req)
	assert.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, "result", result)
}

func TestCORS_Handle_Options(t *testing.T) {
	f := &Filter{}
	_ = f.Init(newTestContext())

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodOptions, "/test", nil)
	c.Request.Header.Set("Origin", "http://example.com")

	mockResp := &mockResponse{ResponseWriter: c.Writer, ctx: c}
	req := web.NewRequestForTest(c, mockResp, web.NewHandlerMeta())

	called := false
	fc := &mockFilterChain{nextFunc: func() (any, error) {
		called = true
		return "should-not-be-called", nil
	}}

	result, err := f.Handle(fc, req)
	assert.NoError(t, err)
	assert.Nil(t, result)
	assert.False(t, called, "filter chain should not be called for OPTIONS")
}

// --- helpers ---

type mockResponse struct {
	gin.ResponseWriter
	ctx *gin.Context
}

func (r *mockResponse) SetAttachmentFileName(fileName string) {
	r.Header().Set("Content-Disposition", `attachment; filename="`+fileName+`"`)
}
func (r *mockResponse) JSON(code int, value any) {
	r.Header().Set("Content-Type", "application/json")
	r.WriteHeader(code)
}
func (r *mockResponse) Abort() {}
func (r *mockResponse) Redirect(code int, location string) {
	r.Header().Set("Location", location)
	r.WriteHeader(code)
}
func (r *mockResponse) FileAttachment(path string, name string) {}
func (r *mockResponse) WriteStatus(code int)                    { r.WriteHeader(code) }
func (r *mockResponse) Message(t *web.Message)                  { r.JSON(t.Code, t) }
func (r *mockResponse) AbortWithMessage(t *web.Message)         { r.Message(t) }
func (r *mockResponse) AbortWithStatusJSON(i int, value any)    { r.JSON(i, value) }
func (r *mockResponse) AbortWithError(err error) error {
	return r.ctx.AbortWithError(http.StatusInternalServerError, err)
}

type mockFilterChain struct {
	nextFunc func() (any, error)
}

func (m *mockFilterChain) Next() (any, error) {
	return m.nextFunc()
}

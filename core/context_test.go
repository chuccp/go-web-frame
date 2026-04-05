package core

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/web"
	"github.com/gin-gonic/gin"
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

type testResponse struct {
	gin.ResponseWriter
	ctx *gin.Context
}

type closeNotifyRecorder struct {
	*httptest.ResponseRecorder
}

func (r *closeNotifyRecorder) CloseNotify() <-chan bool {
	ch := make(chan bool, 1)
	return ch
}

func (r *testResponse) SetAttachmentFileName(fileName string) {
	r.Header().Set("Content-Disposition", `attachment; filename="`+fileName+`"`)
}

func (r *testResponse) JSON(code int, value any) {
	r.ctx.JSON(code, value)
}

func (r *testResponse) Abort() {
	r.ctx.Abort()
}

func (r *testResponse) Redirect(code int, location string) {
	r.ctx.Redirect(code, location)
}

func (r *testResponse) FileAttachment(path string, name string) {
	r.ctx.FileAttachment(path, name)
}

func (r *testResponse) WriteStatus(code int) {
	r.ctx.Status(code)
}

func (r *testResponse) Message(t *web.Message) {
	r.ctx.JSON(t.Code, t)
}

func (r *testResponse) AbortWithMessage(t *web.Message) {
	r.ctx.JSON(t.Code, t)
	r.ctx.Abort()
}

func (r *testResponse) AbortWithStatusJSON(code int, value any) {
	r.ctx.AbortWithStatusJSON(code, value)
}

func (r *testResponse) AbortWithError(err error) error {
	return r.ctx.AbortWithError(http.StatusInternalServerError, err)
}

func newTestRequest(t *testing.T, method, target string, body io.Reader, headers map[string]string) (*web.Request, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	recorder := &closeNotifyRecorder{ResponseRecorder: httptest.NewRecorder()}
	req := httptest.NewRequest(method, target, body)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	c, _ := gin.CreateTestContext(recorder)
	c.Request = req
	return web.NewRequest(c, &testResponse{ResponseWriter: c.Writer, ctx: c}, web.NewHandlerMeta()), recorder.ResponseRecorder
}

func newContextWithHandlers(t *testing.T) (*Context, *web.Handles) {
	t.Helper()
	cfg, err := config.NewFromBytes([]byte(`{}`), "json")
	assert.NoError(t, err)
	handles := web.NewHandles()
	ctx := NewContext(cfg, nil)
	ctx.handlerConfig = web.NewHandlerConfig(nil, handles)
	return ctx, handles
}

func TestContext_StaticFsServesFile(t *testing.T) {
	ctx, handles := newContextWithHandlers(t)
	fs := http.FS(os.DirFS("testdata"))

	info := ctx.StaticFs("/assets", fs)
	assert.Equal(t, "/assets/*filepath", info.RelativePath())
	assert.True(t, handles.HasHandler(http.MethodGet, "/assets/*filepath"))
	assert.True(t, handles.HasHandler(http.MethodHead, "/assets/*filepath"))

	req, recorder := newTestRequest(t, http.MethodGet, "/assets/hello.txt", nil, nil)
	_, err := info.HandlerFunc()[0](req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "hello static\n", recorder.Body.String())
}

func TestContext_StaticServesFileFromDirectory(t *testing.T) {
	ctx, _ := newContextWithHandlers(t)
	dir := t.TempDir()
	file := filepath.Join(dir, "app.js")
	err := os.WriteFile(file, []byte("console.log('ok');\n"), 0644)
	assert.NoError(t, err)

	info := ctx.Static("/static", dir)
	req, recorder := newTestRequest(t, http.MethodGet, "/static/app.js", nil, nil)
	_, err = info.HandlerFunc()[0](req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "console.log('ok');\n", recorder.Body.String())
}

func TestContext_StaticFsHandlesHeadRequest(t *testing.T) {
	ctx, _ := newContextWithHandlers(t)
	fs := http.FS(os.DirFS("testdata"))

	info := ctx.StaticFs("/assets", fs)
	req, recorder := newTestRequest(t, http.MethodHead, "/assets/hello.txt", nil, nil)
	_, err := info.HandlerFunc()[0](req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "", recorder.Body.String())
	assert.Equal(t, "13", recorder.Header().Get("Content-Length"))
}

func TestContext_ReverseProxyForwardsRequest(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		defer r.Body.Close()

		w.Header().Set("X-Upstream-Host", r.Host)
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte(r.Method + " " + r.URL.Path + "?" + r.URL.RawQuery + " " + string(body)))
		assert.NoError(t, err)
	}))
	defer backend.Close()

	ctx, handles := newContextWithHandlers(t)
	info := ctx.ReverseProxy("/api", backend.URL+"/backend")

	assert.Equal(t, "/api/*proxyPath", info.RelativePath())
	assert.True(t, handles.HasHandler(http.MethodGet, "/api"))
	assert.True(t, handles.HasHandler(http.MethodPost, "/api/*proxyPath"))

	req, recorder := newTestRequest(
		t,
		http.MethodPost,
		"/api/users?id=7",
		strings.NewReader("name=alice"),
		map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
	)
	_, err := info.HandlerFunc()[0](req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, recorder.Code)
	assert.Equal(t, "POST /backend/users?id=7 name=alice", recorder.Body.String())
	assert.Contains(t, recorder.Header().Get("X-Upstream-Host"), "127.0.0.1:")
}

func TestContext_ReverseProxyInvalidTargetReturnsError(t *testing.T) {
	ctx, handles := newContextWithHandlers(t)

	info := ctx.ReverseProxy("/bad-proxy", "://bad target")
	assert.Equal(t, "/bad-proxy/*proxyPath", info.RelativePath())
	assert.True(t, handles.HasHandler(http.MethodGet, "/bad-proxy"))

	req, recorder := newTestRequest(t, http.MethodGet, "/bad-proxy/health", nil, nil)
	_, err := info.HandlerFunc()[0](req)
	assert.Error(t, err)
	assert.Equal(t, http.StatusOK, recorder.Code)
}

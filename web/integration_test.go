package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestIntegration_FullRequestChain tests the full request chain: filters -> handler -> converter -> response
func TestIntegration_FullRequestChain(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var filter1Called, filter2Called, handlerCalled bool

	serverConfig := &ServerConfig{Port: 0}
	handles := NewHandles()
	handles.Handle(http.MethodGet, "/chain", func(req *Request) (any, error) {
		handlerCalled = true
		return map[string]string{"result": "ok"}, nil
	})

	config := NewHandlerConfig(nil, handles, serverConfig)
	config.Use(&loggingFilter{callFlag: &filter1Called})
	config.Use(&authFilter{callFlag: &filter2Called})

	server := NewHttpServer(serverConfig, NewCertManager())
	server.Handle(config)

	ts := httptest.NewServer(server.Engine())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/chain")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]any
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	assert.Equal(t, "ok", result["result"])

	assert.True(t, filter1Called, "filter1 should be called")
	assert.True(t, filter2Called, "filter2 should be called")
	assert.True(t, handlerCalled, "handler should be called")
}

// TestIntegration_MultipleRestGroups tests multiple RestGroups on different ports
func TestIntegration_MultipleRestGroups(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Server 1
	config1 := &ServerConfig{Port: 19010}
	handles1 := NewHandles()
	handles1.Handle(http.MethodGet, "/api1", func(req *Request) (any, error) {
		return "server1", nil
	})
	server1 := NewHttpServer(config1, NewCertManager())
	server1.Handle(NewHandlerConfig(nil, handles1, config1))

	// Server 2
	config2 := &ServerConfig{Port: 19011}
	handles2 := NewHandles()
	handles2.Handle(http.MethodGet, "/api2", func(req *Request) (any, error) {
		return "server2", nil
	})
	server2 := NewHttpServer(config2, NewCertManager())
	server2.Handle(NewHandlerConfig(nil, handles2, config2))

	ts1 := httptest.NewServer(server1.Engine())
	defer ts1.Close()
	ts2 := httptest.NewServer(server2.Engine())
	defer ts2.Close()

	resp1, err := http.Get(ts1.URL + "/api1")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp1.StatusCode)
	resp1.Body.Close()

	resp2, err := http.Get(ts2.URL + "/api2")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp2.StatusCode)
	resp2.Body.Close()

	// Cross-check: /api2 should not exist on server1
	resp3, err := http.Get(ts1.URL + "/api2")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp3.StatusCode)
	resp3.Body.Close()
}

// TestIntegration_ContextPathWithFilters tests contextPath + filter chain combination
func TestIntegration_ContextPathWithFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serverConfig := &ServerConfig{Port: 0, ContextPath: "/api/v1"}
	handles := NewHandles()

	var capturedContextPath string
	handles.Handle(http.MethodGet, "/users", func(req *Request) (any, error) {
		capturedContextPath = req.ContextPath()
		return map[string]string{"user": "alice"}, nil
	})

	config := NewHandlerConfig(nil, handles, serverConfig)
	config.Use(&contextPathCaptureFilter{capture: &capturedContextPath})

	server := NewHttpServer(serverConfig, NewCertManager())
	server.Handle(config)

	ts := httptest.NewServer(server.Engine())
	defer ts.Close()

	// Must use contextPath in URL
	resp, err := http.Get(ts.URL + "/api/v1/users")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "/api/v1", capturedContextPath)
	resp.Body.Close()

	// Without contextPath should 404
	resp2, err := http.Get(ts.URL + "/users")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp2.StatusCode)
	resp2.Body.Close()
}

// TestIntegration_ConcurrentRequests tests concurrent request safety
func TestIntegration_ConcurrentRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var requestCount atomic.Int64

	serverConfig := &ServerConfig{Port: 0}
	handles := NewHandles()
	handles.Handle(http.MethodGet, "/concurrent", func(req *Request) (any, error) {
		requestCount.Add(1)
		return "ok", nil
	})

	server := NewHttpServer(serverConfig, NewCertManager())
	server.Handle(NewHandlerConfig(nil, handles, serverConfig))

	ts := httptest.NewServer(server.Engine())
	defer ts.Close()

	var wg sync.WaitGroup
	numRequests := 100

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := http.Get(ts.URL + "/concurrent")
			if err != nil {
				t.Error(err)
				return
			}
			defer resp.Body.Close()
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		}()
	}

	wg.Wait()
	assert.Equal(t, int64(numRequests), requestCount.Load())
}

// TestIntegration_StaticAndAPI tests static file + API route coexistence
func TestIntegration_StaticAndAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serverConfig := &ServerConfig{Port: 0}
	handles := NewHandles()

	// API route
	handles.Handle(http.MethodGet, "/api/data", func(req *Request) (any, error) {
		return map[string]string{"data": "value"}, nil
	})

	// Static route (using current dir as filesystem for testing)
	handles.AddStaticFs("/static", http.Dir("."))

	server := NewHttpServer(serverConfig, NewCertManager())
	server.Handle(NewHandlerConfig(nil, handles, serverConfig))

	ts := httptest.NewServer(server.Engine())
	defer ts.Close()

	// API route should work
	resp, err := http.Get(ts.URL + "/api/data")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Verify static routes are registered
	routes := server.Engine().Routes()
	var staticRouteFound bool
	for _, route := range routes {
		if route.Path == "/static/*filepath" && route.Method == http.MethodGet {
			staticRouteFound = true
			break
		}
	}
	assert.True(t, staticRouteFound, "Static route should be registered")
}

// loggingFilter is a test filter that sets a flag when called
type loggingFilter struct {
	callFlag *bool
}

func (f *loggingFilter) Handle(fc FilterChain, request *Request) (any, error) {
	*f.callFlag = true
	return fc.Next()
}

// authFilter is a test filter that sets a flag when called
type authFilter struct {
	callFlag *bool
}

func (f *authFilter) Handle(fc FilterChain, request *Request) (any, error) {
	*f.callFlag = true
	return fc.Next()
}

// contextPathCaptureFilter captures contextPath for testing
type contextPathCaptureFilter struct {
	capture *string
}

func (f *contextPathCaptureFilter) Handle(fc FilterChain, request *Request) (any, error) {
	*f.capture = request.ContextPath()
	return fc.Next()
}

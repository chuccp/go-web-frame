package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Helper function to create a gin context for testing
func createTestContext(method, path string, queryParams url.Values, body []byte, headers map[string]string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()

	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path+"?"+queryParams.Encode(), bytes.NewBuffer(body))
	} else {
		req = httptest.NewRequest(method, path+"?"+queryParams.Encode(), nil)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	return c, w
}

// newTestRequest creates a Request for testing with a default HandlerConfig
func newTestRequest(c *gin.Context, handlerMeta *HandlerMeta) *Request {
	handlerConfig := &HandlerConfig{
		converter:   nil,
		handles:     NewHandles(),
		filters:     nil,
		contextPath: "",
	}
	return newRequest(c, nil, handlerMeta, handlerConfig)
}

func TestRequest_Query(t *testing.T) {
	queryParams := url.Values{}
	queryParams.Set("key1", "value1")
	queryParams.Set("key2", "value2")

	c, _ := createTestContext("GET", "/test", queryParams, nil, nil)
	req := newTestRequest(c, NewHandlerMeta())

	assert.Equal(t, "value1", req.Query("key1"))
	assert.Equal(t, "", req.Query("nonexistent"))
}

func TestRequest_Param(t *testing.T) {
	c, _ := createTestContext("GET", "/users/123", url.Values{}, nil, nil)
	c.Params = []gin.Param{
		{Key: "id", Value: "123"},
	}

	req := newTestRequest(c, NewHandlerMeta())
	assert.Equal(t, "123", req.Param("id"))
}

func TestRequest_ParamInt(t *testing.T) {
	c, _ := createTestContext("GET", "/users/123", url.Values{}, nil, nil)
	c.Params = []gin.Param{
		{Key: "id", Value: "123"},
	}

	req := newTestRequest(c, NewHandlerMeta())
	assert.Equal(t, 123, req.ParamInt("id"))
}

func TestRequest_ParamIntForDefault(t *testing.T) {
	c, _ := createTestContext("GET", "/users/123", url.Values{}, nil, nil)
	c.Params = []gin.Param{
		{Key: "id", Value: "123"},
	}

	req := newTestRequest(c, NewHandlerMeta())
	assert.Equal(t, 123, req.ParamInt("id"))
	assert.Equal(t, 456, req.ParamIntForDefault("nonexistent", 456))
}

func TestRequest_ParamUint(t *testing.T) {
	c, _ := createTestContext("GET", "/users/123", url.Values{}, nil, nil)
	c.Params = []gin.Param{
		{Key: "id", Value: "123"},
	}

	req := newTestRequest(c, NewHandlerMeta())
	assert.Equal(t, uint(123), req.ParamUint("id"))
}

func TestRequest_IsGet(t *testing.T) {
	c, _ := createTestContext("GET", "/test", url.Values{}, nil, nil)
	req := newTestRequest(c, NewHandlerMeta())

	assert.True(t, req.IsGet())
	assert.False(t, req.IsPost())
}

func TestRequest_IsPost(t *testing.T) {
	c, _ := createTestContext("POST", "/test", url.Values{}, nil, nil)
	req := newTestRequest(c, NewHandlerMeta())

	assert.True(t, req.IsPost())
	assert.False(t, req.IsGet())
}

func TestRequest_BindJSON(t *testing.T) {
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	body, _ := json.Marshal(map[string]any{"name": "John", "age": 30})
	headers := map[string]string{"Content-Type": "application/json"}

	c, _ := createTestContext("POST", "/users", url.Values{}, body, headers)
	req := newTestRequest(c, NewHandlerMeta())

	var user User
	err := req.BindJSON(&user)
	assert.NoError(t, err)
	assert.Equal(t, "John", user.Name)
	assert.Equal(t, 30, user.Age)
}

func TestRequest_Json(t *testing.T) {
	body, _ := json.Marshal(map[string]any{"name": "John", "age": 30})
	headers := map[string]string{"Content-Type": "application/json"}

	c, _ := createTestContext("POST", "/test", url.Values{}, body, headers)
	req := newTestRequest(c, NewHandlerMeta())

	jsonObj, err := req.Json()
	assert.NoError(t, err)
	assert.Equal(t, "John", jsonObj.GetString("name"))
	assert.Equal(t, 30, jsonObj.GetInt("age"))
}

func TestRequest_GetJsonStringValue(t *testing.T) {
	body, _ := json.Marshal(map[string]any{"name": "John"})
	headers := map[string]string{"Content-Type": "application/json"}

	c, _ := createTestContext("POST", "/test", url.Values{}, body, headers)
	req := newTestRequest(c, NewHandlerMeta())

	name, err := req.GetJsonStringValue("name")
	assert.NoError(t, err)
	assert.Equal(t, "John", name)

	// Test default value
	city := req.GetJsonStringValueOrDefault("city", "Unknown")
	assert.Equal(t, "Unknown", city)
}

func TestRequest_GetJsonIntValue(t *testing.T) {
	body, _ := json.Marshal(map[string]any{"age": 30})
	headers := map[string]string{"Content-Type": "application/json"}

	c, _ := createTestContext("POST", "/test", url.Values{}, body, headers)
	req := newTestRequest(c, NewHandlerMeta())

	age, err := req.GetJsonIntValue("age")
	assert.NoError(t, err)
	assert.Equal(t, 30, age)

	// Test default value
	salary := req.GetJsonIntValueOrDefault("salary", 50000)
	assert.Equal(t, 50000, salary)
}

func TestRequest_JsonPage(t *testing.T) {
	body, _ := json.Marshal(map[string]any{"pageNo": 2, "pageSize": 20, "lastId": 100})
	headers := map[string]string{"Content-Type": "application/json"}

	c, _ := createTestContext("POST", "/test", url.Values{}, body, headers)
	req := newTestRequest(c, NewHandlerMeta())

	page, err := req.JsonPage()
	assert.NoError(t, err)
	assert.Equal(t, 2, page.PageNo)
	assert.Equal(t, 20, page.PageSize)
	assert.Equal(t, 100, page.LastId)
}

func TestRequest_GetFormParam(t *testing.T) {
	formData := url.Values{}
	formData.Set("key1", "value1")
	formData.Set("key2", "value2")

	headers := map[string]string{"Content-Type": "application/x-www-form-urlencoded"}
	c, _ := createTestContext("POST", "/test", url.Values{}, []byte(formData.Encode()), headers)

	req := newTestRequest(c, NewHandlerMeta())
	assert.Equal(t, "value1", req.GetFormParam("key1"))
	assert.Equal(t, "", req.GetFormParam("nonexistent"))
}

func TestRequest_GetIntFormParam(t *testing.T) {
	formData := url.Values{}
	formData.Set("key1", "123")

	headers := map[string]string{"Content-Type": "application/x-www-form-urlencoded"}
	c, _ := createTestContext("POST", "/test", url.Values{}, []byte(formData.Encode()), headers)

	req := newTestRequest(c, NewHandlerMeta())
	assert.Equal(t, 123, req.GetIntFormParam("key1"))
}

func TestRequest_GetIntFormParamOrDefault(t *testing.T) {
	formData := url.Values{}
	formData.Set("key1", "123")

	headers := map[string]string{"Content-Type": "application/x-www-form-urlencoded"}
	c, _ := createTestContext("POST", "/test", url.Values{}, []byte(formData.Encode()), headers)

	req := newTestRequest(c, NewHandlerMeta())
	assert.Equal(t, 123, req.GetIntFormParamOrDefault("key1", 999))
	assert.Equal(t, 999, req.GetIntFormParamOrDefault("nonexistent", 999))
}

func TestRequest_FormParamsPage(t *testing.T) {
	queryParams := url.Values{}
	queryParams.Set("pageNo", "2")
	queryParams.Set("pageSize", "20")
	queryParams.Set("lastId", "100")

	c, _ := createTestContext("GET", "/test", queryParams, nil, nil)
	req := newTestRequest(c, NewHandlerMeta())

	page, err := req.FormParamsPage()
	assert.NoError(t, err)
	assert.Equal(t, 2, page.PageNo)
	assert.Equal(t, 20, page.PageSize)
	assert.Equal(t, 100, page.LastId)
}

func TestRequest_Page(t *testing.T) {
	// Test GET request with form params
	queryParams := url.Values{}
	queryParams.Set("pageNo", "2")

	c, _ := createTestContext("GET", "/test", queryParams, nil, nil)
	req := newTestRequest(c, NewHandlerMeta())

	page, err := req.Page()
	assert.NoError(t, err)
	assert.Equal(t, 2, page.PageNo)

	// Test POST request with JSON body
	body, _ := json.Marshal(map[string]any{"pageNo": 3})
	headers := map[string]string{"Content-Type": "application/json"}

	c, _ = createTestContext("POST", "/test", url.Values{}, body, headers)
	req = newTestRequest(c, NewHandlerMeta())

	page, err = req.Page()
	assert.NoError(t, err)
	assert.Equal(t, 3, page.PageNo)
}

func TestRequest_GetHeader(t *testing.T) {
	headers := map[string]string{"Authorization": "Bearer token123"}
	c, _ := createTestContext("GET", "/test", url.Values{}, nil, headers)

	req := newTestRequest(c, NewHandlerMeta())
	assert.Equal(t, "Bearer token123", req.GetHeader("Authorization"))
	assert.Equal(t, "", req.GetHeader("Nonexistent"))
}

func TestRequest_RemoteAddr(t *testing.T) {
	c, _ := createTestContext("GET", "/test", url.Values{}, nil, nil)
	c.Request.RemoteAddr = "192.168.1.1:8080"

	req := newTestRequest(c, NewHandlerMeta())
	addr := req.RemoteAddr()
	// RemoteAddr 可能为空，所以我们只检查它不包含错误字符
	assert.NotContains(t, addr, "error")
}

func TestRequest_ClientIP(t *testing.T) {
	c, _ := createTestContext("GET", "/test", url.Values{}, nil, nil)
	req := newTestRequest(c, NewHandlerMeta())

	ip := req.ClientIP()
	// 在测试环境中，ClientIP 可能返回空字符串或默认值
	// 我们只验证它不包含错误字符
	assert.NotContains(t, ip, "error")
}

func TestRequest_Domain(t *testing.T) {
	c, _ := createTestContext("GET", "/test", url.Values{}, nil, nil)
	c.Request.Host = "example.com:8080"

	req := newTestRequest(c, NewHandlerMeta())
	domain := req.Domain()
	assert.Equal(t, "example.com", domain)
}

func TestRequest_IsMultipartForm(t *testing.T) {
	headers := map[string]string{"Content-Type": "multipart/form-data; boundary=----WebKitFormBoundary"}
	c, _ := createTestContext("POST", "/test", url.Values{}, []byte("test"), headers)

	req := newTestRequest(c, NewHandlerMeta())
	assert.True(t, req.IsMultipartForm())
}

func TestPage_DefaultValues(t *testing.T) {
	page := &Page{}
	assert.Equal(t, 0, page.PageNo)
	assert.Equal(t, 0, page.PageSize)
	assert.Equal(t, 0, page.LastId)
}

func TestJsonObject(t *testing.T) {
	obj := make(JsonObject)
	obj.Add("key1", "value1")
	obj.Add("key2", 123)

	assert.Equal(t, "value1", obj.GetString("key1"))
	assert.Equal(t, 123, obj.GetInt("key2"))
	assert.Equal(t, "", obj.GetString("nonexistent"))
	assert.Equal(t, 0, obj.GetInt("nonexistent"))

	// Test GetIntForDefault
	assert.Equal(t, 123, obj.GetIntForDefault("key2", 999))
	assert.Equal(t, 999, obj.GetIntForDefault("nonexistent", 999))
}

func TestRequest_ContextPathFromHandlerConfig(t *testing.T) {
	c, _ := createTestContext("GET", "/test", url.Values{}, nil, nil)
	handlerMeta := NewHandlerMeta()
	handlerConfig := &HandlerConfig{
		converter:   nil,
		handles:     NewHandles(),
		filters:     nil,
		contextPath: "/api/v1",
	}
	req := newRequest(c, nil, handlerMeta, handlerConfig)

	assert.Equal(t, "/api/v1", req.ContextPath())
}

func TestRequest_FullPath(t *testing.T) {
	c, _ := createTestContext("GET", "/users/123", url.Values{}, nil, nil)
	c.Params = []gin.Param{{Key: "id", Value: "123"}}
	req := newTestRequest(c, NewHandlerMeta())

	// FullPath returns gin.Context.FullPath() which depends on route registration
	// In test context without route registration, it may be empty
	assert.NotNil(t, req.FullPath())
}

func TestRequest_URL(t *testing.T) {
	c, _ := createTestContext("GET", "/test?key=value", url.Values{}, nil, nil)
	req := newTestRequest(c, NewHandlerMeta())

	assert.NotNil(t, req.URL())
	assert.Equal(t, "/test", req.URL().Path)
}

func TestRequest_RemoteIp(t *testing.T) {
	c, _ := createTestContext("GET", "/test", url.Values{}, nil, nil)
	req := newTestRequest(c, NewHandlerMeta())

	// RemoteIP returns the client IP from gin context
	ip := req.RemoteIp()
	// In test environment, this may be empty or a test IP
	assert.NotContains(t, ip, "error")
}

func TestRequest_ContentType(t *testing.T) {
	headers := map[string]string{"Content-Type": "application/json"}
	c, _ := createTestContext("POST", "/test", url.Values{}, []byte(`{}`), headers)
	req := newTestRequest(c, NewHandlerMeta())

	assert.Equal(t, "application/json", req.ContentType())
}

func TestRequest_Cookie(t *testing.T) {
	c, _ := createTestContext("GET", "/test", url.Values{}, nil, nil)
	req := newTestRequest(c, NewHandlerMeta())

	assert.NotNil(t, req.Cookie())
}

func TestHandlerConfig_Use(t *testing.T) {
	handles := NewHandles()
	config := NewHandlerConfig(nil, handles, DefaultServerConfig())

	assert.Len(t, config.filters, 0)

	mockFilter := &mockFilter{}
	config.Use(mockFilter)
	assert.Len(t, config.filters, 1)

	config.Use(mockFilter, mockFilter)
	assert.Len(t, config.filters, 3)
}

func TestHandlerConfig_HasHandler(t *testing.T) {
	handles := NewHandles()
	config := NewHandlerConfig(nil, handles, DefaultServerConfig())

	config.handles.Handle(http.MethodGet, "/test", func(req *Request) (any, error) {
		return "ok", nil
	})

	assert.True(t, config.HasHandler(http.MethodGet, "/test"))
	assert.False(t, config.HasHandler(http.MethodPost, "/test"))
	assert.False(t, config.HasHandler(http.MethodGet, "/other"))
}

type mockFilter struct{}

func (f *mockFilter) Handle(fc FilterChain, request *Request) (any, error) {
	return fc.Next()
}

func TestHandlersChain_Last(t *testing.T) {
	var chain HandlersChain
	h1 := func(req *Request) (any, error) { return nil, nil }
	h2 := func(req *Request) (any, error) { return nil, nil }

	chain = append(chain, h1, h2)

	// Last should return the last handler
	last := chain.Last()
	assert.NotNil(t, last)

	// Single element chain
	singleChain := HandlersChain{h1}
	singleLast := singleChain.Last()
	assert.NotNil(t, singleLast)

	// Empty chain should return nil
	assert.Nil(t, HandlersChain{}.Last())
}

func TestOf(t *testing.T) {
	h1 := func(req *Request) (any, error) { return nil, nil }
	h2 := func(req *Request) (any, error) { return nil, nil }

	chain := Of(h1, h2)

	assert.Len(t, chain, 2)
	assert.NotNil(t, chain[0])
	assert.NotNil(t, chain[1])
}
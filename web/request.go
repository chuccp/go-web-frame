package web

import (
	"context"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/util"
	"github.com/go-viper/mapstructure/v2"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// JsonObject is a convenience type for working with JSON objects as maps.
type JsonObject map[string]any

// GetString returns the value for key as a string.
func (o JsonObject) GetString(key string) string {
	return cast.ToString((o)[key])
}

// GetInt returns the value for key as an int.
func (o JsonObject) GetInt(key string) int {
	return cast.ToInt((o)[key])
}

// GetIntForDefault returns the value for key as an int, or defaultValue if the result is 0.
func (o JsonObject) GetIntForDefault(key string, defaultValue int) int {
	if v := o.GetInt(key); v != 0 {
		return v
	}
	return defaultValue
}

// Add sets the value for key in the JsonObject.
func (o JsonObject) Add(key string, value any) {
	(o)[key] = value
}

// Request wraps the HTTP request with helper methods for accessing
// parameters, query strings, JSON body, headers, and client info.
type Request struct {
	c             *gin.Context
	cookie        *Cookie
	jsonBody      *JsonObject
	handlerMeta   *HandlerMeta
	response      Response
	handlerConfig *HandlerConfig
}

// HandlerMeta returns the metadata attached to the matched route handler.
func (r *Request) HandlerMeta() *HandlerMeta {
	return r.handlerMeta
}

// Ctx 返回当前 HTTP 请求的 context.Context。
// 该 context 在请求完成时自动 cancel，无需用户管理其生命周期。
// 用于将请求级取消、超时和 trace 传播到数据库操作。
func (r *Request) Ctx() context.Context {
	return r.c.Request.Context()
}

// ContextPath returns the configured context path prefix (e.g. "/api/v1").
func (r *Request) ContextPath() string {
	if r.handlerConfig == nil {
		return ""
	}
	return r.handlerConfig.contextPath
}

// FullPath returns the full matched route path.
func (r *Request) FullPath() string {
	return r.c.FullPath()
}

// URL returns the parsed request URL.
func (r *Request) URL() *url.URL {
	return r.c.Request.URL
}

// RemoteAddr returns the raw remote address string (host:port).
func (r *Request) RemoteAddr() string {
	return r.c.Request.RemoteAddr
}

// RemoteIp returns the real client IP, supporting X-Forwarded-For headers.
func (r *Request) RemoteIp() string {
	return r.c.RemoteIP()
}

// ClientIP returns the client IP as resolved by Gin.
func (r *Request) ClientIP() string {
	return r.c.ClientIP()
}

// Domain returns the request host without the port.
func (r *Request) Domain() string {
	host := r.c.Request.Host
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}
	return host
}

// IsGet reports whether the request method is GET.
func (r *Request) IsGet() bool {
	return r.c.Request.Method == "GET"
}

// IsPost reports whether the request method is POST.
func (r *Request) IsPost() bool {
	return r.c.Request.Method == "POST"
}

// Query returns the query parameter value for the given key.
func (r *Request) Query(key string) string {
	return r.c.Query(key)
}

// Param returns the path parameter value for the given key.
func (r *Request) Param(key string) string {
	return r.c.Param(key)
}

// ParamInt returns the path parameter value as an int.
func (r *Request) ParamInt(key string) int {
	return cast.ToInt(r.Param(key))
}

// ParamIntForDefault returns the path parameter value as an int, or defaultValue if 0.
func (r *Request) ParamIntForDefault(key string, defaultValue int) int {
	if cast.ToInt(r.Param(key)) != 0 {
		return cast.ToInt(r.Param(key))
	}
	return defaultValue
}

// ParamUint returns the path parameter value as a uint.
func (r *Request) ParamUint(key string) uint {
	return cast.ToUint(r.Param(key))
}

// Cookie returns the cookie helper for this request.
func (r *Request) Cookie() *Cookie {
	return r.cookie
}

// Json parses the request body as JSON and returns it as a JsonObject.
// The result is cached on subsequent calls.
func (r *Request) Json() (*JsonObject, error) {
	if r.IsGet() {
		return nil, errors.New(GetNotSupportJson)
	}
	if r.jsonBody != nil {
		return r.jsonBody, nil
	}
	var jsonObject JsonObject
	err := r.c.BindJSON(&jsonObject)
	if err != nil {
		return nil, err
	}
	r.jsonBody = &jsonObject
	return &jsonObject, nil
}

// JsonPage extracts pagination parameters (pageNo, pageSize, lastId) from the JSON body.
func (r *Request) JsonPage() (*Page, error) {
	jsonObject, err := r.Json()
	if err != nil {
		return nil, err
	}
	return &Page{
		PageNo:   jsonObject.GetIntForDefault("pageNo", 1),
		PageSize: jsonObject.GetIntForDefault("pageSize", 10),
		LastId:   jsonObject.GetIntForDefault("lastId", 0),
	}, nil
}

// Page extracts pagination parameters from the request.
// For GET requests, it reads from query/form parameters.
// For POST requests, it reads from the JSON body.
func (r *Request) Page() (*Page, error) {
	if r.IsGet() {
		return r.FormParamsPage()
	}
	return r.JsonPage()
}

// GetFormParam returns a form parameter value by key.
func (r *Request) GetFormParam(key string) string {
	if value := r.c.Request.Form.Get(key); len(value) > 0 {
		return value
	}
	if value := r.c.Request.FormValue(key); len(value) > 0 {
		return value
	}
	return ""
}

// GetIntFormParam returns a form parameter value as an int.
func (r *Request) GetIntFormParam(key string) int {
	return cast.ToInt(r.GetFormParam(key))
}

// GetIntFormParamOrDefault returns a form parameter value as an int, or defaultValue if 0.
func (r *Request) GetIntFormParamOrDefault(key string, defaultValue int) int {
	if value := r.GetIntFormParam(key); value != 0 {
		return value
	}
	return defaultValue
}

// FormParamsPage extracts pagination parameters from form/query parameters.
func (r *Request) FormParamsPage() (*Page, error) {
	return &Page{
		PageNo:   r.GetIntFormParamOrDefault("pageNo", 1),
		PageSize: r.GetIntFormParamOrDefault("pageSize", 10),
		LastId:   r.GetIntFormParamOrDefault("lastId", 0),
	}, nil
}

// GetJsonStringValue returns a string value from the JSON body.
func (r *Request) GetJsonStringValue(key string) (string, error) {
	jsonObject, err := r.Json()
	if err != nil {
		return "", err
	}
	return jsonObject.GetString(key), nil
}

// GetJsonStringValueOrDefault returns a string value from the JSON body, or defaultValue if empty.
func (r *Request) GetJsonStringValueOrDefault(key string, defaultValue string) string {
	if value, _ := r.GetJsonStringValue(key); len(value) > 0 {
		return value
	}
	return defaultValue
}

// GetJsonIntValue returns an int value from the JSON body.
func (r *Request) GetJsonIntValue(key string) (int, error) {
	jsonObject, err := r.Json()
	if err != nil {
		return 0, err
	}
	return jsonObject.GetInt(key), nil
}

// GetJsonIntValueOrDefault returns an int value from the JSON body, or defaultValue if 0.
func (r *Request) GetJsonIntValueOrDefault(key string, defaultValue int) int {
	if value, _ := r.GetJsonIntValue(key); value != 0 {
		return value
	}
	return defaultValue
}

// BindJSON binds the request JSON body into the provided struct.
func (r *Request) BindJSON(value any) error {
	if r.IsGet() {
		return errors.New(GetNotSupportJson)
	}
	json, err := r.Json()
	if err != nil {
		return err
	}
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  value,
	})
	if err != nil {
		return err
	}
	return decoder.Decode(json)
}

// ContentType returns the request Content-Type header.
func (r *Request) ContentType() string {
	return r.c.ContentType()
}

// IsMultipartForm reports whether the request is a multipart form.
func (r *Request) IsMultipartForm() bool {
	return util.ContainsAnyIgnoreCase(r.ContentType(), "multipart/form-data")
}

// GetHeader returns the value of the request header for the given key.
func (r *Request) GetHeader(s string) string {
	return r.c.GetHeader(s)
}

// MultipartForm returns the parsed multipart form.
func (r *Request) MultipartForm() (*multipart.Form, error) {
	return r.c.MultipartForm()
}

// Request returns the underlying *http.Request.
func (r *Request) Request() *http.Request {
	return r.c.Request
}

// Response returns the Response writer for this request.
func (r *Request) Response() Response {
	return r.response
}

func (r *Request) GinContext() *gin.Context {
	return r.c
}

func newRequest(c *gin.Context, response Response, handlerMeta *HandlerMeta, handlerConfig *HandlerConfig) *Request {
	return &Request{c: c, cookie: NewCookie(c), response: response, handlerMeta: handlerMeta, handlerConfig: handlerConfig}
}

// NewRequestForTest creates a Request for testing purposes.
// This is exported only for use in tests outside the web package.
func NewRequestForTest(c *gin.Context, response Response, handlerMeta *HandlerMeta) *Request {
	handlerConfig := &HandlerConfig{
		converter:   nil,
		handles:     NewHandles(),
		filters:     nil,
		contextPath: "",
	}
	return newRequest(c, response, handlerMeta, handlerConfig)
}

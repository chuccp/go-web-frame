// Package web: HTTP request wrapper with helper methods for params, query, JSON binding.
package web

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/util"
	"github.com/chuccp/go-web-frame/value"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// GetNotSupportJson is the error message returned when a GET request
// attempts to bind JSON parameters.
const GetNotSupportJson = "JSON parameters are not supported in GET requests"

// Page represents pagination parameters for list queries.
// Alias for util.Page to keep backward compatibility.
type Page = util.Page

// PageAble is a paginated response wrapper containing total count and item list.
// Alias for util.PageAble to keep backward compatibility.
type PageAble[T any] = util.PageAble[T]

// ToPage creates a new PageAble from the given total count and item list.
// Delegates to util.ToPage for backward compatibility.
func ToPage[T any](total int64, list []T) *PageAble[T] {
	return util.ToPage[T](total, list)
}

// JSONObject is a convenience type for working with JSON objects as maps.
// It is an alias for value.Object, inheriting all helper methods (GetString, GetInt, etc.).
type JSONObject = value.Object

// HandlerFunc is the function signature for request handlers.
type HandlerFunc func(*Request) (any, error)

// Request wraps the HTTP request with helper methods for accessing
// parameters, query strings, JSON body, headers, and client info.
type Request struct {
	c            *gin.Context
	cookie       *Cookie
	body         value.Value
	bodyErr      error
	handlerMeta  *HandlerMeta
	response     Response
	serverConfig *ServerConfig
}

// HandlerMeta returns the metadata attached to the matched route handler.
func (r *Request) HandlerMeta() *HandlerMeta {
	return r.handlerMeta
}
func (r *Request) HasMeta(mo ...MetaOption) bool {
	if r.handlerMeta == nil {
		return false
	}
	return slices.ContainsFunc(mo, func(o MetaOption) bool {
		return o.has(r.handlerMeta)
	})
}

// Ctx returns the context.Context for the current HTTP request.
// The context is automatically cancelled when the request completes.
// Use this to propagate request-scoped cancellation, timeouts, and
// tracing to database operations.
func (r *Request) Ctx() context.Context {
	return r.c.Request.Context()
}

// ContextPath returns the configured context path prefix (e.g. "/api/v1").
func (r *Request) ContextPath() string {
	return r.serverConfig.ContextPath
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

// ParamIntForDefault returns the path parameter value as an int, or defaultValue if the param is empty.
func (r *Request) ParamIntForDefault(key string, defaultValue int) int {
	if s := r.Param(key); s != "" {
		return cast.ToInt(s)
	}
	return defaultValue
}

// ParamUint returns the path parameter value as a uint.
func (r *Request) ParamUint(key string) uint {
	return cast.ToUint(r.Param(key))
}

// Json parses the request body as JSON and returns it as a JSONObject.
// The result is cached on subsequent calls.
// The body size is limited by ServerConfig.MaxBodySize (default 10 MB).
func (r *Request) Json() (*JSONObject, error) {
	body, err := r.Value()
	if err != nil {
		return nil, err
	}
	if body != nil {
		if body.IsObject() {
			return body.AsObject(), nil
		}
		return nil, errors.New("invalid json object body")
	}
	return nil, errors.New("invalid json object")
}

func (r *Request) Value() (value.Value, error) {
	if r.IsGet() {
		return nil, errors.New(GetNotSupportJson)
	}
	if r.bodyErr != nil {
		return nil, r.bodyErr
	}
	if r.body != nil {
		return r.body, nil
	}
	r.limitBody()
	r.body, r.bodyErr = value.DecodeJSON(r.c.Request.Body)
	if r.bodyErr != nil {
		return nil, r.bodyErr
	}
	return r.body, nil
}

// limitBody wraps the request body with MaxBytesReader if MaxBodySize is configured.
func (r *Request) limitBody() {
	maxSize := r.maxBodySize()
	if maxSize > 0 {
		r.c.Request.Body = http.MaxBytesReader(r.c.Writer, r.c.Request.Body, maxSize)
	}
}

// maxBodySize returns the effective max body size from serverConfig.
// 0 (default) → DefaultMaxBodySize (10 MB), -1 → unlimited, >0 → custom limit.
func (r *Request) maxBodySize() int64 {
	if r.serverConfig == nil {
		return DefaultMaxBodySize
	}
	if r.serverConfig.MaxBodySize < 0 {
		return 0 // unlimited
	}
	if r.serverConfig.MaxBodySize == 0 {
		return DefaultMaxBodySize
	}
	return r.serverConfig.MaxBodySize
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
	if val := r.c.Request.Form.Get(key); len(val) > 0 {
		return val
	}
	if val := r.c.Request.FormValue(key); len(val) > 0 {
		return val
	}
	return ""
}

// GetIntFormParam returns a form parameter value as an int.
func (r *Request) GetIntFormParam(key string) int {
	return cast.ToInt(r.GetFormParam(key))
}

// GetIntFormParamOrDefault returns a form parameter value as an int, or defaultValue if the param is not set.
func (r *Request) GetIntFormParamOrDefault(key string, defaultValue int) int {
	if s := r.GetFormParam(key); s != "" {
		return cast.ToInt(s)
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
	if val, _ := r.GetJsonStringValue(key); len(val) > 0 {
		return val
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

// GetJsonIntValueOrDefault returns an int value from the JSON body, or defaultValue if the key is not present.
func (r *Request) GetJsonIntValueOrDefault(key string, defaultValue int) int {
	jsonObject, err := r.Json()
	if err != nil {
		return defaultValue
	}
	//if _, ok := (*jsonObject)[key]; ok {
	//	return jsonObject.GetInt(key)
	//}
	return jsonObject.GetIntForDefault(key, defaultValue)
}

// BindJSON binds the request JSON body into the provided struct.
func (r *Request) BindJSON(v any) error {
	if r.IsGet() {
		return errors.New(GetNotSupportJson)
	}
	jsonObj, err := r.Json()
	if err != nil {
		return err
	}
	return jsonObj.Decode(v)
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

// Cookie returns the cookie helper for this request.
func (r *Request) Cookie() *Cookie {
	return r.cookie
}

// Response returns the Response writer for this request.
func (r *Request) Response() Response {
	return r.response
}

// GinContext returns the underlying *gin.Context.
// Not recommended for use — prefer the Request helper methods.
// Only reach for this when you need to reuse libraries or middleware from the gin ecosystem.
func (r *Request) GinContext() *gin.Context {
	return r.c
}

func request(ctx *gin.Context, route *Route, serverConfig *ServerConfig) *Request {
	return &Request{
		c:            ctx,
		cookie:       NewCookie(ctx),
		handlerMeta:  route.handlerMeta,
		response:     newResponse(ctx),
		serverConfig: serverConfig,
	}
}

// NewRequestForTest creates a Request for testing purposes.
func NewRequestForTest(ctx *gin.Context, resp Response, meta *HandlerMeta) *Request {
	return &Request{
		c:           ctx,
		cookie:      NewCookie(ctx),
		handlerMeta: meta,
		response:    resp,
	}
}

// SaveUploadedFile saves an uploaded file to the destination path.
// It creates the destination directory if it does not exist.
func SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer func() {
		closeErr := src.Close()
		if err == nil {
			err = closeErr
		}
	}()

	if err = os.MkdirAll(filepath.Dir(dst), 0775); err != nil {
		return err
	}
	if err = os.Chmod(filepath.Dir(dst), 0775); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		closeErr := out.Close()
		if err == nil {
			err = closeErr
		}
	}()

	if _, err = io.Copy(out, src); err != nil {
		return err
	}
	return out.Sync()
}

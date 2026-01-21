package web

import (
	"math"
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

type JsonObject map[string]any

func (o JsonObject) GetString(key string) string {
	return cast.ToString((o)[key])
}
func (o JsonObject) GetInt(key string) int {
	return cast.ToInt((o)[key])
}
func (o JsonObject) GetIntForDefault(key string, defaultValue int) int {
	if v := o.GetInt(key); v != 0 {
		return v
	}
	return defaultValue
}

func (o JsonObject) Add(key string, value any) {
	(o)[key] = value
}

type HttpContext struct {
	gin.ResponseWriter
	c           *gin.Context
	cookie      *Cookie
	jsonBody    *JsonObject
	handlerMeta *HandlerMeta
	handlers    []HandlerFunc
	index       int8
}

const abortIndex int8 = math.MaxInt8 >> 1

func (r *HttpContext) Next() {
	r.index++
	for r.index < int8(len(r.handlers)) {
		if r.handlers[r.index] != nil {
			r.handlers[r.index](r)
		}
		r.index++
	}
}

func (r *HttpContext) IsAborted() bool {
	return r.index >= abortIndex || r.c.IsAborted()
}

// Abort prevents pending handlers from being called. Note that this will not stop the current handler.
// Let's say you have an authorization middleware that validates that the current request is authorized.
// If the authorization fails (ex: the password does not match), call Abort to ensure the remaining handlers
// for this request are not called.
func (r *HttpContext) Abort() {
	r.index = abortIndex
	r.c.Abort()
}

func (r *HttpContext) HandlerMeta() *HandlerMeta {
	return r.handlerMeta
}
func (r *HttpContext) FullPath() string {
	return r.c.FullPath()
}
func (r *HttpContext) GinContext() *gin.Context {
	return r.c
}

func (r *HttpContext) URL() *url.URL {
	return r.c.Request.URL
}

func (r *HttpContext) RemoteAddr() string {
	return r.c.Request.RemoteAddr
}

func (r *HttpContext) Domain() string {
	host := r.c.Request.Host
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}
	return host

}

func (r *HttpContext) IsGet() bool {
	return r.c.Request.Method == "GET"
}
func (r *HttpContext) IsPost() bool {
	return r.c.Request.Method == "POST"
}

func (r *HttpContext) Query(key string) string {
	return r.c.Query(key)
}
func (r *HttpContext) Param(key string) string {
	return r.c.Param(key)
}
func (r *HttpContext) ParamInt(key string) int {
	return cast.ToInt(r.Param(key))
}
func (r *HttpContext) ParamUint(key string) uint {
	return cast.ToUint(r.Param(key))
}
func (r *HttpContext) Cookie() *Cookie {
	return r.cookie
}

func (r *HttpContext) Json() (*JsonObject, error) {
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
func (r *HttpContext) JsonPage() (*Page, error) {
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
func (r *HttpContext) Page() (*Page, error) {
	if r.IsGet() {
		return r.FormParamsPage()
	}
	return r.JsonPage()
}
func (r *HttpContext) GetFormParam(key string) string {
	if value := r.c.Request.Form.Get(key); len(value) > 0 {
		return value
	}
	if value := r.c.Request.FormValue(key); len(value) > 0 {
		return value
	}
	return ""
}
func (r *HttpContext) GetIntFormParam(key string) int {
	return cast.ToInt(r.GetFormParam(key))
}
func (r *HttpContext) GetIntFormParamOrDefault(key string, defaultValue int) int {
	if value := r.GetIntFormParam(key); value != 0 {
		return value
	}
	return defaultValue
}
func (r *HttpContext) FormParamsPage() (*Page, error) {
	return &Page{
		PageNo:   r.GetIntFormParamOrDefault("pageNo", 1),
		PageSize: r.GetIntFormParamOrDefault("pageSize", 10),
		LastId:   r.GetIntFormParamOrDefault("lastId", 0),
	}, nil
}

func (r *HttpContext) GetJsonStringValue(key string) (string, error) {
	jsonObject, err := r.Json()
	if err != nil {
		return "", err
	}
	return jsonObject.GetString(key), nil
}
func (r *HttpContext) GetJsonStringValueOrDefault(key string, defaultValue string) string {
	if value, _ := r.GetJsonStringValue(key); len(value) > 0 {
		return value
	}
	return defaultValue
}
func (r *HttpContext) GetJsonIntValue(key string) (int, error) {
	jsonObject, err := r.Json()
	if err != nil {
		return 0, err
	}
	return jsonObject.GetInt(key), nil
}
func (r *HttpContext) GetJsonIntValueOrDefault(key string, defaultValue int) int {

	if value, _ := r.GetJsonIntValue(key); value != 0 {
		return value
	}
	return defaultValue
}

func (r *HttpContext) BindJSON(value any) error {
	if r.IsGet() {
		return errors.New(GetNotSupportJson)
	}
	json, err := r.Json()
	if err != nil {
		return err
	}
	return mapstructure.Decode(json, value)
}
func (r *HttpContext) ContentType() string {
	return r.GinContext().ContentType()
}

func (r *HttpContext) IsMultipartForm() bool {
	return util.ContainsAnyIgnoreCase(r.ContentType(), "multipart/form-data")

}

func (r *HttpContext) GetHeader(s string) string {
	return r.c.GetHeader(s)
}

func (r *HttpContext) MultipartForm() (*multipart.Form, error) {
	return r.c.MultipartForm()

}

func (r *HttpContext) Request() *http.Request {
	return r.c.Request
}

func (r *HttpContext) WriteStatus(code int) {
	r.c.Status(code)
}
func (r *HttpContext) AbortWithStatusJSON(i int, value any) {
	r.c.AbortWithStatusJSON(i, value)
}

func (r *HttpContext) Message(t *Message) {
	if t.Code == http.StatusMovedPermanently {
		r.c.Redirect(http.StatusMovedPermanently, t.Data.(string))
		r.Abort()
		return
	}
	r.c.JSON(t.Code, t)
	r.Abort()
}

func (r *HttpContext) AbortWithMessage(t *Message) {
	if t.Code == http.StatusMovedPermanently {
		r.c.Redirect(http.StatusMovedPermanently, t.Data.(string))
		r.Abort()
		return
	}
	r.c.JSON(t.Code, t)
	r.Abort()
}

func (r *HttpContext) SetAttachmentFileName(fileName string) {
	r.Header().Set("Content-Disposition", `attachment; filename="`+fileName+`"`)
}

func (r *HttpContext) JSON(code int, value any) {
	r.c.JSON(code, value)
}

func (r *HttpContext) Redirect(code int, location string) {
	r.c.Redirect(code, location)
}
func (r *HttpContext) FileAttachment(path string, name string) {
	r.c.FileAttachment(path, name)
}
func NewHttpContext(c *gin.Context, handlerMeta *HandlerMeta, handlers ...HandlerFunc) *HttpContext {
	return &HttpContext{c: c, cookie: NewCookie(c), handlerMeta: handlerMeta, handlers: handlers, index: -1}
}

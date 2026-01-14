package web

import (
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
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

type HttpContext struct {
	c        *gin.Context
	cookie   *Cookie
	jsonBody *JsonObject
	//digestAuth *DigestAuth
	handlerConfig *HandlerConfig
}

// Next should be used only inside middleware.
// It executes the pending handlers in the chain inside the calling handler.
// See example in GitHub.
func (r *HttpContext) Next() {
	r.c.Next()
}
func (r *HttpContext) FullPath() string {
	return r.c.FullPath()
}
func (r *HttpContext) GinContext() *gin.Context {
	return r.c
}
func (r *HttpContext) GetDigestAuth() *DigestAuth {
	return r.handlerConfig.digestAuth
}

func (r *HttpContext) SignIn(user any) (any, error) {
	return r.GetDigestAuth().SignIn(user, r)
}
func (r *HttpContext) SignOut() (any, error) {
	return r.GetDigestAuth().SignOut(r)
}

func (r *HttpContext) User() (any, error) {
	return r.GetDigestAuth().User(r)
}
func (r *HttpContext) URL() *url.URL {
	return r.c.Request.URL
}
func User[T any](r *HttpContext) (T, error) {
	u, err := r.User()
	if err != nil {
		return u.(T), err
	}
	v, ok := u.(T)
	if !ok {
		return v, errors.New("Type conversion error. Please check if it is a pointer type." + reflect.TypeOf(u).Name())
	}
	return u.(T), err

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

// Query  returns the keyed url query value if it exists,
// otherwise it returns an empty string `("")`.
// It is shortcut for `c.HttpContext.URL.Query().Get(key)`
//
//	    GET /path?id=1234&name=Manu&value=
//		   c.Query("id") == "1234"
//		   c.Query("name") == "Manu"
//		   c.Query("value") == ""
//		   c.Query("wtf") == ""
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

func (r *HttpContext) JSON(code int, value any) {
	r.c.JSON(code, value)
}
func (r *HttpContext) Message(t *Message) {
	if t.Code == http.StatusMovedPermanently {
		r.c.Redirect(http.StatusMovedPermanently, t.Data.(string))
		r.c.Abort()
		return
	}
	r.c.JSON(t.Code, t)
}
func (r *HttpContext) Abort() {
	r.c.Abort()
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

func (r *HttpContext) Redirect(code int, location string) {
	r.c.Redirect(code, location)
}

func (r *HttpContext) Write(bytes []byte) (int, error) {
	return r.c.Writer.Write(bytes)
}

func (r *HttpContext) FileAttachment(path string, name string) {
	r.c.FileAttachment(path, name)
}
func NewHttpContext(c *gin.Context, handlerConfig *HandlerConfig) *HttpContext {
	return &HttpContext{c: c, cookie: NewCookie(c), handlerConfig: handlerConfig}
}

package web

import (
	"errors"
	"reflect"
	"strings"

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

type Request struct {
	c          *gin.Context
	cookie     *Cookie
	jsonBody   *JsonObject
	digestAuth *DigestAuth
}

func (r *Request) GetDigestAuth() *DigestAuth {
	return r.digestAuth
}

func (r *Request) SignIn(user any) (any, error) {
	return r.digestAuth.SignIn(user, r)
}
func (r *Request) SignOut() (any, error) {
	return r.digestAuth.SignOut(r)
}

func (r *Request) User() (any, error) {
	return r.digestAuth.User(r)
}

func User[T any](r *Request) (T, error) {
	u, err := r.User()
	if err != nil {
		return u.(T), err
	}
	v, ok := u.(T)
	if !ok {
		return v, errors.New("类型转换错误,请检查是否为指针类型" + reflect.TypeOf(u).Name())
	}
	return u.(T), err

}

func (r *Request) RemoteAddr() string {
	return r.c.Request.RemoteAddr
}

func (r *Request) Domain() string {
	host := r.c.Request.Host
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}
	return host

}

func (r *Request) IsGet() bool {
	return r.c.Request.Method == "GET"
}
func (r *Request) IsPost() bool {
	return r.c.Request.Method == "POST"
}

func (r *Request) GetQuery(key string) string {
	return r.c.Query(key)
}
func (r *Request) Cookie() *Cookie {
	return r.cookie
}

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
func (r *Request) Page() (*Page, error) {
	if r.IsGet() {
		return r.FormParamsPage()
	}
	return r.JsonPage()
}
func (r *Request) GetFormParam(key string) string {
	if value := r.c.Request.Form.Get(key); len(value) > 0 {
		return value
	}
	if value := r.c.Request.FormValue(key); len(value) > 0 {
		return value
	}
	return ""
}
func (r *Request) GetIntFormParam(key string) int {
	return cast.ToInt(r.GetFormParam(key))
}
func (r *Request) GetIntFormParamOrDefault(key string, defaultValue int) int {
	if value := r.GetIntFormParam(key); value != 0 {
		return value
	}
	return defaultValue
}
func (r *Request) FormParamsPage() (*Page, error) {
	return &Page{
		PageNo:   r.GetIntFormParamOrDefault("pageNo", 1),
		PageSize: r.GetIntFormParamOrDefault("pageSize", 10),
		LastId:   r.GetIntFormParamOrDefault("lastId", 0),
	}, nil
}

func (r *Request) GetJsonStringValue(key string) (string, error) {
	jsonObject, err := r.Json()
	if err != nil {
		return "", err
	}
	return jsonObject.GetString(key), nil
}
func (r *Request) GetJsonStringValueOrDefault(key string, defaultValue string) string {
	if value, _ := r.GetJsonStringValue(key); len(value) > 0 {
		return value
	}
	return defaultValue
}
func (r *Request) GetJsonIntValue(key string) (int, error) {
	jsonObject, err := r.Json()
	if err != nil {
		return 0, err
	}
	return jsonObject.GetInt(key), nil
}
func (r *Request) GetJsonIntValueOrDefault(key string, defaultValue int) int {

	if value, _ := r.GetJsonIntValue(key); value != 0 {
		return value
	}
	return defaultValue
}

func (r *Request) BindJSON(value any) error {
	if r.IsGet() {
		return errors.New(GetNotSupportJson)
	}
	err := r.c.BindJSON(value)
	return err
}

func NewRequest(c *gin.Context, digestAuth *DigestAuth) *Request {
	return &Request{c: c, cookie: NewCookie(c), digestAuth: digestAuth}
}

package core

import (
	"net/http"
	"sync"

	config2 "github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/model"
	"github.com/chuccp/go-web-frame/util"
	"github.com/chuccp/go-web-frame/web"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Context struct {
	config       config2.IConfig
	httpServer   *web.HttpServer
	restMap      map[string]IRest
	modelMap     map[string]IModel
	rLock        *sync.RWMutex
	serviceMap   map[string]IService
	componentMap map[string]IComponent
	db           *gorm.DB
	transaction  *model.Transaction
	digestAuth   *web.DigestAuth
	schedule     *Schedule
}

func NewContext(config config2.IConfig, db *gorm.DB, schedule *Schedule) *Context {
	context := &Context{
		config:       config,
		restMap:      make(map[string]IRest),
		modelMap:     make(map[string]IModel),
		rLock:        new(sync.RWMutex),
		serviceMap:   make(map[string]IService),
		componentMap: make(map[string]IComponent),
		transaction:  model.NewTransaction(db),
		db:           db,
		schedule:     schedule,
	}
	return context
}

func (c *Context) Copy(digestAuth *web.DigestAuth, httpServer *web.HttpServer) *Context {
	context := &Context{
		config:       c.config,
		httpServer:   httpServer,
		restMap:      c.restMap,
		modelMap:     c.modelMap,
		rLock:        c.rLock,
		serviceMap:   c.serviceMap,
		db:           c.db,
		transaction:  c.transaction,
		digestAuth:   digestAuth,
		componentMap: c.componentMap,
		schedule:     c.schedule,
	}
	return context
}

func (c *Context) GetTransaction() *model.Transaction {
	return c.transaction
}
func (c *Context) GetSchedule() *Schedule {
	return c.schedule
}
func (c *Context) AddRest(rests ...IRest) {
	c.rLock.Lock()
	defer c.rLock.Unlock()
	for _, s := range rests {
		name := util.GetStructFullName(s)
		c.restMap[name] = s
	}
}
func (c *Context) GetRest(name string) IRest {
	return c.restMap[name]
}
func (c *Context) GetDB() *gorm.DB {
	return c.db
}

func (c *Context) AddModel(model ...IModel) {
	c.rLock.Lock()
	defer c.rLock.Unlock()
	for _, m := range model {
		name := util.GetStructFullName(m)
		c.modelMap[name] = m
	}
}
func (c *Context) AddComponent(components ...IComponent) {
	c.rLock.Lock()
	defer c.rLock.Unlock()
	for _, component := range components {
		name := util.GetStructFullName(component)
		c.componentMap[name] = component
	}
}

func GetComponent[T IComponent](c *Context) T {
	var t T
	for _, s := range c.componentMap {
		t, ok := s.(T)
		if ok {
			return t
		}
	}
	return t
}

func GetValueConfig[T any](key string, c *Context) T {
	var t T
	newValue := util.NewPtr(t)
	err := c.config.Unmarshal(key, newValue)
	if err != nil {
		log.Error("GetValueConfig", zap.Error(err))
		return t
	}
	return newValue
}

func (c *Context) AddService(services ...IService) {
	for _, s := range services {
		name := util.GetStructFullName(s)
		c.serviceMap[name] = s
	}
}

func GetService[T IService](c *Context) T {
	var t T
	for _, s := range c.serviceMap {
		t, ok := s.(T)
		if ok {
			return t
		}
	}
	return t
}

func GetModel[T IModel](c *Context) T {
	var v T
	for _, m := range c.modelMap {
		v, ok := m.(T)
		if ok {
			return v
		}
	}
	return v
}
func GetRest[T IRest](c *Context) T {
	var v T
	for _, r := range c.restMap {
		v, ok := r.(T)
		if ok {
			return v
		}
	}
	return v
}

func (c *Context) ginHandler(httpMethod string, relativePath string, handlers ...gin.HandlerFunc) {
	c.httpServer.Handle(httpMethod, relativePath, handlers...)
}

func (c *Context) authHandle(httpMethod, relativePath string, handlers ...web.HandlerFunc) {
	log.Debug("authHandle", zap.String("method", httpMethod), zap.String("path", relativePath), zap.Any("handlers", web.Of(handlers...).GetFuncName()))
	c.ginHandler(httpMethod, relativePath, web.ToGinHandlerFunc(c.digestAuth, web.AuthChecks(handlers...)...)...)
}

func (c *Context) handle(httpMethod, relativePath string, handlers ...web.HandlerFunc) {
	log.Debug("handle", zap.String("method", httpMethod), zap.String("path", relativePath), zap.Any("handlers", web.Of(handlers...).GetFuncName()))
	c.ginHandler(httpMethod, relativePath, web.ToGinHandlerFunc(c.digestAuth, handlers...)...)
}

func (c *Context) handleRaw(httpMethod, relativePath string, handlers ...web.HandlerRawFunc) {
	log.Debug("rawHandle", zap.String("method", httpMethod), zap.String("path", relativePath), zap.Any("handlers", web.OfRaw(handlers...).GetFuncName()))
	c.ginHandler(httpMethod, relativePath, web.ToGinHandlerRawFunc(c.digestAuth, handlers...)...)
}

func (c *Context) authHandleRaw(httpMethod, relativePath string, handlers ...web.HandlerRawFunc) {
	log.Debug("authRawHandle", zap.String("method", httpMethod), zap.String("path", relativePath), zap.Any("handlers", web.OfRaw(handlers...).GetFuncName()))
	c.ginHandler(httpMethod, relativePath, web.ToGinHandlerRawFunc(c.digestAuth, web.AuthRawChecks(handlers...)...)...)
}

func (c *Context) HandleAuth(httpMethod, relativePath string, handlers ...web.HandlerFunc) {
	c.authHandle(httpMethod, relativePath, handlers...)
}
func (c *Context) Handle(httpMethod, relativePath string, handlers ...web.HandlerFunc) {
	c.handle(httpMethod, relativePath, handlers...)
}

func (c *Context) HandleRaw(httpMethod, relativePath string, handlers ...web.HandlerRawFunc) {
	c.handleRaw(httpMethod, relativePath, handlers...)
}

func (c *Context) HandleRawAuth(httpMethod, relativePath string, handlers ...web.HandlerRawFunc) {
	c.authHandleRaw(httpMethod, relativePath, handlers...)
}
func (c *Context) Get(relativePath string, handlers ...web.HandlerFunc) {
	c.handle(http.MethodGet, relativePath, handlers...)
}

func (c *Context) GetAuth(relativePath string, handlers ...web.HandlerFunc) {
	c.authHandle(http.MethodGet, relativePath, handlers...)
}
func (c *Context) GetRaw(relativePath string, handlers ...web.HandlerRawFunc) {
	c.handleRaw(http.MethodGet, relativePath, handlers...)
}

func (c *Context) GetRawAuth(relativePath string, handlers ...web.HandlerRawFunc) {
	c.authHandleRaw(http.MethodGet, relativePath, handlers...)
}
func (c *Context) GetConfig() config2.IConfig {
	return c.config
}

func (c *Context) Post(relativePath string, handlers ...web.HandlerFunc) {
	c.handle(http.MethodPost, relativePath, handlers...)
}
func (c *Context) PostRaw(relativePath string, handlers ...web.HandlerRawFunc) {
	c.handleRaw(http.MethodPost, relativePath, handlers...)
}

func (c *Context) PostRawAuth(relativePath string, handlers ...web.HandlerRawFunc) {
	c.authHandleRaw(http.MethodPost, relativePath, handlers...)
}
func (c *Context) PostAuth(relativePath string, handlers ...web.HandlerFunc) {
	c.authHandle(http.MethodPost, relativePath, handlers...)
}
func (c *Context) AnyAuth(relativePath string, handlers ...web.HandlerFunc) {
	for _, method := range anyMethods {
		c.authHandle(method, relativePath, handlers...)
	}
}

func (c *Context) Any(relativePath string, handlers ...web.HandlerFunc) {
	for _, method := range anyMethods {
		c.handle(method, relativePath, handlers...)
	}
}

func (c *Context) AnyRaw(relativePath string, handlers ...web.HandlerRawFunc) {
	for _, method := range anyMethods {
		c.handleRaw(method, relativePath, handlers...)
	}
}

func (c *Context) AnyRawAuth(relativePath string, handlers ...web.HandlerRawFunc) {
	for _, method := range anyMethods {
		c.authHandleRaw(method, relativePath, handlers...)
	}
}

var (
	// anyMethods for RouterGroup Any method
	anyMethods = []string{
		http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodHead, http.MethodOptions, http.MethodDelete, http.MethodConnect,
		http.MethodTrace,
	}
)

func (c *Context) Delete(relativePath string, handlers ...web.HandlerFunc) {
	c.handle(http.MethodDelete, relativePath, handlers...)
}
func (c *Context) DeleteRaw(relativePath string, handlers ...web.HandlerRawFunc) {
	c.handleRaw(http.MethodDelete, relativePath, handlers...)
}
func (c *Context) DeleteRawAuth(relativePath string, handlers ...web.HandlerRawFunc) {
	c.authHandleRaw(http.MethodDelete, relativePath, handlers...)
}
func (c *Context) DeleteAuth(relativePath string, handlers ...web.HandlerFunc) {
	c.authHandle(http.MethodDelete, relativePath, handlers...)
}
func (c *Context) Put(relativePath string, handlers ...web.HandlerFunc) {
	c.handle(http.MethodPut, relativePath, handlers...)
}
func (c *Context) PutAuth(relativePath string, handlers ...web.HandlerFunc) {
	c.authHandle(http.MethodPut, relativePath, handlers...)
}
func (c *Context) PutRaw(relativePath string, handlers ...web.HandlerRawFunc) {
	c.handleRaw(http.MethodPut, relativePath, handlers...)
}
func (c *Context) PutRawAuth(relativePath string, handlers ...web.HandlerRawFunc) {
	c.authHandleRaw(http.MethodPut, relativePath, handlers...)
}

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
	modelMap     map[string]IModel
	rLock        *sync.RWMutex
	serviceMap   map[string]IService
	componentMap map[string]IComponent
	db           *gorm.DB
	transaction  *model.Transaction
	digestAuth   *web.DigestAuth
	schedule     *Schedule
	routeTree    RouteTree
	runnerMap    map[string]IRunner
}

func NewContext(config config2.IConfig, db *gorm.DB, schedule *Schedule) *Context {
	context := &Context{
		config:       config,
		modelMap:     make(map[string]IModel),
		rLock:        new(sync.RWMutex),
		serviceMap:   make(map[string]IService),
		componentMap: make(map[string]IComponent),
		transaction:  model.NewTransaction(db),
		runnerMap:    make(map[string]IRunner),
		db:           db,
		schedule:     schedule,
		routeTree:    make(RouteTree),
	}
	return context
}

func (c *Context) Copy(digestAuth *web.DigestAuth, httpServer *web.HttpServer) *Context {
	context := &Context{
		config:       c.config,
		httpServer:   httpServer,
		modelMap:     c.modelMap,
		rLock:        c.rLock,
		serviceMap:   c.serviceMap,
		db:           c.db,
		transaction:  c.transaction,
		digestAuth:   digestAuth,
		componentMap: c.componentMap,
		schedule:     c.schedule,
		routeTree:    make(RouteTree),
		runnerMap:    c.runnerMap,
	}
	return context
}

func (c *Context) GetTransaction() *model.Transaction {
	return c.transaction
}
func (c *Context) GetSchedule() *Schedule {
	return c.schedule
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

func (c *Context) AddRunner(runner ...IRunner) {
	c.rLock.Lock()
	defer c.rLock.Unlock()
	for _, r := range runner {
		name := util.GetStructFullName(r)
		c.runnerMap[name] = r
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

func (c *Context) AddService(services ...IService) {
	c.rLock.Lock()
	defer c.rLock.Unlock()
	for _, s := range services {
		name := util.GetStructFullName(s)
		c.serviceMap[name] = s
	}
}
func (c *Context) GetRunner(f func(m IRunner) bool) IRunner {
	c.rLock.RLock()
	defer c.rLock.RUnlock()
	for _, r := range c.runnerMap {
		if f(r) {
			return r
		}
	}
	return nil
}

func (c *Context) GetService(f func(m IService) bool) IService {
	c.rLock.RLock()
	defer c.rLock.RUnlock()
	for _, s := range c.serviceMap {
		if f(s) {
			return s
		}
	}
	return nil
}
func (c *Context) GetComponent(f func(m IComponent) bool) IComponent {
	c.rLock.RLock()
	defer c.rLock.RUnlock()
	for _, s := range c.componentMap {
		if f(s) {
			return s
		}
	}
	return nil
}
func (c *Context) GetModel(f func(m IModel) bool) IModel {
	c.rLock.RLock()
	defer c.rLock.RUnlock()
	for _, m := range c.modelMap {
		if f(m) {
			return m
		}
	}
	return nil
}

func (c *Context) Use(middlewareFunc ...MiddlewareFunc) {
	for _, middlewareFunc := range middlewareFunc {
		c.httpServer.Use(func(ctx *gin.Context) {
			if c.routeTree.Has(ctx.Request.Method, ctx.FullPath()) {
				middlewareFunc(web.NewRequest(ctx, c.digestAuth), c)
			}
		})
	}
}

func (c *Context) ginHandler(httpMethod string, relativePath string, handlers ...gin.HandlerFunc) {
	c.routeTree.Set(httpMethod, relativePath)
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

package core

import (
	"sync"

	config2 "github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/model"
	"github.com/chuccp/go-web-frame/web"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type IContext interface {
	AddModel(model ...IModel)
}
type contextGroup struct {
	contexts []IContext
}

func (cg *contextGroup) addContext(context IContext) {
	cg.contexts = append(cg.contexts, context)
}

func newContextGroup(parent IContext) *contextGroup {
	contexts := make([]IContext, 0)
	contexts = append(contexts, parent)
	return &contextGroup{
		contexts: contexts,
	}
}

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
	contextGroup *contextGroup
	configMap    map[string]IValueConfig
	schedule     *Schedule
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
		contextGroup: c.contextGroup,
		componentMap: c.componentMap,
		configMap:    c.configMap,
		schedule:     c.schedule,
	}
	c.contextGroup.addContext(context)
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
		c.restMap[s.Name()] = s
	}
}
func (c *Context) GetDigestAuth() *web.DigestAuth {
	return c.digestAuth
}

func (c *Context) AddModel(model ...IModel) {
	c.rLock.Lock()
	defer c.rLock.Unlock()
	for _, m := range model {
		c.modelMap[m.Name()] = m
	}
}
func (c *Context) GetRest(name string) IRest {
	return c.restMap[name]
}
func (c *Context) GetDB() *gorm.DB {
	return c.db
}

func (c *Context) addConfig(config ...IValueConfig) {
	c.rLock.Lock()
	defer c.rLock.Unlock()
	for _, config := range config {
		c.configMap[config.Key()] = config
	}

}
func (c *Context) Add() {
}

func (c *Context) addModel(model ...IModel) {
	c.rLock.Lock()
	defer c.rLock.Unlock()
	for _, m := range model {
		c.modelMap[m.Name()] = m
	}
}
func (c *Context) addComponent(components ...IComponent) {
	c.rLock.Lock()
	defer c.rLock.Unlock()
	for _, component := range components {
		c.componentMap[component.Name()] = component
	}
}
func (c *Context) GetComponent(name string) IComponent {
	return c.componentMap[name]
}

func GetComponent[T IComponent](name string, c *Context) T {
	v, _ := c.GetComponent(name).(T)
	return v
}

func (c *Context) GetValueConfig(key string) IValueConfig {
	return c.configMap[key]
}
func GetValueConfig[T IValueConfig](c *Context) T {
	var t T
	for _, v := range c.configMap {
		t, ok := v.(T)
		if ok {
			return t
		}
	}
	return t
}
func GetValueConfigByKey[T IValueConfig](key string, c *Context) T {
	v, _ := c.GetValueConfig(key).(T)
	return v
}

func (c *Context) GetModel(name string) IModel {
	return c.modelMap[name]
}
func (c *Context) AddService(services ...IService) {
	for _, s := range services {
		c.serviceMap[s.Name()] = s
	}
}
func (c *Context) GetService(name string) IService {
	return c.serviceMap[name]
}

func GetServiceByName[T IService](name string, c *Context) T {
	v, _ := c.GetService(name).(T)
	return v
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

func GetModelByName[T IModel](name string, c *Context) T {
	v, _ := c.GetModel(name).(T)
	return v
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
func GetRestByName[T IRest](name string, c *Context) T {
	v, _ := c.GetRest(name).(T)
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

func (c *Context) Get(relativePath string, handlers ...web.HandlerFunc) {
	log.Debug("Get", zap.String("path", relativePath), zap.Any("handlers", web.Of(handlers...).GetFuncName()))
	c.get(relativePath, handlers...)
}
func (c *Context) get(relativePath string, handlers ...web.HandlerFunc) {
	c.httpServer.GET(relativePath, web.ToGinHandlerFunc(c.digestAuth, handlers...)...)
}
func (c *Context) AuthGet(relativePath string, handlers ...web.HandlerFunc) {
	log.Debug("AuthGet", zap.String("path", relativePath), zap.Any("handlers", web.Of(handlers...).GetFuncName()))
	c.get(relativePath, web.AuthChecks(handlers...)...)
}
func (c *Context) GetAuth(relativePath string, handlers ...web.HandlerFunc) {
	c.AuthGet(relativePath, handlers...)
}
func (c *Context) GetRaw(relativePath string, handlers ...web.HandlerRawFunc) {
	log.Debug("GetRaw", zap.String("path", relativePath), zap.Any("handlers", web.OfRaw(handlers...).GetFuncName()))
	c.getRaw(relativePath, handlers...)
}
func (c *Context) getRaw(relativePath string, handlers ...web.HandlerRawFunc) {
	c.httpServer.GET(relativePath, web.ToGinHandlerRawFunc(c.digestAuth, handlers...)...)
}
func (c *Context) AuthGetRaw(relativePath string, handlers ...web.HandlerRawFunc) {
	log.Debug("AuthGetRaw", zap.String("path", relativePath), zap.Any("handlers", web.OfRaw(handlers...).GetFuncName()))
	c.getRaw(relativePath, web.AuthRawChecks(handlers...)...)
}
func (c *Context) GetRawAuth(relativePath string, handlers ...web.HandlerRawFunc) {
	c.AuthGetRaw(relativePath, handlers...)
}
func (c *Context) GetRawConfig() config2.IConfig {
	return c.config
}

func (c *Context) Post(relativePath string, handlers ...web.HandlerFunc) {
	log.Debug("Post", zap.String("path", relativePath), zap.Any("handlers", web.Of(handlers...).GetFuncName()))
	c.post(relativePath, handlers...)
}
func (c *Context) post(relativePath string, handlers ...web.HandlerFunc) {
	c.httpServer.POST(relativePath, web.ToGinHandlerFunc(c.digestAuth, handlers...)...)
}
func (c *Context) PostRaw(relativePath string, handlers ...web.HandlerRawFunc) {
	log.Debug("PostRaw", zap.String("path", relativePath), zap.Any("handlers", web.OfRaw(handlers...).GetFuncName()))
	c.postRaw(relativePath, handlers...)
}
func (c *Context) postRaw(relativePath string, handlers ...web.HandlerRawFunc) {
	c.httpServer.POST(relativePath, web.ToGinHandlerRawFunc(c.digestAuth, handlers...)...)
}
func (c *Context) AuthPostRaw(relativePath string, handlers ...web.HandlerRawFunc) {
	log.Debug("AuthPostRaw", zap.String("path", relativePath), zap.Any("handlers", web.OfRaw(handlers...).GetFuncName()))
	c.postRaw(relativePath, web.AuthRawChecks(handlers...)...)
}
func (c *Context) PostRawAuth(relativePath string, handlers ...web.HandlerRawFunc) {
	c.AuthPostRaw(relativePath, handlers...)
}
func (c *Context) AuthPost(relativePath string, handlers ...web.HandlerFunc) {
	log.Debug("AuthPost", zap.String("path", relativePath), zap.Any("handlers", web.Of(handlers...).GetFuncName()))
	c.post(relativePath, web.AuthChecks(handlers...)...)
}
func (c *Context) PostAuth(relativePath string, handlers ...web.HandlerFunc) {
	c.AuthPost(relativePath, handlers...)
}

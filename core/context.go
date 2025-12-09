package core

import (
	"sync"

	config2 "github.com/chuccp/go-web-frame/config"
	log2 "github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/web"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Context struct {
	config       *config2.Config
	log          *log2.Logger
	engine       *gin.Engine
	restMap      map[string]IRest
	modelMap     map[string]IModel
	rLock        *sync.RWMutex
	serviceMap   map[string]IService
	db           *gorm.DB
	transaction  *Transaction
	localCache   *web.LocalCache
	digestAuth   *web.DigestAuth
	componentMap map[string]IComponent
}

func (c *Context) Copy(digestAuth *web.DigestAuth, engine *gin.Engine) *Context {
	return &Context{
		config:       c.config,
		log:          c.log,
		engine:       engine,
		restMap:      c.restMap,
		modelMap:     c.modelMap,
		rLock:        c.rLock,
		serviceMap:   c.serviceMap,
		db:           c.db,
		transaction:  c.transaction,
		digestAuth:   digestAuth,
		localCache:   c.localCache,
		componentMap: c.componentMap,
	}
}

func (c *Context) GetLocalCache() *web.LocalCache {
	return c.localCache
}
func (c *Context) GetTransaction() *Transaction {
	return c.transaction
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

func (c *Context) GetModel(name string) IModel {
	return c.modelMap[name]
}
func (c *Context) AddService(services ...IService) {
	c.rLock.RLock()
	defer c.rLock.RUnlock()
	for _, s := range services {
		c.serviceMap[s.Name()] = s
	}
}
func (c *Context) GetService(name string) IService {
	return c.serviceMap[name]
}

func GetGetService[T IService](name string, c *Context) T {
	return c.GetService(name).(T)
}

func GetModel[T IModel](name string, c *Context) T {
	return c.GetModel(name).(T)
}

func GetRest[T IRest](name string, c *Context) T {
	return c.GetRest(name).(T)
}

func (c *Context) Get(relativePath string, handlers ...web.HandlerFunc) {
	c.engine.GET(relativePath, web.ToGinHandlerFunc(c.digestAuth, handlers...)...)
}
func (c *Context) AuthGet(relativePath string, handlers ...web.HandlerFunc) {
	c.Get(relativePath, web.AuthChecks(handlers...)...)
}
func (c *Context) GetAuth(relativePath string, handlers ...web.HandlerFunc) {
	c.AuthGet(relativePath, handlers...)
}
func (c *Context) GetRaw(relativePath string, handlers ...web.HandlerRawFunc) {
	c.engine.GET(relativePath, web.ToGinHandlerRawFunc(c.digestAuth, handlers...)...)
}
func (c *Context) AuthGetRaw(relativePath string, handlers ...web.HandlerRawFunc) {
	c.GetRaw(relativePath, web.AuthRawChecks(handlers...)...)
}
func (c *Context) GetRawAuth(relativePath string, handlers ...web.HandlerRawFunc) {
	c.AuthGetRaw(relativePath, handlers...)
}

func (c *Context) GetLogger() *log2.Logger {
	return c.log
}
func (c *Context) GetConfig() *config2.Config {
	return c.config
}

func (c *Context) Post(relativePath string, handlers ...web.HandlerFunc) {
	c.engine.POST(relativePath, web.ToGinHandlerFunc(c.digestAuth, handlers...)...)
}
func (c *Context) PostRaw(relativePath string, handlers ...web.HandlerRawFunc) {
	c.engine.POST(relativePath, web.ToGinHandlerRawFunc(c.digestAuth, handlers...)...)
}
func (c *Context) AuthPostRaw(relativePath string, handlers ...web.HandlerRawFunc) {
	c.PostRaw(relativePath, web.AuthRawChecks(handlers...)...)
}
func (c *Context) PostRawAuth(relativePath string, handlers ...web.HandlerRawFunc) {
	c.AuthPostRaw(relativePath, handlers...)
}
func (c *Context) AuthPost(relativePath string, handlers ...web.HandlerFunc) {
	c.Post(relativePath, web.AuthChecks(handlers...)...)
}
func (c *Context) PostAuth(relativePath string, handlers ...web.HandlerFunc) {
	c.AuthPost(relativePath, handlers...)
}

func (c *Context) Rest(name string) IRest {
	c.rLock.RLock()
	defer c.rLock.RUnlock()
	return c.restMap[name]
}

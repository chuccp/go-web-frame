package core

import (
	"net/http"
	"sync"

	config2 "github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/model"
	"github.com/chuccp/go-web-frame/util"
	"github.com/chuccp/go-web-frame/web"
	"github.com/sourcegraph/conc/panics"
	"go.uber.org/zap"
)

type Context struct {
	config            config2.IConfig
	modelMap          map[string]IModel
	rLock             *sync.RWMutex
	serviceMap        map[string]IService
	componentMap      map[string]IComponent
	runnerMap         map[string]IRunner
	defaultModelGroup IModelGroup
	modelGroup        map[string]IModelGroup
	handlerConfig     *web.HandlerConfig
	filters           []IFilter
	certManager       *web.CertManager
}

func NewContext(config config2.IConfig, defaultModelGroup IModelGroup) *Context {
	context := &Context{
		config:       config,
		modelMap:     make(map[string]IModel),
		rLock:        new(sync.RWMutex),
		serviceMap:   make(map[string]IService),
		componentMap: make(map[string]IComponent),
		//transaction:  model.NewTransaction(db),
		runnerMap:         make(map[string]IRunner),
		modelGroup:        make(map[string]IModelGroup),
		defaultModelGroup: defaultModelGroup,
		filters:           make([]IFilter, 0),
		certManager:       web.NewCertManager(),
	}
	return context
}
func (c *Context) CertManager() *web.CertManager {
	return c.certManager
}

func (c *Context) Copy(handlerConfig *web.HandlerConfig, filters []IFilter) *Context {
	context := &Context{
		config:            c.config,
		modelMap:          c.modelMap,
		rLock:             c.rLock,
		serviceMap:        c.serviceMap,
		componentMap:      c.componentMap,
		runnerMap:         c.runnerMap,
		modelGroup:        c.modelGroup,
		defaultModelGroup: c.defaultModelGroup,
		handlerConfig:     handlerConfig,
		filters:           filters,
		certManager:       c.certManager,
	}
	return context
}

func (c *Context) GetTransaction() *model.Transaction {
	return c.defaultModelGroup.GetTransaction()
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

func (c *Context) AddModelGroup(modelGroup ...IModelGroup) {
	c.rLock.Lock()
	defer c.rLock.Unlock()
	for _, m := range modelGroup {
		c.modelGroup[m.Name()] = m
	}
}
func (c *Context) DefaultModelGroup() IModelGroup {
	c.rLock.RLock()
	defer c.rLock.RUnlock()
	return c.defaultModelGroup
}

func (c *Context) GetModelGroup(name string) IModelGroup {
	c.rLock.RLock()
	defer c.rLock.RUnlock()
	if m, ok := c.modelGroup[name]; ok {
		return m
	}
	return nil
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
func (c *Context) GetFilter(f func(m IFilter) bool) IFilter {
	c.rLock.RLock()
	defer c.rLock.RUnlock()
	for _, filter := range c.filters {
		if f(filter) {
			return filter
		}
	}
	return nil
}
func (c *Context) Get(relativePath string, handlers ...web.HandlerFunc) *web.HandlerInfo {
	return c.handle(http.MethodGet, relativePath, handlers...)
}

func (c *Context) Post(relativePath string, handlers ...web.HandlerFunc) *web.HandlerInfo {
	return c.handle(http.MethodPost, relativePath, handlers...)
}
func (c *Context) Delete(relativePath string, handlers ...web.HandlerFunc) *web.HandlerInfo {
	return c.handle(http.MethodDelete, relativePath, handlers...)
}
func (c *Context) Put(relativePath string, handlers ...web.HandlerFunc) *web.HandlerInfo {
	return c.handle(http.MethodPut, relativePath, handlers...)
}

func (c *Context) Any(relativePath string, handlers ...web.HandlerFunc) *web.HandlerInfo {
	return c.handles(anyMethods, relativePath, handlers...)
}
func (c *Context) Go(f func(c *Context)) {
	catcher := panics.Try(func() {
		f(c)
	})
	err := catcher.AsError()
	if err != nil {
		log.Error("Context Go", zap.Error(err))
		log.PrintPanic(err)
	}

}

func (c *Context) handle(httpMethod string, relativePath string, handlers ...web.HandlerFunc) *web.HandlerInfo {
	return c.handles([]string{httpMethod}, relativePath, handlers...)
}
func (c *Context) handles(httpMethod []string, relativePath string, handlers ...web.HandlerFunc) *web.HandlerInfo {
	return c.handlerConfig.Handle(httpMethod, relativePath, handlers...)
}

func (c *Context) GetConfig() config2.IConfig {
	return c.config
}

var (
	// anyMethods for RouterGroup Any method
	anyMethods = []string{
		http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodHead, http.MethodOptions, http.MethodDelete, http.MethodConnect,
		http.MethodTrace,
	}
)

func GetService[T IService](c *Context) T {
	t, _ := c.GetService(func(m IService) bool {
		_, ok := m.(T)
		return ok
	}).(T)
	return t
}

func GetModel[T IModel](c *Context) T {
	t, _ := c.GetModel(func(m IModel) bool {
		_, ok := m.(T)
		return ok
	}).(T)
	return t
}

func GetFilter[T IFilter](c *Context) T {
	t, _ := c.GetFilter(func(m IFilter) bool {
		_, ok := m.(T)
		return ok
	}).(T)
	return t
}

func GetReNewModel[T IModel](db *db.DB, c *Context) T {
	t, ok := c.GetModel(func(m IModel) bool {
		_, ok := m.(T)
		return ok
	}).(T)
	if ok {
		t, _ = t.ReNew(db, c).(T)
		return t
	}
	return t
}
func GetComponent[T IComponent](c *Context) T {
	t, _ := c.GetComponent(func(m IComponent) bool {
		_, ok := m.(T)
		return ok
	}).(T)
	return t
}

func GetRunner[T IRunner](c *Context) T {
	t, _ := c.GetRunner(func(m IRunner) bool {
		_, ok := m.(T)
		return ok
	}).(T)
	return t
}

func UnmarshalKeyConfig[T any](key string, c *Context) (T, error) {
	var t T
	newValue := util.NewPtr(t)
	err := c.GetConfig().UnmarshalKey(key, newValue)
	if err != nil {
		log.Error("GetValueConfig", zap.Error(err))
		return t, err
	}
	return newValue, nil
}

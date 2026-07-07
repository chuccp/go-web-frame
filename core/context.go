package core

import (
	"context"
	"errors"
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

// Context is the core dependency injection container for the framework.
// It embeds context.Context for lifecycle control and manages all services,
// models, runners, and filters.
type Context struct {
	context.Context
	config        config2.IConfig
	modelMap      map[string]IModel
	rLock         *sync.RWMutex
	serviceMap    map[string]IService
	runnerMap     map[string]IRunner
	modelGroup    map[string]IModelGroup
	server        *web.Server
	filters       []IFilter
	allServiceMap map[string]IService
}

// NewContext creates a new Context with the given server, config, and parent context.
func NewContext(server *web.Server, config config2.IConfig, ctx context.Context) *Context {
	context := &Context{
		Context:       ctx,
		config:        config,
		modelMap:      make(map[string]IModel),
		rLock:         new(sync.RWMutex),
		serviceMap:    make(map[string]IService),
		allServiceMap: make(map[string]IService),
		runnerMap:     make(map[string]IRunner),
		modelGroup:    make(map[string]IModelGroup),
		filters:       make([]IFilter, 0),
		server:        server,
	}
	return context
}

// Server returns the web2 server for this context.
func (c *Context) Server() *web.Server {
	return c.server
}

// Copy creates a shallow copy of the Context with a new server and filters,
// sharing all maps and locks with the original for concurrent access.
func (c *Context) Copy(server *web.Server, filters []IFilter) *Context {
	context2 := &Context{
		Context:       c.Context,
		config:        c.config,
		modelMap:      c.modelMap,
		rLock:         c.rLock,
		serviceMap:    c.serviceMap,
		allServiceMap: c.allServiceMap,
		runnerMap:     c.runnerMap,
		modelGroup:    c.modelGroup,
		server:        server,
		filters:       filters,
	}
	return context2
}

// GetTransaction returns the transaction manager for the default model group.
func (c *Context) GetTransaction() *model.Transaction {
	return c.DefaultModelGroup().GetTransaction()
}

// GetTransactionByName returns the transaction manager for the named model group.
func (c *Context) GetTransactionByName(name string) *model.Transaction {
	return c.GetModelGroup(name).GetTransaction()
}

// AddModel registers one or more models into the context.
func (c *Context) AddModel(model ...IModel) {
	c.rLock.Lock()
	defer c.rLock.Unlock()
	for _, m := range model {
		name := util.GetStructFullQualifiedName(m)
		c.modelMap[name] = m
	}
}

// AddRunner registers one or more background runners into the context.
func (c *Context) AddRunner(runner ...IRunner) {
	c.rLock.Lock()
	defer c.rLock.Unlock()
	for _, r := range runner {
		name := util.GetStructFullQualifiedName(r)
		c.runnerMap[name] = r
		c.allServiceMap[name] = r
	}
}

// AddModelGroup registers one or more model groups into the context.
func (c *Context) AddModelGroup(modelGroup ...IModelGroup) {
	c.rLock.Lock()
	defer c.rLock.Unlock()
	for _, m := range modelGroup {
		c.modelGroup[m.Name()] = m
		c.allServiceMap[m.Name()] = m
	}
}

// DefaultModelGroup returns the default model group registered under ModelDefaultName.
func (c *Context) DefaultModelGroup() IModelGroup {
	return c.GetModelGroup(ModelDefaultName)
}

// GetModelGroup returns the model group with the given name, or nil if not found.
func (c *Context) GetModelGroup(name string) IModelGroup {
	c.rLock.RLock()
	defer c.rLock.RUnlock()
	if m, ok := c.modelGroup[name]; ok {
		return m
	}
	return nil
}

// AddService registers one or more services into the context.
func (c *Context) AddService(services ...IService) {
	c.rLock.Lock()
	defer c.rLock.Unlock()

	for _, s := range services {
		name := util.GetStructFullQualifiedName(s)
		c.serviceMap[name] = s
		c.allServiceMap[name] = s
		// If the service also implements IRunner, register it in the runner map as well.
		if runner, ok := s.(IRunner); ok {
			c.runnerMap[name] = runner
		}
	}
}

// GetRunner finds the first runner matching the predicate function.
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

// GetService finds the first service (including runners and model groups) matching the predicate.
func (c *Context) GetService(f func(m IService) bool) IService {
	c.rLock.RLock()
	defer c.rLock.RUnlock()
	for _, s := range c.allServiceMap {
		if f(s) {
			return s
		}
	}
	return nil
}

// GetModel finds the first model matching the predicate.
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

// GetFilter finds the first filter matching the predicate.
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

// Get registers a GET route handler and returns the route.
func (c *Context) Get(relativePath string, handlers ...web.HandlerFunc) *web.Route {
	return c.server.Get(relativePath, handlers...)
}

// Static registers a static file serving route and returns the route.
func (c *Context) Static(relativePath string, filepath string) *web.Route {
	return c.StaticFs(relativePath, http.Dir(filepath))
}

// ReverseProxy registers a reverse proxy to the target URL and returns the route.
func (c *Context) ReverseProxy(relativePath string, targetUrl string) *web.Route {
	return c.server.AddReverseProxy(relativePath, targetUrl)
}

// StaticFs registers a static file serving route with a custom http.FileSystem and returns the route.
func (c *Context) StaticFs(relativePath string, fs http.FileSystem) *web.Route {
	return c.server.AddStaticFs(relativePath, fs)
}

// WebSocket registers a WebSocket endpoint and returns the route.
func (c *Context) WebSocket(relativePath string, handler web.WebSocketHandler) *web.Route {
	return c.server.AddWebSocket(relativePath, handler)
}

// SSE registers a Server-Sent Events endpoint and returns the route.
func (c *Context) SSE(relativePath string, handler web.SSEHandler) *web.Route {
	return c.server.AddSSE(relativePath, handler)
}

// Post registers a POST route handler and returns the route.
func (c *Context) Post(relativePath string, handlers ...web.HandlerFunc) *web.Route {
	return c.server.Post(relativePath, handlers...)
}

// Delete registers a DELETE route handler and returns the route.
func (c *Context) Delete(relativePath string, handlers ...web.HandlerFunc) *web.Route {
	return c.server.Delete(relativePath, handlers...)
}

// Put registers a PUT route handler and returns the route.
func (c *Context) Put(relativePath string, handlers ...web.HandlerFunc) *web.Route {
	return c.server.Put(relativePath, handlers...)
}

// Any registers a route handler for all HTTP methods and returns the route.
func (c *Context) Any(relativePath string, handlers ...web.HandlerFunc) *web.Route {
	return c.server.Any(relativePath, handlers...)
}

// Go runs the given function in a goroutine with panic recovery.
// Errors are logged instead of crashing the application.
func (c *Context) Go(f func(c *Context)) {
	go func() {
		catcher := panics.Try(func() {
			f(c)
		})
		err := catcher.AsError()
		if err != nil {
			log.Error("Context Go", zap.Error(err))
			log.PrintPanic(err)
		}
	}()
}

func (c *Context) handle(httpMethod string, relativePath string, handlers ...web.HandlerFunc) *web.Route {
	return c.server.Handle(httpMethod, relativePath, handlers...)
}

// Handle registers a route handler for the given HTTP method and returns the route.
func (c *Context) Handle(httpMethod string, relativePath string, handlers ...web.HandlerFunc) *web.Route {
	return c.handle(httpMethod, relativePath, handlers...)
}

// GetConfig returns the configuration object associated with this context.
func (c *Context) GetConfig() config2.IConfig {
	return c.config
}

// GetRunners returns all registered runners in the context.
func (c *Context) GetRunners() []IRunner {
	c.rLock.RLock()
	defer c.rLock.RUnlock()
	runners := make([]IRunner, 0, len(c.runnerMap))
	for _, r := range c.runnerMap {
		runners = append(runners, r)
	}
	return runners
}


// GetService retrieves a service of the specified type from the context.
// Panics if the service is not registered.
func GetService[T IService](c *Context) T {
	t, ok := c.GetService(func(m IService) bool {
		_, ok := m.(T)
		return ok
	}).(T)
	if !ok {
		log.PanicErrors("GetService error", errors.New(util.GetStructFullQualifiedName(t)+" no register"))
	}
	return t
}

// GetModel retrieves a model of the specified type from the context.
// Panics if the model is not registered.
func GetModel[T IModel](c *Context) T {
	t, ok := c.GetModel(func(m IModel) bool {
		_, ok := m.(T)
		return ok
	}).(T)
	if !ok {
		log.PanicErrors("GetModel error", errors.New(util.GetStructFullQualifiedName(t)+" no register"))
	}
	return t
}

// GetFilter retrieves a filter of the specified type from the context.
// Panics if the filter is not registered.
func GetFilter[T IFilter](c *Context) T {
	t, ok := c.GetFilter(func(m IFilter) bool {
		_, ok := m.(T)
		return ok
	}).(T)
	if !ok {
		log.PanicErrors("GetFilter error", errors.New(util.GetStructFullQualifiedName(t)+" no register"))
	}
	return t
}

// GetReNewModel retrieves a model of the specified type and creates a fresh instance
// with the given database connection, useful for transaction handling.
// Panics if the model is not registered.
func GetReNewModel[T IModel](db *db.DB, c *Context) T {
	t, ok := c.GetModel(func(m IModel) bool {
		_, ok := m.(T)
		return ok
	}).(T)
	if !ok {
		log.PanicErrors("GetModel error", errors.New(util.GetStructFullQualifiedName(t)+" no register"))
	}
	renewed, ok := t.ReNew(db, c).(T)
	if !ok {
		log.PanicErrors("ReNew model type mismatch", errors.New(util.GetStructFullQualifiedName(t)+" ReNew returned wrong type"))
	}
	return renewed
}

// GetRunner retrieves a runner of the specified type from the context.
// Panics if the runner is not registered.
func GetRunner[T IRunner](c *Context) T {
	t, ok := c.GetRunner(func(m IRunner) bool {
		_, ok := m.(T)
		return ok
	}).(T)
	if !ok {
		log.PanicErrors("GetRunner error", errors.New(util.GetStructFullQualifiedName(t)+" no register"))
	}
	return t
}

// UnmarshalKeyConfig unmarshals configuration under the given key into the specified type.
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

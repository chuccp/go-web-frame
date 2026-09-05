// Package web: route definition, handler metadata, and meta options.
package web

import (
	"reflect"
	"runtime"
	"sync"

	"github.com/chuccp/go-web-frame/value"
)

// HandlerMeta holds key-value metadata attached to a route handler.
type HandlerMeta struct {
	meta *value.Object
	lock sync.RWMutex
}

// Add sets a key-value pair in the handler metadata.
func (hm *HandlerMeta) Add(key string, value any) {
	hm.lock.Lock()
	defer hm.lock.Unlock()
	hm.meta.PutAny(key, value)
}

func (hm *HandlerMeta) AddKeys(keys ...string) {
	hm.lock.Lock()
	defer hm.lock.Unlock()
	for _, key := range keys {
		hm.meta.PutAny(key, true)
	}
}

func (hm *HandlerMeta) HasAnyKey(key ...string) bool {
	hm.lock.RLock()
	defer hm.lock.RUnlock()
	return hm.meta.HasAnyKey(key...)
}
func (hm *HandlerMeta) HasKeyValue(key string, value any) bool {
	hm.lock.RLock()
	defer hm.lock.RUnlock()
	return hm.meta.HasKeyValue(key, value)
}

// NewHandlerMeta creates a new empty HandlerMeta.
func NewHandlerMeta() *HandlerMeta {
	return &HandlerMeta{
		meta: value.NewObject(),
	}
}

// Route represents an HTTP route with its path, methods, and handlers.
type Route struct {
	relativePath string
	handlerMeta  *HandlerMeta
	handlers     []HandlerFunc
	httpMethods  []string
}

func route(relativePath string, httpMethods []string, handlers ...HandlerFunc) *Route {
	return &Route{
		relativePath: relativePath,
		handlers:     handlers,
		httpMethods:  httpMethods,
		handlerMeta:  NewHandlerMeta(),
	}
}

// LastFuncName returns the fully-qualified function name of the last handler in the route.
// Used for debug logging; returns empty string if the route has no handlers.
func (r *Route) LastFuncName() string {
	if len(r.handlers) == 0 {
		return ""
	}
	last := r.handlers[len(r.handlers)-1]
	return runtime.FuncForPC(reflect.ValueOf(last).Pointer()).Name()
}

// WithMeta adds the given MetaOption values to the route's metadata.
// If handlerMeta is nil, a new one is created automatically.
func (r *Route) WithMeta(mo ...MetaOption) *Route {
	if r.handlerMeta == nil {
		r.handlerMeta = NewHandlerMeta()
	}
	for _, o := range mo {
		o.add(r.handlerMeta)
	}
	return r
}

// MetaOption is an option that can be applied to a HandlerMeta.
type MetaOption interface {
	add(o *HandlerMeta)
	has(o *HandlerMeta) bool
}

type funcOption struct {
	a func(o *HandlerMeta)
	h func(o *HandlerMeta) bool
}

func (fo *funcOption) add(o *HandlerMeta) {
	fo.a(o)
}

func (fo *funcOption) has(o *HandlerMeta) bool {
	return fo.h(o)
}

func newFunMetaOption(a func(o *HandlerMeta), h func(o *HandlerMeta) bool) *funcOption {
	return &funcOption{a: a, h: h}
}

// WithKey creates a MetaOption that sets the given keys to true in the handler metadata.
func WithKey(keys ...string) MetaOption {
	return newFunMetaOption(func(o *HandlerMeta) {

		o.AddKeys(keys...)
	}, func(o *HandlerMeta) bool {
		return o.HasAnyKey(keys...)
	})
}

// WithValue creates a MetaOption that sets a key-value pair in the handler metadata.
func WithValue(key string, value any) MetaOption {
	return newFunMetaOption(func(o *HandlerMeta) {
		o.Add(key, value)
	}, func(o *HandlerMeta) bool {

		return o.HasKeyValue(key, value)
	})
}

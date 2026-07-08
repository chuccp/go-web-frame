// Package web: route definition, handler metadata, and meta options.
package web

import (
	"reflect"
	"runtime"
)

// HandlerMeta holds key-value metadata attached to a route handler.
type HandlerMeta struct {
	data KV
}

// Add sets a key-value pair in the handler metadata.
func (hm *HandlerMeta) Add(key string, value any) {
	hm.data.Add(key, value)
}

// Has reports whether the given key exists in the handler metadata.
func (hm *HandlerMeta) Has(key string) bool {
	_, ok := hm.data[key]
	return ok
}

// Get returns the value for the given key, or nil if not found.
func (hm *HandlerMeta) Get(key string) any {
	v, ok := hm.data[key]
	if ok {
		return v
	}
	return nil
}

// NewHandlerMeta creates a new empty HandlerMeta.
func NewHandlerMeta() *HandlerMeta {
	return &HandlerMeta{
		data: make(KV),
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
	a func(oo *HandlerMeta)
	h func(oo *HandlerMeta) bool
}

func (fo *funcOption) add(oo *HandlerMeta) {
	fo.a(oo)
}

func (fo *funcOption) has(oo *HandlerMeta) bool {
	return fo.h(oo)
}

func newFunMetaOption(a func(o *HandlerMeta), h func(oo *HandlerMeta) bool) *funcOption {
	return &funcOption{a: a, h: h}
}

// WithKey creates a MetaOption that sets the given keys to true in the handler metadata.
func WithKey(keys ...string) MetaOption {
	return newFunMetaOption(func(oo *HandlerMeta) {
		for _, key := range keys {
			oo.Add(key, true)
		}
	}, func(oo *HandlerMeta) bool {
		for _, key := range keys {
			if oo.Has(key) {
				return true
			}
		}
		return false
	})
}

// WithValue creates a MetaOption that sets a key-value pair in the handler metadata.
func WithValue(key string, value any) MetaOption {
	return newFunMetaOption(func(oo *HandlerMeta) {
		oo.Add(key, value)
	}, func(oo *HandlerMeta) bool {
		if oo.Has(key) {
			return oo.Get(key) == value
		}
		return false
	})
}

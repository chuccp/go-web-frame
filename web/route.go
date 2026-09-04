// Package web: route definition, handler metadata, and meta options.
package web

import (
	"reflect"
	"runtime"
	"slices"

	"github.com/chuccp/go-web-frame/value"
)

// HandlerMeta holds key-value metadata attached to a route handler.
type HandlerMeta struct {
	data *value.Object
}

// Add sets a key-value pair in the handler metadata.
func (hm *HandlerMeta) Add(key string, value any) {
	hm.data.PutAny(key, value)
}

// Has reports whether the given key exists in the handler metadata.
func (hm *HandlerMeta) Has(key string) bool {
	return hm.data.HasKey(key)
}

// Get returns the value for the given key, or nil if not found.
func (hm *HandlerMeta) Get(key string) any {
	v := hm.data.Get(key)
	if v == nil {
		return nil
	}
	return v
}

// NewHandlerMeta creates a new empty HandlerMeta.
func NewHandlerMeta() *HandlerMeta {
	return &HandlerMeta{
		data: value.NewObject(),
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
		for _, key := range keys {
			o.Add(key, true)
		}
	}, func(o *HandlerMeta) bool {
		return slices.ContainsFunc(keys, o.Has)
	})
}

// WithValue creates a MetaOption that sets a key-value pair in the handler metadata.
func WithValue(key string, value any) MetaOption {
	return newFunMetaOption(func(o *HandlerMeta) {
		o.Add(key, value)
	}, func(o *HandlerMeta) bool {
		if o.Has(key) {
			return reflect.DeepEqual(o.Get(key), value)
		}
		return false
	})
}

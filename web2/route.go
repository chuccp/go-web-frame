package web2

import (
	"reflect"
	"runtime"
)

type HandlerMeta struct {
	data        KV
	contextPath string
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

type Route struct {
	relativePath string
	handlerMeta  *HandlerMeta
	handlers     []HandlerFunc
	httpMethods  []string
}

func newHandlerRoute(relativePath string, httpMethods []string, handlers ...HandlerFunc) *Route {
	return &Route{relativePath: relativePath, handlers: handlers, httpMethods: httpMethods}
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

// WithMeta applies the given MetaOption values to the route's metadata.
// If handlerMeta is nil, a new one is created automatically.
func (r *Route) WithMeta(mo ...MetaOption) *Route {
	if r.handlerMeta == nil {
		r.handlerMeta = NewHandlerMeta()
	}
	for _, o := range mo {
		o.apply(r.handlerMeta)
	}
	return r
}

// MetaOption is an option that can be applied to a HandlerMeta.
type MetaOption interface {
	apply(o *HandlerMeta)
}

type funcOption struct {
	f func(oo *HandlerMeta)
}

func (fo *funcOption) apply(oo *HandlerMeta) {
	fo.f(oo)
}

func newFunMetaOption(f func(o *HandlerMeta)) *funcOption {
	return &funcOption{f: f}
}

// WithKey creates a MetaOption that sets the given keys to true in the handler metadata.
func WithKey(keys ...string) MetaOption {
	return newFunMetaOption(func(oo *HandlerMeta) {
		for _, key := range keys {
			oo.Add(key, true)
		}
	})
}

// WithValue creates a MetaOption that sets a key-value pair in the handler metadata.
func WithValue(key string, value any) MetaOption {
	return newFunMetaOption(func(oo *HandlerMeta) {
		oo.Add(key, value)
	})
}

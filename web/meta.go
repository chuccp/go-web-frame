package web

// HandlerMeta holds arbitrary metadata attached to a registered route handler.
// It can be used to store and retrieve per-route configuration or flags.
type HandlerMeta struct {
	data        JsonObject
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
		data: make(JsonObject),
	}
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

// WithMeta applies the given MetaOption values to the handler's metadata.
func (hi *HandlerInfo) WithMeta(mo ...MetaOption) *HandlerInfo {
	for _, o := range mo {
		o.apply(hi.handlerMeta)
	}
	return hi
}

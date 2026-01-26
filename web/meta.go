package web

type HandlerMeta struct {
	data JsonObject
}

func (hm *HandlerMeta) Add(key string, value any) {
	hm.data.Add(key, value)
}
func (hm *HandlerMeta) Has(key string) bool {
	_, ok := hm.data[key]
	return ok
}
func (hm *HandlerMeta) Get(key string) any {
	v, ok := hm.data[key]
	if ok {
		return v
	}
	return nil
}
func NewHandlerMeta() *HandlerMeta {
	return &HandlerMeta{
		data: make(JsonObject),
	}
}

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

func WithKey(keys ...string) MetaOption {
	return newFunMetaOption(func(oo *HandlerMeta) {
		for _, key := range keys {
			oo.Add(key, true)
		}
	})
}
func WithValue(key string, value any) MetaOption {
	return newFunMetaOption(func(oo *HandlerMeta) {
		oo.Add(key, value)
	})
}

func (hi *HandlerInfo) WithMeta(mo ...MetaOption) *HandlerInfo {
	for _, o := range mo {
		o.apply(hi.handlerMeta)
	}
	return hi
}

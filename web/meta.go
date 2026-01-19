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
type HandlerInfo struct {
	HttpMethod   []string
	RelativePath string
	HandlerMeta  *HandlerMeta
	Handlers     []HandlerFunc
	routeTree    RouteTree
}

func NewHandlerInfo(httpMethod []string, relativePath string, handlers ...HandlerFunc) *HandlerInfo {
	return &HandlerInfo{
		HttpMethod:   httpMethod,
		RelativePath: relativePath,
		HandlerMeta:  NewHandlerMeta(),
		Handlers:     handlers,
		routeTree:    make(RouteTree),
	}
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

const (
	PemKey   = "pem"
	LoginKey = "login"
	RawKey   = "raw"
)

func WithPem(pem ...string) MetaOption {
	return newFunMetaOption(func(oo *HandlerMeta) {
		oo.Add(PemKey, pem)
	})
}

func WithKey(keys ...string) MetaOption {
	return newFunMetaOption(func(oo *HandlerMeta) {
		for _, key := range keys {
			oo.Add(key, true)
		}
	})
}

func WithLogin() MetaOption {
	return newFunMetaOption(func(oo *HandlerMeta) {
		oo.Add(LoginKey, true)
	})
}
func WithRaw() MetaOption {
	return newFunMetaOption(func(oo *HandlerMeta) {
		oo.Add(RawKey, true)
	})
}
func (h *HandlerInfo) WithMeta(mo ...MetaOption) *HandlerInfo {
	for _, o := range mo {
		o.apply(h.HandlerMeta)
	}
	return h
}

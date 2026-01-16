package web

type HandlerMeta struct {
	data JsonObject
}

func (hm *HandlerMeta) Add(key string, value any) {
	hm.data.Add(key, value)
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
	httpMethod   string
	relativePath string
	handlerMeta  *HandlerMeta
	handlers     []HandlerFunc
}

func NewHandlerInfo(httpMethod string, relativePath string, handlers ...HandlerFunc) *HandlerInfo {
	return &HandlerInfo{
		httpMethod:   httpMethod,
		relativePath: relativePath,
		handlerMeta:  NewHandlerMeta(),
		handlers:     handlers,
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
		o.apply(h.handlerMeta)
	}
	return h
}

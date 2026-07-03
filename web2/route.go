package web2

import "github.com/spf13/cast"

type KV map[string]any

// GetString returns the value for key as a string.
func (o KV) GetString(key string) string {
	return cast.ToString((o)[key])
}

// GetInt returns the value for key as an int.
func (o KV) GetInt(key string) int {
	return cast.ToInt((o)[key])
}

// GetIntForDefault returns the value for key as an int, or defaultValue if the result is 0.
func (o KV) GetIntForDefault(key string, defaultValue int) int {
	if v := o.GetInt(key); v != 0 {
		return v
	}
	return defaultValue
}

// Add sets the value for key in the JsonObject.
func (o KV) Add(key string, value any) {
	(o)[key] = value
}

type HandlerMeta struct {
	data        KV
	contextPath string
}

type HandlerInfo struct {
	relativePath string
	handlerMeta  *HandlerMeta
	handlers     []HandlerFunc
}

func NewHandlerInfo(relativePath string, handlers ...HandlerFunc) *HandlerInfo {
	return &HandlerInfo{relativePath: relativePath, handlers: handlers}
}

func (hi *HandlerInfo) IsHandler() bool {
	return len(hi.handlers) > 0
}

type routeTree map[string][]*HandlerInfo

func (rt routeTree) add(httpMethod string, handler *HandlerInfo) {
	rt[httpMethod] = append(rt[httpMethod], handler)
}

func (rt routeTree) each(f func(httpMethod string, handler []*HandlerInfo)) {
	for key, infos := range rt {
		f(key, infos)
	}
}

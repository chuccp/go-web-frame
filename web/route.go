package web

import "github.com/gin-gonic/gin"

type HandlerInfo struct {
	path        string
	handlerMeta *HandlerMeta
	handlers    []gin.HandlerFunc
}

func (hi *HandlerInfo) RelativePath() string {
	return hi.path
}
func (hi *HandlerInfo) HandlerFunc() []gin.HandlerFunc {
	return hi.handlers
}

type RouteInfo []*HandlerInfo

func NewHandlerInfo(path string, handlers ...gin.HandlerFunc) *HandlerInfo {
	return &HandlerInfo{handlerMeta: NewHandlerMeta(), path: path, handlers: handlers}
}

type RouteTree map[string]RouteInfo

func (rt RouteTree) Set(method string, handlerInfo *HandlerInfo) {
	rt[method] = append(rt[method], handlerInfo)
}

func (rt RouteTree) Has(method, path string) bool {
	if rt[method] != nil {
		for _, info := range rt[method] {
			if info.path == path {
				return true
			}
		}
	}
	return false
}
func (rt RouteTree) GetHandlerMeta(method, path string) *HandlerMeta {
	if rt[method] != nil {
		for _, info := range rt[method] {
			if info.path == path {
				return info.handlerMeta
			}
		}
	}
	return NewHandlerMeta()
}

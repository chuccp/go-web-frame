package web

import "net/http"

type HandlerInfo struct {
	path        string
	handlerMeta *HandlerMeta
	handlers    []HandlerFunc
	fs          http.FileSystem // 静态文件系统，用于 StaticFs
	targetUrl   string          // 反向代理目标 URL，用于 ReverseProxy
}

func (hi *HandlerInfo) RelativePath() string {
	return hi.path
}
func (hi *HandlerInfo) HandlerFunc() []HandlerFunc {
	return hi.handlers
}
func (hi *HandlerInfo) FileSystem() http.FileSystem {
	return hi.fs
}
func (hi *HandlerInfo) TargetUrl() string {
	return hi.targetUrl
}
func (hi *HandlerInfo) IsStaticFs() bool {
	return hi.fs != nil
}
func (hi *HandlerInfo) IsReverseProxy() bool {
	return hi.targetUrl != ""
}

type RouteInfo []*HandlerInfo

func NewHandlerInfo(path string, handlers ...HandlerFunc) *HandlerInfo {
	return &HandlerInfo{handlerMeta: NewHandlerMeta(), path: path, handlers: handlers}
}

func NewStaticFsHandlerInfo(relativePath string, fs http.FileSystem) *HandlerInfo {
	return &HandlerInfo{handlerMeta: NewHandlerMeta(), path: relativePath, fs: fs}
}

func NewReverseProxyHandlerInfo(relativePath string, targetUrl string) *HandlerInfo {
	return &HandlerInfo{handlerMeta: NewHandlerMeta(), path: relativePath, targetUrl: targetUrl}
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

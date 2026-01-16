package web

type handlerInfo struct {
	path        string
	HandlerMeta *HandlerMeta
}
type RouteInfo []*handlerInfo

type RouteTree map[string]RouteInfo

func (rt RouteTree) Set(method, path string, HandlerMeta *HandlerMeta) {
	rt[method] = append(rt[method], &handlerInfo{path: path, HandlerMeta: HandlerMeta})
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
				return info.HandlerMeta
			}
		}
	}
	return NewHandlerMeta()
}

package core

type RouteInfo []string

type RouteTree map[string]RouteInfo

func (rt RouteTree) Set(method, path string) {
	rt[method] = append(rt[method], path)
}

func (rt RouteTree) Has(method, path string) bool {
	if rt[method] != nil {
		for _, v := range rt[method] {
			if v == path {
				return true
			}
		}
	}
	return false
}

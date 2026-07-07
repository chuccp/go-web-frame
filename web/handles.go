// Package web: route collection that can be transferred to a Server.
package web

import (
	"net/http"
	"strings"
)

// Handles collects route definitions before they are registered on a Server.
// Use NewHandles() to create, then register routes via Get/Post/etc,
// and finally call server.AddHandles(handles) to apply them.
type Handles struct {
	routes []*Route
}

// NewHandles creates a new empty Handles.
func NewHandles() *Handles {
	return &Handles{
		routes: make([]*Route, 0),
	}
}

// Empty reports whether no routes have been registered.
func (h *Handles) Empty() bool {
	return len(h.routes) == 0
}

// Get registers a GET route handler.
func (h *Handles) Get(relativePath string, handlers ...HandlerFunc) *Route {
	return h.Handle(http.MethodGet, relativePath, handlers...)
}

// Post registers a POST route handler.
func (h *Handles) Post(relativePath string, handlers ...HandlerFunc) *Route {
	return h.Handle(http.MethodPost, relativePath, handlers...)
}

// Delete registers a DELETE route handler.
func (h *Handles) Delete(relativePath string, handlers ...HandlerFunc) *Route {
	return h.Handle(http.MethodDelete, relativePath, handlers...)
}

// Put registers a PUT route handler.
func (h *Handles) Put(relativePath string, handlers ...HandlerFunc) *Route {
	return h.Handle(http.MethodPut, relativePath, handlers...)
}

// Patch registers a PATCH route handler.
func (h *Handles) Patch(relativePath string, handlers ...HandlerFunc) *Route {
	return h.Handle(http.MethodPatch, relativePath, handlers...)
}

// Any registers a route handler for all HTTP methods.
func (h *Handles) Any(relativePath string, handlers ...HandlerFunc) *Route {
	return h.Handlers(allMethods, relativePath, handlers...)
}

// Handle registers a handler for a single HTTP method.
func (h *Handles) Handle(httpMethod string, relativePath string, handlers ...HandlerFunc) *Route {
	return h.Handlers([]string{httpMethod}, relativePath, handlers...)
}

// Handlers registers a handler for multiple HTTP methods.
func (h *Handles) Handlers(httpMethods []string, relativePath string, handlers ...HandlerFunc) *Route {
	route := route(relativePath, httpMethods, handlers...)
	h.routes = append(h.routes, route)
	return route
}

// AddSSE registers a Server-Sent Events endpoint.
func (h *Handles) AddSSE(relativePath string, handler SSEHandler) *Route {
	return h.Get(relativePath, func(r *Request) (any, error) {
		return &SSEResponse{Handler: handler}, nil
	})
}

// AddWebSocket registers a WebSocket endpoint.
func (h *Handles) AddWebSocket(relativePath string, handler WebSocketHandler) *Route {
	return h.Get(relativePath, func(r *Request) (any, error) {
		return &WSResponse{Handler: handler}, nil
	})
}

// AddReverseProxy registers a reverse proxy to the target URL for all HTTP methods.
// The relativePath should be a prefix (e.g. "/api"); a wildcard is appended automatically
// so that "/api/hello" matches in addition to "/api".
func (h *Handles) AddReverseProxy(relativePath string, target string) *Route {
	proxyPath := strings.TrimSuffix(relativePath, "/") + "/*path"
	return h.Handlers(allMethods, proxyPath, func(r *Request) (any, error) {
		return &ReverseProxyResponse{Target: target}, nil
	})
}

// AddStaticFs registers a static file server at the given path.
func (h *Handles) AddStaticFs(relativePath string, fs http.FileSystem) *Route {
	return h.Get(relativePath+"/*filepath", func(r *Request) (any, error) {
		return &FileSystemResponse{Filepath: r.Param("filepath"), FS: fs}, nil
	})
}

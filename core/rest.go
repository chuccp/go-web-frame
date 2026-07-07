package core

import (
	"github.com/chuccp/go-web-frame/web"
)

// RestGroup groups REST controllers, filters, and converter for a single HTTP server.
type RestGroup struct {
	rests        []IRest
	port         int
	converter    IConverter
	filters      []IFilter
	serverConfig *web.ServerConfig
	ContextPath  string
}

// Port returns the HTTP server port for this REST group.
func (rg *RestGroup) Port() int {
	return rg.port
}
// AddRest adds one or more REST controllers to this group.
func (rg *RestGroup) AddRest(rest ...IRest) *RestGroup {
	rg.rests = append(rg.rests, rest...)
	return rg
}

// AddFilter adds one or more filters/middleware to this group.
func (rg *RestGroup) AddFilter(filter ...IFilter) *RestGroup {
	rg.filters = append(rg.filters, filter...)
	return rg
}
// Converter sets the response converter for this group.
func (rg *RestGroup) Converter(converter IConverter) *RestGroup {
	rg.converter = converter
	return rg
}

// Merge combines another RestGroup into this one, merging rests, filters, and server config.
func (rg *RestGroup) Merge(restGroup *RestGroup) *RestGroup {
	rg.rests = append(rg.rests, restGroup.rests...)
	if rg.port == 0 {
		rg.port = restGroup.port
	}
	if rg.port == restGroup.port {
		if rg.serverConfig == nil || !rg.serverConfig.SSLEnabled() {
			rg.serverConfig = restGroup.serverConfig
		}
	}
	rg.filters = append(rg.filters, restGroup.filters...)
	return rg
}

func restGroup(serverConfig *web.ServerConfig, converter IConverter) *RestGroup {
	return &RestGroup{
		rests:        make([]IRest, 0),
		port:         serverConfig.Port,
		serverConfig: serverConfig,
		converter:    converter,
		filters:      make([]IFilter, 0),
	}
}

// DefaultConverter wraps web.DefaultConverter to satisfy the IConverter interface.
// All response handling logic lives in web.DefaultConverter; this type only
// provides the Init method required by IConverter (IService).
type DefaultConverter struct {
	inner *web.DefaultConverter
}

// Init initializes the converter. The built-in converter is stateless; this
// method exists to satisfy the IConverter (IService) interface.
func (c *DefaultConverter) Init(_ *Context) error {
	c.inner = &web.DefaultConverter{}
	return nil
}

// Request delegates to web.DefaultConverter.
func (c *DefaultConverter) Request(filterChain web.FilterChain, request *web.Request) {
	c.inner.Request(filterChain, request)
}

// RestGroupBuilder provides a fluent API for constructing RestGroup configurations.
type RestGroupBuilder struct {
	converter    IConverter
	serverConfig *web.ServerConfig
	rests        []IRest
	filters      []IFilter
	port         int
	contextPath  string
}

// Converter sets a custom response converter for this REST group.
func (b *RestGroupBuilder) Converter(converter IConverter) *RestGroupBuilder {
	b.converter = converter
	return b
}

// ServerConfig sets the HTTP server configuration for this REST group.
func (b *RestGroupBuilder) ServerConfig(serverConfig *web.ServerConfig) *RestGroupBuilder {
	b.serverConfig = serverConfig
	return b
}

// Port sets the HTTP server port for this REST group.
func (b *RestGroupBuilder) Port(port int) *RestGroupBuilder {
	b.port = port
	return b
}

// ContextPath sets the URL prefix for all routes in this REST group.
func (b *RestGroupBuilder) ContextPath(contextPath string) *RestGroupBuilder {
	b.contextPath = contextPath
	return b
}

// Handles is kept for backward compatibility but is now a no-op.
// Routes are registered directly on the server during Init.
func (b *RestGroupBuilder) Handles(handles any) *RestGroupBuilder {
	return b
}

// Rest adds one or more REST controllers to this REST group.
func (b *RestGroupBuilder) Rest(rest ...IRest) *RestGroupBuilder {
	b.rests = append(b.rests, rest...)
	return b
}
// Filter adds one or more filters/middleware to this REST group.
func (b *RestGroupBuilder) Filter(filters ...IFilter) *RestGroupBuilder {
	b.filters = append(b.filters, filters...)
	return b
}

// Build creates a RestGroup from the builder configuration.
// Applies defaults for any unset fields (default server config, converter, etc.).
func (b *RestGroupBuilder) Build() *RestGroup {
	if b.serverConfig == nil {
		b.serverConfig = web.DefaultServerConfig()
	}
	if b.port != 0 {
		b.serverConfig.Port = b.port
	}
	if b.converter == nil {
		b.converter = &DefaultConverter{}
	}
	if b.contextPath != "" {
		b.serverConfig.ContextPath = b.contextPath
	}
	group := restGroup(b.serverConfig, b.converter)
	group.AddRest(b.rests...)
	group.AddFilter(b.filters...)
	return group
}
// NewRestGroupBuilder creates a new RestGroupBuilder for fluent REST group construction.
func NewRestGroupBuilder() *RestGroupBuilder {
	return &RestGroupBuilder{}
}

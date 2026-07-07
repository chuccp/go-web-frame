package core

import (
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/chuccp/go-web-frame/util"
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

// DefaultConverter is the built-in response converter that handles JSON, strings, files, redirects, and messages.
type DefaultConverter struct {
	ctx *Context
}

// Init initializes the converter with the given context.
func (receiver *DefaultConverter) Init(ctx *Context) error {
	receiver.ctx = ctx
	return nil
}

// Request processes the filter chain result and writes the appropriate HTTP response.
// It supports JSON, plain text, file download, redirect, and message response types.
func (receiver *DefaultConverter) Request(filterChain web.FilterChain, request *web.Request) {
	value, err := filterChain.Next()
	resp := request.Response()
	if err != nil {
		err0 := web.Errors(value, err)
		resp.JSON(err0.Code, err0)
		resp.Abort()
	} else {
		if value != nil {
			switch t := value.(type) {
			case *web.Message:
				if t.Code == http.StatusMovedPermanently {
					resp.Redirect(http.StatusMovedPermanently, t.Data.(string))
					resp.Abort()
					return
				}
				resp.JSON(t.Code, value)
			case string:
				_, err2 := resp.Write([]byte(t))
				if err2 != nil {
					resp.Abort()
					return
				}
			case *web.FileResponse:
				if len(t.FileName) == 0 {
					_, filename := path.Split(t.Path)
					t.FileName = filename
				}
				if util.IsNotBlank(t.Suffix) && !strings.HasSuffix(t.FileName, t.Suffix) {
					if !strings.HasPrefix(t.Suffix, ".") {
						t.Suffix = "." + t.Suffix
					}
					t.FileName = t.FileName + t.Suffix
				}
				resp.FileAttachment(t.Path, t.FileName)
			case *os.File:
				resp.FileAttachment(t.Name(), t.Name())
			default:
				resp.JSON(200, web.Data(value))
			}
		}
	}
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

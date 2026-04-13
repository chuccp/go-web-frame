package core

import (
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/chuccp/go-web-frame/util"
	"github.com/chuccp/go-web-frame/web"
)

type RestGroup struct {
	rests        []IRest
	port         int
	converter    IConverter
	filters      []IFilter
	serverConfig *web.ServerConfig
	ContextPath  string
	handles      *web.Handles
}

func (rg *RestGroup) Port() int {
	return rg.port
}
func (rg *RestGroup) AddRest(rest ...IRest) *RestGroup {
	rg.rests = append(rg.rests, rest...)
	return rg
}

func (rg *RestGroup) AddFilter(filter ...IFilter) *RestGroup {
	rg.filters = append(rg.filters, filter...)
	return rg
}
func (rg *RestGroup) Converter(converter IConverter) *RestGroup {
	rg.converter = converter
	return rg
}

func (rg *RestGroup) Merge(restGroup *RestGroup) *RestGroup {
	rg.rests = append(rg.rests, restGroup.rests...)
	if rg.port == 0 {
		rg.port = restGroup.port
	}
	if rg.port == restGroup.port {
		if rg.serverConfig == nil || (!rg.serverConfig.SSLEnabled()) {
			rg.serverConfig = restGroup.serverConfig
		}
	}
	rg.filters = append(rg.filters, restGroup.filters...)
	return rg
}

func restGroup(serverConfig *web.ServerConfig, converter IConverter, handles *web.Handles) *RestGroup {
	return &RestGroup{
		rests:        make([]IRest, 0),
		port:         serverConfig.Port,
		serverConfig: serverConfig,
		converter:    converter,
		filters:      make([]IFilter, 0),
		handles:      handles,
	}
}

type DefaultConverter struct {
	ctx *Context
}

func (receiver *DefaultConverter) Init(ctx *Context) error {
	receiver.ctx = ctx
	return nil
}

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
			case *web.File:
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

type RestGroupBuilder struct {
	converter    IConverter
	serverConfig *web.ServerConfig
	handles      *web.Handles
	rests        []IRest
	filters      []IFilter
	port         int
	contextPath  string
}

func (b *RestGroupBuilder) Converter(converter IConverter) *RestGroupBuilder {
	b.converter = converter
	return b
}
func (b *RestGroupBuilder) ServerConfig(serverConfig *web.ServerConfig) *RestGroupBuilder {
	b.serverConfig = serverConfig
	return b
}
func (b *RestGroupBuilder) Port(port int) *RestGroupBuilder {
	b.port = port
	return b
}

func (b *RestGroupBuilder) ContextPath(contextPath string) *RestGroupBuilder {
	b.contextPath = contextPath
	return b
}
func (b *RestGroupBuilder) Handles(handles *web.Handles) *RestGroupBuilder {
	b.handles = handles
	return b
}
func (b *RestGroupBuilder) Rest(rest ...IRest) *RestGroupBuilder {
	b.rests = append(b.rests, rest...)
	return b
}
func (b *RestGroupBuilder) Filter(filters ...IFilter) *RestGroupBuilder {
	b.filters = append(b.filters, filters...)
	return b
}
func (b *RestGroupBuilder) Build() *RestGroup {
	if b.serverConfig == nil {
		b.serverConfig = web.DefaultServerConfig()
	}
	if b.port != 0 {
		b.serverConfig.Port = b.port
	}
	if b.handles == nil {
		b.handles = web.NewHandles()
	}
	if b.converter == nil {
		b.converter = &DefaultConverter{}
	}
	if b.contextPath != "" {
		b.serverConfig.ContextPath = b.contextPath
	}
	group := restGroup(b.serverConfig, b.converter, b.handles)
	group.AddRest(b.rests...)
	group.AddFilter(b.filters...)
	return group
}
func NewRestGroupBuilder() *RestGroupBuilder {
	return &RestGroupBuilder{}
}

package cors

import (
	"net/http"

	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Filter is a CORS middleware filter that handles preflight and cross-origin requests.
type Filter struct {
	handlerFunc gin.HandlerFunc
}

func (s *Filter) Init(ctx *core.Context) error {

	config := cors.DefaultConfig()
	config.AllowAllOrigins = false
	config.AllowCredentials = true
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	config.AllowOriginFunc = func(origin string) bool {
		return true
	}
	s.handlerFunc = cors.New(config)

	// Register as a gin global middleware so OPTIONS preflight requests
	// are handled before routing. go-web-frame filters only execute after
	// route matching, but OPTIONS requests don't match method-specific
	// routes and would otherwise return 404.
	//if engine, ok := ctx.Server().GetHandler().(*gin.Engine); ok {
	//	engine.Use(s.handlerFunc)
	//}

	return nil
}
func (s *Filter) Handle(filterChain web.FilterChain, request *web.Request) (any, error) {
	if request.Request().Method == http.MethodOptions {
		s.handlerFunc(request.GinContext())
		return nil, nil
	}
	s.handlerFunc(request.GinContext())
	return filterChain.Next()

}

// NewCrosFilter creates a new CORS filter with permissive defaults.
func NewCrosFilter() *Filter {
	return &Filter{}
}

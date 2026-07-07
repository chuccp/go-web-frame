// Package web: reverse proxy handler.
package web

import (
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/util"
	"go.uber.org/zap"
)

// ReverseProxyResponse is a handler return value that signals the converter
// to reverse-proxy the request to the given target URL.
type ReverseProxyResponse struct {
	Target string
}

// reverseProxy handles the actual reverse proxy forwarding.
func reverseProxy(request *Request, target string) {
	targetURL, err := url.Parse(target)
	if err != nil {
		log.Error("ReverseProxy: parse target URL", zap.Error(err), zap.String("target", target))
		request.response.AbortWithError(err)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	// NewSingleHostReverseProxy sets Director; clear it so we can use Rewrite instead.
	proxy.Director = nil
	// FullPath() returns the route pattern (e.g. "/api/*path").
	// Strip the wildcard suffix so we can match actual request paths.
	routePath := strings.TrimSuffix(request.FullPath(), "/*path")
	proxy.Rewrite = func(pr *httputil.ProxyRequest) {
		pr.SetXForwarded()
		pr.Out.URL.Scheme = targetURL.Scheme
		pr.Out.URL.Host = targetURL.Host
		inPath := pr.In.URL.Path
		if strings.HasPrefix(inPath, routePath) {
			remainder := strings.TrimPrefix(inPath, routePath)
			pr.Out.URL.Path = util.JoinUrl(targetURL.Path, remainder)
		}
	}
	proxy.ServeHTTP(request.response, request.Request())
}

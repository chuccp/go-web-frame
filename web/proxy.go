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
	path := request.FullPath()
	proxy.Rewrite = func(pr *httputil.ProxyRequest) {
		pr.SetXForwarded()
		if strings.HasPrefix(pr.In.URL.Path, path) {
			pr.Out.URL.Path = util.JoinUrl(targetURL.Path, strings.TrimPrefix(pr.In.URL.Path, path))
		}
	}
	proxy.ServeHTTP(request.response, request.Request())
}

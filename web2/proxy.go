package web2

import (
	"net/http"
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
	baseDirector := proxy.Director
	path := request.FullPath()
	proxy.Director = func(r *http.Request) {
		originalPath := r.URL.Path
		baseDirector(r)
		requestPath := originalPath
		if !strings.HasPrefix(path, "/") && strings.HasPrefix(originalPath, path) {
			requestPath = strings.TrimPrefix(originalPath, path)
		}
		r.URL.Path = util.JoinUrl(targetURL.Path, requestPath)
		r.URL.RawPath = r.URL.EscapedPath()
		r.Host = targetURL.Host
	}

	proxy.ServeHTTP(request.GinContext().Writer, request.Request())
}

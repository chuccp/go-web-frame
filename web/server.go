package web

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/util"
	//"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/net/http2"
)

const MaxHeaderBytes = 8192

const MaxReadHeaderTimeout = time.Second * 30

const MaxReadTimeout = time.Minute * 10

// SSLCert represents a local certificate file pair for a specific host.
// Configure Host, CertFile, and KeyFile to use a local TLS certificate
// instead of Let's Encrypt auto-certification.
type SSLCert struct {
	Host     string // Domain name for this certificate
	CertFile string // Path to the certificate file (PEM format)
	KeyFile  string // Path to the private key file (PEM format)
}


// SSLConfig holds the HTTPS/TLS configuration for the server.
// It supports two modes:
//   - Auto-cert: configure Hosts for Let's Encrypt automatic certificate management
//   - Local cert: configure Certs with pre-obtained certificate files
//
// Both modes can be combined. Local certs take priority over auto-certs
type SSLConfig struct {
	Enabled bool       // Whether HTTPS is enabled
	Hosts   []string   // Domain names for Let's Encrypt auto-certification
	Certs   []SSLCert  // Local certificate entries for pre-obtained certificates
}

// HasLocalCert returns true if Certs has any entries with both CertFile and KeyFile
func (s *SSLConfig) HasLocalCert() bool {
	if s == nil {
		return false
	}
	for _, c := range s.Certs {
		if c.CertFile != "" && c.KeyFile != "" {
			return true
		}
	}
	return false
}
// ServerConfig holds the HTTP server configuration.
type ServerConfig struct {
	Port        int        // Listen port (default: 19009)
	ContextPath string     // Route prefix applied to all routes (e.g., "/api")
	Locations   []string   // Static file directories to serve
	Page404     string     // Fallback page for 404 responses (useful for SPA)
	SSL         *SSLConfig // HTTPS/TLS configuration
}

const ServerConfigKey = "web.server"

// SSLEnabled reports whether HTTPS is configured and enabled.
func (s *ServerConfig) SSLEnabled() bool {
	return s.SSL != nil && s.SSL.Enabled
}

// DefaultServerConfig returns a ServerConfig with sensible defaults.
// Default port is 19009, SSL disabled.
func DefaultServerConfig() *ServerConfig {

	return &ServerConfig{
		Port: 19009,
		SSL: &SSLConfig{
			Enabled: false,
		},
	}
}

// HttpServer manages an HTTP server instance with Gin engine,
// route handling, static file serving, and TLS support.
type HttpServer struct {
	httpServer     *http.Server
	engine         *gin.Engine
	serverConfig   *ServerConfig
	certManager    *CertManager
	memFileSystem  *MemFileSystem
	handlerConfigs []*HandlerConfig
}

func defaultEngine() *gin.Engine {
	engine := gin.Default()
	// Trust all proxies to resolve real IP from X-Forwarded-For
	engine.SetTrustedProxies([]string{"0.0.0.0/0"})
	// Enable resolving real IP from client IP header
	engine.ForwardedByClientIP = true
	//config := cors.DefaultConfig()
	//config.AllowAllOrigins = false
	//config.AllowCredentials = true
	//config.AllowOriginFunc = func(origin string) bool {
	//	return true
	//}
	//engine.Use(cors.New(config))
	return engine
}

// NewHttpServer creates a new HttpServer with the given server config and certificate manager.
func NewHttpServer(serverConfig *ServerConfig, certManager *CertManager) *HttpServer {
	engine := defaultEngine()
	return &HttpServer{
		engine:         engine,
		serverConfig:   serverConfig,
		certManager:    certManager,
		memFileSystem:  DefaultMemFileSystem(serverConfig),
		handlerConfigs: make([]*HandlerConfig, 0),
	}
}
// Port returns the configured listen port.
func (httpServer *HttpServer) Port() int {
	return httpServer.serverConfig.Port
}

// Engine returns the underlying Gin engine.
func (httpServer *HttpServer) Engine() *gin.Engine {
	return httpServer.engine
}

// joinContextPath joins the context path prefix with the relative path
func joinContextPath(contextPath string, relativePath string) string {

	if contextPath == "" {
		return relativePath
	}
	// Ensure contextPath starts with /
	if !strings.HasPrefix(contextPath, "/") {
		contextPath = "/" + contextPath
	}
	// Remove trailing slash from contextPath
	contextPath = strings.TrimSuffix(contextPath, "/")

	// Handle root path
	if relativePath == "/" {
		return contextPath + "/"
	}

	// Ensure relativePath starts with /
	if !strings.HasPrefix(relativePath, "/") {
		relativePath = "/" + relativePath
	}

	return contextPath + relativePath
}
// AddHandle registers a HandlerConfig to be processed when Handle is called.
func (httpServer *HttpServer) AddHandle(handlerConfig *HandlerConfig) {
	httpServer.handlerConfigs = append(httpServer.handlerConfigs, handlerConfig)
}

// Handle processes all registered HandlerConfigs, binding filters and routes
// to the Gin engine. Must be called after all handlers are added via AddHandle.
func (httpServer *HttpServer) Handle() {
	allFilters := make([]Filter, 0)
	for _, handlerConfig := range httpServer.handlerConfigs {
		allFilters = append(allFilters, handlerConfig.filters...)
	}
	httpServer.engine.Use(func(ctx *gin.Context) {
		fullPath := ctx.FullPath()
		if len(fullPath) == 0 {
			resp := newResponse(ctx)
			request := newRequest(ctx, resp, NewHandlerMeta(), nil)
			mock := newMockFilterChain(request, nil, allFilters, nil)
			mock.Converter()
		}

	})

	for _, handlerConfig := range httpServer.handlerConfigs {
		// Process API routes
		for httpMethod, routeInfo := range handlerConfig.handles.RouteTree() {
			for _, handlerInfo := range routeInfo {
				// Set contextPath on HandlerMeta
				fullPath := joinContextPath(handlerConfig.contextPath, handlerInfo.path)
				handlerInfo.fullPath = fullPath
				if handlerInfo.IsWebSocket() {
					log.Debug("Handle WebSocket", zap.String("path", fullPath), zap.Any("handlers", Of(handlerInfo.handlers...).GetFuncName()))
					httpServer.handleWebSocket(fullPath, handlerConfig, handlerInfo)
				} else if handlerInfo.IsSSE() {
					log.Debug("Handle SSE", zap.String("path", fullPath), zap.Any("handlers", Of(handlerInfo.handlers...).GetFuncName()))
					httpServer.handleSSE(fullPath, handlerConfig, handlerInfo)
				} else if handlerInfo.IsReverseProxy() {
					log.Debug("Handle ReverseProxy", zap.String("path", fullPath))
					httpServer.handleReverseProxy(httpMethod, fullPath, handlerInfo)
				} else if handlerInfo.IsStaticFs() {
					log.Debug("Handle StaticFs", zap.String("path", fullPath))
					httpServer.handleStaticFs(fullPath, handlerInfo)
				} else if len(handlerInfo.handlers) > 0 {
					log.Debug("handle", zap.String("method", httpMethod), zap.String("path", fullPath), zap.Any("handlers", Of(handlerInfo.handlers...).GetFuncName()))
					httpServer.engine.Handle(httpMethod, fullPath, httpServer.ToGinHandlerFunc(handlerConfig, handlerInfo.handlers...)...)
				}
			}
		}
	}
}

func (httpServer *HttpServer) handleStaticFs(relativePath string, handlerInfo *HandlerInfo) {
	fs := handlerInfo.FileSystem()
	fileServer := http.StripPrefix(relativePath, http.FileServer(fs))
	pattern := path.Join(relativePath, "*filepath")
	if relativePath == "/" {
		pattern = "/*filepath"
	}
	httpServer.engine.GET(pattern, func(ctx *gin.Context) {
		// HandlerMeta with contextPath is available via handlerInfo.HandlerMeta()
		fileServer.ServeHTTP(ctx.Writer, ctx.Request)
	})
	httpServer.engine.HEAD(pattern, func(ctx *gin.Context) {
		fileServer.ServeHTTP(ctx.Writer, ctx.Request)
	})
}

func (httpServer *HttpServer) handleWebSocket(relativePath string, handlerConfig *HandlerConfig, handlerInfo *HandlerInfo) {
	upgrader := handlerInfo.Upgrader()
	handler := handlerInfo.WebSocketHandler()
	httpServer.engine.GET(relativePath, httpServer.toGinHandlerFunc(handlerConfig, func(request *Request) (any, error) {

		conn, err := upgrader.Upgrade(request.c.Writer, request.c.Request, nil)
		if err != nil {
			log.Error("WebSocket upgrade failed", zap.Error(err), zap.String("path", relativePath))
		}
		defer conn.Close()
		if err := handler(conn); err != nil {
			log.Debug("WebSocket handler error", zap.Error(err), zap.String("path", relativePath))
		}
		return nil, err
	}))
}

func (httpServer *HttpServer) handleSSE(relativePath string, handlerConfig *HandlerConfig, handlerInfo *HandlerInfo) {
	handler := handlerInfo.SSEHandler()
	httpServer.engine.GET(relativePath, httpServer.toGinHandlerFunc(handlerConfig, func(request *Request) (any, error) {
		// HandlerMeta with contextPath is available via handlerInfo.HandlerMeta()
		stream := NewSSEStream(request.c.Writer)
		if stream == nil {
			log.Error("SSE stream creation failed", zap.String("path", relativePath))
			return nil, errors.New("SSE stream creation failed")
		}
		defer stream.Close()
		stream.SetHeaders()
		if err := handler(stream); err != nil {
			log.Debug("SSE handler error", zap.Error(err), zap.String("path", relativePath))
		}
		return nil, nil
	}))
}

func (httpServer *HttpServer) handleReverseProxy(httpMethod string, relativePath string, handlerInfo *HandlerInfo) {
	targetUrl := handlerInfo.TargetUrl()
	// HandlerMeta with contextPath is available via handlerInfo.HandlerMeta()
	target, err := url.Parse(targetUrl)
	if err != nil {
		log.Error("handleReverseProxy targetUrl", zap.Error(err), zap.String("targetUrl", targetUrl))
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	baseDirector := proxy.Director
	proxy.Director = func(r *http.Request) {
		originalPath := r.URL.Path
		baseDirector(r)
		requestPath := originalPath
		if relativePath != "/" && strings.HasPrefix(originalPath, relativePath) {
			requestPath = strings.TrimPrefix(originalPath, relativePath)
		}
		r.URL.Path = util.JoinUrl(target.Path, requestPath)
		r.URL.RawPath = r.URL.EscapedPath()
		r.Host = target.Host
	}

	// 注册两个路由：精确匹配和通配符匹配
	if relativePath == "/" {
		httpServer.engine.Handle(httpMethod, "/*proxyPath", func(ctx *gin.Context) {
			proxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
	} else {
		httpServer.engine.Handle(httpMethod, relativePath, func(ctx *gin.Context) {
			proxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		httpServer.engine.Handle(httpMethod, path.Join(relativePath, "/*proxyPath"), func(ctx *gin.Context) {
			proxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
	}
}
// ToGinHandlerFunc converts framework HandlerFunc values to Gin handler functions.
func (httpServer *HttpServer) ToGinHandlerFunc(handlerConfig *HandlerConfig, handlers ...HandlerFunc) []gin.HandlerFunc {
	var handlerFunc = make([]gin.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		handlerFunc[i] = httpServer.toGinHandlerFunc(handlerConfig, handler)
	}
	return handlerFunc
}
func (httpServer *HttpServer) toGinHandlerFunc(handlerConfig *HandlerConfig, handler HandlerFunc) gin.HandlerFunc {
	handlerFunc := func(ctx *gin.Context) {
		resp := newResponse(ctx)
		handlerMeta := handlerConfig.HandlerMeta(ctx.Request.Method, ctx.FullPath())
		request := newRequest(ctx, resp, handlerMeta, handlerConfig)
		mock := newMockFilterChain(request, handlerConfig.converter, handlerConfig.filters, &lastFilter{handler})
		mock.Converter()
	}
	return handlerFunc
}

// Run starts the HTTP server and blocks until the context is cancelled.
// It handles static file serving, TLS configuration, and graceful shutdown.
func (httpServer *HttpServer) Run(ctx context.Context) error {
	log.Info("Start the service：", zap.Any("serverConfig", httpServer.serverConfig))
	serverConfig := httpServer.serverConfig
	engine := httpServer.engine
	if serverConfig.Locations != nil {
		for _, dir := range serverConfig.Locations {
			log.Info("Static Files Directory", zap.String("dir", dir))
		}
		engine.NoRoute(func(context *gin.Context) {
			_path_ := context.Request.URL.Path
			info, err := httpServer.memFileSystem.Stat(_path_)
			if info != nil && err == nil {
				if info.IsDir() {
					indexPage := filepath.Join(_path_, "index.html")
					exists, err := httpServer.memFileSystem.Exists(indexPage)
					if exists && err == nil {
						context.FileFromFS(_path_, httpServer.memFileSystem)
						return
					}
				} else {
					context.FileFromFS(_path_, httpServer.memFileSystem)
					return
				}
			}

			accepted := context.Request.Header.Get("Accept")
			if strings.Contains(accepted, "html") && !util.IsImagePath(_path_) {
				exists, err := httpServer.memFileSystem.Exists(serverConfig.Page404)
				if err != nil {
					log.Error("File not found", zap.String("file", serverConfig.Page404))
					return
				}
				if exists {
					context.FileFromFS(serverConfig.Page404, httpServer.memFileSystem)
				}
			}
		})

	}
	if httpServer.serverConfig.SSLEnabled() {
		return httpServer.startTLS(ctx)
	}

	var engine2 http.Handler = engine
	if httpServer.certManager.HasTLS() && (httpServer.serverConfig.Port == 80 || httpServer.serverConfig.Port == 443) {
		certManager, err := httpServer.certManager.GetCertManager()
		if err != nil {
			return err
		}
		engine2 = certManager.HTTPHandler(engine2)
	}

	httpServer.httpServer = &http.Server{
		Addr:              ":" + strconv.Itoa(httpServer.serverConfig.Port),
		Handler:           engine2,
		ReadHeaderTimeout: MaxReadHeaderTimeout,
		MaxHeaderBytes:    MaxHeaderBytes,
		ReadTimeout:       MaxReadTimeout,
	}

	go func() {
		<-ctx.Done()
		err := httpServer.httpServer.Shutdown(ctx)
		if err != nil {
			log.Error("stop the service", zap.Error(err))
		}
	}()

	log.Info("Start the service：", zap.String("address", "http://127.0.0.1:"+strconv.Itoa(httpServer.serverConfig.Port)))
	return errors.WithStackIf(httpServer.httpServer.ListenAndServe())
}

func (httpServer *HttpServer) startTLS(ctx context.Context) error {
	ssl := httpServer.serverConfig.SSL

	// If local certificate files are configured, use them directly without auto-certification
	if ssl.HasLocalCert() {
		return httpServer.startTLSWithLocalCert(ctx)
	}

	certManager, err := httpServer.certManager.GetCertManager()
	if err != nil {
		return err
	}
	var engine http.Handler = httpServer.engine
	if httpServer.serverConfig.Port == 80 || httpServer.serverConfig.Port == 443 {
		engine = certManager.HTTPHandler(engine)
	}

	httpServer.httpServer = &http.Server{
		Addr:              ":" + strconv.Itoa(httpServer.serverConfig.Port),
		Handler:           engine,
		ReadHeaderTimeout: MaxReadHeaderTimeout,
		MaxHeaderBytes:    MaxHeaderBytes,
		ReadTimeout:       MaxReadTimeout,
		//TLSConfig:         certManager.TLSConfig(),
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
			NextProtos:     []string{http2.NextProtoTLS, "http/1.1"},
			MinVersion:     tls.VersionTLS12,
		},
	}
	go func() {
		<-ctx.Done()
		err := httpServer.httpServer.Shutdown(ctx)
		if err != nil {
			log.Error("stop the service", zap.Error(err))
		}
	}()
	for _, host := range httpServer.serverConfig.SSL.Hosts {
		log.Info("Start the service：", zap.String("address", "https://"+host+":"+strconv.Itoa(httpServer.serverConfig.Port)))
	}
	if httpServer.serverConfig.Port == 443 {
		httpServer.httpServer.TLSConfig = certManager.TLSConfig()
		listener := certManager.Listener()
		err := httpServer.httpServer.Serve(listener)
		if err != nil {
			return errors.WithStackIf(err)
		}
		return nil
	}
	return errors.WithStackIf(httpServer.httpServer.ListenAndServeTLS("", ""))
}

// startTLSWithLocalCert starts HTTPS using locally provided certificate files.
// It loads all configured local certificates into a map keyed by host name,
// and optionally falls back to Let's Encrypt autocert for unconfigured hosts.
func (httpServer *HttpServer) startTLSWithLocalCert(ctx context.Context) error {
	ssl := httpServer.serverConfig.SSL

	// Build certEntry map keyed by host for on-demand reloading
	certMap := make(map[string]*certEntry)
	var defaultEntry *certEntry
	for _, c := range ssl.Certs {
		if c.CertFile == "" || c.KeyFile == "" {
			continue
		}
		if c.Host == "" {
			log.Warn("Skipping local cert with empty Host", zap.String("cert", c.CertFile), zap.String("key", c.KeyFile))
			continue
		}
		host := strings.ToLower(c.Host)
		if _, exists := certMap[host]; exists {
			log.Warn("Duplicate host in local certificate config, overwriting", zap.String("host", host))
		}
		entry := &certEntry{
			host:     host,
			certFile: c.CertFile,
			keyFile:  c.KeyFile,
		}
		// Load the certificate eagerly on startup to fail fast on errors
		if _, err := entry.get(); err != nil {
			return errors.Wrapf(err, "failed to load certificate: host=%s, cert=%s, key=%s", c.Host, c.CertFile, c.KeyFile)
		}
		log.Info("Loaded local certificate", zap.String("host", host), zap.String("cert", c.CertFile), zap.String("key", c.KeyFile))
		certMap[host] = entry
		if defaultEntry == nil {
			defaultEntry = entry
		}
	}

	// If Hosts (auto-cert) are also configured, create autocert manager as fallback
	var autocertManager *autocert.Manager
	if len(ssl.Hosts) > 0 {
		var err error
		autocertManager, err = httpServer.certManager.GetCertManager()
		if err != nil {
			log.Warn("Failed to init autocert manager, only local certs available", zap.Error(err))
		}
	}

	httpServer.httpServer = &http.Server{
		Addr:              ":" + strconv.Itoa(httpServer.serverConfig.Port),
		Handler:           httpServer.engine,
		ReadHeaderTimeout: MaxReadHeaderTimeout,
		MaxHeaderBytes:    MaxHeaderBytes,
		ReadTimeout:       MaxReadTimeout,
		TLSConfig: &tls.Config{
			NextProtos: []string{http2.NextProtoTLS, "http/1.1"},
			MinVersion: tls.VersionTLS12,
			GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
				serverName := strings.ToLower(info.ServerName)
				// Prefer local certificate match (reloads automatically if files changed)
				if entry, ok := certMap[serverName]; ok {
					return entry.get()
				}
				// Try autocert fallback
				if autocertManager != nil {
					if c, err := autocertManager.GetCertificate(info); err == nil && c != nil {
						return c, nil
					}
				}
				// Fall back to the first local certificate as default (reloads automatically if files changed)
				if defaultEntry != nil {
					return defaultEntry.get()
				}
				return nil, errors.Errorf("no certificate found for host: %s", serverName)
			},
		},
	}
	go func() {
		<-ctx.Done()
		err := httpServer.httpServer.Shutdown(ctx)
		if err != nil {
			log.Error("stop the service", zap.Error(err))
		}
	}()
	for _, c := range ssl.Certs {
		if c.Host == "" {
			continue
		}
		log.Info("Start the service:", zap.String("address", "https://"+c.Host+":"+strconv.Itoa(httpServer.serverConfig.Port)))
	}
	for _, host := range ssl.Hosts {
		log.Info("Start the service:", zap.String("address", "https://"+host+":"+strconv.Itoa(httpServer.serverConfig.Port)))
	}
	return errors.WithStackIf(httpServer.httpServer.ListenAndServeTLS("", ""))
}

// Close immediately closes the underlying HTTP server.
func (httpServer *HttpServer) Close() error {
	if httpServer.httpServer == nil {
		return nil
	}
	return httpServer.httpServer.Close()
}


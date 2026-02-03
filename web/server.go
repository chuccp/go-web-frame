package web

import (
	"context"
	"crypto/tls"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/util"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sourcegraph/conc/pool"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/net/http2"
)

const MaxHeaderBytes = 8192

const MaxReadHeaderTimeout = time.Second * 30

const MaxReadTimeout = time.Minute * 10

type SSLConfig struct {
	Enabled bool
	Hosts   []string
}
type ServerConfig struct {
	Port      int
	Locations []string
	Page404   string
	SSL       *SSLConfig
}

const ServerConfigKey = "web.server"

func (s *ServerConfig) SSLEnabled() bool {
	return s.SSL != nil && s.SSL.Enabled
}

func DefaultServerConfig() *ServerConfig {

	return &ServerConfig{
		Port: 19009,
		SSL: &SSLConfig{
			Enabled: false,
		},
	}
}

type HttpServer struct {
	httpServer    *http.Server
	engine        *gin.Engine
	serverConfig  *ServerConfig
	certManager   *CertManager
	memFileSystem *MemFileSystem
}

func defaultEngine() *gin.Engine {
	engine := gin.Default()
	config := cors.DefaultConfig()
	config.AllowAllOrigins = false
	config.AllowCredentials = true
	config.AllowOriginFunc = func(origin string) bool {
		return true
	}
	engine.Use(cors.New(config))
	return engine
}

func NewHttpServer(serverConfig *ServerConfig, certManager *CertManager) *HttpServer {
	engine := defaultEngine()
	return &HttpServer{
		engine:        engine,
		serverConfig:  serverConfig,
		certManager:   certManager,
		memFileSystem: DefaultMemFileSystem(serverConfig),
	}
}
func (httpServer *HttpServer) Port() int {
	return httpServer.serverConfig.Port
}

func (httpServer *HttpServer) Handle(handlerConfig *HandlerConfig) {
	for httpMethod, routeInfo := range handlerConfig.handles.RouteTree() {
		for _, handlerInfo := range routeInfo {
			httpServer.engine.Handle(httpMethod, handlerInfo.RelativePath(), httpServer.ToGinHandlerFunc(handlerConfig, handlerInfo.handlers...)...)
		}
	}
}
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
		request := NewRequest(ctx, resp, handlerConfig.HandlerMeta(ctx.Request.Method, ctx.FullPath()))
		mock := newMockFilterChain(request, handlerConfig.converter, handlerConfig.filters, &lastFilter{handler})
		mock.Converter()
	}
	return handlerFunc
}

func (httpServer *HttpServer) Run(ctx context.Context) error {
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

func (httpServer *HttpServer) Close() error {
	if httpServer.httpServer == nil {
		return nil
	}
	return httpServer.httpServer.Close()
}

type CertManager struct {
	certManager *autocert.Manager
	hosts       []string
	port        []int
	lock        *sync.RWMutex
}

func NewCertManager() *CertManager {
	return &CertManager{
		hosts: []string{},
		port:  []int{},
		lock:  new(sync.RWMutex),
	}
}
func (cm *CertManager) HasTLS() bool {
	return len(cm.hosts) > 0
}
func (cm *CertManager) AddHost(host string) {
	if strings.Contains(host, ":") {
		host = host[:strings.Index(host, ":")]
	}
	host = strings.ToLower(strings.TrimSpace(host))
	if util.IsDomain(host) {
		if util.EqualsAnyIgnoreCase(host, cm.hosts...) {
			return
		}
		cm.hosts = append(cm.hosts, host)
	}
}
func (cm *CertManager) AddPort(port int) {
	if port > 0 {
		if util.ArrayIntContains(cm.port, port) {
			return
		}
		cm.port = append(cm.port, port)
	}
}
func (cm *CertManager) GetCertManager() (*autocert.Manager, error) {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	if len(cm.hosts) == 0 {
		return &autocert.Manager{}, nil
	}
	if cm.certManager != nil {
		return cm.certManager, nil
	}
	certsPath := "certs"
	err := util.CreateDirIfNoExists(certsPath)
	if err != nil {
		return nil, err
	}
	m := &autocert.Manager{
		Prompt: autocert.AcceptTOS,
		// 缓存证书的路径
		Cache: autocert.DirCache(certsPath),
		// 需要自动获取证书的域名
		HostPolicy: autocert.HostWhitelist(cm.hosts...),
	}
	cm.certManager = m
	return m, nil
}
func (cm *CertManager) Run(ctx context.Context) error {

	if len(cm.hosts) > 0 && (!util.ArrayIntContains(cm.port, 80) || !util.ArrayIntContains(cm.port, 443)) {
		var wg = pool.New()
		errorsPool := wg.WithContext(ctx).WithFirstError()
		if !util.ArrayIntContains(cm.port, 80) {
			errorsPool.Go(func(ctx context.Context) error {
				manager, err := cm.GetCertManager()
				if err != nil {
					log.Errors("Failed to obtain certificate management：", err)
					return err
				}
				server := http.Server{
					Addr:    ":80",
					Handler: manager.HTTPHandler(nil),
				}

				go func() {
					<-ctx.Done()
					err = server.Shutdown(ctx)
					if err != nil {
						log.Errors("Shutdown service on port 80", err)
					}
				}()
				err = server.ListenAndServe()
				if err != nil {
					log.Errors("Failed to start the certificate service on port 80", err)
				}
				return err

			})
		}
		if !util.ArrayIntContains(cm.port, 443) {

			errorsPool.Go(func(ctx context.Context) error {

				manager, err := cm.GetCertManager()
				if err != nil {
					log.Errors("证书获取管理失败：", err)
					return err
				}
				server := http.Server{
					Addr:    ":443",
					Handler: manager.HTTPHandler(nil),
				}
				go func() {
					<-ctx.Done()
					err = server.Shutdown(ctx)
					if err != nil {
						log.Errors("Shutdown service on port 80", err)
					}
				}()
				err = server.ListenAndServe()
				if err != nil {
					log.Errors("Failed to start the certificate service on port 443", err)
				}
				return err
			})
		}
		return errors.WithStackIf(errorsPool.Wait())
	}
	return nil
}

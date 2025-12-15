package web

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const MaxHeaderBytes = 8192

const MaxReadHeaderTimeout = time.Second * 30

const MaxReadTimeout = time.Minute * 10

type HttpServer struct {
	httpServer *http.Server
	engine     *gin.Engine
	port       int
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

func NewHttpServer(port int) *HttpServer {
	return &HttpServer{
		engine: defaultEngine(),
		port:   port,
	}
}
func (httpServer *HttpServer) Port() int {
	return httpServer.port
}
func (httpServer *HttpServer) GET(relativePath string, handlers ...gin.HandlerFunc) {
	httpServer.engine.GET(relativePath, handlers...)
}
func (httpServer *HttpServer) POST(relativePath string, handlers ...gin.HandlerFunc) {
	httpServer.engine.POST(relativePath, handlers...)
}

func (httpServer *HttpServer) Run() error {
	httpServer.httpServer = &http.Server{
		Addr:              ":" + strconv.Itoa(httpServer.port),
		Handler:           httpServer.engine,
		ReadHeaderTimeout: MaxReadHeaderTimeout,
		MaxHeaderBytes:    MaxHeaderBytes,
		ReadTimeout:       MaxReadTimeout,
	}
	return httpServer.httpServer.ListenAndServe()
}
func (httpServer *HttpServer) Close() error {
	if httpServer.httpServer == nil {
		return nil
	}
	return httpServer.httpServer.Close()
}

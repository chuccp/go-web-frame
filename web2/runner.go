package web2

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"strconv"

	"github.com/chuccp/go-web-frame/log"
	"github.com/sourcegraph/conc/pool"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
)

// ---- ServerRunner ----

type ServerRunner struct {
	servers    []*Server
	serversMap map[int]*Server
	certs      *certStore
	ctx        context.Context
}

func newServerRunner(ctx context.Context, certsPath string, servers []*Server) *ServerRunner {
	sr := &ServerRunner{
		servers:    servers,
		serversMap: make(map[int]*Server),
		certs:      newCertStore(certsPath),
		ctx:        ctx,
	}
	for _, server := range servers {
		sr.serversMap[server.serverConfig.Port] = server
	}
	return sr
}

func (sr *ServerRunner) Start() error {
	if err := sr.certs.init(sr.servers); err != nil {
		return err
	}

	errorPool := pool.New().WithContext(sr.ctx).WithFirstError()

	if sr.certs.hasAutoCert() {
		if _, ok := sr.serversMap[80]; !ok {
			errorPool.Go(func(ctx context.Context) error {
				return sr.startHTTPChallengeServer(ctx)
			})
		}
		if _, ok := sr.serversMap[443]; !ok {
			errorPool.Go(func(ctx context.Context) error {
				return sr.startTLSChallengeServer(ctx)
			})
		}
	}
	for _, server := range sr.servers {
		errorPool.Go(func(_ context.Context) error {
			return sr.startServer(server)
		})
	}
	return errorPool.Wait()
}

func (sr *ServerRunner) startServer(server *Server) error {
	server.initRoute()
	if server.isTls() {
		return sr.listenTLS(server)
	}
	return sr.listen(server)
}

func (sr *ServerRunner) listen(server *Server) error {
	var engine http.Handler = server.engine

	if sr.certs.hasAutoCert() {
		if server.serverConfig.Port == 80 {
			engine = sr.certs.autoCertManager.HTTPHandler(engine)
		}
	}
	addr := ":" + strconv.Itoa(server.serverConfig.Port)
	httpServer := &http.Server{
		BaseContext: func(listener net.Listener) context.Context {
			return server.ctx
		},
		Addr:              addr,
		Handler:           engine,
		ReadHeaderTimeout: MaxReadHeaderTimeout,
		MaxHeaderBytes:    MaxHeaderBytes,
		ReadTimeout:       MaxReadTimeout,
	}
	log.Info("server listening", zap.String("url", "http://localhost"+addr))
	return httpServer.ListenAndServe()
}

func (sr *ServerRunner) listenTLS(server *Server) error {
	var engine http.Handler = server.engine
	if sr.certs.hasAutoCert() {
		if server.serverConfig.Port == 443 {
			engine = sr.certs.autoCertManager.HTTPHandler(engine)
		}
	}
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		NextProtos: []string{http2.NextProtoTLS, "http/1.1"},
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			return sr.certs.getCertificate(info.ServerName)
		},
	}
	if server.isAuto() {
		tlsConfig.GetCertificate = sr.certs.autoCertManager.GetCertificate
	}
	addr := ":" + strconv.Itoa(server.serverConfig.Port)
	httpServer := &http.Server{
		BaseContext: func(listener net.Listener) context.Context {
			return server.ctx
		},
		Addr:              addr,
		Handler:           engine,
		TLSConfig:         tlsConfig,
		ReadHeaderTimeout: MaxReadHeaderTimeout,
		MaxHeaderBytes:    MaxHeaderBytes,
		ReadTimeout:       MaxReadTimeout,
	}
	sr.logTLSListen(server, addr)
	return httpServer.ListenAndServeTLS("", "")
}

func (sr *ServerRunner) logTLSListen(server *Server, addr string) {
	if server.isAuto() {
		log.Info("server listening (auto-cert)",
			zap.Strings("hosts", server.serverConfig.SSL.Hosts),
			zap.String("url", "https://"+server.serverConfig.SSL.Hosts[0]+addr),
		)
		return
	}
	if domain := sr.certs.matchingDomain(); domain != "" {
		log.Info("server listening",
			zap.String("url", "https://"+domain+addr),
		)
		return
	}
	log.Info("server listening", zap.String("url", "https://localhost"+addr))
}

func (sr *ServerRunner) startHTTPChallengeServer(ctx context.Context) error {
	server := &http.Server{
		Addr:    ":80",
		Handler: sr.certs.autoCertManager.HTTPHandler(nil),
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		},
	}
	log.Info("starting ACME HTTP-01 challenge server on :80")
	if err := server.ListenAndServe(); err != nil {
		log.Error("ACME HTTP-01 challenge server error", zap.Error(err))
		return err
	}
	return nil
}

func (sr *ServerRunner) startTLSChallengeServer(ctx context.Context) error {
	tlsConfig := sr.certs.autoCertManager.TLSConfig()
	tlsConfig.MinVersion = tls.VersionTLS12
	tlsConfig.NextProtos = []string{http2.NextProtoTLS, "http/1.1"}

	server := &http.Server{
		Addr:      ":443",
		TLSConfig: tlsConfig,
		Handler:   sr.certs.autoCertManager.HTTPHandler(nil),
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		},
	}
	log.Info("starting ACME TLS-ALPN-01 challenge + auto-cert HTTPS server on :443")
	if err := server.ListenAndServeTLS("", ""); err != nil {
		log.Error("ACME TLS challenge server error", zap.Error(err))
		return err
	}
	return nil
}

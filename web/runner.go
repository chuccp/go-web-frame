// Package web: server runner that manages HTTP/TLS listeners and auto-cert.
package web

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"

	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/log"
	"github.com/sourcegraph/conc/pool"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
)

// ---- ServerRunner ----

// ServerRunner manages HTTP and TLS listeners with auto-certification support.
type ServerRunner struct {
	servers      []*Server
	serversMap   map[int]*Server
	certs        *certStore
	serverConfig *ServerConfig
	ctx          context.Context
}

func newServerRunner(ctx context.Context, certsPath string, servers []*Server) *ServerRunner {
	sr := &ServerRunner{
		servers:    servers,
		serversMap: make(map[int]*Server),
		certs:      newCertStore(certsPath),
		ctx:        ctx,
	}
	if len(servers) > 0 {
		sr.serverConfig = servers[0].serverConfig
	}
	for _, server := range servers {
		sr.serversMap[server.serverConfig.Port] = server
	}
	return sr
}

func (sr *ServerRunner) Start() error {
	if err := sr.certs.init(sr.servers); err != nil {
		log.PrintPanic(err)
		log.Error("certs init", zap.Error(err))
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
		errorPool.Go(func(ctx context.Context) error {
			return sr.startServer(ctx, server)
		})
	}
	return errorPool.Wait()
}

func (sr *ServerRunner) startServer(ctx context.Context, server *Server) error {
	if server.isTls() {
		return server.ListenTLS(ctx, sr.certs)
	}
	return server.Listen(ctx, sr.certs)
}

func (sr *ServerRunner) logTLSListen(server *Server, addr string) {

	for _, host := range server.serverConfig.SSL.Hosts {
		if server.isAuto(host) {
			log.Info("server listening (auto-cert)",
				zap.Strings("hosts", server.serverConfig.SSL.Hosts),
				zap.String("url", "https://"+host+addr),
			)
		} else {
			log.Info("server listening",
				zap.String("url", "https://"+host+addr),
			)
		}
	}
	if domains := sr.certs.matchingDomains(); len(domains) > 0 {
		for _, domain := range domains {
			log.Info("server listening",
				zap.String("url", "https://"+domain+addr),
			)
		}
		return
	}
	log.Info("server listening", zap.String("url", "https://localhost"+addr))
}

func (sr *ServerRunner) startHTTPChallengeServer(ctx context.Context) error {
	server := &http.Server{
		Addr:              ":80",
		Handler:           sr.certs.autoCertManager.HTTPHandler(nil),
		ReadHeaderTimeout: sr.serverConfig.GetReadHeaderTimeout(),
		MaxHeaderBytes:    sr.serverConfig.GetMaxHeaderBytes(),
		ReadTimeout:       sr.serverConfig.GetReadTimeout(),
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		},
	}
	go func() {
		<-sr.ctx.Done()
		if err := server.Shutdown(ctx); err != nil {
			log.Error("Failed to shutdown ACME HTTP-01 challenge server", zap.Error(err))
		}
	}()
	log.Info("starting ACME HTTP-01 challenge server on :80")
	if err := errors.WithStackIf(server.ListenAndServe()); err != nil {
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
		Addr:              ":443",
		TLSConfig:         tlsConfig,
		ReadHeaderTimeout: sr.serverConfig.GetReadHeaderTimeout(),
		MaxHeaderBytes:    sr.serverConfig.GetMaxHeaderBytes(),
		ReadTimeout:       sr.serverConfig.GetReadTimeout(),
		Handler:           sr.certs.autoCertManager.HTTPHandler(nil),
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		},
	}
	go func() {
		<-sr.ctx.Done()
		if err := server.Shutdown(ctx); err != nil {
			log.Error("Failed to shutdown ACME TLS challenge server", zap.Error(err))
		}
	}()
	log.Info("starting ACME TLS-ALPN-01 challenge + auto-cert HTTPS server on :443")
	if err := errors.WithStackIf(server.ListenAndServeTLS("", "")); err != nil {
		log.Error("ACME TLS challenge server error", zap.Error(err))
		return err
	}
	return nil
}

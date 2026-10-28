// Package web: server runner that manages HTTP/TLS listeners and auto-cert.
package web

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"strconv"
	"strings"

	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/log"
	"github.com/sourcegraph/conc/pool"
	"go.uber.org/zap"
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

	if target, ok := sr.httpRedirectTarget(); ok {
		if _, occupied := sr.serversMap[80]; occupied {
			log.Warn("bind_http ignored because port 80 has an explicit server")
		} else {
			errorPool.Go(func(ctx context.Context) error {
				return sr.startHTTPRedirectServer(ctx, target)
			})
		}
	}
	if sr.certs.hasAutoCert() {
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

func (sr *ServerRunner) httpRedirectTarget() (*Server, bool) {
	for _, server := range sr.servers {
		if server.isTls() && server.serverConfig.SSL.BindHTTP {
			return server, true
		}
	}
	return nil, false
}

func (sr *ServerRunner) startHTTPRedirectServer(ctx context.Context, target *Server) error {
	server := &http.Server{
		Addr:              ":80",
		Handler:           httpRedirectHandler(target.serverConfig.Port),
		ReadHeaderTimeout: target.serverConfig.GetReadHeaderTimeout(),
		MaxHeaderBytes:    target.serverConfig.GetMaxHeaderBytes(),
		ReadTimeout:       target.serverConfig.GetReadTimeout(),
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		},
	}
	go func() {
		<-ctx.Done()
		if err := shutdownServer(ctx, server); err != nil {
			log.Error("Failed to shutdown HTTP redirect server", zap.Error(err))
		}
	}()
	log.Info("starting HTTP-to-HTTPS redirect server on :80")
	if err := errors.WithStackIf(server.ListenAndServe()); err != nil {
		log.Error("HTTP redirect server error", zap.Error(err))
		return err
	}
	return nil
}

func httpRedirectHandler(httpsPort int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		if parsedHost, _, err := net.SplitHostPort(host); err == nil {
			host = parsedHost
		}
		host = strings.TrimSpace(host)
		if host == "" {
			host = "localhost"
		}
		if httpsPort != 443 {
			host = net.JoinHostPort(host, strconv.Itoa(httpsPort))
		} else if ip := net.ParseIP(host); ip != nil && strings.Contains(host, ":") {
			host = "[" + host + "]"
		}
		http.Redirect(w, r, "https://"+host+r.URL.RequestURI(), http.StatusPermanentRedirect)
	})
}

func (sr *ServerRunner) startTLSChallengeServer(ctx context.Context) error {
	tlsConfig := sr.certs.autoCertManager.TLSConfig()
	tlsConfig.MinVersion = tls.VersionTLS12

	server := &http.Server{
		Addr:              ":443",
		TLSConfig:         tlsConfig,
		ReadHeaderTimeout: sr.serverConfig.GetReadHeaderTimeout(),
		MaxHeaderBytes:    sr.serverConfig.GetMaxHeaderBytes(),
		ReadTimeout:       sr.serverConfig.GetReadTimeout(),
		Handler:           http.NotFoundHandler(),
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		},
	}
	go func() {
		<-sr.ctx.Done()
		if err := shutdownServer(sr.ctx, server); err != nil {
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

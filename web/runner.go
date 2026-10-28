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
	if err := sr.validateChallengePorts(); err != nil {
		return errors.WithStackIf(err)
	}

	errorPool := pool.New().WithContext(sr.ctx).WithFirstError()

	httpHandler, needsHTTP80 := sr.http80Handler()
	if needsHTTP80 {
		if httpServer, occupied := sr.serversMap[80]; occupied {
			if sr.certs.hasAutoCertForMode(ACMEChallengeHTTP01) {
				httpServer.httpHandler = sr.certs.httpChallengeHandler(httpServer.engine)
			}
			log.Warn("automatic HTTP redirect binding ignored because port 80 has an explicit server")
		} else {
			errorPool.Go(func(ctx context.Context) error {
				return sr.startHTTP80Server(ctx, httpHandler)
			})
		}
	}
	if sr.certs.hasAutoCertForMode(ACMEChallengeTLSALPN01) {
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

func (sr *ServerRunner) validateChallengePorts() error {
	if sr.certs.hasAutoCertForMode(ACMEChallengeHTTP01) {
		if server, ok := sr.serversMap[80]; ok && server.isTls() {
			return errors.New("HTTP-01 requires port 80 to serve plain HTTP, but port 80 is configured for TLS")
		}
	}
	if sr.certs.hasAutoCertForMode(ACMEChallengeTLSALPN01) {
		if server, ok := sr.serversMap[443]; ok && !server.isTls() {
			return errors.New("TLS-ALPN-01 requires port 443 to serve TLS, but port 443 is configured for plain HTTP")
		}
	}
	return nil
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

func (sr *ServerRunner) http80Handler() (http.Handler, bool) {
	httpChallenge := sr.certs.hasAutoCertForMode(ACMEChallengeHTTP01)
	defaultHTTPSPort, hasDefaultHTTPSPort := sr.defaultHTTPSPort()
	if !httpChallenge && !hasDefaultHTTPSPort {
		return nil, false
	}

	challengeHandler := sr.certs.httpChallengeHandler(http.NotFoundHandler())
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := normalizeRequestHost(r.Host)
		route, configured := sr.certs.hostRoute(host)
		if strings.HasPrefix(r.URL.Path, "/.well-known/acme-challenge/") {
			if configured && route.challenge == ACMEChallengeHTTP01 {
				challengeHandler.ServeHTTP(w, r)
				return
			}
			http.NotFound(w, r)
			return
		}
		if configured && route.bindHTTP {
			writeHTTPSRedirect(w, r, host, route.httpsPort)
			return
		}
		if !configured && hasDefaultHTTPSPort {
			writeHTTPSRedirect(w, r, host, defaultHTTPSPort)
			return
		}
		http.NotFound(w, r)
	}), true
}

func (sr *ServerRunner) defaultHTTPSPort() (int, bool) {
	port := 0
	for _, server := range sr.servers {
		if !server.isTls() || !server.serverConfig.SSL.BindHTTP {
			continue
		}
		if port != 0 && port != server.serverConfig.Port {
			return 0, false
		}
		port = server.serverConfig.Port
	}
	return port, port != 0
}

func (sr *ServerRunner) startHTTP80Server(ctx context.Context, handler http.Handler) error {
	server := &http.Server{
		Addr:              ":80",
		Handler:           handler,
		ReadHeaderTimeout: sr.serverConfig.GetReadHeaderTimeout(),
		MaxHeaderBytes:    sr.serverConfig.GetMaxHeaderBytes(),
		ReadTimeout:       sr.serverConfig.GetReadTimeout(),
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		},
	}
	go func() {
		<-ctx.Done()
		if err := shutdownServer(ctx, server); err != nil {
			log.Error("Failed to shutdown HTTP port 80 server", zap.Error(err))
		}
	}()
	log.Info("starting HTTP port 80 server for ACME and/or HTTPS redirects")
	if err := errors.WithStackIf(server.ListenAndServe()); err != nil {
		log.Error("HTTP port 80 server error", zap.Error(err))
		return err
	}
	return nil
}

func httpRedirectHandler(httpsPort int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeHTTPSRedirect(w, r, normalizeRequestHost(r.Host), httpsPort)
	})
}

func normalizeRequestHost(host string) string {
	if parsedHost, _, err := net.SplitHostPort(host); err == nil {
		host = parsedHost
	}
	host = strings.TrimSpace(host)
	if host == "" {
		return "localhost"
	}
	return host
}

func writeHTTPSRedirect(w http.ResponseWriter, r *http.Request, host string, httpsPort int) {
	if httpsPort != 443 {
		host = net.JoinHostPort(host, strconv.Itoa(httpsPort))
	} else if ip := net.ParseIP(host); ip != nil && strings.Contains(host, ":") {
		host = "[" + host + "]"
	}
	http.Redirect(w, r, "https://"+host+r.URL.RequestURI(), http.StatusPermanentRedirect)
}

func (sr *ServerRunner) startTLSChallengeServer(ctx context.Context) error {
	manager := sr.certs.autoCertManager(ACMEChallengeTLSALPN01)
	if manager == nil {
		return errors.New("TLS-ALPN-01 manager is not configured")
	}
	tlsConfig := manager.TLSConfig()
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

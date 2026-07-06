package web2

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/chuccp/go-web-frame/log"
	"github.com/sourcegraph/conc/pool"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"
)

const MaxHeaderBytes = 8192
const MaxReadHeaderTimeout = time.Second * 30
const MaxReadTimeout = time.Minute * 10

type localCert struct {
	cert    *tls.Certificate
	domains []string
}

type CertServer struct {
	servers         []*Server
	serversMap      map[int]*Server
	autoCertManager *autocert.Manager
	autoHosts       []string
	localCerts      []*localCert
	certMap         map[string]*tls.Certificate
	certsPath       string
	ctx             context.Context
}

func newCertServer(ctx context.Context, certsPath string, servers []*Server) *CertServer {
	cs := &CertServer{
		servers:    servers,
		serversMap: make(map[int]*Server),
		localCerts: make([]*localCert, 0),
		certMap:    make(map[string]*tls.Certificate),
		certsPath:  certsPath,
		ctx:        ctx,
	}
	for _, server := range servers {
		cs.serversMap[server.serverConfig.Port] = server
	}
	return cs
}

func (cs *CertServer) parseCert(certFile, keyFile string) (*localCert, error) {
	certPEM, err := os.ReadFile(certFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file %s: %w", certFile, err)
	}
	keyPEM, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file %s: %w", keyFile, err)
	}

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate/key pair (%s, %s): %w", certFile, keyFile, err)
	}

	domains := extractDomains(&tlsCert)
	if len(domains) == 0 {
		return nil, fmt.Errorf("no domain names found in certificate %s", certFile)
	}

	return &localCert{
		cert:    &tlsCert,
		domains: domains,
	}, nil
}

func (cs *CertServer) addLocalCert(lc *localCert) {
	cs.localCerts = append(cs.localCerts, lc)
	for _, domain := range lc.domains {
		cs.certMap[strings.ToLower(domain)] = lc.cert
		log.Debug("registered local certificate for domain", zap.String("domain", domain))
	}
}

func extractDomains(tlsCert *tls.Certificate) []string {
	domains := make([]string, 0)

	leaf, err := x509.ParseCertificate(tlsCert.Certificate[0])
	if err != nil {
		log.Error("failed to parse x509 certificate for domain extraction", zap.Error(err))
		return domains
	}

	if len(leaf.DNSNames) > 0 {
		domains = append(domains, leaf.DNSNames...)
	}

	if len(domains) == 0 && leaf.Subject.CommonName != "" {
		domains = append(domains, leaf.Subject.CommonName)
	}

	return domains
}

func (cs *CertServer) GetCertificate(host string) (*tls.Certificate, error) {
	host = strings.ToLower(host)

	if cert, ok := cs.certMap[host]; ok {
		return cert, nil
	}

	if cert := cs.matchWildcard(host); cert != nil {
		return cert, nil
	}

	if cs.autoCertManager == nil {
		return nil, nil
	}
	return cs.autoCertManager.GetCertificate(&tls.ClientHelloInfo{ServerName: host})
}

func (cs *CertServer) matchWildcard(host string) *tls.Certificate {
	for _, lc := range cs.localCerts {
		for _, domain := range lc.domains {
			if strings.HasPrefix(domain, "*.") {
				suffix := domain[1:]
				if strings.HasSuffix(host, suffix) {
					prefix := strings.TrimSuffix(host, suffix)
					if !strings.Contains(prefix, ".") {
						return lc.cert
					}
				}
			}
		}
	}
	return nil
}

func (cs *CertServer) initCert() error {
	hosts := make([]string, 0)
	for _, server := range cs.servers {
		if server.serverConfig.SSL != nil && server.serverConfig.SSL.Enabled {
			if len(server.serverConfig.SSL.Certs) > 0 {
				for _, certCfg := range server.serverConfig.SSL.Certs {
					lc, err := cs.parseCert(certCfg.CertFile, certCfg.KeyFile)
					if err != nil {
						return fmt.Errorf("server port %d: %w", server.serverConfig.Port, err)
					}
					cs.addLocalCert(lc)
				}
				continue
			}
			if len(server.serverConfig.SSL.Hosts) > 0 {
				hosts = append(hosts, server.serverConfig.SSL.Hosts...)
			}
		}
	}
	cs.autoHosts = hosts

	if len(hosts) > 0 {
		cs.autoCertManager = &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			Cache:      autocert.DirCache(cs.certsPath),
			HostPolicy: autocert.HostWhitelist(hosts...),
		}
	}

	return nil
}
func (cs *CertServer) startHTTPChallengeServer(ctx context.Context) error {
	server := &http.Server{
		Addr:    ":80",
		Handler: cs.autoCertManager.HTTPHandler(nil),
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

func (cs *CertServer) startTLSChallengeServer(ctx context.Context) error {
	tlsConfig := cs.autoCertManager.TLSConfig()
	tlsConfig.MinVersion = tls.VersionTLS12

	server := &http.Server{
		Addr:      ":443",
		TLSConfig: tlsConfig,
		Handler:   cs.autoCertManager.HTTPHandler(nil),
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

func (cs *CertServer) Start() error {
	if err := cs.initCert(); err != nil {
		return fmt.Errorf("certificate initialization failed: %w", err)
	}

	errorPool := pool.New().WithContext(cs.ctx).WithFirstError()

	if len(cs.autoHosts) > 0 {
		if _, ok := cs.serversMap[80]; !ok {
			errorPool.Go(func(ctx context.Context) error {
				return cs.startHTTPChallengeServer(ctx)
			})
		}
		if _, ok := cs.serversMap[443]; !ok {
			errorPool.Go(func(ctx context.Context) error {
				return cs.startTLSChallengeServer(ctx)
			})
		}
	}
	for _, server := range cs.servers {
		errorPool.Go(func(ctx context.Context) error {
			return cs.startServer(ctx, server)
		})
	}
	return errorPool.Wait()
}
func (cs *CertServer) startServer(ctx context.Context, server *Server) error {
	server.initRoute()
	if server.serverConfig.SSL != nil && server.serverConfig.SSL.Enabled {
		err := cs.listenTLS(ctx, server)
		if err != nil {
			return err
		}
	} else {
		err := cs.listen(ctx, server)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cs *CertServer) listen(ctx context.Context, server *Server) error {
	var engine http.Handler = server.engine

	if len(cs.autoHosts) > 0 {
		if server.serverConfig.Port == 80 {
			engine = cs.autoCertManager.HTTPHandler(engine)
		}
	}
	httpServer := &http.Server{
		BaseContext: func(listener net.Listener) context.Context {
			return server.ctx
		},
		Addr:              ":" + strconv.Itoa(server.serverConfig.Port),
		Handler:           engine,
		ReadHeaderTimeout: MaxReadHeaderTimeout,
		MaxHeaderBytes:    MaxHeaderBytes,
		ReadTimeout:       MaxReadTimeout,
	}
	return httpServer.ListenAndServe()
}

func (cs *CertServer) listenTLS(ctx context.Context, server *Server) error {
	var engine http.Handler = server.engine
	if len(cs.autoHosts) > 0 {
		if server.serverConfig.Port == 443 {
			engine = cs.autoCertManager.HTTPHandler(engine)
		}
	}
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			return cs.GetCertificate(info.ServerName)
		},
	}
	httpServer := &http.Server{
		BaseContext: func(listener net.Listener) context.Context {
			return server.ctx
		},
		Addr:              ":" + strconv.Itoa(server.serverConfig.Port),
		Handler:           engine,
		TLSConfig:         tlsConfig,
		ReadHeaderTimeout: MaxReadHeaderTimeout,
		MaxHeaderBytes:    MaxHeaderBytes,
		ReadTimeout:       MaxReadTimeout,
	}
	return httpServer.ListenAndServeTLS("", "")
}

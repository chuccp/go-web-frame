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
	"sync"
	"time"

	"github.com/chuccp/go-web-frame/log"
	"github.com/sourcegraph/conc/pool"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/net/http2"
)

const MaxHeaderBytes = 8192
const MaxReadHeaderTimeout = time.Second * 30
const MaxReadTimeout = time.Minute * 10

const certCheckInterval = 5 * time.Minute

type certEntry struct {
	host      string
	certFile  string
	keyFile   string
	domains   []string
	cert      *tls.Certificate
	certMod   time.Time
	keyMod    time.Time
	nextCheck time.Time
	mu        sync.RWMutex
}

func (e *certEntry) get() (*tls.Certificate, error) {
	e.mu.RLock()
	if e.cert != nil && time.Now().Before(e.nextCheck) {
		c := e.cert
		e.mu.RUnlock()
		return c, nil
	}
	e.mu.RUnlock()

	certStat, err := os.Stat(e.certFile)
	if err != nil {
		return nil, fmt.Errorf("stat cert file %s: %w", e.certFile, err)
	}
	keyStat, err := os.Stat(e.keyFile)
	if err != nil {
		return nil, fmt.Errorf("stat key file %s: %w", e.keyFile, err)
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if e.cert != nil && time.Now().Before(e.nextCheck) {
		return e.cert, nil
	}

	if e.cert != nil && certStat.ModTime().Equal(e.certMod) && keyStat.ModTime().Equal(e.keyMod) {
		e.nextCheck = time.Now().Add(certCheckInterval)
		return e.cert, nil
	}

	cert, err := tls.LoadX509KeyPair(e.certFile, e.keyFile)
	if err != nil {
		return nil, fmt.Errorf("reload certificate host=%s cert=%s key=%s: %w", e.host, e.certFile, e.keyFile, err)
	}
	e.cert = &cert
	e.certMod = certStat.ModTime()
	e.keyMod = keyStat.ModTime()
	e.nextCheck = time.Now().Add(certCheckInterval)

	log.Info("certificate reloaded from disk", zap.String("host", e.host), zap.String("cert", e.certFile))
	return e.cert, nil
}

type CertServer struct {
	servers         []*Server
	serversMap      map[int]*Server
	autoCertManager *autocert.Manager
	autoHosts       []string
	certEntries     []*certEntry
	certMap         map[string]*certEntry
	wildcardCache   map[string]*certEntry
	mu              sync.RWMutex
	certsPath       string
	ctx             context.Context
}

func newCertServer(ctx context.Context, certsPath string, servers []*Server) *CertServer {
	cs := &CertServer{
		servers:       servers,
		serversMap:    make(map[int]*Server),
		certEntries:   make([]*certEntry, 0),
		certMap:       make(map[string]*certEntry),
		wildcardCache: make(map[string]*certEntry),
		certsPath:     certsPath,
		ctx:           ctx,
	}
	for _, server := range servers {
		cs.serversMap[server.serverConfig.Port] = server
	}
	return cs
}

func (cs *CertServer) parseCert(certFile, keyFile string) (*certEntry, error) {
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

	entry := &certEntry{
		certFile: certFile,
		keyFile:  keyFile,
		domains:  domains,
		cert:     &tlsCert,
	}
	entry.nextCheck = time.Now().Add(certCheckInterval)
	return entry, nil
}

func (cs *CertServer) addCertEntry(entry *certEntry) {
	cs.certEntries = append(cs.certEntries, entry)
	for _, domain := range entry.domains {
		key := strings.ToLower(domain)
		cs.certMap[key] = entry
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

	if entry, ok := cs.certMap[host]; ok {
		return entry.get()
	}

	cs.mu.RLock()
	if entry, ok := cs.wildcardCache[host]; ok {
		cs.mu.RUnlock()
		return entry.get()
	}
	cs.mu.RUnlock()

	if entry := cs.matchWildcard(host); entry != nil {
		cs.mu.Lock()
		cs.wildcardCache[host] = entry
		cs.mu.Unlock()
		return entry.get()
	}

	if cs.autoCertManager == nil {
		return nil, nil
	}
	return cs.autoCertManager.GetCertificate(&tls.ClientHelloInfo{ServerName: host})
}

func (cs *CertServer) matchWildcard(host string) *certEntry {
	for _, entry := range cs.certEntries {
		for _, domain := range entry.domains {
			if strings.HasPrefix(domain, "*.") {
				suffix := domain[1:]
				if strings.HasSuffix(host, suffix) {
					prefix := strings.TrimSuffix(host, suffix)
					if !strings.Contains(prefix, ".") {
						return entry
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
					entry, err := cs.parseCert(certCfg.CertFile, certCfg.KeyFile)
					if err != nil {
						return fmt.Errorf("server port %d: %w", server.serverConfig.Port, err)
					}
					cs.addCertEntry(entry)
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
	tlsConfig.NextProtos = []string{http2.NextProtoTLS, "http/1.1"}

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
		errorPool.Go(func(_ context.Context) error {
			return cs.startServer(server)
		})
	}
	return errorPool.Wait()
}

func (cs *CertServer) startServer(server *Server) error {
	server.initRoute()
	if server.isTls() {
		return cs.listenTLS(server)
	}
	return cs.listen(server)
}

func (cs *CertServer) listen(server *Server) error {
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

func (cs *CertServer) listenTLS(server *Server) error {
	var engine http.Handler = server.engine
	if len(cs.autoHosts) > 0 {
		if server.serverConfig.Port == 443 {
			engine = cs.autoCertManager.HTTPHandler(engine)
		}
	}
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		NextProtos: []string{http2.NextProtoTLS, "http/1.1"},
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			return cs.GetCertificate(info.ServerName)
		},
	}
	if server.isAuto() {
		tlsConfig.GetCertificate = cs.autoCertManager.GetCertificate
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

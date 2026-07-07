// Package web: TLS certificate store and auto-certification.
package web

import (
	"crypto/tls"
	"crypto/x509"
	"os"
	"strings"
	"sync"
	"time"

	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/log"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"
)

// MaxHeaderBytes is the maximum size of request headers in bytes.
const MaxHeaderBytes = 8192

// MaxReadHeaderTimeout is the maximum duration for reading request headers.
const MaxReadHeaderTimeout = time.Second * 30

// MaxReadTimeout is the maximum duration for reading the entire request.
const MaxReadTimeout = time.Minute * 10

const certCheckInterval = 5 * time.Minute

// ---- certEntry ----

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
		return nil, errors.Wrapf(err, "stat cert file %s", e.certFile)
	}
	keyStat, err := os.Stat(e.keyFile)
	if err != nil {
		return nil, errors.Wrapf(err, "stat key file %s", e.keyFile)
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
		return nil, errors.Wrapf(err, "reload certificate host=%s cert=%s key=%s", e.host, e.certFile, e.keyFile)
	}
	e.cert = &cert
	e.certMod = certStat.ModTime()
	e.keyMod = keyStat.ModTime()
	e.nextCheck = time.Now().Add(certCheckInterval)

	log.Info("certificate reloaded from disk", zap.String("host", e.host), zap.String("cert", e.certFile))
	return e.cert, nil
}

// ---- certStore ----

type certStore struct {
	autoCertManager *autocert.Manager
	autoHosts       []string
	certEntries     []*certEntry
	certMap         map[string]*certEntry
	wildcardCache   map[string]*certEntry
	mu              sync.RWMutex
	certsPath       string
}

func newCertStore(certsPath string) *certStore {
	return &certStore{
		certsPath:     certsPath,
		certEntries:   make([]*certEntry, 0),
		certMap:       make(map[string]*certEntry),
		wildcardCache: make(map[string]*certEntry),
	}
}

func (cs *certStore) init(servers []*Server) error {
	hosts := make([]string, 0)
	for _, server := range servers {
		if server.isTls() {
			if !server.isAuto() {
				for _, certCfg := range server.serverConfig.SSL.Certs {
					entry, err := cs.parseCert(certCfg.CertFile, certCfg.KeyFile)
					if err != nil {
						return errors.Wrapf(err, "server port %d", server.serverConfig.Port)
					}
					cs.addCertEntry(entry)
				}
				continue
			} else {
				if len(server.serverConfig.SSL.Hosts) > 0 {
					hosts = append(hosts, server.serverConfig.SSL.Hosts...)
				}
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

func (cs *certStore) parseCert(certFile, keyFile string) (*certEntry, error) {
	certPEM, err := os.ReadFile(certFile)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read certificate file %s", certFile)
	}
	keyPEM, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read key file %s", keyFile)
	}

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse certificate/key pair (%s, %s)", certFile, keyFile)
	}

	domains := extractDomains(&tlsCert)
	if len(domains) == 0 {
		return nil, errors.Errorf("no domain names found in certificate %s", certFile)
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

func (cs *certStore) addCertEntry(entry *certEntry) {
	cs.certEntries = append(cs.certEntries, entry)
	for _, domain := range entry.domains {
		key := strings.ToLower(domain)
		cs.certMap[key] = entry
		log.Debug("registered local certificate for domain", zap.String("domain", domain))
	}
}

func (cs *certStore) getCertificate(host string) (*tls.Certificate, error) {
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

func (cs *certStore) matchWildcard(host string) *certEntry {
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

func (cs *certStore) hasAutoCert() bool {
	return len(cs.autoHosts) > 0
}

func (cs *certStore) matchingDomain() string {
	for _, entry := range cs.certEntries {
		for _, domain := range entry.domains {
			if !strings.HasPrefix(domain, "*.") {
				return domain
			}
		}
	}
	return ""
}

// ---- package helpers ----

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


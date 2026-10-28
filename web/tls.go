// Package web: TLS certificate store and auto-certification.
package web

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
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

// ---- certEntry interface ----

type certEntry interface {
	get() (*tls.Certificate, error)
	domains() []string
}

// ---- fileCertEntry: loaded from cert/key files on disk, auto-reloads on change ----

type fileCertEntry struct {
	certFile   string
	keyFile    string
	cert       *tls.Certificate
	domainList []string
	certMod    time.Time
	keyMod     time.Time
	nextCheck  time.Time
	mu         sync.RWMutex
}

func (e *fileCertEntry) domains() []string {
	return e.domainList
}

func (e *fileCertEntry) get() (*tls.Certificate, error) {
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
		return nil, errors.Wrapf(err, "reload certificate cert=%s key=%s", e.certFile, e.keyFile)
	}
	e.cert = &cert
	e.certMod = certStat.ModTime()
	e.keyMod = keyStat.ModTime()
	e.nextCheck = time.Now().Add(certCheckInterval)

	log.Info("certificate reloaded from disk", zap.String("cert", e.certFile))
	return e.cert, nil
}

// ---- selfCertEntry: in-memory self-signed certificate, no file I/O ----

type selfCertEntry struct {
	host      string
	cert      *tls.Certificate
	nextCheck time.Time
	mu        sync.RWMutex
}

func (e *selfCertEntry) domains() []string {
	return []string{e.host}
}

func (e *selfCertEntry) get() (*tls.Certificate, error) {
	e.mu.RLock()
	if e.cert != nil && time.Now().Before(e.nextCheck) {
		c := e.cert
		e.mu.RUnlock()
		return c, nil
	}
	e.mu.RUnlock()

	e.mu.Lock()
	defer e.mu.Unlock()

	// double-check after acquiring write lock
	if e.cert != nil && time.Now().Before(e.nextCheck) {
		return e.cert, nil
	}

	if e.cert != nil {
		e.nextCheck = time.Now().Add(24 * time.Hour)
		return e.cert, nil
	}

	// first access: generate self-signed certificate
	cert, err := e.generate()
	if err != nil {
		return nil, err
	}
	e.cert = cert
	e.nextCheck = time.Now().Add(24 * time.Hour)
	log.Info("generated self-signed certificate", zap.String("host", e.host))
	return e.cert, nil
}

// generate creates a new self-signed certificate for the host.
func (e *selfCertEntry) generate() (*tls.Certificate, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, errors.Wrapf(err, "generate self-signed key for %s", e.host)
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, errors.Wrapf(err, "generate serial number for %s", e.host)
	}

	template := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{Organization: []string{"go-web-frame"}, CommonName: e.host},
		NotBefore:    time.Now().Add(-1 * time.Hour),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	ipStr := strings.Trim(e.host, "[]")
	if ip := net.ParseIP(ipStr); ip != nil {
		template.IPAddresses = []net.IP{ip}
	} else {
		template.DNSNames = []string{e.host}
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return nil, errors.Wrapf(err, "create self-signed certificate for %s", e.host)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, errors.Wrapf(err, "marshal private key for %s", e.host)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, errors.Wrapf(err, "load self-signed key pair for %s", e.host)
	}
	return &tlsCert, nil
}

// ---- certStore ----

type certStore struct {
	autoCertManagers map[string]*autocert.Manager
	autoHostModes    map[string]string
	hostRoutes       map[string]acmeHostRoute
	certEntries      []certEntry
	certMap          map[string]certEntry
	wildcardCache    map[string]certEntry
	mu               sync.RWMutex
	certsPath        string
}

type acmeHostRoute struct {
	challenge string
	httpsPort int
	bindHTTP  bool
}

func newCertStore(certsPath string) *certStore {
	return &certStore{
		certsPath:        certsPath,
		autoCertManagers: make(map[string]*autocert.Manager),
		autoHostModes:    make(map[string]string),
		hostRoutes:       make(map[string]acmeHostRoute),
		certEntries:      make([]certEntry, 0),
		certMap:          make(map[string]certEntry),
		wildcardCache:    make(map[string]certEntry),
	}
}

func (server *Server) acmeChallengeMode() (string, error) {
	if server.serverConfig.SSL == nil {
		return "", errors.New("TLS configuration is nil")
	}
	mode := strings.ToLower(strings.TrimSpace(server.serverConfig.SSL.ACMEChallenge))
	switch mode {
	case "", ACMEChallengeAuto:
		if server.serverConfig.Port == 443 {
			return ACMEChallengeTLSALPN01, nil
		}
		return ACMEChallengeHTTP01, nil
	case ACMEChallengeTLSALPN01, ACMEChallengeHTTP01:
		return mode, nil
	default:
		return "", errors.Errorf(
			"invalid acme_challenge %q: want %s, %s, or %s",
			server.serverConfig.SSL.ACMEChallenge,
			ACMEChallengeAuto,
			ACMEChallengeTLSALPN01,
			ACMEChallengeHTTP01,
		)
	}
}

func (server *Server) hasAutoCertConfig() bool {
	return server.isTls() &&
		server.serverConfig.SSL != nil &&
		len(server.serverConfig.SSL.Certs) == 0 &&
		len(server.serverConfig.SSL.Hosts) > 0
}

func (cs *certStore) init(servers []*Server) error {
	hostsByMode := make(map[string][]string)
	errs := make([]error, 0)
	for _, server := range servers {
		if server.isTls() {
			route := acmeHostRoute{
				httpsPort: server.serverConfig.Port,
				bindHTTP:  server.serverConfig.SSL.BindHTTP,
			}
			var hasFile = false
			for _, certCfg := range server.serverConfig.SSL.Certs {
				entry, err := cs.parseCert(certCfg.CertFile, certCfg.KeyFile)
				if err != nil {
					errs = append(errs, err)
					continue
				}
				hasFile = true
				cs.addCertEntry(entry)
				for _, host := range entry.domains() {
					if err := cs.registerHostRoute(host, route); err != nil {
						errs = append(errs, err)
					}
				}
			}
			if hasFile {
				continue
			}
			if len(server.serverConfig.SSL.Hosts) > 0 {
				mode, err := server.acmeChallengeMode()
				if err != nil {
					errs = append(errs, err)
					continue
				}
				route.challenge = mode
				for _, host := range server.serverConfig.SSL.Hosts {
					host = strings.ToLower(host)
					if err := cs.registerHostRoute(host, route); err != nil {
						errs = append(errs, err)
						continue
					}
					if _, exists := cs.autoHostModes[host]; exists {
						continue
					}
					cs.autoHostModes[host] = mode
					hostsByMode[mode] = append(hostsByMode[mode], host)
				}
			}
		}
	}

	for mode, hosts := range hostsByMode {
		cs.autoCertManagers[mode] = &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			Cache:      autocert.DirCache(cs.certsPath),
			HostPolicy: autocert.HostWhitelist(hosts...),
		}
	}
	if len(errs) > 0 {
		return errors.Combine(errs...)
	}
	return nil
}

func (cs *certStore) registerHostRoute(host string, route acmeHostRoute) error {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" {
		return errors.New("TLS host is empty")
	}
	if existing, ok := cs.hostRoutes[host]; ok {
		if existing != route {
			return errors.Errorf(
				"TLS host %q has conflicting configurations: %+v and %+v",
				host,
				existing,
				route,
			)
		}
		return nil
	}
	cs.hostRoutes[host] = route
	return nil
}

func (cs *certStore) parseCert(certFile, keyFile string) (certEntry, error) {
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

	entry := &fileCertEntry{
		certFile:   certFile,
		keyFile:    keyFile,
		domainList: domains,
		cert:       &tlsCert,
		nextCheck:  time.Now().Add(certCheckInterval),
	}
	return entry, nil
}

func (cs *certStore) addCertEntry(entry certEntry) {
	cs.certEntries = append(cs.certEntries, entry)
	for _, domain := range entry.domains() {
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
	mode, ok := cs.autoHostModes[host]
	manager := cs.autoCertManagers[mode]
	if !ok || manager == nil {
		if len(cs.certEntries) > 0 {
			certEntry0 := cs.certEntries[0]
			cs.mu.Lock()
			cs.wildcardCache[host] = certEntry0
			cs.mu.Unlock()
			return certEntry0.get()
		}
		return nil, nil
	}
	return manager.GetCertificate(&tls.ClientHelloInfo{ServerName: host})
}

func (cs *certStore) matchWildcard(host string) certEntry {
	for _, entry := range cs.certEntries {
		for _, domain := range entry.domains() {
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
	return cs.generateSelfSigned(host)
}

// generateSelfSigned creates a self-cert entry for the given host.
// The actual certificate is generated lazily on first get() call.
func (cs *certStore) generateSelfSigned(host string) certEntry {
	entry := &selfCertEntry{host: host}

	cs.mu.Lock()
	cs.certEntries = append(cs.certEntries, entry)
	cs.certMap[strings.ToLower(host)] = entry
	cs.mu.Unlock()

	return entry
}

func (cs *certStore) hasAutoCert() bool {
	return len(cs.autoHostModes) > 0
}

func (cs *certStore) hasAutoCertForMode(mode string) bool {
	return cs.autoCertManagers[mode] != nil
}

func (cs *certStore) autoCertManager(mode string) *autocert.Manager {
	return cs.autoCertManagers[mode]
}

func (cs *certStore) autoHostMode(host string) (string, bool) {
	mode, ok := cs.autoHostModes[strings.ToLower(host)]
	return mode, ok
}

func (cs *certStore) hostRoute(host string) (acmeHostRoute, bool) {
	host = strings.ToLower(host)
	if route, ok := cs.hostRoutes[host]; ok {
		return route, true
	}
	for configuredHost, route := range cs.hostRoutes {
		if strings.HasPrefix(configuredHost, "*.") {
			suffix := configuredHost[1:]
			if strings.HasSuffix(host, suffix) {
				prefix := strings.TrimSuffix(host, suffix)
				if prefix != "" && !strings.Contains(prefix, ".") {
					return route, true
				}
			}
		}
	}
	return acmeHostRoute{}, false
}

func (cs *certStore) httpChallengeHandler(fallback http.Handler) http.Handler {
	manager := cs.autoCertManager(ACMEChallengeHTTP01)
	if manager == nil {
		return fallback
	}
	return manager.HTTPHandler(fallback)
}

func (cs *certStore) matchingDomains() []string {
	domains := make([]string, 0)
	for _, entry := range cs.certEntries {
		for _, domain := range entry.domains() {
			if !strings.HasPrefix(domain, "*.") {
				domains = append(domains, domain)
			}
		}
	}
	return domains
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

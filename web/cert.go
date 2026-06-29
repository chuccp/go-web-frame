package web

import (
	"bytes"
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/util"
	"github.com/sourcegraph/conc/pool"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"
)

// certCheckInterval is the minimum interval between file stat checks for
// certificate changes. Under high traffic, every TLS handshake would
// otherwise stat the cert files, causing unnecessary I/O.
const certCheckInterval = 5 * time.Minute

// certEntry holds a dynamically reloadable local certificate.
// It tracks file modification times and reloads the certificate
// automatically when the cert or key file changes on disk.
// File checks are throttled to at most once per certCheckInterval.
type certEntry struct {
	host      string
	certFile  string
	keyFile   string
	cert      *tls.Certificate
	certMod   time.Time
	keyMod    time.Time
	nextCheck time.Time
	mu        sync.RWMutex
}

// get returns the current certificate, reloading from disk if either
// the cert file or key file has been modified since the last load.
// File stat calls are throttled to at most once per certCheckInterval.
func (e *certEntry) get() (*tls.Certificate, error) {
	// Fast path: within cooldown period, return cached cert immediately
	e.mu.RLock()
	if e.cert != nil && time.Now().Before(e.nextCheck) {
		c := e.cert
		e.mu.RUnlock()
		return c, nil
	}
	e.mu.RUnlock()

	// Outside cooldown — stat files to check for changes
	certStat, err := os.Stat(e.certFile)
	if err != nil {
		return nil, errors.Wrapf(err, "stat cert file %s", e.certFile)
	}
	keyStat, err := os.Stat(e.keyFile)
	if err != nil {
		return nil, errors.Wrapf(err, "stat key file %s", e.keyFile)
	}

	// Acquire write lock to update state
	e.mu.Lock()
	defer e.mu.Unlock()

	// Double-check cooldown (another goroutine may have just checked)
	if e.cert != nil && time.Now().Before(e.nextCheck) {
		return e.cert, nil
	}

	// Files unchanged — extend cooldown, return cached cert
	if e.cert != nil && certStat.ModTime().Equal(e.certMod) && keyStat.ModTime().Equal(e.keyMod) {
		e.nextCheck = time.Now().Add(certCheckInterval)
		return e.cert, nil
	}

	// Files changed — reload the certificate
	cert, err := tls.LoadX509KeyPair(e.certFile, e.keyFile)
	if err != nil {
		return nil, errors.Wrapf(err, "reload certificate host=%s cert=%s key=%s", e.host, e.certFile, e.keyFile)
	}
	e.cert = &cert
	e.certMod = certStat.ModTime()
	e.keyMod = keyStat.ModTime()
	e.nextCheck = time.Now().Add(certCheckInterval)

	log.Info("Certificate reloaded from disk", zap.String("host", e.host), zap.String("cert", e.certFile))
	return e.cert, nil
}

// CertManager manages TLS certificates, supporting both Let's Encrypt
// auto-certification and local certificate files.
type CertManager struct {
	certManager *autocert.Manager
	hosts       []string
	port        []int
	lock        *sync.RWMutex
}

// NewCertManager creates a new CertManager with empty host and port lists.
func NewCertManager() *CertManager {
	return &CertManager{
		hosts: []string{},
		port:  []int{},
		lock:  new(sync.RWMutex),
	}
}

// HasTLS reports whether any hosts have been registered for TLS.
func (cm *CertManager) HasTLS() bool {
	return len(cm.hosts) > 0
}

// AddHost registers a domain for Let's Encrypt auto-certification.
// The host is normalized to lowercase and validated as a domain name.
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

// AddPort registers a port number to track which ports the server listens on.
func (cm *CertManager) AddPort(port int) {
	if port > 0 {
		if util.ArrayIntContains(cm.port, port) {
			return
		}
		cm.port = append(cm.port, port)
	}
}

// GetCertManager returns the autocert.Manager, creating one if needed.
// Certificates are cached in the "certs" directory.
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
		// Path to cache certificates
		Cache: autocert.DirCache(certsPath),
		// Domains requiring automatic certificate acquisition
		HostPolicy: autocert.HostWhitelist(cm.hosts...),
	}
	cm.certManager = m
	return m, nil
}

// GetPEM retrieves the PEM-encoded certificate chain and private key for the given host.
// It supports RSA, ECDSA, and Ed25519 private key types.
func (cm *CertManager) GetPEM(host string) (certPEM []byte, keyPEM []byte, err error) {
	manager, err := cm.GetCertManager()
	if err != nil {
		return nil, nil, errors.WithStackIf(err)
	}
	hello := &tls.ClientHelloInfo{ServerName: host}
	tlsCert, err := manager.GetCertificate(hello)
	if err != nil {
		return nil, nil, errors.WithStackIf(err)
	}
	if tlsCert == nil {
		return nil, nil, errors.New("no certificate found")
	}

	// 1. Certificate chain (leaf + intermediates, typically excluding root)
	var certBuf bytes.Buffer
	for i, der := range tlsCert.Certificate {
		err = pem.Encode(&certBuf, &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: der,
		})
		if err != nil {
			return nil, nil, errors.Wrapf(err, "failed to encode certificate #%d", i)
		}
	}
	certPEM = certBuf.Bytes()
	if tlsCert.PrivateKey == nil {
		return certPEM, nil, nil
	}

	var derBytes []byte

	switch pk := tlsCert.PrivateKey.(type) {
	case *rsa.PrivateKey:
		derBytes = x509.MarshalPKCS1PrivateKey(pk)
		pemType := "RSA PRIVATE KEY"

		keyBuf := new(bytes.Buffer)
		if err := pem.Encode(keyBuf, &pem.Block{Type: pemType, Bytes: derBytes}); err != nil {
			return certPEM, nil, errors.Wrap(err, "failed to encode RSA private key")
		}
		return certPEM, keyBuf.Bytes(), nil

	case *ecdsa.PrivateKey:
		derBytes, err = x509.MarshalPKCS8PrivateKey(pk)
		if err != nil {
			return certPEM, nil, errors.Wrap(err, "failed to marshal ECDSA private key to PKCS#8")
		}
		pemType := "PRIVATE KEY"

		keyBuf := new(bytes.Buffer)
		if err := pem.Encode(keyBuf, &pem.Block{Type: pemType, Bytes: derBytes}); err != nil {
			return certPEM, nil, errors.Wrap(err, "failed to encode ECDSA private key")
		}
		return certPEM, keyBuf.Bytes(), nil

	case ed25519.PrivateKey:
		derBytes, err = x509.MarshalPKCS8PrivateKey(pk)
		if err != nil {
			return certPEM, nil, errors.Wrap(err, "failed to marshal Ed25519 private key to PKCS#8")
		}
		pemType := "PRIVATE KEY"

		keyBuf := new(bytes.Buffer)
		if err := pem.Encode(keyBuf, &pem.Block{Type: pemType, Bytes: derBytes}); err != nil {
			return certPEM, nil, errors.Wrap(err, "failed to encode Ed25519 private key")
		}
		return certPEM, keyBuf.Bytes(), nil

	case crypto.Signer, interface{ MarshalPKCS8PrivateKey() ([]byte, error) }:
		if marshaler, ok := pk.(interface{ MarshalPKCS8PrivateKey() ([]byte, error) }); ok {
			derBytes, err = marshaler.MarshalPKCS8PrivateKey()
			if err != nil {
				return certPEM, nil, errors.Wrap(err, "MarshalPKCS8PrivateKey failed")
			}
		} else {
			return certPEM, nil, errors.New("private key implements crypto.Signer but no marshal method")
		}

		keyBuf := new(bytes.Buffer)
		if err := pem.Encode(keyBuf, &pem.Block{Type: "PRIVATE KEY", Bytes: derBytes}); err != nil {
			return certPEM, nil, errors.Wrap(err, "failed to encode private key")
		}
		return certPEM, keyBuf.Bytes(), nil

	default:
		return certPEM, nil, errors.Errorf("unsupported private key type: %T", pk)
	}
}

// Run starts auxiliary HTTP servers on ports 80 and/or 443 for Let's Encrypt
// ACME HTTP-01 challenge handling, if those ports are not already in use.
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

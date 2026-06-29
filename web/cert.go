package web

import (
	"crypto/tls"
	"os"
	"sync"
	"time"

	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/log"
	"go.uber.org/zap"
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

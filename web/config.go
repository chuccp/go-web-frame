// Package web: server configuration types and defaults.
package web

import "time"

// ServerConfigKey is the configuration key for server settings.
const ServerConfigKey = "web.server"

// DefaultMaxBodySize is the default maximum request body size (10 MB).
const DefaultMaxBodySize int64 = 10 << 20

const (
	// ACMEChallengeAuto selects TLS-ALPN-01 on port 443 and HTTP-01 otherwise.
	ACMEChallengeAuto = "auto"
	// ACMEChallengeTLSALPN01 forces the TLS-ALPN-01 challenge.
	ACMEChallengeTLSALPN01 = "tls-alpn-01"
	// ACMEChallengeHTTP01 forces the HTTP-01 challenge.
	ACMEChallengeHTTP01 = "http-01"
)

// ServerConfig holds the HTTP server configuration including port, context path, and TLS.
type ServerConfig struct {
	Port              int        // Listen port (default: 19009)
	ContextPath       string     // Route prefix applied to all routes (e.g., "/api")
	Locations         []string   // Static file directories to serve
	Page404           string     // Fallback page for 404 responses (useful for SPA)
	MaxBodySize       int64      // Maximum request body size in bytes (0 = default 10 MB, -1 = unlimited)
	ReadHeaderTimeout int        // Timeout in seconds for reading request headers (0 = default 30s)
	ReadTimeout       int        // Timeout in seconds for reading the entire request (0 = default 10min)
	MaxHeaderBytes    int        // Maximum size of request headers in bytes (0 = default 8192)
	SSL               *SSLConfig // HTTPS/TLS configuration
}

// SSLConfig holds TLS/HTTPS configuration with optional auto-certification hosts.
type SSLConfig struct {
	Enabled       bool             // Whether HTTPS is enabled
	Hosts         []string         // Domain names for Let's Encrypt auto-certification
	Certs         []SSLCertificate // Local certificate entries for pre-obtained certificates
	BindHTTP      bool             // Listen on port 80 and redirect HTTP requests to HTTPS
	ACMEChallenge string           `json:"acme_challenge"` // auto (default), tls-alpn-01, or http-01
}

// SSLCertificate holds paths to a TLS certificate and private key file.
type SSLCertificate struct {
	CertFile string // Path to the certificate file (PEM format)
	KeyFile  string // Path to the private key file (PEM format)
}

// DefaultServerConfig returns a ServerConfig with default values.
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Port: 19009,
	}
}

// SSLEnabled reports whether HTTPS is enabled in this configuration.
func (sc *ServerConfig) SSLEnabled() bool {
	return sc.SSL != nil && sc.SSL.Enabled
}

// GetReadHeaderTimeout returns the configured ReadHeaderTimeout (seconds) or the default (30s).
func (sc *ServerConfig) GetReadHeaderTimeout() time.Duration {
	if sc.ReadHeaderTimeout > 0 {
		return time.Duration(sc.ReadHeaderTimeout) * time.Second
	}
	return MaxReadHeaderTimeout
}

// GetReadTimeout returns the configured ReadTimeout (seconds) or the default (10min).
func (sc *ServerConfig) GetReadTimeout() time.Duration {
	if sc.ReadTimeout > 0 {
		return time.Duration(sc.ReadTimeout) * time.Second
	}
	return MaxReadTimeout
}

// GetMaxHeaderBytes returns the configured MaxHeaderBytes or the default (8192).
func (sc *ServerConfig) GetMaxHeaderBytes() int {
	if sc.MaxHeaderBytes > 0 {
		return sc.MaxHeaderBytes
	}
	return MaxHeaderBytes
}

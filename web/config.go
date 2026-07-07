// Package web: server configuration types and defaults.
package web

// ServerConfigKey is the configuration key for server settings.
const ServerConfigKey = "web.server"

// DefaultMaxBodySize is the default maximum request body size (10 MB).
const DefaultMaxBodySize int64 = 10 << 20

// ServerConfig holds the HTTP server configuration including port, context path, and TLS.
type ServerConfig struct {
	Port        int        // Listen port (default: 19009)
	ContextPath string     // Route prefix applied to all routes (e.g., "/api")
	Locations   []string   // Static file directories to serve
	Page404     string     // Fallback page for 404 responses (useful for SPA)
	MaxBodySize int64      // Maximum request body size in bytes (0 = default 10 MB, -1 = unlimited)
	SSL         *SSLConfig // HTTPS/TLS configuration
}

// SSLConfig holds TLS/HTTPS configuration with optional auto-certification hosts.
type SSLConfig struct {
	Enabled bool      // Whether HTTPS is enabled
	Hosts   []string  // Domain names for Let's Encrypt auto-certification
	Certs   []SSLCertificate // Local certificate entries for pre-obtained certificates
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

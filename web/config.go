// Package web: server configuration types and defaults.
package web

// ServerConfigKey is the configuration key for server settings.
const ServerConfigKey = "web.server"

// ServerConfig holds the HTTP server configuration including port, context path, and TLS.
type ServerConfig struct {
	Port        int        // Listen port (default: 19009)
	ContextPath string     // Route prefix applied to all routes (e.g., "/api")
	Locations   []string   // Static file directories to serve
	Page404     string     // Fallback page for 404 responses (useful for SPA)
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

package web2

type ServerConfig struct {
	Port        int        // Listen port (default: 19009)
	ContextPath string     // Route prefix applied to all routes (e.g., "/api")
	Locations   []string   // Static file directories to serve
	Page404     string     // Fallback page for 404 responses (useful for SPA)
	SSL         *SSLConfig // HTTPS/TLS configuration
}

type SSLConfig struct {
	Enabled bool      // Whether HTTPS is enabled
	Hosts   []string  // Domain names for Let's Encrypt auto-certification
	Certs   []SSLCert // Local certificate entries for pre-obtained certificates
}

type SSLCert struct {
	CertFile string // Path to the certificate file (PEM format)
	KeyFile  string // Path to the private key file (PEM format)
}

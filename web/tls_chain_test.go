package web

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	config2 "github.com/chuccp/go-web-frame/config"
	"golang.org/x/crypto/acme"
)

func TestTLSCertificateChainHandshake(t *testing.T) {
	root, intermediate, leaf, leafKey := generateTLSTestChain(t)
	certPool := x509.NewCertPool()
	certPool.AddCert(root)

	dir := t.TempDir()
	certFile := filepath.Join(dir, "server-chain.pem")
	keyFile := filepath.Join(dir, "server-key.pem")
	writeCertificateChain(t, certFile, leaf, intermediate)
	writeECPrivateKey(t, keyFile, leafKey)

	certs := newCertStore(t.TempDir())
	entry, err := certs.parseCert(certFile, keyFile)
	if err != nil {
		t.Fatalf("parseCert() error = %v", err)
	}
	certs.addCertEntry(entry)

	server := DefaultServer()
	server.Get("/", func(request *Request) (any, error) {
		return "tls-ok", nil
	})
	httpServer := httptest.NewUnstartedServer(server.GetHandler())
	httpServer.TLS = &tls.Config{
		MinVersion: tls.VersionTLS12,
		GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
			return certs.getCertificate(hello.ServerName)
		},
	}
	httpServer.StartTLS()
	defer httpServer.Close()

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
				RootCAs:    certPool,
				ServerName: "service.test",
			},
		},
	}
	response, err := client.Get(httpServer.URL)
	if err != nil {
		t.Fatalf("HTTPS request failed: %v", err)
	}
	defer func() { _ = response.Body.Close() }()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	if response.StatusCode != http.StatusOK || string(body) != "tls-ok" {
		t.Fatalf("response = %d %q, want 200 tls-ok", response.StatusCode, body)
	}
	if response.TLS == nil {
		t.Fatal("response TLS state is nil")
	}
	if len(response.TLS.PeerCertificates) != 2 {
		t.Fatalf("server sent %d certificates, want leaf + intermediate", len(response.TLS.PeerCertificates))
	}
	if len(response.TLS.VerifiedChains) == 0 {
		t.Fatal("client did not construct a verified certificate chain")
	}
	verified := response.TLS.VerifiedChains[0]
	if len(verified) != 3 {
		t.Fatalf("verified chain length = %d, want leaf + intermediate + root", len(verified))
	}
	if verified[1].Subject.CommonName != "Test Intermediate CA" {
		t.Errorf("intermediate CN = %q", verified[1].Subject.CommonName)
	}
	if verified[2].Subject.CommonName != "Test Root CA" {
		t.Errorf("root CN = %q", verified[2].Subject.CommonName)
	}
	if err := verified[0].VerifyHostname("service.test"); err != nil {
		t.Errorf("VerifyHostname() error = %v", err)
	}
}

func TestTLSCertificateChainRejectsWrongOrder(t *testing.T) {
	_, intermediate, leaf, leafKey := generateTLSTestChain(t)
	dir := t.TempDir()
	certFile := filepath.Join(dir, "wrong-order.pem")
	keyFile := filepath.Join(dir, "wrong-order-key.pem")
	writeCertificateChain(t, certFile, intermediate, leaf)
	writeECPrivateKey(t, keyFile, leafKey)

	certs := newCertStore(t.TempDir())
	if _, err := certs.parseCert(certFile, keyFile); err == nil {
		t.Fatal("parseCert() accepted a certificate chain whose leaf is not first")
	}
}

func TestAutoCertTLSConfigEnablesTLSALPN(t *testing.T) {
	server := DefaultServer()
	server.serverConfig.SSL = &SSLConfig{
		Enabled: true,
		Hosts:   []string{"auto.example.com"},
	}
	certs := newCertStore(t.TempDir())
	if err := certs.init([]*Server{server}); err != nil {
		t.Fatalf("certStore.init() error = %v", err)
	}

	config := server.serverTLSConfig(certs)
	if !slices.Contains(config.NextProtos, acme.ALPNProto) {
		t.Fatalf("TLS NextProtos = %v, missing %q", config.NextProtos, acme.ALPNProto)
	}
	if !slices.Contains(config.NextProtos, "h2") {
		t.Errorf("TLS NextProtos = %v, missing HTTP/2", config.NextProtos)
	}
}

func TestSSLConfigBindHTTP(t *testing.T) {
	raw := []byte(`
web:
  server:
    ssl:
      enabled: true
      bind_http: true
`)
	cfg, err := config2.NewFromBytes(raw, "yaml")
	if err != nil {
		t.Fatalf("NewFromBytes() error = %v", err)
	}
	serverConfig := DefaultServerConfig()
	if err := cfg.UnmarshalKey(ServerConfigKey, serverConfig); err != nil {
		t.Fatalf("UnmarshalKey() error = %v", err)
	}
	if serverConfig.SSL == nil || !serverConfig.SSL.BindHTTP {
		t.Fatalf("SSLConfig.BindHTTP = %v, want true", serverConfig.SSL)
	}
}

func TestHTTPRedirectHandler(t *testing.T) {
	tests := []struct {
		name      string
		httpsPort int
		url       string
		want      string
	}{
		{
			name:      "default HTTPS port",
			httpsPort: 443,
			url:       "http://example.com:80/users?page=2",
			want:      "https://example.com/users?page=2",
		},
		{
			name:      "custom HTTPS port",
			httpsPort: 8443,
			url:       "http://example.com:80/users",
			want:      "https://example.com:8443/users",
		},
		{
			name:      "IPv6 default HTTPS port",
			httpsPort: 443,
			url:       "http://[2001:db8::1]:80/users",
			want:      "https://[2001:db8::1]/users",
		},
		{
			name:      "IPv6 custom HTTPS port",
			httpsPort: 8443,
			url:       "http://[2001:db8::1]:80/users",
			want:      "https://[2001:db8::1]:8443/users",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.url, nil)
			recorder := httptest.NewRecorder()
			httpRedirectHandler(tt.httpsPort).ServeHTTP(recorder, request)
			if recorder.Code != http.StatusPermanentRedirect {
				t.Errorf("status = %d, want %d", recorder.Code, http.StatusPermanentRedirect)
			}
			if location := recorder.Header().Get("Location"); location != tt.want {
				t.Errorf("Location = %q, want %q", location, tt.want)
			}
		})
	}
}

func TestHTTPRedirectTarget(t *testing.T) {
	plain := DefaultServer()
	plain.serverConfig.Port = 8080

	tlsServer := DefaultServer()
	tlsServer.serverConfig.Port = 8443
	tlsServer.serverConfig.SSL = &SSLConfig{Enabled: true, BindHTTP: true}

	runner := &ServerRunner{servers: []*Server{plain, tlsServer}}
	target, ok := runner.httpRedirectTarget()
	if !ok || target != tlsServer {
		t.Fatalf("httpRedirectTarget() = %v, %v; want TLS server, true", target, ok)
	}

	tlsServer.serverConfig.SSL.BindHTTP = false
	if target, ok := runner.httpRedirectTarget(); ok || target != nil {
		t.Fatalf("httpRedirectTarget() = %v, %v; want nil, false", target, ok)
	}
}

func generateTLSTestChain(
	t *testing.T,
) (*x509.Certificate, *x509.Certificate, *x509.Certificate, *ecdsa.PrivateKey) {
	t.Helper()
	now := time.Now()

	rootKey := newTLSTestKey(t)
	rootTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test Root CA"},
		NotBefore:             now.Add(-time.Hour),
		NotAfter:              now.Add(24 * time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		MaxPathLen:            1,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
	}
	root := createTLSTestCertificate(t, rootTemplate, rootTemplate, &rootKey.PublicKey, rootKey)

	intermediateKey := newTLSTestKey(t)
	intermediateTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(2),
		Subject:               pkix.Name{CommonName: "Test Intermediate CA"},
		NotBefore:             now.Add(-time.Hour),
		NotAfter:              now.Add(24 * time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		MaxPathLen:            0,
		MaxPathLenZero:        true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
	}
	intermediate := createTLSTestCertificate(
		t,
		intermediateTemplate,
		root,
		&intermediateKey.PublicKey,
		rootKey,
	)

	leafKey := newTLSTestKey(t)
	leafTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject:      pkix.Name{CommonName: "service.test"},
		DNSNames:     []string{"service.test"},
		NotBefore:    now.Add(-time.Hour),
		NotAfter:     now.Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	leaf := createTLSTestCertificate(
		t,
		leafTemplate,
		intermediate,
		&leafKey.PublicKey,
		intermediateKey,
	)
	return root, intermediate, leaf, leafKey
}

func newTLSTestKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	return key
}

func createTLSTestCertificate(
	t *testing.T,
	template *x509.Certificate,
	parent *x509.Certificate,
	publicKey any,
	signer any,
) *x509.Certificate {
	t.Helper()
	der, err := x509.CreateCertificate(rand.Reader, template, parent, publicKey, signer)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parse certificate: %v", err)
	}
	return cert
}

func writeCertificateChain(t *testing.T, path string, certs ...*x509.Certificate) {
	t.Helper()
	var output bytes.Buffer
	for _, cert := range certs {
		if err := pem.Encode(&output, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}); err != nil {
			t.Fatalf("encode certificate: %v", err)
		}
	}
	if err := os.WriteFile(path, output.Bytes(), 0o600); err != nil {
		t.Fatalf("write certificate chain: %v", err)
	}
}

func writeECPrivateKey(t *testing.T, path string, key *ecdsa.PrivateKey) {
	t.Helper()
	der, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatalf("marshal private key: %v", err)
	}
	data := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: der})
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write private key: %v", err)
	}
}

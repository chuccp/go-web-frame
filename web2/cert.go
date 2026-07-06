package web2

import (
	"crypto/tls"
	"net/http"

	"github.com/chuccp/go-web-frame/log"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"
)

type certManager struct {
	serversMap      map[int]*Server
	autoCertManager *autocert.Manager
	autoHosts       []string
}

func newCertManager(certsPath string, servers []*Server) *certManager {
	hosts := make([]string, 0)
	serversMap := make(map[int]*Server)
	for _, server := range servers {
		serversMap[server.serverConfig.Port] = server
		if server.serverConfig.SSL != nil && server.serverConfig.SSL.Enabled {
			if server.serverConfig.SSL.Certs != nil && len(server.serverConfig.SSL.Certs) > 0 {
				continue
			} else {
				if server.serverConfig.SSL.Hosts != nil && len(server.serverConfig.SSL.Hosts) > 0 {
					hosts = append(hosts, server.serverConfig.SSL.Hosts...)
				}
			}

		}
	}
	autoCertManager := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache(certsPath),
		HostPolicy: autocert.HostWhitelist(hosts...),
	}
	return &certManager{serversMap, autoCertManager, hosts}
}
func (certManager *certManager) get(host string) (*tls.Certificate, error) {
	return certManager.autoCertManager.GetCertificate(&tls.ClientHelloInfo{ServerName: host})
}
func (certManager *certManager) start() {
	if len(certManager.autoHosts) == 0 {
		return
	}
	_, ok := certManager.serversMap[80]
	if !ok {
		server := &http.Server{
			Addr:    ":80",
			Handler: certManager.autoCertManager.HTTPHandler(nil),
		}
		go func() {
			log.Info("starting ACME HTTP-01 challenge server on :80")
			if err := server.ListenAndServe(); err != nil {
				log.Error("ACME challenge server error", zap.Error(err))
			}
		}()
	}
}

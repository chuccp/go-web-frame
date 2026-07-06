package web2

type listener struct {
	tlsServers   []*Server
	noTlsServers []*Server
	certManager  *certManager
}

func newListener(servers []*Server) *listener {
	tlsServers := make([]*Server, 0)
	noTlsServers := make([]*Server, 0)
	for _, server := range servers {
		if server.serverConfig.SSL != nil && server.serverConfig.SSL.Enabled {
			tlsServers = append(tlsServers, server)
		} else {
			noTlsServers = append(noTlsServers, server)
		}
	}
	return &listener{tlsServers: tlsServers, noTlsServers: noTlsServers, certManager: newCertManager("./autocert-cache", servers)}
}

func (l *listener) start() {
	l.certManager.start()

}

func (l *listener) startTLS() {

}
func (l *listener) startNoTls() {

}

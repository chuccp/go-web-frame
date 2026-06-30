package web

type CertServer struct {
	serverConfig []*ServerConfig
}

func New(serverConfig []*ServerConfig) *CertServer {
	return &CertServer{serverConfig: serverConfig}
}

package web2

import (
	"testing"

	"github.com/chuccp/go-web-frame/web"
)

func TestTslServer(t *testing.T) {

	var certServer = NewServer()

	server, err := certServer.CreateServer(&web.ServerConfig{Port: 1256})
	if err != nil {
		t.Fatal(err)
	}
	server.Get("/", func(request *web.Request) (any, error) {
		return "hello", nil
	})

	err = certServer.Start()
	if err != nil {
		t.Fatal(err)
		return
	}

}

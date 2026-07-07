package web

import (
	"testing"
)

func TestTslServer(t *testing.T) {

	var servers = NewServers()

	server, err := servers.CreateServer(&ServerConfig{Port: 1256})
	if err != nil {
		t.Fatal(err)
	}
	server.Get("/", func(request *Request) (any, error) {
		return "hello", nil
	})

	err = servers.Start()
	if err != nil {
		t.Fatal(err)
		return
	}

}

package web

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTslServer(t *testing.T) {

	server := DefaultServer()
	server.serverConfig.Locations = []string{"C:\\Users\\cao\\Documents"}
	server.Get("/", func(r *Request) (any, error) {
		return "hello", nil
	})

	ts := httptest.NewServer(server.GetHandler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	t.Log(string(body))

}

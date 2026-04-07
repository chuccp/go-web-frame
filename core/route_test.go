package core

import (
	"testing"

	"github.com/chuccp/go-web-frame/web"
)

func TestRoute(t *testing.T) {

	rt := make(web.RouteTree)

	//rt.Set("GET", "/api/user/:id")

	t.Log(rt.Has("GET", "/api2/user/:id"))
}

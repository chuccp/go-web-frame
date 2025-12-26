package core

import "testing"

func TestRoute(t *testing.T) {

	rt := make(RouteTree)

	//rt.Set("GET", "/api/user/:id")

	t.Log(rt.Has("GET", "/api2/user/:id"))
}

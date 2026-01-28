package main

import (
	wf "github.com/chuccp/go-web-frame"
	"github.com/chuccp/go-web-frame/web"
)

func main() {

	webFrame := wf.NewWithAutoConfig()
	webFrame.Get("/", func(c *web.Request) (any, error) {
		return "hello world", nil
	})
	err := webFrame.Start()
	if err != nil {
		return
	}

}

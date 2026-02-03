package main

import (
	"context"
	"time"

	wf "github.com/chuccp/go-web-frame"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/web"
)

func main() {

	webFrame := wf.NewWithAutoConfig()
	webFrame.Get("/", func(c *web.Request) (any, error) {
		return "hello world", nil
	})

	ctx, ctxFun := context.WithCancel(context.Background())
	go func() {
		time.Sleep(time.Second * 10)
		ctxFun()
	}()

	err := webFrame.Run(ctx)
	if err != nil {
		log.PrintPanic(err)
		return
	}

}

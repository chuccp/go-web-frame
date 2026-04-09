package main

import (
	"context"
	"time"

	wf "github.com/chuccp/go-web-frame"
	config2 "github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/web"
)

func main() {

	builder := wf.NewBuilder(config2.LoadAutoConfig())
	builder.Get("/", func(c *web.Request) (any, error) {
		return "hello world", nil
	})
	webFrame := builder.Build()

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

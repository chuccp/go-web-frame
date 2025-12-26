package main

import (
	"log"

	"github.com/chuccp/go-web-frame"
	"github.com/chuccp/go-web-frame/config"
)

func main() {
	loadConfig, err := config.LoadConfig("application.yml")
	if err != nil {
		return
	}
	web := wf.New(loadConfig)
	err = web.Start()
	if err != nil {
		log.Printf("启动失败 %v", err)
		return
	}
}

package main

import (
	"log"

	"github.com/chuccp/go-web-frame"
	"github.com/chuccp/go-web-frame/cache"
	"github.com/chuccp/go-web-frame/component/captcha"
	"github.com/chuccp/go-web-frame/config"
)

func main() {
	loadConfig, err := config.LoadConfig("application.yml")
	if err != nil {
		return
	}
	web := wf.New(loadConfig)
	web.AddRest(&Api{})
	web.AddComponent(&cache.Component{})
	web.AddComponent(&captcha.captcha{})
	err = web.Start()
	if err != nil {
		log.Printf("启动失败 %v", err)
		return
	}
}

package main

import (
	"github.com/chuccp/go-web-frame/captcha"
	"github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/core"
)

func main() {
	loadConfig, err := config.LoadConfig("application.yml")
	if err != nil {
		return
	}
	web := core.New(loadConfig)
	web.AddRest(&Api{})
	web.AddComponent(&captcha.Component{})
	err = web.Start()
	if err != nil {
		return
	}
}

package main

import (
	"github.com/chuccp/go-web-frame/captcha"
	"github.com/chuccp/go-web-frame/core"
)

func main() {
	web := core.CreateWebFrame("./application.yml")
	web.AddRest(&Api{})
	web.AddComponent(&captcha.Component{})
	err := web.Start()
	if err != nil {
		return
	}
}

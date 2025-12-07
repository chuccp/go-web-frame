package main

import (
	"log"

	"github.com/chuccp/go-web-frame/core"
)

func main() {
	web := core.CreateWeb("application.yml")
	web.GetRestGroup().Authentication(&Authentication{})
	web.AddRest(&Api{})
	err := web.Start()
	if err != nil {
		log.Printf("启动失败 %v", err)
		return
	}
}

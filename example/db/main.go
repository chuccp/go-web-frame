package main

import (
	"log"

	"github.com/chuccp/go-web-frame/core"
)

func main() {
	web := core.CreateWebFrame("application.yml")
	err := web.Start()
	if err != nil {
		log.Printf("启动失败 %v", err)
		return
	}
}

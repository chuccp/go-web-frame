package main

import "github.com/chuccp/go-web-frame/core"

func main() {
	web := core.CreateWeb("./application.yml")
	web.AddRest(&Api{})
	err := web.Start()
	if err != nil {
		return
	}
}

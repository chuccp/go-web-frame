package main

import wf "github.com/chuccp/go-web-frame"

func main() {

	webFrame := wf.NewWithAutoConfig()
	err := webFrame.Start()
	if err != nil {
		return
	}

}

package main

import (
	"fmt"
	"time"

	"github.com/sourcegraph/conc/panics"
)

func main() {

	var catcher panics.Catcher
	catcher.Try(func() {

		time.Sleep(5 * time.Second)
		panic("panic")
	})

	//log.Debug("panic", zap.Error(catcher.Recovered().AsError()))

	fmt.Println("主协程继续执行，不会 re-panic")
}

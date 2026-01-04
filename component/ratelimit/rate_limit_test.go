package rate_limit

import (
	"testing"

	wf "github.com/chuccp/go-web-frame"
	"github.com/chuccp/go-web-frame/log"
)

func TestName(t *testing.T) {
	frame := wf.New(wf.LoadAutoConfig())
	err := frame.Start()
	if err != nil {
		log.Errors("启动失败 %v", err)
		return
	}

}

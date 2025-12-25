package web

import (
	"testing"
)

func TestMemfs(t *testing.T) {
	v := DefaultMemFileSystem()
	stat, err := v.Stat("message.go")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(stat.Name())
}

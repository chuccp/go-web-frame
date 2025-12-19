package core

import config2 "github.com/chuccp/go-web-frame/config"

type IComponent interface {
	Init(Config config2.IConfig) error
	Name() string
}

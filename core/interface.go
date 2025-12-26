package core

import (
	config2 "github.com/chuccp/go-web-frame/config"
)

type IModel interface {
	IsExist() bool
	CreateTable() error
	DeleteTable() error
	GetTableName() string
	Init(context *Context) error
}
type IRest interface {
	IService
}
type IComponent interface {
	Init(Config config2.IConfig) error
}
type IService interface {
	Init(Config *Context) error
}

type IRunner interface {
	Init(Config *Context) error
	Run() error
	Stop() error
}

package core

import (
	config2 "github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/db"
)

type IService interface {
	Init(Config *Context) error
}
type IModel interface {
	IService
	IsExist() bool
	CreateTable() error
	DeleteTable() error
	GetTableName() string
	ReNew(db *db.DB, c *Context) IModel
}
type IDestroy interface {
	Destroy() error
}
type IRest interface {
	IService
}
type IComponent interface {
	Init(Config config2.IConfig) error
	IDestroy
}

type IRunner interface {
	IService
	IDestroy
	Run() error
}

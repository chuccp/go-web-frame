package core

import (
	config2 "github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/go-web-frame/model"
)

type IService interface {
	Init(ctx *Context) error
}
type IModel interface {
	Init(db *db.DB, c *Context) error
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
	Init(config config2.IConfig) error
	IDestroy
}

type IRunner interface {
	IService
	IDestroy
	Run() error
}

type IModelGroup interface {
	AddModel(model ...IModel)
	GetModel() []IModel
	Init(context *Context) error
	SwitchDB(db *db.DB, context *Context) error
	Name() string
	GetTransaction() *model.Transaction
}

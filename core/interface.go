package core

import (
	"context"

	config2 "github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/go-web-frame/model"
	"github.com/chuccp/go-web-frame/web"
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
type IRest interface {
	IService
}
type IComponent interface {
	Init(ctx context.Context, config config2.IConfig) error
}
type IRun interface {
	Run(ctx context.Context) error
}
type IRunner interface {
	IService
	Run(ctx context.Context) error
}
type IModelGroup interface {
	IService
	AddModel(model ...IModel)
	GetModel() []IModel
	AutoCreateTable(autoCreateTable bool)
	SwitchDB(db *db.DB, context *Context) error
	SetDefaultDB(db *db.DB)
	Name() string
	GetTransaction() *model.Transaction
}

type IFilter interface {
	IService
	web.Filter
}
type IConverter interface {
	IService
	web.Converter
}

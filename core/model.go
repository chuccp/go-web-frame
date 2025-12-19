package core

type IModel interface {
	IsExist() bool
	CreateTable() error
	DeleteTable() error
	GetTableName() string
	Name() string
	Init(context *Context)
}

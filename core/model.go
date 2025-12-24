package core

type IModel interface {
	IsExist() bool
	CreateTable() error
	DeleteTable() error
	GetTableName() string
	Init(context *Context) error
}

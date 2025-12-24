package core

type IModel interface {
	IService
	IsExist() bool
	CreateTable() error
	DeleteTable() error
	GetTableName() string
}

package model

import (
	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/util"
	"gorm.io/gorm"
)

type Model[T any] struct {
	db        *gorm.DB
	tableName string
	entry     T
}

func (a *Model[T]) IsExist() bool {
	return a.db.Migrator().HasTable(a.tableName)
}
func (a *Model[T]) CreateTable() error {
	if a.IsExist() {
		return nil
	}
	t := util.NewPtr(a.entry)
	return errors.WithStackIf(a.db.Table(a.tableName).AutoMigrate(t))
}
func (a *Model[T]) DeleteTable() error {
	t := util.NewPtr(a.entry)
	tx := a.db.Table(a.tableName).Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(t)
	return errors.WithStackIf(tx.Error)
}
func (a *Model[T]) GetTableName() string {
	return a.tableName
}
func (a *Model[T]) Save(entry T) error {
	return a.db.Table(a.tableName).Save(entry).Error
}

func (a *Model[T]) Saves(entry []T) error {
	return a.db.Table(a.tableName).Create(&entry).Error
}

func (a *Model[T]) Query() *Query[T] {
	tx := a.db.Table(a.tableName)
	return &Query[T]{tx: tx, entry: a.entry}
}

func (a *Model[T]) Update() *Update[T] {
	tx := a.db.Table(a.tableName)
	return &Update[T]{tx: tx, model: a.entry, wheres: NewUpdateWheres[T](tx)}
}
func (a *Model[T]) Delete() *Delete[T] {
	tx := a.db.Table(a.tableName)
	return &Delete[T]{tx: tx, model: a.entry, wheres: NewDeleteWheres[T](tx, a.entry)}
}

func NewModel[T any](db *gorm.DB, tableName string) *Model[T] {
	var entryPtr T
	return &Model[T]{db: db, tableName: tableName, entry: util.NewPtr(entryPtr)}
}

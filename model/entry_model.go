package model

import (
	"time"

	"github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/go-web-frame/util"
	"github.com/chuccp/go-web-frame/web"
)

type IEntry interface {
	SetCreateTime(createTime time.Time)
	SetUpdateTime(updateTIme time.Time)
	GetId() uint
	SetId(id uint)
}

type EntryModel[T IEntry] struct {
	model *Model[T]
}

func NewEntryModel[T IEntry](db *db.DB, tableName string) *EntryModel[T] {
	return &EntryModel[T]{NewModel[T](db, tableName)}
}

func (a *EntryModel[T]) IsExist() bool {
	return a.model.IsExist()
}
func (a *EntryModel[T]) CreateTable() error {
	return a.model.CreateTable()
}
func (a *EntryModel[T]) DeleteTable() error {
	return a.model.DeleteTable()
}

func (a *EntryModel[T]) Save(t T) error {
	t.SetCreateTime(time.Now())
	t.SetUpdateTime(time.Now())
	return a.model.Save(t)
}
func (a *EntryModel[T]) Saves(ts []T) error {
	for _, t := range ts {
		t.SetCreateTime(time.Now())
		t.SetUpdateTime(time.Now())
	}
	return a.model.Saves(ts)

}
func (a *EntryModel[T]) FindById(id uint) (T, error) {
	t := util.NewPtr(a.model.entry)
	t.SetId(id)
	err := a.model.db.Table(a.model.tableName).First(&t)
	return t, err
}

func (a *EntryModel[T]) FindOne(query interface{}, args ...interface{}) (T, error) {
	t := util.NewPtr(a.model.entry)
	err := a.model.db.Table(a.model.tableName).Where(query, args...).First(&t)
	return t, err
}

func (a *EntryModel[T]) FindAllByIds(id ...uint) ([]T, error) {
	return a.model.Query().Where("`id` in (?) ", id).All()
}
func (a *EntryModel[T]) FindAll() ([]T, error) {
	return a.model.Query().All()
}
func (a *EntryModel[T]) DeleteOne(id uint) error {
	t := util.NewPtr(a.model.entry)
	err := a.model.db.Table(a.model.tableName).Where("`id` = ? ", id).Delete(t)
	return err
}

func (a *EntryModel[T]) UpdateById(t T) error {
	t.SetUpdateTime(time.Now())
	return a.model.Update().Where("`id` = ? ", t.GetId()).Update(t)
}
func (a *EntryModel[T]) UpdateColumn(id uint, column string, value interface{}) error {
	return a.model.Update().Where("`id` = ? ", id).UpdateColumn(column, value)
}
func (a *EntryModel[T]) UpdateForMap(id uint, data map[string]interface{}) error {
	return a.model.Update().Where("`id` = ? ", id).UpdateForMap(data)

}

func (a *EntryModel[T]) NewEntryModel(db *db.DB) *EntryModel[T] {
	return &EntryModel[T]{&Model[T]{db, a.model.tableName, a.model.entry}}
}
func (a *EntryModel[T]) Page(page *web.Page) ([]T, int, error) {
	return a.model.Query().Order("`id` desc").Page(page)
}
func (a *EntryModel[T]) QueryPage(page *web.Page, query interface{}, args ...interface{}) ([]T, int, error) {
	return a.model.Query().Where(query, args...).Order("`id` desc").Page(page)
}

func (a *EntryModel[T]) Query() *Query[T] {

	return a.model.Query()
}

func (a *EntryModel[T]) Update() *Update[T] {
	return a.model.Update()
}
func (a *EntryModel[T]) Delete() *Delete[T] {
	return a.model.Delete()
}

func (a *EntryModel[T]) GetTableName() string {
	return a.model.tableName
}

type Transaction struct {
	db *db.DB
}

func (t *Transaction) Exec(fc func(tx *db.DB) error) error {
	return t.db.Transaction(fc)
}

func NewTransaction(db *db.DB) *Transaction {
	return &Transaction{db: db}
}

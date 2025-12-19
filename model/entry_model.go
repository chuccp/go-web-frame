package model

import (
	"reflect"
	"time"

	"github.com/chuccp/go-web-frame/web"
	"gorm.io/gorm"
)

type IEntry interface {
	SetCreateTime(createTime time.Time)
	SetUpdateTime(updateTIme time.Time)
	GetId() uint
	SetId(id uint)
}

type EntryModel[T IEntry] struct {
	*Model[T]
}

func NewPtr[T any](v T) T {
	type_ := reflect.TypeOf(v)
	switch type_.Kind() {
	case reflect.Ptr:
		return newPtr(type_.Elem()).(T)
	case reflect.Struct:
		return newPtr(type_).(T)
	default:
		panic("unhandled default case")
	}
}

func NewSlice[T any](v T) []T {
	elemType := reflect.TypeOf(v)
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}
	ptrType := reflect.PointerTo(elemType)
	sliceType := reflect.SliceOf(ptrType)
	slice := reflect.MakeSlice(sliceType, 0, 0)
	return slice.Interface().([]T)
}

func newPtr(type_ reflect.Type) interface{} {
	value := reflect.New(type_)
	u := value.Interface()
	return u
}

func NewEntryModel[T IEntry](db *gorm.DB, tableName string, entry T) *EntryModel[T] {
	return &EntryModel[T]{NewModel(db, tableName, entry)}
}

//func (a *EntryModel[T]) IsExist() bool {
//	return a.db.Migrator().HasTable(a.tableName)
//}
//func (a *EntryModel[T]) CreateTable() error {
//	if a.IsExist() {
//		return nil
//	}
//	t := NewPtr(a.entry)
//	err := a.db.Table(a.tableName).AutoMigrate(t)
//	return err
//}
//func (a *EntryModel[T]) DeleteTable() error {
//	t := NewPtr(a.entry)
//	tx := a.db.Table(a.tableName).Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(t)
//	return tx.Error
//}

func (a *EntryModel[T]) Save(t T) error {
	t.SetCreateTime(time.Now())
	t.SetUpdateTime(time.Now())
	return a.Model.Save(t)
}

func (a *EntryModel[T]) FindById(id uint) (T, error) {
	t := NewPtr(a.entry)
	t.SetId(id)
	tx := a.db.Table(a.tableName).First(&t)
	if tx.Error == nil {
		return t, nil
	}
	return t, tx.Error
}

func (a *EntryModel[T]) FindOne(query interface{}, args ...interface{}) (T, error) {
	t := NewPtr(a.entry)
	tx := a.db.Table(a.tableName).Where(query, args...).First(&t)
	if tx.Error == nil {
		return t, nil
	}
	return t, tx.Error
}

func (a *EntryModel[T]) FindAllByIds(id ...uint) ([]T, error) {
	return a.Query().Where("`id` in (?) ", id).All()
}
func (a *EntryModel[T]) DeleteOne(id uint) error {
	t := NewPtr(a.entry)
	tx := a.db.Table(a.tableName).Where("`id` = ? ", id).Delete(t)
	return tx.Error
}

func (a *EntryModel[T]) UpdateById(t T) error {
	t.SetUpdateTime(time.Now())
	return a.Update().Where("`id` = ? ", t.GetId()).Update(t)
}
func (a *EntryModel[T]) UpdateColumn(id uint, column string, value interface{}) error {
	return a.Update().Where("`id` = ? ", id).UpdateColumn(column, value)
}
func (a *EntryModel[T]) UpdateForMap(id uint, data map[string]interface{}) error {
	return a.Update().Where("`id` = ? ", id).UpdateForMap(data)

}

func (a *EntryModel[T]) NewEntryModel(db *gorm.DB) *EntryModel[T] {
	return &EntryModel[T]{&Model[T]{db, a.tableName, a.entry}}
}
func (a *EntryModel[T]) Page(page *web.Page) ([]T, int, error) {
	return a.Query().Order("`id` desc").Page(page)
}
func (a *EntryModel[T]) QueryPage(page *web.Page, query interface{}, args ...interface{}) ([]T, int, error) {
	return a.Query().Where(query, args...).Order("`id` desc").Page(page)
}

//func (a *EntryModel[T]) Query() *Query[T] {
//	tx := a.db.Table(a.tableName)
//	return &Query[T]{tx: tx, entry: a.entry}
//}

//func (a *EntryModel[T]) Update() *Update[T] {
//	tx := a.db.Table(a.tableName)
//	return &Update[T]{tx: tx, model: a.entry, wheres: NewUpdateWheres[T](tx)}
//}

//func (a *EntryModel[T]) GetTableName() string {
//	return a.tableName
//}

type Transaction struct {
	db *gorm.DB
}

func (t *Transaction) Exec(fc func(tx *gorm.DB) error) error {
	return t.db.Transaction(fc)
}

func NewTransaction(db *gorm.DB) *Transaction {
	return &Transaction{db: db}
}

package core

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

type IModel interface {
	IsExist() bool
	CreateTable() error
	DeleteTable() error
	GetTableName() string
	Name() string
	Init(context *Context)
}

type Query[T IEntry] struct {
	tx    *gorm.DB
	model T
}

func (q *Query[T]) Where(query interface{}, args ...interface{}) *Query[T] {
	q.tx = q.tx.Where(query, args...)
	return q
}
func (q *Query[T]) Order(query interface{}) *Query[T] {
	q.tx = q.tx.Order(query)
	return q
}
func (q *Query[T]) List(page *web.Page) ([]T, error) {
	ts := NewSlice(q.model)
	tx := q.tx.Offset((page.PageNo - 1) * page.PageSize).Limit(page.PageSize).Find(&ts)
	if tx.Error == nil {
		return ts, nil
	}
	return nil, tx.Error

}
func (q *Query[T]) All() ([]T, error) {
	ts := NewSlice(q.model)
	tx := q.tx.Find(&ts)
	if tx.Error == nil {
		return ts, nil
	}
	return nil, tx.Error
}
func (q *Query[T]) One() (T, error) {
	t := NewPtr(q.model)
	tx := q.tx.Limit(1).First(&t)
	if tx.Error == nil {
		return t, nil
	}
	return t, tx.Error
}

func (q *Query[T]) Page(page *web.Page) ([]T, int, error) {
	ts := NewSlice(q.model)
	tx := q.tx.Offset((page.PageNo - 1) * page.PageSize).Limit(page.PageSize).Find(&ts)
	if tx.Error == nil {
		var num int64
		tx := q.tx.Count(&num)
		if tx.Error == nil {
			return ts, int(num), nil
		}
	}
	return nil, 0, tx.Error

}
func (q *Query[T]) Size(size int) ([]T, int, error) {
	ts := NewSlice(q.model)
	tx := q.tx.Limit(size).Find(&ts)
	if tx.Error == nil {
		var num int64
		tx := q.tx.Count(&num)
		if tx.Error == nil {
			return ts, int(num), nil
		}
	}
	return nil, 0, tx.Error
}

type Model[T IEntry] struct {
	db        *gorm.DB
	tableName string
	model     T
}

func NewPtr[T IEntry](v T) T {
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

func NewSlice[T IEntry](v T) []T {
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

func NewModel[T IEntry](db *gorm.DB, tableName string, model T) *Model[T] {

	return &Model[T]{db: db, tableName: tableName, model: model}
}

func (a *Model[T]) IsExist() bool {
	return a.db.Migrator().HasTable(a.tableName)
}
func (a *Model[T]) CreateTable() error {
	if a.IsExist() {
		return nil
	}
	t := NewPtr(a.model)
	err := a.db.Table(a.tableName).AutoMigrate(t)
	return err
}
func (a *Model[T]) DeleteTable() error {
	t := NewPtr(a.model)
	tx := a.db.Table(a.tableName).Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(t)
	return tx.Error
}

func (a *Model[T]) Save(t T) error {
	t.SetCreateTime(time.Now())
	t.SetUpdateTime(time.Now())
	tx := a.db.Table(a.tableName).Create(t)
	return tx.Error
}

func (a *Model[T]) FindById(id uint) (T, error) {
	t := NewPtr(a.model)
	t.SetId(id)
	tx := a.db.Table(a.tableName).First(&t)
	if tx.Error == nil {
		return t, nil
	}
	return t, tx.Error
}

func (a *Model[T]) FindOne(query interface{}, args ...interface{}) (T, error) {
	t := NewPtr(a.model)
	tx := a.db.Table(a.tableName).Where(query, args...).First(&t)
	if tx.Error == nil {
		return t, nil
	}
	return t, tx.Error
}

func (a *Model[T]) FindAllByIds(id ...uint) ([]T, error) {
	return a.Query().Where("`id` in (?) ", id).All()
}
func (a *Model[T]) DeleteOne(id uint) error {
	t := NewPtr(a.model)
	tx := a.db.Table(a.tableName).Where("`id` = ? ", id).Delete(t)
	return tx.Error
}

func (a *Model[T]) Edit(t T) error {
	t.SetUpdateTime(time.Now())
	tx := a.db.Table(a.tableName).Updates(t)
	return tx.Error
}
func (a *Model[T]) Update(id uint, column string, value interface{}) error {
	tx := a.db.Table(a.tableName).Update(column, value).Where("`id` = ? ", id)
	return tx.Error
}
func (a *Model[T]) EditForMap(id uint, data map[string]interface{}) error {
	tx := a.db.Table(a.tableName).Where("`id` = ? ", id).Updates(data)
	return tx.Error
}

func (a *Model[T]) NewModel(db *gorm.DB) *Model[T] {
	return &Model[T]{db: db, tableName: a.tableName}
}
func (a *Model[T]) Page(page *web.Page) ([]T, int, error) {
	return a.Query().Order("`id` desc").Page(page)
}
func (a *Model[T]) QueryPage(page *web.Page, query interface{}, args ...interface{}) ([]T, int, error) {
	return a.Query().Where(query, args...).Order("`id` desc").Page(page)
}
func (a *Model[T]) Query() *Query[T] {
	tx := a.db.Table(a.tableName)
	return &Query[T]{tx: tx, model: a.model}
}
func (a *Model[T]) GetTableName() string {
	return a.tableName
}

type Transaction struct {
	db *gorm.DB
}

func (t *Transaction) Exec(fc func(tx *gorm.DB) error) error {
	return t.db.Transaction(fc)
}

func NewTransaction(db *gorm.DB) *Transaction {
	return &Transaction{db: db}
}

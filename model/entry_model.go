package model

import (
	"fmt"

	"github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/go-web-frame/web"
)

type PKConstraint interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~string
}

type EntryModel[T any, PK PKConstraint] struct {
	*Model[T]
}

func NewEntryModel[T any, PK PKConstraint](db *db.DB, tableName string) *EntryModel[T, PK] {
	return &EntryModel[T, PK]{NewModel[T](db, tableName)}
}

func (a *EntryModel[T, PK]) FindByPK(id PK) (T, error) {
	q := a.Query()
	pkCol := a.GetPkColumn()
	return q.Where(fmt.Sprintf("`%s` = (?)", pkCol), id).One()
}

// FindByPKWithPreload finds a record by PK with preloaded associations
// Usage: model.FindByPKWithPreload(id, "Profile", "Role")
func (a *EntryModel[T, PK]) FindByPKWithPreload(id PK, preloads ...string) (T, error) {
	q := a.Query()
	for _, p := range preloads {
		q = q.Preload(p)
	}
	pkCol := a.GetPkColumn()
	return q.Where(fmt.Sprintf("`%s` = (?)", pkCol), id).One()
}

func (a *EntryModel[T, PK]) FindOne(query interface{}, args ...interface{}) (T, error) {
	return a.Query().Where(query, args...).One()
}

// FindOneWithPreload finds one record with preloaded associations
// Usage: model.FindOneWithPreload("status = ?", 1, "Profile", "Role")
func (a *EntryModel[T, PK]) FindOneWithPreload(query interface{}, args []interface{}, preloads ...string) (T, error) {
	q := a.Query().Where(query, args...)
	for _, p := range preloads {
		q = q.Preload(p)
	}
	return q.One()
}

func (a *EntryModel[T, PK]) FindAllByPK(id ...PK) ([]T, error) {
	q := a.Query()
	pkCol := a.GetPkColumn()
	return q.Where(fmt.Sprintf("`%s` in (?)", pkCol), id).All()
}

// FindAllByPKWithPreload finds all records by PKs with preloaded associations
// Usage: model.FindAllByPKWithPreload([]uint{1, 2, 3}, "Profile", "Role")
func (a *EntryModel[T, PK]) FindAllByPKWithPreload(ids []PK, preloads ...string) ([]T, error) {
	q := a.Query()
	for _, p := range preloads {
		q = q.Preload(p)
	}
	pkCol := a.GetPkColumn()
	return q.Where(fmt.Sprintf("`%s` in (?)", pkCol), ids).All()
}

func (a *EntryModel[T, PK]) FindAll() ([]T, error) {
	q := a.Query()
	return q.All()
}

// FindAllWithPreload finds all records with preloaded associations
// Usage: model.FindAllWithPreload("Profile", "Role")
func (a *EntryModel[T, PK]) FindAllWithPreload(preloads ...string) ([]T, error) {
	q := a.Query()
	for _, p := range preloads {
		q = q.Preload(p)
	}
	return q.All()
}
func (a *EntryModel[T, PK]) DeleteByPK(id PK) error {
	pkCol := a.GetPkColumn()
	return a.Delete().Where(fmt.Sprintf("`%s` = ?", pkCol), id).Delete()
}

func (a *EntryModel[T, PK]) UpdateByPK(t T) error {
	u := a.Update()
	pkCol := a.GetPkColumn()
	pkValue := getPrimaryKeyValue(t)
	return u.Where(fmt.Sprintf("`%s` = ?", pkCol), pkValue).Update(t)
}
func (a *EntryModel[T, PK]) UpdateColumn(id PK, column string, value interface{}) error {
	u := a.Update()
	pkCol := a.GetPkColumn()
	return u.Where(fmt.Sprintf("`%s` = ?", pkCol), id).UpdateColumn(column, value)
}
func (a *EntryModel[T, PK]) UpdateForMap(id PK, data map[string]interface{}) error {
	u := a.Update()
	pkCol := a.GetPkColumn()
	return u.Where(fmt.Sprintf("`%s` = ?", pkCol), id).UpdateForMap(data)
}

func (a *EntryModel[T, PK]) NewEntryModel(db *db.DB) *EntryModel[T, PK] {
	return &EntryModel[T, PK]{NewModel[T](db, a.tableName)}
}
func (a *EntryModel[T, PK]) Page(page *web.Page) ([]T, int, error) {
	q := a.Query()
	pkCol := a.GetPkColumn()
	return q.Order(fmt.Sprintf("`%s` desc", pkCol)).Page(page)
}
func (a *EntryModel[T, PK]) PageForWeb(page *web.Page) (*web.PageAble[T], error) {
	q := a.Query()
	pkCol := a.GetPkColumn()
	return q.Order(fmt.Sprintf("`%s` desc", pkCol)).PageForWeb(page)
}
func (a *EntryModel[T, PK]) QueryPage(page *web.Page, query interface{}, args ...interface{}) ([]T, int, error) {
	q := a.Query()
	pkCol := a.GetPkColumn()
	return q.Where(query, args...).Order(fmt.Sprintf("`%s` desc", pkCol)).Page(page)
}
func (a *EntryModel[T, PK]) GetTableName() string {
	return a.tableName
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

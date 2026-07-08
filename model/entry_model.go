package model

import (
	"context"
	"fmt"

	"github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/go-web-frame/util"
)

// PKConstraint defines the set of types that can be used as primary keys.
type PKConstraint interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~string
}

// EntryModel extends Model with primary-key-based lookup methods.
// PK is the type of the primary key column.
type EntryModel[T any, PK PKConstraint] struct {
	*Model[T]
}

// NewEntryModel creates a new EntryModel for the given table.
func NewEntryModel[T any, PK PKConstraint](db *db.DB, tableName string) *EntryModel[T, PK] {
	return &EntryModel[T, PK]{NewModel[T](db, tableName)}
}

// WithContext returns a shallow copy of the EntryModel with the given context,
// preserving the correct generic type. The original instance is unchanged.
func (a *EntryModel[T, PK]) WithContext(ctx context.Context) *EntryModel[T, PK] {
	return &EntryModel[T, PK]{Model: a.Model.WithContext(ctx)}
}

// FindByPK finds a single record by its primary key.
func (a *EntryModel[T, PK]) FindByPK(id PK) (T, error) {
	q := a.Query()
	pkCol, err := a.GetPkColumn()
	if err != nil {
		var zero T
		return zero, err
	}
	return q.Where(fmt.Sprintf("`%s` = (?)", pkCol), id).One()
}

// FindByPKWithPreload finds a record by PK with preloaded associations
// Usage: model.FindByPKWithPreload(id, "Profile", "Role")
func (a *EntryModel[T, PK]) FindByPKWithPreload(id PK, preloads ...string) (T, error) {
	q := a.Query()
	for _, p := range preloads {
		q = q.Preload(p)
	}
	pkCol, err := a.GetPkColumn()
	if err != nil {
		var zero T
		return zero, err
	}
	return q.Where(fmt.Sprintf("`%s` = (?)", pkCol), id).One()
}

// FindOne finds a single record matching the given query condition.
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

// FindAllByPK finds all records matching the given primary keys.
func (a *EntryModel[T, PK]) FindAllByPK(id ...PK) ([]T, error) {
	q := a.Query()
	pkCol, err := a.GetPkColumn()
	if err != nil {
		var zero []T
		return zero, err
	}
	return q.Where(fmt.Sprintf("`%s` in (?)", pkCol), id).All()
}

// FindAllByPKWithPreload finds all records by PKs with preloaded associations
// Usage: model.FindAllByPKWithPreload([]uint{1, 2, 3}, "Profile", "Role")
func (a *EntryModel[T, PK]) FindAllByPKWithPreload(ids []PK, preloads ...string) ([]T, error) {
	q := a.Query()
	for _, p := range preloads {
		q = q.Preload(p)
	}
	pkCol, err := a.GetPkColumn()
	if err != nil {
		var zero []T
		return zero, err
	}
	return q.Where(fmt.Sprintf("`%s` in (?)", pkCol), ids).All()
}

// FindAll returns all records in the table.
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
// DeleteByPK deletes a record by its primary key.
func (a *EntryModel[T, PK]) DeleteByPK(id PK) error {
	pkCol, err := a.GetPkColumn()
	if err != nil {
		return err
	}
	return a.Delete().Where(fmt.Sprintf("`%s` = ?", pkCol), id).Delete()
}

// UpdateByPK updates a record using the primary key value embedded in the entity.
func (a *EntryModel[T, PK]) UpdateByPK(t T) error {
	u := a.Update()
	pkCol, err := a.GetPkColumn()
	if err != nil {
		return err
	}
	pkValue := getPrimaryKeyValue(t)
	return u.Where(fmt.Sprintf("`%s` = ?", pkCol), pkValue).Update(t)
}
// UpdateColumn updates a single column value for the record with the given primary key.
func (a *EntryModel[T, PK]) UpdateColumn(id PK, column string, value interface{}) error {
	u := a.Update()
	pkCol, err := a.GetPkColumn()
	if err != nil {
		return err
	}
	return u.Where(fmt.Sprintf("`%s` = ?", pkCol), id).UpdateColumn(column, value)
}
// UpdateForMap updates columns from a map for the record with the given primary key.
func (a *EntryModel[T, PK]) UpdateForMap(id PK, data map[string]interface{}) error {
	u := a.Update()
	pkCol, err := a.GetPkColumn()
	if err != nil {
		return err
	}
	return u.Where(fmt.Sprintf("`%s` = ?", pkCol), id).UpdateForMap(data)
}

// NewEntryModel creates a new EntryModel with the given database connection,
// reusing the same table name.
func (a *EntryModel[T, PK]) NewEntryModel(db *db.DB) *EntryModel[T, PK] {
	return &EntryModel[T, PK]{NewModel[T](db, a.tableName)}
}
// Page returns a paginated list of records ordered by primary key descending.
func (a *EntryModel[T, PK]) Page(page *util.Page) ([]T, int, error) {
	q := a.Query()
	pkCol, err := a.GetPkColumn()
	if err != nil {
		var zero []T
		return zero, 0, err
	}
	return q.Order(fmt.Sprintf("`%s` desc", pkCol)).Page(page)
}
// PageForWeb returns a paginated response suitable for web API responses.
func (a *EntryModel[T, PK]) PageForWeb(page *util.Page) (*util.PageAble[T], error) {
	q := a.Query()
	pkCol, err := a.GetPkColumn()
	if err != nil {
		return nil, err
	}
	return q.Order(fmt.Sprintf("`%s` desc", pkCol)).PageForWeb(page)
}
// QueryPage returns a paginated list matching the given query condition.
func (a *EntryModel[T, PK]) QueryPage(page *util.Page, query interface{}, args ...interface{}) ([]T, int, error) {
	q := a.Query()
	pkCol, err := a.GetPkColumn()
	if err != nil {
		var zero []T
		return zero, 0, err
	}
	return q.Where(query, args...).Order(fmt.Sprintf("`%s` desc", pkCol)).Page(page)
}
// GetTableName returns the database table name for this model.
func (a *EntryModel[T, PK]) GetTableName() string {
	return a.tableName
}

// Transaction wraps database transaction execution.
type Transaction struct {
	db *db.DB
}

// Exec executes the given function within a database transaction.
// The transaction is committed if fc returns nil, rolled back otherwise.
func (t *Transaction) Exec(fc func(tx *db.DB) error) error {
	return t.db.Transaction(fc)
}

// NewTransaction creates a new Transaction with the given database connection.
func NewTransaction(db *db.DB) *Transaction {
	return &Transaction{db: db}
}

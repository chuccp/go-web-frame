package model

import (
	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/go-web-frame/util"
	"github.com/chuccp/go-web-frame/web"
)

type Query[T any] struct {
	tx    *db.Table
	entry T
}

func (q *Query[T]) Where(query interface{}, args ...interface{}) *Query[T] {
	q.tx = q.tx.Where(query, args...)
	return q
}
func (q *Query[T]) Order(query interface{}) *Query[T] {
	q.tx = q.tx.Order(query)
	return q
}
func (q *Query[T]) List(size int) ([]T, error) {
	ts := util.NewSlice(q.entry)
	err := q.tx.Limit(size).Find(&ts)
	return ts, errors.WithStackIf(err)

}
func (q *Query[T]) ListPage(page *web.Page) ([]T, error) {
	ts := util.NewSlice(q.entry)
	err := q.tx.Offset((page.PageNo - 1) * page.PageSize).Limit(page.PageSize).Find(&ts)
	return ts, errors.WithStackIf(err)

}
func (q *Query[T]) All() ([]T, error) {
	ts := util.NewSlice(q.entry)
	err := q.tx.Find(&ts)
	return ts, errors.WithStackIf(err)
}
func (q *Query[T]) One() (T, error) {
	t := util.NewPtr(q.entry)
	err := q.tx.Limit(1).First(&t)
	return t, errors.WithStackIf(err)
}

func (q *Query[T]) Page(page *web.Page) ([]T, int, error) {
	ts := util.NewSlice(q.entry)
	err := q.tx.Offset((page.PageNo - 1) * page.PageSize).Limit(page.PageSize).Find(&ts)
	if err == nil {
		var num int64
		err = q.tx.Count(&num)
		if err == nil {
			return ts, int(num), nil
		}
	}
	return nil, 0, errors.WithStackIf(err)

}

func (q *Query[T]) PageForWeb(page *web.Page) (*web.PageAble[T], error) {
	values, num, err := q.Page(page)
	if err != nil {
		return nil, errors.WithStackIf(err)
	}
	return web.ToPage[T](int64(num), values), nil
}

func (q *Query[T]) Size(size int) ([]T, int, error) {
	ts := util.NewSlice(q.entry)
	err := q.tx.Limit(size).Find(&ts)
	if err == nil {
		var num int64
		err = q.tx.Count(&num)
		if err == nil {
			return ts, int(num), nil
		}
	}
	return nil, 0, errors.WithStackIf(err)
}
func (q *Query[T]) Count() (int, error) {
	var num int64
	err := q.tx.Count(&num)
	return int(num), errors.WithStackIf(err)
}

type where struct {
	query interface{}
	args  []interface{}
}
type UpdateWheres[T any] struct {
	wheres []*where
	tx     *db.Table
}
type UpdateSet struct {
	tx  *db.Table
	set map[string]any
}

func (w *UpdateSet) Set(s string, value any) *UpdateSet {
	w.set[s] = value
	return w
}
func (w *UpdateSet) Exec() error {
	return w.tx.Updates(w.set)
}
func NewUpdateWheres[T any](tx *db.Table) *UpdateWheres[T] {
	return &UpdateWheres[T]{wheres: make([]*where, 0), tx: tx}
}
func (w *UpdateWheres[T]) Where(query interface{}, args ...interface{}) *UpdateWheres[T] {
	w.wheres = append(w.wheres, &where{query: query, args: args})
	return w
}

func (w *UpdateWheres[T]) buildWhere() *db.Table {
	for _, w2 := range w.wheres {
		w.tx = w.tx.Where(w2.query, w2.args...)
	}
	return w.tx
}
func (w *UpdateWheres[T]) UpdateForMap(mapValue map[string]any) error {
	return w.buildWhere().Updates(mapValue)
}

func (w *UpdateWheres[T]) UpdateColumn(column string, value any) error {
	return w.buildWhere().UpdateColumn(column, value)
}

func (w *UpdateWheres[T]) Update(t T) error {
	return w.buildWhere().Updates(t)
}

func (w *UpdateWheres[T]) Set(s string, value any) *UpdateSet {
	set_map := map[string]any{s: value}
	return &UpdateSet{tx: w.buildWhere(), set: set_map}
}

type DeleteWheres[T any] struct {
	wheres []*where
	tx     *db.Table
	entry  T
}

func (w *DeleteWheres[T]) buildWhere() *db.Table {
	for _, w2 := range w.wheres {
		w.tx = w.tx.Where(w2.query, w2.args...)
	}
	return w.tx
}
func (w *DeleteWheres[T]) Delete() error {
	return w.buildWhere().Delete(w.entry)
}
func (w *DeleteWheres[T]) Where(query interface{}, args ...interface{}) *DeleteWheres[T] {
	w.wheres = append(w.wheres, &where{query: query, args: args})
	return w
}
func NewDeleteWheres[T any](tx *db.Table, entry T) *DeleteWheres[T] {
	return &DeleteWheres[T]{wheres: make([]*where, 0), entry: entry, tx: tx}
}

type Update[T any] struct {
	tx     *db.Table
	model  T
	wheres *UpdateWheres[T]
}

func (u *Update[T]) Where(query any, args ...any) *UpdateWheres[T] {
	return u.wheres.Where(query, args...)
}

type Delete[T any] struct {
	tx     *db.Table
	model  T
	wheres *DeleteWheres[T]
}

func (d *Delete[T]) Where(query any, args ...any) *DeleteWheres[T] {
	return d.wheres.Where(query, args...)
}

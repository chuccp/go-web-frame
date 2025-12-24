package model

import (
	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/web"
	"gorm.io/gorm"
)

type Query[T any] struct {
	tx    *gorm.DB
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
	ts := NewSlice(q.entry)
	tx := q.tx.Limit(size).Find(&ts)
	if tx.Error == nil {
		return ts, nil
	}
	return nil, errors.WithStackIf(tx.Error)

}
func (q *Query[T]) ListPage(page *web.Page) ([]T, error) {
	ts := NewSlice(q.entry)
	tx := q.tx.Offset((page.PageNo - 1) * page.PageSize).Limit(page.PageSize).Find(&ts)
	if tx.Error == nil {
		return ts, nil
	}
	return nil, errors.WithStackIf(tx.Error)

}
func (q *Query[T]) All() ([]T, error) {
	ts := NewSlice(q.entry)
	tx := q.tx.Find(&ts)
	if tx.Error == nil {
		return ts, nil
	}
	return nil, errors.WithStackIf(tx.Error)
}
func (q *Query[T]) One() (T, error) {
	t := NewPtr(q.entry)
	tx := q.tx.Limit(1).First(&t)
	if tx.Error == nil {
		return t, nil
	}
	return t, errors.WithStackIf(tx.Error)
}

func (q *Query[T]) Page(page *web.Page) ([]T, int, error) {
	ts := NewSlice(q.entry)
	tx := q.tx.Offset((page.PageNo - 1) * page.PageSize).Limit(page.PageSize).Find(&ts)
	if tx.Error == nil {
		var num int64
		tx := q.tx.Count(&num)
		if tx.Error == nil {
			return ts, int(num), nil
		}
	}
	return nil, 0, errors.WithStackIf(tx.Error)

}
func (q *Query[T]) Size(size int) ([]T, int, error) {
	ts := NewSlice(q.entry)
	tx := q.tx.Limit(size).Find(&ts)
	if tx.Error == nil {
		var num int64
		tx := q.tx.Count(&num)
		if tx.Error == nil {
			return ts, int(num), nil
		}
	}
	return nil, 0, errors.WithStackIf(tx.Error)
}

type where struct {
	query interface{}
	args  []interface{}
}
type UpdateWheres[T any] struct {
	wheres []*where
	tx     *gorm.DB
}

func NewUpdateWheres[T any](tx *gorm.DB) *UpdateWheres[T] {
	return &UpdateWheres[T]{wheres: make([]*where, 0), tx: tx}
}
func (w *UpdateWheres[T]) Where(query interface{}, args ...interface{}) *UpdateWheres[T] {
	w.wheres = append(w.wheres, &where{query: query, args: args})
	return w
}

func (w *UpdateWheres[T]) buildWhere() *gorm.DB {
	for _, w2 := range w.wheres {
		w.tx = w.tx.Where(w2.query, w2.args...)
	}
	return w.tx
}
func (w *UpdateWheres[T]) UpdateForMap(mapValue map[string]any) error {
	return w.buildWhere().Updates(mapValue).Error
}

func (w *UpdateWheres[T]) UpdateColumn(column string, value any) error {
	return w.buildWhere().UpdateColumn(column, value).Error
}

func (w *UpdateWheres[T]) Update(t T) error {
	return w.buildWhere().Updates(t).Error
}

type DeleteWheres[T any] struct {
	wheres []*where
	tx     *gorm.DB
	entry  T
}

func (w *DeleteWheres[T]) buildWhere() *gorm.DB {
	for _, w2 := range w.wheres {
		w.tx = w.tx.Where(w2.query, w2.args...)
	}
	return w.tx
}
func (w *DeleteWheres[T]) Delete() error {
	return w.buildWhere().Delete(w.entry).Error
}
func (w *DeleteWheres[T]) Where(query interface{}, args ...interface{}) *DeleteWheres[T] {
	w.wheres = append(w.wheres, &where{query: query, args: args})
	return w
}
func NewDeleteWheres[T any](tx *gorm.DB, entry T) *DeleteWheres[T] {
	return &DeleteWheres[T]{wheres: make([]*where, 0), entry: entry, tx: tx}
}

type Update[T any] struct {
	tx     *gorm.DB
	model  T
	wheres *UpdateWheres[T]
}

func (u *Update[T]) Where(query any, args ...any) *UpdateWheres[T] {
	return u.wheres.Where(query, args...)
}

type Delete[T any] struct {
	tx     *gorm.DB
	model  T
	wheres *DeleteWheres[T]
}

func (d *Delete[T]) Where(query any, args ...any) *DeleteWheres[T] {
	return d.wheres.Where(query, args...)
}

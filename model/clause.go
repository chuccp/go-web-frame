package model

import (
	"strings"

	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/go-web-frame/util"
	"github.com/chuccp/go-web-frame/web"
)

type Query[T any] struct {
	db        *db.DB
	tableName string
	entry     T
	wheres    []*where
	orders    []interface{}
}

// 构建tx实例，惰性创建，仅在实际执行时调用
func (q *Query[T]) buildTx() (*db.Table, error) {
	if q.db == nil {
		return nil, errors.New("db is nil")
	}
	tx := q.db.Table(q.tableName)
	for _, w := range q.wheres {
		tx = tx.Where(w.query, w.args...)
	}
	for _, o := range q.orders {
		tx = tx.Order(o)
	}
	return tx, nil
}

func (q *Query[T]) Where(query interface{}, args ...interface{}) *Query[T] {
	q.wheres = append(q.wheres, &where{query: query, args: args})
	return q
}
func (q *Query[T]) Order(query interface{}) *Query[T] {
	q.orders = append(q.orders, query)
	return q
}
func (q *Query[T]) List(size int) ([]T, error) {
	ts := util.NewSlice(q.entry)
	tx, err := q.buildTx()
	if err != nil {
		return nil, errors.WithStackIf(err)
	}
	err = tx.Limit(size).Find(&ts)
	return ts, errors.WithStackIf(err)

}
func (q *Query[T]) ListPage(page *web.Page) ([]T, error) {
	ts := util.NewSlice(q.entry)
	tx, err := q.buildTx()
	if err != nil {
		return nil, errors.WithStackIf(err)
	}
	err = tx.Offset((page.PageNo - 1) * page.PageSize).Limit(page.PageSize).Find(&ts)
	return ts, errors.WithStackIf(err)

}
func (q *Query[T]) All() ([]T, error) {
	ts := util.NewSlice(q.entry)
	tx, err := q.buildTx()
	if err != nil {
		return nil, errors.WithStackIf(err)
	}
	err = tx.Find(&ts)
	return ts, errors.WithStackIf(err)
}
func (q *Query[T]) One() (T, error) {
	t := util.NewPtr(q.entry)
	tx, err := q.buildTx()
	if err != nil {
		return t, errors.WithStackIf(err)
	}
	err = tx.Limit(1).First(&t)
	return t, errors.WithStackIf(err)
}

func (q *Query[T]) Exec(sql string, args ...interface{}) ([]T, error) {
	ts := util.NewSlice(q.entry)
	if q.db == nil {
		return nil, errors.New("db is nil")
	}
	tx := q.db.Table(q.tableName)
	err := tx.Raw(sql, args...).Scan(&ts)
	return ts, errors.WithStackIf(err)
}
func (q *Query[T]) ExecPage(sql string, args ...interface{}) ([]T, int, error) {
	ts := util.NewSlice(q.entry)
	if q.db == nil {
		return nil, 0, errors.New("db is nil")
	}
	tx := q.db.Table(q.tableName)
	err := tx.Raw(sql, args...).Scan(&ts)
	if err != nil {
		return nil, 0, errors.WithStackIf(err)
	}
	var num int64
	countSql := toCountSql(sql)
	tx2 := q.db.Table(q.tableName)
	err = tx2.Raw(countSql, args...).Scan(&num)
	if err != nil {
		return nil, 0, errors.WithStackIf(err)
	}
	return ts, int(num), nil
}

// toCountSql 将查询 SQL 转换为 COUNT SQL
// 例如: SELECT * FROM users WHERE status = ? LIMIT 10 -> SELECT COUNT(*) FROM users WHERE status = ?
func toCountSql(sql string) string {
	upper := strings.ToUpper(sql)
	// 找到 SELECT 和 FROM 的位置
	selectIdx := strings.Index(upper, "SELECT")
	fromIdx := strings.Index(upper, "FROM")
	if selectIdx == -1 || fromIdx == -1 || fromIdx <= selectIdx {
		return sql
	}

	// 构建基础 count SQL
	countSql := "SELECT COUNT(*) " + sql[fromIdx:]

	// 移除 LIMIT, OFFSET, ORDER BY 子句
	upper = strings.ToUpper(countSql)
	keywords := []string{" LIMIT", " OFFSET", " ORDER BY"}
	for _, kw := range keywords {
		if idx := strings.Index(upper, kw); idx != -1 {
			countSql = countSql[:idx]
			upper = strings.ToUpper(countSql)
		}
	}
	return countSql
}

func (q *Query[T]) Page(page *web.Page) ([]T, int, error) {
	ts := util.NewSlice(q.entry)
	tx, err := q.buildTx()
	if err != nil {
		return nil, 0, errors.WithStackIf(err)
	}
	err = tx.Offset((page.PageNo - 1) * page.PageSize).Limit(page.PageSize).Find(&ts)
	if err == nil {
		var num int64
		err = tx.NewTable().Count(&num)
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
	tx, err := q.buildTx()
	if err != nil {
		return nil, 0, errors.WithStackIf(err)
	}
	err = tx.Limit(size).Find(&ts)
	if err == nil {
		var num int64
		err = tx.NewTable().Count(&num)
		if err == nil {
			return ts, int(num), nil
		}
	}
	return nil, 0, errors.WithStackIf(err)
}
func (q *Query[T]) Count() (int, error) {
	var num int64
	tx, err := q.buildTx()
	if err != nil {
		return 0, errors.WithStackIf(err)
	}
	err = tx.Count(&num)
	return int(num), errors.WithStackIf(err)
}

type where struct {
	query interface{}
	args  []interface{}
}
type Update[T any] struct {
	db        *db.DB
	tableName string
	model     T
	wheres    []*where
}

// 构建tx实例，惰性创建
func (u *Update[T]) buildTx() (*db.Table, error) {
	if u.db == nil {
		return nil, errors.New("db is nil")
	}
	tx := u.db.Table(u.tableName)
	for _, w := range u.wheres {
		tx = tx.Where(w.query, w.args...)
	}
	return tx, nil
}

func (u *Update[T]) Where(query interface{}, args ...interface{}) *Update[T] {
	u.wheres = append(u.wheres, &where{query: query, args: args})
	return u
}

func (u *Update[T]) UpdateForMap(mapValue map[string]any) error {
	tx, err := u.buildTx()
	if err != nil {
		return errors.WithStackIf(err)
	}
	return tx.Updates(mapValue)
}

func (u *Update[T]) UpdateColumn(column string, value any) error {
	tx, err := u.buildTx()
	if err != nil {
		return errors.WithStackIf(err)
	}
	return tx.UpdateColumn(column, value)
}

func (u *Update[T]) Update(t T) error {
	tx, err := u.buildTx()
	if err != nil {
		return errors.WithStackIf(err)
	}
	return tx.Updates(t)
}

func (u *Update[T]) Set(s string, value any) *UpdateSet {
	tx, err := u.buildTx()
	if err != nil {
		return &UpdateSet{err: errors.WithStackIf(err)}
	}
	setMap := map[string]any{s: value}
	return &UpdateSet{tx: tx, set: setMap}
}

type UpdateSet struct {
	tx  *db.Table
	set map[string]any
	err error
}

func (w *UpdateSet) Set(s string, value any) *UpdateSet {
	if w.err != nil {
		return w
	}
	w.set[s] = value
	return w
}
func (w *UpdateSet) Exec() error {
	if w.err != nil {
		return w.err
	}
	return w.tx.Updates(w.set)
}

type Delete[T any] struct {
	db        *db.DB
	tableName string
	model     T
	wheres    []*where
}

// 构建tx实例，惰性创建
func (d *Delete[T]) buildTx() (*db.Table, error) {
	if d.db == nil {
		return nil, errors.New("db is nil")
	}
	tx := d.db.Table(d.tableName)
	for _, w := range d.wheres {
		tx = tx.Where(w.query, w.args...)
	}
	return tx, nil
}

func (d *Delete[T]) Where(query interface{}, args ...interface{}) *Delete[T] {
	d.wheres = append(d.wheres, &where{query: query, args: args})
	return d
}

func (d *Delete[T]) Delete() error {
	tx, err := d.buildTx()
	if err != nil {
		return errors.WithStackIf(err)
	}
	return tx.Delete(d.model)
}

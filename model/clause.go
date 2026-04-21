package model

import (
	"fmt"
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
	preloads  []string
	joins     []string
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
	for _, p := range q.preloads {
		tx = tx.Preload(p)
	}
	for _, j := range q.joins {
		tx = tx.Joins(j)
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

// Preload adds a preload clause for eager loading associations (GORM foreign key support)
// Usage: query.Preload("Profile").Preload("Role").All()
// Supports nested preloading: query.Preload("Profile.Addresses")
func (q *Query[T]) Preload(query string) *Query[T] {
	q.preloads = append(q.preloads, query)
	return q
}

// Joins adds a join clause for association loading (GORM join support)
// Usage: query.Joins("Profile").All()
func (q *Query[T]) Joins(query string) *Query[T] {
	q.joins = append(q.joins, query)
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

	// 合并 Where 条件
	finalSql, finalArgs := q.mergeWheres(sql, args...)

	tx := q.db.Table(q.tableName)
	err := tx.Raw(finalSql, finalArgs...).Scan(&ts)
	return ts, errors.WithStackIf(err)
}

// mergeWheres 合并 Where 条件到 SQL 和参数中
func (q *Query[T]) mergeWheres(sql string, args ...interface{}) (string, []interface{}) {
	finalSql := sql
	finalArgs := args

	if len(q.wheres) > 0 {
		// 检查 SQL 是否已有 WHERE 子句
		upperSql := strings.ToUpper(sql)
		hasWhere := strings.Contains(upperSql, " WHERE ")

		// 构建 WHERE 子句
		var whereClause strings.Builder
		whereArgs := make([]interface{}, 0)

		for i, w := range q.wheres {
			if i > 0 {
				whereClause.WriteString(" AND ")
			}
			whereClause.WriteString(fmt.Sprintf("%v", w.query))
			whereArgs = append(whereArgs, w.args...)
		}

		// 追加到 SQL
		if hasWhere {
			finalSql = sql + " AND " + whereClause.String()
		} else {
			finalSql = sql + " WHERE " + whereClause.String()
		}
		finalArgs = append(finalArgs, whereArgs...)
	}

	return finalSql, finalArgs
}

func (q *Query[T]) ExecPage(page *web.Page, sql string, args ...interface{}) ([]T, int, error) {
	ts := util.NewSlice(q.entry)
	if q.db == nil {
		return nil, 0, errors.New("db is nil")
	}

	// 合并 Where 条件
	finalSql, finalArgs := q.mergeWheres(sql, args...)

	// 添加 ORDER BY（如果有多个排序字段）
	if len(q.orders) > 0 {
		orderParts := make([]string, len(q.orders))
		for i, o := range q.orders {
			orderParts[i] = fmt.Sprintf("%v", o)
		}
		finalSql = finalSql + " ORDER BY " + strings.Join(orderParts, ", ")
	}

	// 构建 count SQL
	countSql := toCountSql(finalSql)
	var num int64
	tx := q.db.Table(q.tableName)
	err := tx.Raw(countSql, finalArgs...).Scan(&num)
	if err != nil {
		return nil, 0, errors.WithStackIf(err)
	}

	// 构建分页 SQL
	pageSql := finalSql + fmt.Sprintf(" LIMIT %d OFFSET %d", page.PageSize, (page.PageNo-1)*page.PageSize)
	tx2 := q.db.Table(q.tableName)
	err = tx2.Raw(pageSql, finalArgs...).Scan(&ts)
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

	// 构建基础 count SQL，确保 FROM 前有空格
	countSql := "SELECT COUNT(*) FROM" + sql[fromIdx+4:]

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
		// 重新构建带有相同 WHERE 条件的查询来统计总数
		countTx, err := q.buildTx()
		if err != nil {
			return nil, 0, errors.WithStackIf(err)
		}
		err = countTx.Count(&num)
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
		// 重新构建带有相同 WHERE 条件的查询来统计总数
		countTx, err := q.buildTx()
		if err != nil {
			return nil, 0, errors.WithStackIf(err)
		}
		err = countTx.Count(&num)
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
package model

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/go-web-frame/util"
	"gorm.io/gorm"
)

// Query is a type-safe query builder for constructing read operations on type T.
type Query[T any] struct {
	db        *db.DB
	tableName string
	entry     T
	wheres    []*where
	orders    []any
	preloads  []string
	joins     []string
	selects   []any
	having    []*where
	groups    []string
}

// buildTx constructs the database transaction, lazily created on first execution.
// buildTxWith builds a query using the given DB connection (e.g. within a transaction).
func (q *Query[T]) buildTxWith(d *db.DB) (*db.Table, error) {
	if d == nil {
		return nil, errors.New("db is nil")
	}
	tx := d.Table(q.tableName)
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
	for _, s := range q.selects {
		tx = tx.Select(s)
	}
	for _, g := range q.groups {
		tx = tx.Group(g)
	}
	for _, h := range q.having {
		tx = tx.Having(h.query, h.args...)
	}
	return tx, nil
}

// buildTx builds a query using the Query's DB connection.
func (q *Query[T]) buildTx() (*db.Table, error) {
	return q.buildTxWith(q.db)
}

// Where adds a WHERE condition to the query.
func (q *Query[T]) Where(query any, args ...any) *Query[T] {
	q.wheres = append(q.wheres, &where{query: query, args: args})
	return q
}

// Order adds an ORDER BY clause to the query.
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

// Select specifies select columns for the query.
// Usage: query.Select("name, age").All()
func (q *Query[T]) Select(query any, args ...any) *Query[T] {
	q.selects = append(q.selects, query)
	return q
}

// Group adds a GROUP BY clause to the query.
// Usage: query.Group("category").All()
func (q *Query[T]) Group(name string) *Query[T] {
	q.groups = append(q.groups, name)
	return q
}

// Having adds a HAVING condition to the query (used with Group).
// Usage: query.Group("category").Having("COUNT(*) > ?", 5).All()
func (q *Query[T]) Having(query any, args ...any) *Query[T] {
	q.having = append(q.having, &where{query: query, args: args})
	return q
}

// List returns up to size records matching the query conditions.
func (q *Query[T]) List(size int) ([]T, error) {
	ts := util.NewSlice(q.entry)
	tx, err := q.buildTx()
	if err != nil {
		return nil, errors.WithStackIf(err)
	}
	err = tx.Limit(size).Find(&ts)
	return ts, errors.WithStackIf(err)

}

// ListPage returns a page of records matching the query conditions.
func (q *Query[T]) ListPage(page *util.Page) ([]T, error) {
	if page == nil || page.PageNo < 1 || page.PageSize < 1 {
		return nil, errors.New("invalid page parameters: PageNo and PageSize must be >= 1")
	}
	ts := util.NewSlice(q.entry)
	tx, err := q.buildTx()
	if err != nil {
		return nil, errors.WithStackIf(err)
	}
	err = tx.Offset((page.PageNo - 1) * page.PageSize).Limit(page.PageSize).Find(&ts)
	return ts, errors.WithStackIf(err)

}

// All returns all records matching the query conditions.
func (q *Query[T]) All() ([]T, error) {
	ts := util.NewSlice(q.entry)
	tx, err := q.buildTx()
	if err != nil {
		return nil, errors.WithStackIf(err)
	}
	err = tx.Find(&ts)
	return ts, errors.WithStackIf(err)
}

// One returns a single record matching the query conditions.
func (q *Query[T]) One() (T, error) {
	t := util.NewPtr(q.entry)
	tx, err := q.buildTx()
	if err != nil {
		return t, errors.WithStackIf(err)
	}
	err = tx.Limit(1).First(&t)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		var t T
		return t, errors.WithStackIf(err)
	}
	return t, errors.WithStackIf(err)
}

// Exec executes a raw SQL query with the accumulated WHERE conditions.
func (q *Query[T]) Exec(sql string, args ...interface{}) ([]T, error) {
	ts := util.NewSlice(q.entry)
	if q.db == nil {
		return nil, errors.New("db is nil")
	}

	// Merge Where conditions
	finalSql, finalArgs := q.mergeWheres(sql, args...)

	tx := q.db.Table(q.tableName)
	err := tx.Raw(finalSql, finalArgs...).Scan(&ts)
	return ts, errors.WithStackIf(err)
}

// mergeWheres merges Where conditions into the SQL and arguments.
func (q *Query[T]) mergeWheres(sql string, args ...interface{}) (string, []interface{}) {
	finalSql := sql
	finalArgs := args

	if len(q.wheres) > 0 {
		// Check if SQL already has a WHERE clause
		upperSql := strings.ToUpper(sql)
		hasWhere := whereRe.MatchString(upperSql)

		// Build WHERE clause
		var whereClause strings.Builder
		whereArgs := make([]interface{}, 0)

		for i, w := range q.wheres {
			if i > 0 {
				whereClause.WriteString(" AND ")
			}
			whereClause.WriteString(fmt.Sprintf("%v", w.query))
			whereArgs = append(whereArgs, w.args...)
		}

		// Append to SQL
		if hasWhere {
			finalSql = sql + " AND " + whereClause.String()
		} else {
			finalSql = sql + " WHERE " + whereClause.String()
		}
		finalArgs = append(finalArgs, whereArgs...)
	}

	return finalSql, finalArgs
}

// ExecPage executes a paginated raw SQL query with the accumulated WHERE conditions.
//
// Deprecated: This method is not recommended for use. Use Page() or PageForWeb() instead.
func (q *Query[T]) ExecPage(page *util.Page, sql string, args ...interface{}) ([]T, int, error) {
	ts := util.NewSlice(q.entry)
	if q.db == nil {
		return nil, 0, errors.New("db is nil")
	}
	if page == nil || page.PageNo < 1 || page.PageSize < 1 {
		return nil, 0, errors.New("invalid page parameters: PageNo and PageSize must be >= 1")
	}

	// Merge Where conditions
	finalSql, finalArgs := q.mergeWheres(sql, args...)

	// Add ORDER BY if there are ordering fields
	if len(q.orders) > 0 {
		orderParts := make([]string, len(q.orders))
		for i, o := range q.orders {
			orderParts[i] = fmt.Sprintf("%v", o)
		}
		finalSql = finalSql + " ORDER BY " + strings.Join(orderParts, ", ")
	}

	// Build count SQL
	countSql := toCountSql(finalSql)
	var num int64
	tx := q.db.Table(q.tableName)
	err := tx.Raw(countSql, finalArgs...).Scan(&num)
	if err != nil {
		return nil, 0, errors.WithStackIf(err)
	}

	// Build paginated SQL
	pageSql := finalSql + fmt.Sprintf(" LIMIT %d OFFSET %d", page.PageSize, (page.PageNo-1)*page.PageSize)
	tx2 := q.db.Table(q.tableName)
	err = tx2.Raw(pageSql, finalArgs...).Scan(&ts)
	if err != nil {
		return nil, 0, errors.WithStackIf(err)
	}
	return ts, int(num), nil
}

// toCountSql converts a SELECT query to a COUNT query using a subquery.
// Works correctly with DISTINCT, subqueries, CTEs, and complex SQL.
// Example: SELECT DISTINCT name FROM users WHERE status = ? ORDER BY id LIMIT 10
//
//	-> SELECT COUNT(*) FROM (SELECT DISTINCT name FROM users WHERE status = ?) AS _t
func toCountSql(sql string) string {
	return "SELECT COUNT(*) FROM (" + sql + ") AS _t"
}

// Page returns a paginated list with total count.
// Both queries run in a single transaction for consistency.
func (q *Query[T]) Page(page *util.Page) ([]T, int, error) {
	if page == nil || page.PageNo < 1 || page.PageSize < 1 {
		return nil, 0, errors.New("invalid page parameters: PageNo and PageSize must be >= 1")
	}
	ts := util.NewSlice(q.entry)
	var num int64

	err := q.db.Transaction(func(txDB *db.DB) error {
		// Data query
		dataTx, err := q.buildTxWith(txDB)
		if err != nil {
			return err
		}
		if err := dataTx.Offset((page.PageNo - 1) * page.PageSize).Limit(page.PageSize).Find(&ts); err != nil {
			return err
		}
		// Count query (same transaction snapshot)
		countTx, err := q.buildTxWith(txDB)
		if err != nil {
			return err
		}
		return countTx.Count(&num)
	})
	if err != nil {
		return nil, 0, errors.WithStackIf(err)
	}
	return ts, int(num), nil
}

// PageForWeb returns a paginated web response suitable for API responses.
func (q *Query[T]) PageForWeb(page *util.Page) (*util.PageAble[T], error) {
	values, num, err := q.Page(page)
	if err != nil {
		return nil, errors.WithStackIf(err)
	}
	return util.ToPage[T](int64(num), values), nil
}

// Size returns up to size records with total count.
func (q *Query[T]) Size(size int) ([]T, int, error) {
	ts := util.NewSlice(q.entry)
	tx, err := q.buildTx()
	if err != nil {
		return nil, 0, errors.WithStackIf(err)
	}
	err = tx.Limit(size).Find(&ts)
	if err == nil {
		var num int64
		// Re-build query with same WHERE conditions to count total
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

// Count returns the number of records matching the query conditions.
func (q *Query[T]) Count() (int, error) {
	var num int64
	tx, err := q.buildTx()
	if err != nil {
		return 0, errors.WithStackIf(err)
	}
	err = tx.Count(&num)
	return int(num), errors.WithStackIf(err)
}

// WithContext returns a shallow copy of the query builder with the given context.
// The original instance is unchanged; safe for concurrent use.
func (q *Query[T]) WithContext(ctx context.Context) *Query[T] {
	return &Query[T]{
		db:        q.db.WithContext(ctx),
		tableName: q.tableName,
		entry:     q.entry,
		wheres:    q.wheres,
		orders:    q.orders,
		preloads:  q.preloads,
		joins:     q.joins,
		selects:   q.selects,
		having:    q.having,
		groups:    q.groups,
	}
}

// Aggregate is a type-safe builder for constructing aggregate queries on type T.
// T is the entity type used for table name resolution; result is scanned into a separate type.
type Aggregate[T any] struct {
	db        *db.DB
	tableName string
	wheres    []*where
	orders    []any
	groups    []string
	having    []*where
	selects   []any
	joins     []string
}

// buildTx constructs the database transaction for the aggregate query.
func (a *Aggregate[T]) buildTx() (*db.Table, error) {
	if a.db == nil {
		return nil, errors.New("db is nil")
	}
	tx := a.db.Table(a.tableName)
	for _, w := range a.wheres {
		tx = tx.Where(w.query, w.args...)
	}
	for _, j := range a.joins {
		tx = tx.Joins(j)
	}
	for _, s := range a.selects {
		tx = tx.Select(s)
	}
	for _, g := range a.groups {
		tx = tx.Group(g)
	}
	for _, h := range a.having {
		tx = tx.Having(h.query, h.args...)
	}
	for _, o := range a.orders {
		tx = tx.Order(o)
	}
	return tx, nil
}

// Where adds a WHERE condition to the aggregate query.
func (a *Aggregate[T]) Where(query any, args ...any) *Aggregate[T] {
	a.wheres = append(a.wheres, &where{query: query, args: args})
	return a
}

// Order adds an ORDER BY clause to the aggregate query.
func (a *Aggregate[T]) Order(query any) *Aggregate[T] {
	a.orders = append(a.orders, query)
	return a
}

// Select specifies select columns or aggregate expressions for the query.
// Usage: a.Select("category, SUM(amount) as total").Aggregate(&results)
func (a *Aggregate[T]) Select(query any, args ...any) *Aggregate[T] {
	a.selects = append(a.selects, query)
	return a
}

// Group adds a GROUP BY clause to the aggregate query.
// Usage: a.Group("category").Aggregate(&results)
func (a *Aggregate[T]) Group(name string) *Aggregate[T] {
	a.groups = append(a.groups, name)
	return a
}

// Having adds a HAVING condition to the aggregate query (used with Group).
// Usage: a.Group("category").Having("COUNT(*) > ?", 5).Aggregate(&results)
func (a *Aggregate[T]) Having(query any, args ...any) *Aggregate[T] {
	a.having = append(a.having, &where{query: query, args: args})
	return a
}

// Joins adds a join clause to the aggregate query.
func (a *Aggregate[T]) Joins(query string) *Aggregate[T] {
	a.joins = append(a.joins, query)
	return a
}

// Aggregate executes the aggregate query and scans the result into the provided pointer.
// For scalar results (e.g. *float64), a LIMIT 1 is applied automatically.
// For slice results (e.g. *[]T), all matching rows are returned.
//
// Usage:
//
//	var total float64
//	err := model.Aggregate().Select("SUM(amount)").Where("status = ?", 1).Aggregate(&total)
//
//	var stats []CategoryStat
//	err := model.Aggregate().Select("category, SUM(amount) as total").Group("category").Aggregate(&stats)
func (a *Aggregate[T]) Aggregate(result any) error {
	tx, err := a.buildTx()
	if err != nil {
		return errors.WithStackIf(err)
	}
	if isScalarResult(result) {
		return errors.WithStackIf(tx.Limit(1).Scan(result))
	}
	return errors.WithStackIf(tx.Find(result))
}

// WithContext returns a shallow copy of the aggregate builder with the given context.
// The original instance is unchanged; safe for concurrent use.
func (a *Aggregate[T]) WithContext(ctx context.Context) *Aggregate[T] {
	var newDB *db.DB
	if a.db != nil {
		newDB = a.db.WithContext(ctx)
	}
	return &Aggregate[T]{
		db:        newDB,
		tableName: a.tableName,
		wheres:    a.wheres,
		orders:    a.orders,
		groups:    a.groups,
		having:    a.having,
		selects:   a.selects,
		joins:     a.joins,
	}
}

// isScalarResult returns true if result is a pointer to a non-slice type
// (e.g. *float64, *int, *string, *struct{...}).
func isScalarResult(result any) bool {
	if result == nil {
		return false
	}
	rv := reflect.ValueOf(result)
	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		rv = rv.Elem()
	}
	return rv.Kind() != reflect.Slice
}

var whereRe = regexp.MustCompile(`(?i)\sWHERE\s`)

type where struct {
	query any
	args  []any
}

// Update is a type-safe update builder for constructing update operations on type T.
type Update[T any] struct {
	db        *db.DB
	tableName string
	model     T
	wheres    []*where
}

// buildTx constructs the database transaction, lazily created.
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

// Where adds a WHERE condition to the update operation.
func (u *Update[T]) Where(query interface{}, args ...interface{}) *Update[T] {
	u.wheres = append(u.wheres, &where{query: query, args: args})
	return u
}

// UpdateForMap updates records matching the WHERE conditions using a column-value map.
func (u *Update[T]) UpdateForMap(mapValue map[string]any) error {
	tx, err := u.buildTx()
	if err != nil {
		return errors.WithStackIf(err)
	}
	return tx.Updates(mapValue)
}

// UpdateColumn updates a single column for records matching the WHERE conditions.
func (u *Update[T]) UpdateColumn(column string, value any) error {
	tx, err := u.buildTx()
	if err != nil {
		return errors.WithStackIf(err)
	}
	return tx.UpdateColumn(column, value)
}

// Update applies the entity values to records matching the WHERE conditions.
func (u *Update[T]) Update(t T) error {
	tx, err := u.buildTx()
	if err != nil {
		return errors.WithStackIf(err)
	}
	return tx.Updates(t)
}

// Set begins a chain of column-value updates.
func (u *Update[T]) Set(s string, value any) *UpdateSet {
	tx, err := u.buildTx()
	if err != nil {
		return &UpdateSet{err: errors.WithStackIf(err)}
	}
	setMap := map[string]any{s: value}
	return &UpdateSet{tx: tx, set: setMap}
}

// WithContext returns a shallow copy of the update builder with the given context.
// The original instance is unchanged; safe for concurrent use.
func (u *Update[T]) WithContext(ctx context.Context) *Update[T] {
	return &Update[T]{
		db:        u.db.WithContext(ctx),
		tableName: u.tableName,
		model:     u.model,
		wheres:    u.wheres,
	}
}

// UpdateSet accumulates column-value pairs for a batch update.
type UpdateSet struct {
	tx  *db.Table
	set map[string]any
	err error
}

// Set adds a column-value pair to the update set.
func (w *UpdateSet) Set(s string, value any) *UpdateSet {
	if w.err != nil {
		return w
	}
	w.set[s] = value
	return w
}

// Exec executes the accumulated updates.
func (w *UpdateSet) Exec() error {
	if w.err != nil {
		return w.err
	}
	return w.tx.Updates(w.set)
}

// Delete is a type-safe delete builder for constructing delete operations on type T.
type Delete[T any] struct {
	db        *db.DB
	tableName string
	model     T
	wheres    []*where
}

// buildTx constructs the database transaction, lazily created.
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

// Where adds a WHERE condition to the delete operation.
func (d *Delete[T]) Where(query interface{}, args ...interface{}) *Delete[T] {
	d.wheres = append(d.wheres, &where{query: query, args: args})
	return d
}

// Delete executes the delete operation for records matching the WHERE conditions.
func (d *Delete[T]) Delete() error {
	tx, err := d.buildTx()
	if err != nil {
		return errors.WithStackIf(err)
	}
	return tx.Delete(d.model)
}

// WithContext returns a shallow copy of the delete builder with the given context.
// The original instance is unchanged; safe for concurrent use.
func (d *Delete[T]) WithContext(ctx context.Context) *Delete[T] {
	return &Delete[T]{
		db:        d.db.WithContext(ctx),
		tableName: d.tableName,
		model:     d.model,
		wheres:    d.wheres,
	}
}

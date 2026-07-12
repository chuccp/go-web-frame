package model

import (
	"context"
	"testing"
	"time"

	"github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/go-web-frame/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- test entities ----

type TestEntity struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"size:255" json:"name"`
	Age       int       `json:"age"`
	Status    int       `gorm:"default:1" json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type TestStringPK struct {
	Code string `gorm:"primaryKey;size:64"`
	Val  string `gorm:"size:255"`
}

// ---- helpers ----

func setupDB(t *testing.T) *db.DB {
	t.Helper()
	d, err := db.ConnectionSQLite(":memory:")
	require.NoError(t, err)
	return d
}

func setupModel[T any](t *testing.T, d *db.DB, table string) *Model[T] {
	t.Helper()
	m := NewModel[T](d, table)
	require.NoError(t, m.CreateTable())
	t.Cleanup(func() { _ = m.DropTable() })
	return m
}

func setupEntryModel[T any, PK PKConstraint](t *testing.T, d *db.DB, table string) *EntryModel[T, PK] {
	t.Helper()
	m := NewEntryModel[T, PK](d, table)
	require.NoError(t, m.CreateTable())
	t.Cleanup(func() { _ = m.DropTable() })
	return m
}

// ---- nil-db constructor tests (builder-only, no SQL) ----

func TestNewModel(t *testing.T) {
	model := NewModel[*TestEntity](nil, "t_test")
	assert.NotNil(t, model)
	assert.Equal(t, "t_test", model.GetTableName())
}

func TestModel_TableName(t *testing.T) {
	model := NewModel[*TestEntity](nil, "custom_table")
	assert.Equal(t, "custom_table", model.GetTableName())
}

func TestModel_DefaultTableName(t *testing.T) {
	model := NewModel[*TestEntity](nil, "")
	assert.Equal(t, "", model.GetTableName())
}

func TestModel_Query_BuilderOnly(t *testing.T) {
	model := NewModel[*TestEntity](nil, "t_test")
	query := model.Query()
	assert.Equal(t, "t_test", query.tableName)
}

func TestModel_Update_BuilderOnly(t *testing.T) {
	model := NewModel[*TestEntity](nil, "t_test")
	update := model.Update()
	assert.Equal(t, "t_test", update.tableName)
}

func TestModel_Delete_BuilderOnly(t *testing.T) {
	model := NewModel[*TestEntity](nil, "t_test")
	del := model.Delete()
	assert.Equal(t, "t_test", del.tableName)
}

// ---- SQLite CRUD ----

func TestModel_Save(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestEntity](t, d, "t_test")

	err := m.Save(&TestEntity{Name: "alice", Age: 25})
	require.NoError(t, err)
}

func TestModel_QueryAll(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestEntity](t, d, "t_test")

	require.NoError(t, m.Save(&TestEntity{Name: "alice", Age: 25}))
	require.NoError(t, m.Save(&TestEntity{Name: "bob", Age: 30}))

	users, err := m.Query().All()
	require.NoError(t, err)
	assert.Len(t, users, 2)
}

func TestModel_QueryOne(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestEntity](t, d, "t_test")

	require.NoError(t, m.Save(&TestEntity{Name: "alice", Age: 25}))

	user, err := m.Query().Where("name = ?", "alice").One()
	require.NoError(t, err)
	assert.Equal(t, "alice", user.Name)
	assert.Equal(t, 25, user.Age)
}

func TestModel_QueryCount(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestEntity](t, d, "t_test")

	require.NoError(t, m.Save(&TestEntity{Name: "a", Age: 10}))
	require.NoError(t, m.Save(&TestEntity{Name: "b", Age: 20}))
	require.NoError(t, m.Save(&TestEntity{Name: "c", Age: 30}))

	count, err := m.Query().Where("age > ?", 15).Count()
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestModel_QueryOrder(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestEntity](t, d, "t_test")

	require.NoError(t, m.Save(&TestEntity{Name: "c", Age: 30}))
	require.NoError(t, m.Save(&TestEntity{Name: "a", Age: 10}))
	require.NoError(t, m.Save(&TestEntity{Name: "b", Age: 20}))

	users, err := m.Query().Order("age asc").All()
	require.NoError(t, err)
	assert.Len(t, users, 3)
	assert.Equal(t, "a", users[0].Name)
	assert.Equal(t, "b", users[1].Name)
	assert.Equal(t, "c", users[2].Name)
}

func TestModel_QueryList(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestEntity](t, d, "t_test")

	for i := 0; i < 10; i++ {
		require.NoError(t, m.Save(&TestEntity{Name: "user", Age: 20 + i}))
	}

	users, err := m.Query().List(3)
	require.NoError(t, err)
	assert.Len(t, users, 3)
}

func TestModel_UpdateForMap(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestEntity](t, d, "t_test")

	require.NoError(t, m.Save(&TestEntity{Name: "alice", Age: 25}))

	err := m.Update().Where("name = ?", "alice").UpdateForMap(map[string]any{
		"name": "alice_updated",
		"age":  26,
	})
	require.NoError(t, err)

	user, err := m.Query().Where("name = ?", "alice_updated").One()
	require.NoError(t, err)
	assert.Equal(t, 26, user.Age)
}

func TestModel_UpdateColumn(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestEntity](t, d, "t_test")

	require.NoError(t, m.Save(&TestEntity{Name: "alice", Age: 25}))

	err := m.Update().Where("name = ?", "alice").UpdateColumn("age", 99)
	require.NoError(t, err)

	user, err := m.Query().Where("name = ?", "alice").One()
	require.NoError(t, err)
	assert.Equal(t, 99, user.Age)
}

func TestModel_Delete_ByWhere(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestEntity](t, d, "t_test")

	require.NoError(t, m.Save(&TestEntity{Name: "keep", Age: 10}))
	require.NoError(t, m.Save(&TestEntity{Name: "trash", Age: 20}))

	err := m.Delete().Where("name = ?", "trash").Delete()
	require.NoError(t, err)

	all, err := m.Query().All()
	require.NoError(t, err)
	assert.Len(t, all, 1)
	assert.Equal(t, "keep", all[0].Name)
}

// ---- EntryModel ----

func TestEntryModel_FindByPK(t *testing.T) {
	d := setupDB(t)
	m := setupEntryModel[*TestEntity, uint](t, d, "t_test")

	require.NoError(t, m.Save(&TestEntity{Name: "alice", Age: 25}))

	user, err := m.FindByPK(uint(1))
	require.NoError(t, err)
	assert.Equal(t, "alice", user.Name)
}

func TestEntryModel_FindAll(t *testing.T) {
	d := setupDB(t)
	m := setupEntryModel[*TestEntity, uint](t, d, "t_test")

	require.NoError(t, m.Save(&TestEntity{Name: "a", Age: 1}))
	require.NoError(t, m.Save(&TestEntity{Name: "b", Age: 2}))

	users, err := m.FindAll()
	require.NoError(t, err)
	assert.Len(t, users, 2)
}

func TestEntryModel_DeleteByPK(t *testing.T) {
	d := setupDB(t)
	m := setupEntryModel[*TestEntity, uint](t, d, "t_test")

	require.NoError(t, m.Save(&TestEntity{Name: "alice", Age: 25}))

	err := m.DeleteByPK(uint(1))
	require.NoError(t, err)

	all, err := m.FindAll()
	require.NoError(t, err)
	assert.Len(t, all, 0)
}

func TestEntryModel_UpdateByPK(t *testing.T) {
	d := setupDB(t)
	m := setupEntryModel[*TestEntity, uint](t, d, "t_test")

	require.NoError(t, m.Save(&TestEntity{Name: "alice", Age: 25}))

	user, err := m.FindByPK(uint(1))
	require.NoError(t, err)
	user.Name = "alice_new"
	user.Age = 30
	require.NoError(t, m.UpdateByPK(user))

	updated, err := m.FindByPK(uint(1))
	require.NoError(t, err)
	assert.Equal(t, "alice_new", updated.Name)
	assert.Equal(t, 30, updated.Age)
}

func TestEntryModel_UpdateColumn(t *testing.T) {
	d := setupDB(t)
	m := setupEntryModel[*TestEntity, uint](t, d, "t_test")

	require.NoError(t, m.Save(&TestEntity{Name: "alice", Age: 25}))
	require.NoError(t, m.UpdateColumn(uint(1), "age", 50))

	user, err := m.FindByPK(uint(1))
	require.NoError(t, err)
	assert.Equal(t, 50, user.Age)
}

func TestEntryModel_StringPK(t *testing.T) {
	d := setupDB(t)
	m := setupEntryModel[*TestStringPK, string](t, d, "t_string_pk")

	require.NoError(t, m.Save(&TestStringPK{Code: "key1", Val: "value1"}))
	require.NoError(t, m.Save(&TestStringPK{Code: "key2", Val: "value2"}))

	found, err := m.FindByPK("key1")
	require.NoError(t, err)
	assert.Equal(t, "value1", found.Val)

	require.NoError(t, m.DeleteByPK("key2"))
	all, err := m.FindAll()
	require.NoError(t, err)
	assert.Len(t, all, 1)
}

// ---- Page ----

func TestEntryModel_Page(t *testing.T) {
	d := setupDB(t)
	m := setupEntryModel[*TestEntity, uint](t, d, "t_test")

	for i := 0; i < 25; i++ {
		require.NoError(t, m.Save(&TestEntity{Name: "user", Age: 10 + i}))
	}

	users, total, err := m.Page(&util.Page{PageNo: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Len(t, users, 10)
	assert.Equal(t, 25, total)

	users2, total2, err := m.Page(&util.Page{PageNo: 3, PageSize: 10})
	require.NoError(t, err)
	assert.Len(t, users2, 5)
	assert.Equal(t, 25, total2)
}

func TestQuery_Page(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestEntity](t, d, "t_test")

	for i := 0; i < 50; i++ {
		require.NoError(t, m.Save(&TestEntity{Name: "user", Age: 10 + i}))
	}

	users, total, err := m.Query().Where("age >= ?", 20).Page(&util.Page{PageNo: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, 40, total)
	assert.Len(t, users, 10)
}

// ---- WithContext ----

func TestModel_WithContext(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestEntity](t, d, "t_test")

	require.NoError(t, m.Save(&TestEntity{Name: "alice", Age: 25}))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mCtx := m.WithContext(ctx)
	assert.NotSame(t, m, mCtx, "WithContext should return a new instance")

	user, err := mCtx.Query().Where("name = ?", "alice").One()
	require.NoError(t, err)
	assert.Equal(t, "alice", user.Name)
}

func TestEntryModel_WithContext(t *testing.T) {
	d := setupDB(t)
	m := setupEntryModel[*TestEntity, uint](t, d, "t_test")

	require.NoError(t, m.Save(&TestEntity{Name: "alice", Age: 25}))

	ctx := context.Background()
	mCtx := m.WithContext(ctx)
	assert.NotSame(t, m, mCtx, "WithContext should return a new instance")

	user, err := mCtx.FindByPK(uint(1))
	require.NoError(t, err)
	assert.Equal(t, "alice", user.Name)
}

func TestQuery_WithContext_ReturnsCopy(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestEntity](t, d, "t_test")

	q1 := m.Query()
	q2 := q1.WithContext(context.Background())

	assert.NotSame(t, q1, q2, "WithContext should return a new instance")

	q1.Where("name = ?", "alice")
	assert.Len(t, q1.wheres, 1)
	assert.Len(t, q2.wheres, 0, "original query wheres should not leak into copy")
}

func TestUpdate_WithContext_ReturnsCopy(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestEntity](t, d, "t_test")

	u1 := m.Update()
	u2 := u1.WithContext(context.Background())

	assert.NotSame(t, u1, u2, "WithContext should return a new instance")

	u1.Where("id = ?", 1)
	assert.Len(t, u1.wheres, 1)
	assert.Len(t, u2.wheres, 0, "original update wheres should not leak into copy")
}

func TestDelete_WithContext_ReturnsCopy(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestEntity](t, d, "t_test")

	d1 := m.Delete()
	d2 := d1.WithContext(context.Background())

	assert.NotSame(t, d1, d2, "WithContext should return a new instance")

	d1.Where("id = ?", 1)
	assert.Len(t, d1.wheres, 1)
	assert.Len(t, d2.wheres, 0, "original delete wheres should not leak into copy")
}

// ---- table management ----

func TestModel_CreateAndDropTable(t *testing.T) {
	d := setupDB(t)
	m := NewModel[*TestEntity](d, "t_dynamic")

	require.NoError(t, m.CreateTable())

	err := m.Save(&TestEntity{Name: "test", Age: 1})
	require.NoError(t, err)

	require.NoError(t, m.DropTable())
}

// ---- aggregate test entity ----

type TestOrder struct {
	ID       uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	Category string  `gorm:"size:64" json:"category"`
	Product  string  `gorm:"size:128" json:"product"`
	Amount   float64 `json:"amount"`
	Quantity int     `json:"quantity"`
	Status   int     `gorm:"default:1" json:"status"`
}

func setupAggregateData(t *testing.T, m *Model[*TestOrder]) {
	t.Helper()
	orders := []*TestOrder{
		{Category: "electronics", Product: "phone", Amount: 800, Quantity: 2, Status: 1},
		{Category: "electronics", Product: "laptop", Amount: 1500, Quantity: 1, Status: 1},
		{Category: "electronics", Product: "tablet", Amount: 500, Quantity: 3, Status: 2},
		{Category: "clothing", Product: "shirt", Amount: 50, Quantity: 5, Status: 1},
		{Category: "clothing", Product: "pants", Amount: 80, Quantity: 3, Status: 1},
		{Category: "food", Product: "bread", Amount: 10, Quantity: 10, Status: 1},
	}
	for _, o := range orders {
		require.NoError(t, m.Save(o))
	}
}

// ---- Aggregate builder-only (nil db) ----

func TestAggregate_BuilderOnly(t *testing.T) {
	m := NewModel[*TestOrder](nil, "t_orders")
	agg := m.Aggregate()
	assert.Equal(t, "t_orders", agg.tableName)
}

// ---- Aggregate SUM ----

func TestAggregate_Sum(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestOrder](t, d, "t_orders")
	setupAggregateData(t, m)

	var total float64
	err := m.Aggregate().Select("SUM(amount)").Aggregate(&total)
	require.NoError(t, err)
	assert.Equal(t, 2940.0, total)
}

// ---- Aggregate SUM with WHERE ----

func TestAggregate_SumWithWhere(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestOrder](t, d, "t_orders")
	setupAggregateData(t, m)

	var total float64
	err := m.Aggregate().Select("SUM(amount)").Where("status = ?", 1).Aggregate(&total)
	require.NoError(t, err)
	// electronics(800+1500) + clothing(50+80) + food(10) = 2440
	assert.Equal(t, 2440.0, total)
}

// ---- Aggregate COUNT ----

func TestAggregate_Count(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestOrder](t, d, "t_orders")
	setupAggregateData(t, m)

	var cnt int
	err := m.Aggregate().Select("COUNT(*)").Aggregate(&cnt)
	require.NoError(t, err)
	assert.Equal(t, 6, cnt)
}

// ---- Aggregate AVG ----

func TestAggregate_Avg(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestOrder](t, d, "t_orders")
	setupAggregateData(t, m)

	var avg float64
	err := m.Aggregate().Select("AVG(amount)").Aggregate(&avg)
	require.NoError(t, err)
	assert.Equal(t, 2940.0/6.0, avg)
}

// ---- Aggregate MAX / MIN ----

func TestAggregate_MaxMin(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestOrder](t, d, "t_orders")
	setupAggregateData(t, m)

	var maxAmt float64
	err := m.Aggregate().Select("MAX(amount)").Aggregate(&maxAmt)
	require.NoError(t, err)
	assert.Equal(t, 1500.0, maxAmt)

	var minAmt float64
	err = m.Aggregate().Select("MIN(amount)").Aggregate(&minAmt)
	require.NoError(t, err)
	assert.Equal(t, 10.0, minAmt)
}

// ---- Aggregate GROUP BY ----

type CategoryStat struct {
	Category string  `json:"category"`
	Total    float64 `json:"total"`
}

func TestAggregate_GroupBy(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestOrder](t, d, "t_orders")
	setupAggregateData(t, m)

	var stats []CategoryStat
	err := m.Aggregate().
		Select("category, SUM(amount) as total").
		Group("category").
		Aggregate(&stats)
	require.NoError(t, err)
	assert.Len(t, stats, 3)

	statsMap := make(map[string]float64)
	for _, s := range stats {
		statsMap[s.Category] = s.Total
	}
	assert.Equal(t, 2800.0, statsMap["electronics"])
	assert.Equal(t, 130.0, statsMap["clothing"])
	assert.Equal(t, 10.0, statsMap["food"])
}

// ---- Aggregate GROUP BY + HAVING ----

func TestAggregate_GroupByHaving(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestOrder](t, d, "t_orders")
	setupAggregateData(t, m)

	var stats []CategoryStat
	err := m.Aggregate().
		Select("category, SUM(amount) as total").
		Group("category").
		Having("SUM(amount) > ?", 200).
		Aggregate(&stats)
	require.NoError(t, err)
	assert.Len(t, stats, 1)
	assert.Equal(t, "electronics", stats[0].Category)
	assert.Equal(t, 2800.0, stats[0].Total)
}

// ---- Aggregate GROUP BY + WHERE ----

func TestAggregate_GroupByWithWhere(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestOrder](t, d, "t_orders")
	setupAggregateData(t, m)

	var stats []CategoryStat
	err := m.Aggregate().
		Select("category, SUM(amount) as total").
		Where("status = ?", 1).
		Group("category").
		Aggregate(&stats)
	require.NoError(t, err)

	statsMap := make(map[string]float64)
	for _, s := range stats {
		statsMap[s.Category] = s.Total
	}
	// electronics: 800+1500=2300 (status=2 的 tablet 被排除)
	assert.Equal(t, 2300.0, statsMap["electronics"])
	assert.Equal(t, 130.0, statsMap["clothing"])
	assert.Equal(t, 10.0, statsMap["food"])
}

// ---- Aggregate GROUP BY + ORDER ----

func TestAggregate_GroupByOrder(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestOrder](t, d, "t_orders")
	setupAggregateData(t, m)

	var stats []CategoryStat
	err := m.Aggregate().
		Select("category, SUM(amount) as total").
		Group("category").
		Order("SUM(amount) DESC").
		Aggregate(&stats)
	require.NoError(t, err)
	require.Len(t, stats, 3)
	assert.Equal(t, "electronics", stats[0].Category)
	assert.Equal(t, "clothing", stats[1].Category)
	assert.Equal(t, "food", stats[2].Category)
}

// ---- Aggregate multi-column result ----

type CategoryDetail struct {
	Category string  `json:"category"`
	Total    float64 `json:"total"`
	Cnt      int     `json:"cnt"`
	AvgAmt   float64 `json:"avgAmt"`
}

func TestAggregate_MultiColumn(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestOrder](t, d, "t_orders")
	setupAggregateData(t, m)

	var details []CategoryDetail
	err := m.Aggregate().
		Select("category, SUM(amount) as total, COUNT(*) as cnt, AVG(amount) as avg_amt").
		Group("category").
		Aggregate(&details)
	require.NoError(t, err)

	detailMap := make(map[string]CategoryDetail)
	for _, d := range details {
		detailMap[d.Category] = d
	}

	assert.Equal(t, 2800.0, detailMap["electronics"].Total)
	assert.Equal(t, 3, detailMap["electronics"].Cnt)
}

// ---- Aggregate empty result (scalar) ----

func TestAggregate_EmptyScalar(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestOrder](t, d, "t_orders")
	setupAggregateData(t, m)

	// Use COALESCE to handle NULL when no rows match
	var total float64
	err := m.Aggregate().Select("COALESCE(SUM(amount), 0)").Where("category = ?", "nonexistent").Aggregate(&total)
	require.NoError(t, err)
	assert.Equal(t, 0.0, total)
}

// ---- Aggregate empty result (grouped) ----

func TestAggregate_EmptyGrouped(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestOrder](t, d, "t_orders")
	setupAggregateData(t, m)

	var stats []CategoryStat
	err := m.Aggregate().
		Select("category, SUM(amount) as total").
		Where("category = ?", "nonexistent").
		Group("category").
		Aggregate(&stats)
	require.NoError(t, err)
	assert.Len(t, stats, 0)
}

// ---- Aggregate with int result ----

func TestAggregate_SumInt(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestOrder](t, d, "t_orders")
	setupAggregateData(t, m)

	var totalQty int
	err := m.Aggregate().Select("SUM(quantity)").Aggregate(&totalQty)
	require.NoError(t, err)
	assert.Equal(t, 24, totalQty)
}

// ---- Aggregate WithContext ----

func TestAggregate_WithContext(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestOrder](t, d, "t_orders")
	setupAggregateData(t, m)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var total float64
	err := m.Aggregate().
		WithContext(ctx).
		Select("SUM(amount)").
		Aggregate(&total)
	require.NoError(t, err)
	assert.Equal(t, 2940.0, total)
}

// ---- Aggregate WithContext returns copy ----

func TestAggregate_WithContext_ReturnsCopy(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestOrder](t, d, "t_orders")

	a1 := m.Aggregate()
	a2 := a1.WithContext(context.Background())

	assert.NotSame(t, a1, a2, "WithContext should return a new instance")

	a1.Where("id = ?", 1)
	assert.Len(t, a1.wheres, 1)
	assert.Len(t, a2.wheres, 0, "original aggregate wheres should not leak into copy")
}

// ---- Aggregate nil db ----

func TestAggregate_NilDB(t *testing.T) {
	m := NewModel[*TestOrder](nil, "t_orders")
	var total float64
	err := m.Aggregate().Select("SUM(amount)").Aggregate(&total)
	assert.Error(t, err)
}

// ---- Query Select/Group/Having ----

func TestQuery_GroupBy(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestOrder](t, d, "t_orders")
	setupAggregateData(t, m)

	// Verify Select/Group/Having chain methods compile and build correctly.
	// Query().All() returns []*TestOrder, not grouped structs — use Aggregate() for that.
	q := m.Query().
		Select("category, SUM(amount) as total").
		Group("category").
		Having("COUNT(*) > ?", 1).
		Order("category")

	assert.Len(t, q.selects, 1)
	assert.Len(t, q.groups, 1)
	assert.Len(t, q.having, 1)
}

// ---- Aggregate Select with args ----

func TestAggregate_SelectWithArgs(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestOrder](t, d, "t_orders")
	setupAggregateData(t, m)

	// COALESCE with a parameterized fallback value
	var total float64
	err := m.Aggregate().
		Select("COALESCE(SUM(amount), ?)", 0).
		Where("category = ?", "nonexistent").
		Aggregate(&total)
	require.NoError(t, err)
	assert.Equal(t, 0.0, total)
}

func TestAggregate_SelectWithArgs_Multiple(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestOrder](t, d, "t_orders")
	setupAggregateData(t, m)

	// Multiple selects with args
	var result float64
	err := m.Aggregate().
		Select("SUM(amount) * ? + ?", 2, 100).
		Aggregate(&result)
	require.NoError(t, err)
	// SUM = 2940, 2940*2+100 = 5980
	assert.Equal(t, 5980.0, result)
}

// ---- Aggregate Having with args ----

func TestAggregate_HavingWithArgs(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestOrder](t, d, "t_orders")
	setupAggregateData(t, m)

	var stats []CategoryStat
	err := m.Aggregate().
		Select("category, SUM(amount) as total").
		Group("category").
		Having("SUM(amount) > ? AND SUM(amount) < ?", 100, 2000).
		Aggregate(&stats)
	require.NoError(t, err)
	require.Len(t, stats, 1)
	assert.Equal(t, "clothing", stats[0].Category)
	assert.Equal(t, 130.0, stats[0].Total)
}

// ---- Aggregate chained selects ----

func TestAggregate_MultipleSelects(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestOrder](t, d, "t_orders")
	setupAggregateData(t, m)

	// Multiple Select() calls — last one wins in GORM
	var total float64
	err := m.Aggregate().
		Select("SUM(quantity)").
		Select("SUM(amount)").
		Aggregate(&total)
	require.NoError(t, err)
	assert.Equal(t, 2940.0, total)
}

// ---- Aggregate Distinct ----

func TestAggregate_Distinct(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestOrder](t, d, "t_orders")
	setupAggregateData(t, m)

	// Distinct categories count
	var cnt int
	err := m.Aggregate().
		Distinct().
		Select("COUNT(DISTINCT category)").
		Aggregate(&cnt)
	require.NoError(t, err)
	assert.Equal(t, 3, cnt)
}

func TestAggregate_DistinctWithColumns(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestOrder](t, d, "t_orders")
	setupAggregateData(t, m)

	// Distinct("category") sets DISTINCT + SELECT category
	type CatRow struct {
		Category string
	}
	var rows []CatRow
	err := m.Aggregate().
		Distinct("category").
		Aggregate(&rows)
	require.NoError(t, err)
	assert.Len(t, rows, 3)

	categories := make(map[string]bool)
	for _, r := range rows {
		categories[r.Category] = true
	}
	assert.True(t, categories["electronics"])
	assert.True(t, categories["clothing"])
	assert.True(t, categories["food"])
}

func TestAggregate_DistinctWithGroupBy(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestOrder](t, d, "t_orders")
	setupAggregateData(t, m)

	// Distinct + GroupBy: count distinct products per category
	type CatProductCnt struct {
		Category string
		Cnt      int
	}
	var stats []CatProductCnt
	err := m.Aggregate().
		Select("category, COUNT(*) as cnt").
		Distinct().
		Group("category").
		Aggregate(&stats)
	require.NoError(t, err)
	require.Len(t, stats, 3)

	statMap := make(map[string]int)
	for _, s := range stats {
		statMap[s.Category] = s.Cnt
	}
	assert.Equal(t, 3, statMap["electronics"])
	assert.Equal(t, 2, statMap["clothing"])
	assert.Equal(t, 1, statMap["food"])
}

// ---- Query Distinct ----

func TestQuery_Distinct(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestOrder](t, d, "t_orders")
	setupAggregateData(t, m)

	// Distinct categories via Query
	orders, err := m.Query().Select("category").Distinct().All()
	require.NoError(t, err)
	assert.Len(t, orders, 3)
}

func TestQuery_DistinctWithArgs(t *testing.T) {
	d := setupDB(t)
	m := setupModel[*TestOrder](t, d, "t_orders")
	setupAggregateData(t, m)

	// Distinct("category") via Query
	orders, err := m.Query().Distinct("category").All()
	require.NoError(t, err)
	assert.Len(t, orders, 3)
}

// ---- isScalarResult ----

func TestIsScalarResult(t *testing.T) {
	var f float64
	var i int
	var s string
	type st struct{ Name string }
	var ps st

	assert.True(t, isScalarResult(&f))
	assert.True(t, isScalarResult(&i))
	assert.True(t, isScalarResult(&s))
	assert.True(t, isScalarResult(&ps))

	var slice []CategoryStat
	assert.False(t, isScalarResult(&slice))
	assert.False(t, isScalarResult(nil))
}

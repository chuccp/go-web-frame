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

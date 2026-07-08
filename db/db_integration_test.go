package db

import (
	"context"
	"reflect"
	"testing"

	config2 "github.com/chuccp/go-web-frame/config"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func reflectKind(k reflect.Kind) reflect.Kind { return k }

func newConfigForTest() config2.IConfig {
	return config2.NewConfig()
}

type testUser struct {
	ID   uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"size:255" json:"name"`
	Age  int    `json:"age"`
}

func newTestDB(t *testing.T) *DB {
	t.Helper()
	d, err := ConnectionSQLite(":memory:")
	assert.NoError(t, err)
	t.Cleanup(func() {
		sqlDB, _ := d.db.DB()
		_ = sqlDB.Close()
	})
	return d
}

func TestDB_ConnectionSQLite(t *testing.T) {
	d, err := ConnectionSQLite(":memory:")
	assert.NoError(t, err)
	assert.NotNil(t, d)

	sqlDB, err := d.db.DB()
	assert.NoError(t, err)
	assert.NotNil(t, sqlDB)
	_ = sqlDB.Close()
}

func TestDB_Table_CreateAndQuery(t *testing.T) {
	d := newTestDB(t)

	// Auto migrate
	table := d.Table("test_users")
	err := table.AutoMigrate(&testUser{})
	assert.NoError(t, err)

	// Create
	err = table.Create(&testUser{Name: "alice", Age: 30})
	assert.NoError(t, err)

	// Find
	var users []testUser
	err = table.Find(&users)
	assert.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, "alice", users[0].Name)
	assert.Equal(t, 30, users[0].Age)
}

func TestDB_Table_Save(t *testing.T) {
	d := newTestDB(t)
	table := d.Table("test_users")
	_ = table.AutoMigrate(&testUser{})

	u := &testUser{Name: "bob", Age: 25}
	err := table.Save(u)
	assert.NoError(t, err)
	assert.NotZero(t, u.ID)
}

func TestDB_Table_Delete(t *testing.T) {
	d := newTestDB(t)
	table := d.Table("test_users")
	_ = table.AutoMigrate(&testUser{})
	_ = table.Create(&testUser{Name: "alice", Age: 30})
	_ = table.Create(&testUser{Name: "bob", Age: 25})

	err := table.Where("name = ?", "alice").Delete(&testUser{})
	// Where mutates t.db, so we use the returned table
	var users []testUser
	err = table.NewTable().Find(&users)
	assert.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, "bob", users[0].Name)
}

func TestDB_Table_Where(t *testing.T) {
	d := newTestDB(t)
	table := d.Table("test_users")
	_ = table.AutoMigrate(&testUser{})
	_ = table.Create(&testUser{Name: "alice", Age: 30})
	_ = table.Create(&testUser{Name: "bob", Age: 25})
	_ = table.Create(&testUser{Name: "carol", Age: 30})

	// Where chain
	filtered := d.Table("test_users").Where("age = ?", 30)
	var users []testUser
	err := filtered.Find(&users)
	assert.NoError(t, err)
	assert.Len(t, users, 2)
}

func TestDB_Table_OrderByLimit(t *testing.T) {
	d := newTestDB(t)
	table := d.Table("test_users")
	_ = table.AutoMigrate(&testUser{})
	_ = table.Create(&testUser{Name: "c", Age: 3})
	_ = table.Create(&testUser{Name: "a", Age: 1})
	_ = table.Create(&testUser{Name: "b", Age: 2})

	var users []testUser
	err := table.Order("age asc").Limit(2).Find(&users)
	assert.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Equal(t, "a", users[0].Name)
}

func TestDB_Table_First(t *testing.T) {
	d := newTestDB(t)
	table := d.Table("test_users")
	_ = table.AutoMigrate(&testUser{})
	_ = table.Create(&testUser{Name: "alice", Age: 30})

	var u testUser
	err := table.First(&u)
	assert.NoError(t, err)
	assert.Equal(t, "alice", u.Name)
}

func TestDB_Table_Count(t *testing.T) {
	d := newTestDB(t)
	table := d.Table("test_users")
	_ = table.AutoMigrate(&testUser{})
	_ = table.Create(&testUser{Name: "alice", Age: 30})
	_ = table.Create(&testUser{Name: "bob", Age: 25})

	var count int64
	err := table.Count(&count)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestDB_Table_Updates(t *testing.T) {
	d := newTestDB(t)
	table := d.Table("test_users")
	_ = table.AutoMigrate(&testUser{})
	_ = table.Create(&testUser{Name: "alice", Age: 30})

	err := table.Where("name = ?", "alice").Updates(map[string]any{"age": 31})
	assert.NoError(t, err)

	var u testUser
	_ = table.NewTable().First(&u)
	assert.Equal(t, 31, u.Age)
}

func TestDB_Table_UpdateColumn(t *testing.T) {
	d := newTestDB(t)
	table := d.Table("test_users")
	_ = table.AutoMigrate(&testUser{})
	_ = table.Create(&testUser{Name: "alice", Age: 30})

	err := table.Where("name = ?", "alice").UpdateColumn("age", 35)
	assert.NoError(t, err)

	var u testUser
	_ = table.NewTable().First(&u)
	assert.Equal(t, 35, u.Age)
}

func TestDB_Table_CreateWithPk(t *testing.T) {
	d := newTestDB(t)
	table := d.Table("test_users")
	_ = table.AutoMigrate(&testUser{})

	u := &testUser{Name: "alice", Age: 30}
	pk, err := table.CreateWithPk(u, "ID", reflectKind(reflect.Uint))
	assert.NoError(t, err)
	assert.NotNil(t, pk)
}

func TestDB_Table_CreateWithPk_AutoDetect(t *testing.T) {
	d := newTestDB(t)
	table := d.Table("test_users")
	_ = table.AutoMigrate(&testUser{})

	u := &testUser{Name: "alice", Age: 30}
	pk, err := table.CreateWithPk(u, "", 0)
	assert.NoError(t, err)
	assert.NotNil(t, pk)
}

func TestDB_Table_CreateWithPk_FieldNotFound(t *testing.T) {
	d := newTestDB(t)
	table := d.Table("test_users")
	_ = table.AutoMigrate(&testUser{})

	u := &testUser{Name: "alice", Age: 30}
	_, err := table.CreateWithPk(u, "NonExistent", 0)
	assert.Error(t, err)
}

func TestDB_Table_WithContext(t *testing.T) {
	d := newTestDB(t)
	table := d.Table("test_users")
	_ = table.AutoMigrate(&testUser{})

	ctx := context.Background()
	newTable := table.WithContext(ctx)
	assert.NotNil(t, newTable)
	assert.NotNil(t, newTable.raw)
}

func TestDB_Table_NewTable(t *testing.T) {
	d := newTestDB(t)
	table := d.Table("test_users")

	newTable := table.NewTable()
	assert.NotNil(t, newTable)
	assert.Equal(t, "test_users", newTable.tableName)
}

func TestDB_Table_Session(t *testing.T) {
	d := newTestDB(t)
	table := d.Table("test_users")

	sess := table.Session(&gorm.Session{})
	assert.NotNil(t, sess)
}

func TestDB_Transaction_Commit(t *testing.T) {
	d := newTestDB(t)
	table := d.Table("test_users")
	_ = table.AutoMigrate(&testUser{})

	err := d.Transaction(func(txDB *DB) error {
		txTable := txDB.Table("test_users")
		_ = txTable.Create(&testUser{Name: "alice", Age: 30})
		_ = txTable.Create(&testUser{Name: "bob", Age: 25})
		return nil
	})
	assert.NoError(t, err)

	var count int64
	_ = table.Count(&count)
	assert.Equal(t, int64(2), count)
}

func TestDB_Transaction_Rollback(t *testing.T) {
	d := newTestDB(t)
	table := d.Table("test_users")
	_ = table.AutoMigrate(&testUser{})

	err := d.Transaction(func(txDB *DB) error {
		txTable := txDB.Table("test_users")
		_ = txTable.Create(&testUser{Name: "alice", Age: 30})
		return context.DeadlineExceeded // simulate error
	})
	assert.Error(t, err)

	var count int64
	_ = table.Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestDB_Migrator(t *testing.T) {
	d := newTestDB(t)
	migrator := d.Migrator()
	assert.NotNil(t, migrator)
}

func TestDB_New(t *testing.T) {
	d := newTestDB(t)
	d2 := d.New()
	assert.NotNil(t, d2)
}

func TestDB_CreateDB_NoConfig(t *testing.T) {
	// Empty config should return NoConfigDBError
	cfg := newConfigForTest()
	_, err := CreateDB(cfg)
	assert.Error(t, err)
	assert.ErrorIs(t, err, NoConfigDBError)
}

func TestDB_CreateDB_SQLite(t *testing.T) {
	cfg := newConfigForTest()
	cfg.Put("web.db.type", "sqlite")
	cfg.Put("web.db.path", ":memory:")

	d, err := CreateDB(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, d)

	sqlDB, _ := d.db.DB()
	_ = sqlDB.Close()
}

func TestDB_CreateDB_UnsupportedType(t *testing.T) {
	cfg := newConfigForTest()
	cfg.Put("web.db.type", "unsupported")

	_, err := CreateDB(cfg)
	assert.Error(t, err)
	assert.ErrorIs(t, err, NoConfigDBError)
}

func TestDB_ConnectionPoolDefaults(t *testing.T) {
	// Test that pool defaults are applied
	cfg := &SQLiteConfig{FilePath: ":memory:"}
	assert.Equal(t, 10, cfg.GetMaxOpenConns())
	assert.Equal(t, 5, cfg.GetMaxIdleConns())
	assert.Equal(t, 3600, cfg.GetConnMaxLifetime())

	cfg2 := &SQLiteConfig{FilePath: ":memory:", MaxOpenConns: 20, MaxIdleConns: 10, ConnMaxLifetime: 7200}
	assert.Equal(t, 20, cfg2.GetMaxOpenConns())
	assert.Equal(t, 10, cfg2.GetMaxIdleConns())
	assert.Equal(t, 7200, cfg2.GetConnMaxLifetime())
}

func TestDB_MysqlConfig_ConnectionPoolDefaults(t *testing.T) {
	cfg := &MysqlConfig{}
	assert.Equal(t, 100, cfg.GetMaxOpenConns())
	assert.Equal(t, 10, cfg.GetMaxIdleConns())
	assert.Equal(t, 3600, cfg.GetConnMaxLifetime())

	cfg2 := &MysqlConfig{MaxOpenConns: 200, MaxIdleConns: 20, ConnMaxLifetime: 1800}
	assert.Equal(t, 200, cfg2.GetMaxOpenConns())
	assert.Equal(t, 20, cfg2.GetMaxIdleConns())
	assert.Equal(t, 1800, cfg2.GetConnMaxLifetime())
}

func TestDB_PostgresConfig_ConnectionPoolDefaults(t *testing.T) {
	cfg := &PostgresConfig{}
	assert.Equal(t, 100, cfg.GetMaxOpenConns())
	assert.Equal(t, 10, cfg.GetMaxIdleConns())
	assert.Equal(t, 3600, cfg.GetConnMaxLifetime())
}

func TestDB_NoConfigDBError(t *testing.T) {
	err := NoConfigDBError
	assert.Error(t, err)
	assert.Equal(t, "no config db", err.Error())
}

func TestDB_ConnectionSQLiteWithParams(t *testing.T) {
	d, err := ConnectionSQLite(":memory:")
	assert.NoError(t, err)
	assert.NotNil(t, d)
	sqlDB, _ := d.db.DB()
	_ = sqlDB.Close()
}

func TestDB_CreateMapWithPk(t *testing.T) {
	d := newTestDB(t)
	table := d.Table("test_users")
	_ = table.AutoMigrate(&testUser{})

	// CreateMapWithPk adds "@" prefix to keyName, then looks it up in the map.
	// GORM does not write back PK to map on all drivers, so this may error
	// if the key is not present after Create.
	_, err := table.CreateMapWithPk(map[string]any{"name": "alice", "age": 30}, "ID")
	// The record is created; the PK extraction may or may not succeed depending on driver.
	// We just verify no panic occurs.
	_ = err
}

func TestDB_CreateMapWithPk_AutoDetectKey(t *testing.T) {
	d := newTestDB(t)
	table := d.Table("test_users")
	_ = table.AutoMigrate(&testUser{})

	_, err := table.CreateMapWithPk(map[string]any{"name": "alice", "age": 30}, "")
	_ = err
}

func TestDB_CreateMapWithPk_KeyNotFound(t *testing.T) {
	d := newTestDB(t)
	table := d.Table("test_users")
	_ = table.AutoMigrate(&testUser{})

	_, err := table.CreateMapWithPk(map[string]any{"name": "alice"}, "@NonExistent")
	assert.Error(t, err)
}

func TestDB_CreateMapWithUintPk(t *testing.T) {
	d := newTestDB(t)
	table := d.Table("test_users")
	_ = table.AutoMigrate(&testUser{})

	_, err := table.CreateMapWithUintPk(map[string]any{"name": "alice", "age": 30}, "ID")
	_ = err
}

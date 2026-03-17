package model

import (
	"testing"

	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/db"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// TestEntity is a test entity for generic model testing
type TestEntity struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"size:255"`
	Age  int
}

// TestModel extends the generic Model
type TestModel struct {
	*Model[*TestEntity]
}

func (m *TestModel) Init(db *core.DB, ctx *core.Context) error {
	m.Model = NewModel[*TestEntity](db, "t_test")
	return m.CreateTable()
}

func TestNewModel(t *testing.T) {
	// Given: create a new model with in-memory sqlite
	// We can't actually open a real database in unit tests without dependencies,
	// so this test just verifies the model can be created.

	// This test documents the API usage
	var entity *TestEntity
	model := NewModel[*TestEntity](nil, "t_test")

	assert.NotNil(t, model)
	assert.Equal(t, "t_test", model.TableName())
}

func TestModel_TableName(t *testing.T) {
	// Given: a model with custom table name
	model := NewModel[*TestEntity](nil, "custom_table")

	// When: getting table name
	tableName := model.TableName()

	// Then: should return the custom table name
	assert.Equal(t, "custom_table", tableName)
}

func TestModel_DefaultTableName(t *testing.T) {
	// Given: a model created with default table name
	// When tableName is empty, gorm would infer it, but our API requires explicit table name
	model := NewModel[*TestEntity](nil, "")

	// Then: should return empty string
	assert.Equal(t, "", model.TableName())
}

func TestModel_Query(t *testing.T) {
	// Given: a model
	model := NewModel[*TestEntity](nil, "t_test")

	// When: creating a new query
	query := model.Query()

	// Then: should return a Table with the correct table name
	assert.Equal(t, "t_test", query.tableName)
}

func TestModel_Update(t *testing.T) {
	// Given: a model
	model := NewModel[*TestEntity](nil, "t_test")

	// When: creating an update query
	update := model.Update()

	// Then: should return a Table with the correct table name
	assert.Equal(t, "t_test", update.tableName)
}

func TestModel_Delete(t *testing.T) {
	// Given: a model
	model := NewModel[*TestEntity](nil, "t_test")

	// When: creating a delete query
	delete := model.Delete()

	// Then: should return a Table with the correct table name
	assert.Equal(t, "t_test", delete.tableName)
}

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEntity is a test entity for generic model testing
type TestEntity struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"size:255"`
	Age  int
}

func TestNewModel(t *testing.T) {
	model := NewModel[*TestEntity](nil, "t_test")

	assert.NotNil(t, model)
	assert.Equal(t, "t_test", model.GetTableName())
}

func TestModel_TableName(t *testing.T) {
	// Given: a model with custom table name
	model := NewModel[*TestEntity](nil, "custom_table")

	tableName := model.GetTableName()
	assert.Equal(t, "custom_table", tableName)
}

func TestModel_DefaultTableName(t *testing.T) {
	model := NewModel[*TestEntity](nil, "")

	assert.Equal(t, "", model.GetTableName())
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

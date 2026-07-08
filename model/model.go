package model

import (
	"context"
	"reflect"
	"strings"

	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/go-web-frame/util"
	"gorm.io/gorm"
)

// Model is a generic database model that provides CRUD operations for type T.
// It wraps GORM with type-safe query building and table management.
type Model[T any] struct {
	db        *db.DB
	tableName string
	entry     T
	pkColumn  string
}

// getPrimaryKeyColumn extracts the primary key column name from struct's gorm tag
func getPrimaryKeyColumn[T any]() string {
	var entryPtr T
	t := util.NewPtr(entryPtr)
	rt := reflect.TypeOf(t).Elem()

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		gormTag := field.Tag.Get("gorm")
		if gormTag == "" {
			continue
		}

		// Check if this field has primaryKey
		tagParts := strings.Split(gormTag, ";")
		isPrimaryKey := false
		columnName := ""

		for _, part := range tagParts {
			part = strings.TrimSpace(part)
			if part == "primaryKey" {
				isPrimaryKey = true
			}
			if strings.HasPrefix(part, "column:") {
				columnName = strings.TrimPrefix(part, "column:")
			}
		}

		if isPrimaryKey {
			if columnName != "" {
				return columnName
			}
			return field.Name
		}
	}

	// Fallback to "id" if no primaryKey tag found
	return "id"
}

// getPrimaryKeyValue extracts the primary key value from entity struct via reflection
func getPrimaryKeyValue[T any](entity T) interface{} {
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	rt := v.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		gormTag := field.Tag.Get("gorm")
		if strings.Contains(gormTag, "primaryKey") {
			return v.Field(i).Interface()
		}
	}
	return nil
}

// IsExist reports whether the table for this model exists in the database.
func (a *Model[T]) IsExist() (bool, error) {
	dbConn, err := a.getDB()
	if err != nil {
		return false, errors.WithStackIf(err)
	}
	return dbConn.Migrator().HasTable(a.tableName), nil
}

// CreateTable creates the database table for this model if it does not already exist.
func (a *Model[T]) CreateTable() error {
	dbConn, err := a.getDB()
	if err != nil {
		return errors.WithStackIf(err)
	}
	exist, err := a.IsExist()
	if err != nil {
		return err
	}
	if exist {
		return nil
	}
	t := util.NewPtr(a.entry)
	return errors.WithStackIf(dbConn.Table(a.tableName).AutoMigrate(t))
}

// DeleteTable deletes all rows from the table (does not drop the table structure).
func (a *Model[T]) DeleteTable() error {
	dbConn, err := a.getDB()
	if err != nil {
		return errors.WithStackIf(err)
	}
	t := util.NewPtr(a.entry)
	err = dbConn.Table(a.tableName).Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(t)
	return errors.WithStackIf(err)
}

// DropTable drops the database table for this model, removing the table structure and all data.
func (a *Model[T]) DropTable() error {
	dbConn, err := a.getDB()
	if err != nil {
		return errors.WithStackIf(err)
	}
	return errors.WithStackIf(dbConn.Migrator().DropTable(a.tableName))
}

// GetTableName returns the database table name for this model.
func (a *Model[T]) GetTableName() string {
	return a.tableName
}

// Save creates or updates a single record in the database.
func (a *Model[T]) Save(entry T) error {
	dbConn, err := a.getDB()
	if err != nil {
		return errors.WithStackIf(err)
	}
	return dbConn.Table(a.tableName).Save(entry)
}
func (a *Model[T]) getDB() (*db.DB, error) {
	if a.db == nil {
		return nil, errors.New("db is nil")
	}
	return a.db, nil
}

// Saves creates multiple records in a single batch operation.
func (a *Model[T]) Saves(entry []T) error {
	dbConn, err := a.getDB()
	if err != nil {
		return errors.WithStackIf(err)
	}
	return dbConn.Table(a.tableName).Create(&entry)
}

// SaveForMap creates a record from a map of column-value pairs.
func (a *Model[T]) SaveForMap(mapValue map[string]any) error {
	dbConn, err := a.getDB()
	if err != nil {
		return errors.WithStackIf(err)
	}
	return dbConn.Table(a.tableName).Create(mapValue)
}

// SavesForMap creates multiple records from a slice of column-value maps.
func (a *Model[T]) SavesForMap(mapValues []map[string]any) error {
	dbConn, err := a.getDB()
	if err != nil {
		return errors.WithStackIf(err)
	}
	return dbConn.Table(a.tableName).Create(mapValues)
}

// SaveForMapWithPk saves a record from a map and returns the generated primary key
func (a *Model[T]) SaveForMapWithPk(mapValue map[string]any, keyName string) (any, error) {
	dbConn, err := a.getDB()
	if err != nil {
		return nil, errors.WithStackIf(err)
	}
	return dbConn.Table(a.tableName).CreateMapWithPk(mapValue, keyName)
}

// SaveForMapWithUintPk saves a record from a map and returns the generated uint primary key.
func (a *Model[T]) SaveForMapWithUintPk(mapValue map[string]any, keyName string) (uint, error) {
	dbConn, err := a.getDB()
	if err != nil {
		return 0, errors.WithStackIf(err)
	}
	return dbConn.Table(a.tableName).CreateMapWithUintPk(mapValue, keyName)
}

// CreateWithPk creates a record and returns the generated primary key
func (a *Model[T]) CreateWithPk(entry T, keyName string, kind reflect.Kind) (any, error) {
	dbConn, err := a.getDB()
	if err != nil {
		return nil, errors.WithStackIf(err)
	}
	return dbConn.Table(a.tableName).CreateWithPk(entry, keyName, kind)
}

// Query returns a new Query builder for constructing type-safe read queries.
func (a *Model[T]) Query() *Query[T] {

	return &Query[T]{db: a.db, tableName: a.tableName, entry: a.entry}
}

// Update returns a new Update builder for constructing type-safe update operations.
func (a *Model[T]) Update() *Update[T] {
	return &Update[T]{
		db:        a.db,
		tableName: a.tableName,
		model:     a.entry,
		wheres:    make([]*where, 0),
	}
}

// Delete returns a new Delete builder for constructing type-safe delete operations.
func (a *Model[T]) Delete() *Delete[T] {
	return &Delete[T]{
		db:        a.db,
		tableName: a.tableName,
		model:     a.entry,
		wheres:    make([]*where, 0),
	}
}

// NewModel creates a new Model instance for the given table name.
func NewModel[T any](db *db.DB, tableName string) *Model[T] {
	var entryPtr T
	entry := util.NewPtr(entryPtr)
	return &Model[T]{db: db, tableName: tableName, entry: entry, pkColumn: getPrimaryKeyColumn[T]()}
}

// GetPkColumn returns the primary key column name for this model.
func (a *Model[T]) GetPkColumn() string {

	// getPrimaryKeyColumn[T]()
	return a.pkColumn
}

// WithContext returns a shallow copy of the model with the given context.
// All database operations on the returned model will propagate the context.
// The original instance is unchanged; safe for concurrent use.
func (a *Model[T]) WithContext(ctx context.Context) *Model[T] {
	if a.db == nil {
		return &Model[T]{tableName: a.tableName, entry: a.entry, pkColumn: a.pkColumn}
	}
	return &Model[T]{
		db:        a.db.WithContext(ctx),
		tableName: a.tableName,
		entry:     a.entry,
		pkColumn:  a.pkColumn,
	}
}

package model

import (
	"reflect"
	"strings"

	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/db"
	"github.com/chuccp/go-web-frame/util"
	"gorm.io/gorm"
)

type Model[T any] struct {
	db           *db.DB
	tableName    string
	entry        T
	pkColumn     string
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

		if isPrimaryKey && columnName != "" {
			return columnName
		}
	}

	// Fallback to "id" if no primaryKey found with column tag
	return "id"
}

func (a *Model[T]) IsExist() (bool, error) {
	dbConn, err := a.getBb()
	if err != nil {
		return false, errors.WithStackIf(err)
	}
	return dbConn.Migrator().HasTable(a.tableName), nil
}
func (a *Model[T]) CreateTable() error {
	dbConn, err := a.getBb()
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
func (a *Model[T]) DeleteTable() error {
	dbConn, err := a.getBb()
	if err != nil {
		return errors.WithStackIf(err)
	}
	t := util.NewPtr(a.entry)
	err = dbConn.Table(a.tableName).Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(t)
	return errors.WithStackIf(err)
}
func (a *Model[T]) GetTableName() string {
	return a.tableName
}
func (a *Model[T]) Save(entry T) error {
	dbConn, err := a.getBb()
	if err != nil {
		return errors.WithStackIf(err)
	}
	return dbConn.Table(a.tableName).Save(entry)
}
func (a *Model[T]) getBb() (*db.DB, error) {
	if a.db == nil {
		return nil, errors.New("db is nil")
	}
	return a.db, nil
}

func (a *Model[T]) Saves(entry []T) error {
	dbConn, err := a.getBb()
	if err != nil {
		return errors.WithStackIf(err)
	}
	return dbConn.Table(a.tableName).Create(&entry)
}

func (a *Model[T]) SaveForMap(mapValue map[string]any) error {
	dbConn, err := a.getBb()
	if err != nil {
		return errors.WithStackIf(err)
	}
	return dbConn.Table(a.tableName).Create(mapValue)
}

func (a *Model[T]) SavesForMap(mapValues []map[string]any) error {
	dbConn, err := a.getBb()
	if err != nil {
		return errors.WithStackIf(err)
	}
	return dbConn.Table(a.tableName).Create(mapValues)
}

// SaveForMapWithPk saves a record from a map and returns the generated primary key
func (a *Model[T]) SaveForMapWithPk(mapValue map[string]any, keyName string) (any, error) {
	dbConn, err := a.getBb()
	if err != nil {
		return nil, errors.WithStackIf(err)
	}
	return dbConn.Table(a.tableName).CreateMapWithPk(mapValue, keyName)
}

func (a *Model[T]) SaveForMapWithUintPk(mapValue map[string]any, keyName string) (uint, error) {
	dbConn, err := a.getBb()
	if err != nil {
		return 0, errors.WithStackIf(err)
	}
	return dbConn.Table(a.tableName).CreateMapWithUintPk(mapValue, keyName)
}

// CreateWithPk creates a record and returns the generated primary key
func (a *Model[T]) CreateWithPk(entry T, keyName string, kind reflect.Kind) (any, error) {
	dbConn, err := a.getBb()
	if err != nil {
		return nil, errors.WithStackIf(err)
	}
	return dbConn.Table(a.tableName).CreateWithPk(entry, keyName, kind)
}

func (a *Model[T]) Query() *Query[T] {

	return &Query[T]{db: a.db, tableName: a.tableName, entry: a.entry}
}

func (a *Model[T]) Update() *Update[T] {
	return &Update[T]{
		db:        a.db,
		tableName: a.tableName,
		model:     a.entry,
		wheres:    make([]*where, 0),
	}
}
func (a *Model[T]) Delete() *Delete[T] {
	return &Delete[T]{
		db:        a.db,
		tableName: a.tableName,
		model:     a.entry,
		wheres:    make([]*where, 0),
	}
}

func NewModel[T any](db *db.DB, tableName string) *Model[T] {
	var entryPtr T
	entry := util.NewPtr(entryPtr)
	return &Model[T]{db: db, tableName: tableName, entry: entry, pkColumn: getPrimaryKeyColumn[T]()}
}

func (a *Model[T]) GetPkColumn() string {
	return a.pkColumn
}

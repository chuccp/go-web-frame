package db

import (
	"reflect"

	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/util"
	"gorm.io/gorm"
)

const (
	MYSQL  = "mysql"
	SQLITE = "sqlite"
)

type Source interface {
	Connection(cfg config.IConfig) (db *gorm.DB, err error)
}
type Session struct {
	db *gorm.DB
}

func (s *Session) Delete(value any, conds ...any) error {
	tx := s.db.Delete(value, conds...)
	return tx.Error
}

type Table struct {
	db *gorm.DB
}

func (t *Table) Session(g *gorm.Session) *Session {
	return &Session{db: t.db.Session(g)}
}

func (t *Table) AutoMigrate(v ...any) error {
	return t.db.AutoMigrate(v...)
}

func (t *Table) Delete(value any, conds ...any) error {
	tx := t.db.Delete(value, conds...)
	return tx.Error
}

func (t *Table) Save(entry any) error {
	tx := t.db.Save(entry)
	return tx.Error
}

func (t *Table) Create(value any) error {
	return t.db.Create(value).Error
}

// CreateWithPk creates a record and returns the generated primary key
// If keyName is empty, it will try common ID field names (ID, Id, id)
// If kind is reflect.Invalid, it will try to determine the type automatically
func (t *Table) CreateWithPk(value any, keyName string, kind reflect.Kind) (any, error) {
	if err := t.db.Create(value).Error; err != nil {
		return nil, err
	}

	// Extract primary key using reflection
	val := reflect.ValueOf(value)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	var idField reflect.Value

	// Use specified key name or try common names
	if keyName != "" {
		idField = val.FieldByName(keyName)
		if !idField.IsValid() {
			return nil, errors.Errorf("cannot extract primary key: field '%s' not found", keyName)
		}
	} else {
		// Try common ID field names
		idField = val.FieldByName("ID")
		if !idField.IsValid() {
			idField = val.FieldByName("Id")
		}
		if !idField.IsValid() {
			idField = val.FieldByName("id")
		}
		if !idField.IsValid() {
			return nil, errors.New("cannot extract primary key: no common ID field found (tried ID, Id, id)")
		}
	}

	// Check if the field type matches the specified kind (if provided)
	if kind != reflect.Invalid && idField.Kind() != kind {
		return nil, errors.Errorf(
			"cannot extract primary key: field '%s' is of type %s, expected %s",
			idField.Type().Name(), idField.Kind(), kind)
	}

	// Extract the value based on type
	switch idField.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return idField.Uint(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return idField.Int(), nil
	case reflect.Float32, reflect.Float64:
		return idField.Float(), nil
	default:
		return nil, errors.Errorf(
			"cannot extract primary key: unsupported field type '%s'",
			idField.Kind())
	}
}

// CreateMapWithPk creates a record from a map and returns the generated primary key
// If keyName is empty, it will try common ID field names (ID, Id, id)
func (t *Table) CreateMapWithPk(mapValue map[string]any, keyName string) (any, error) {
	// Create the record
	if err := t.db.Create(mapValue).Error; err != nil {
		return nil, err
	}

	// Use specified key name or try common names
	var idValue any
	var found bool

	if keyName != "" {
		idValue, found = mapValue[keyName]
		if !found {
			return nil, errors.Errorf("cannot extract primary key: field '%s' not found in map", keyName)
		}
	} else {
		// Try common ID field names
		idValue, found = mapValue["ID"]
		if !found {
			idValue, found = mapValue["Id"]
		}
		if !found {
			idValue, found = mapValue["id"]
		}
		if !found {
			return nil, errors.New("cannot extract primary key: no common ID field found in map (tried ID, Id, id)")
		}
	}

	// Return the value in its original type
	return idValue, nil
}

// CreateMapWithUintPk creates a record from a map and returns the primary key as uint
func (t *Table) CreateMapWithUintPk(mapValue map[string]any, keyName string) (uint, error) {
	value, err := t.CreateMapWithPk(mapValue, keyName)
	if err != nil {
		return 0, err
	}

	// Convert to uint
	switch v := value.(type) {
	case uint:
		return v, nil
	case int:
		return uint(v), nil
	case int64:
		return uint(v), nil
	case float64:
		return uint(v), nil
	case uint8:
		return uint(v), nil
	case uint16:
		return uint(v), nil
	case uint32:
		return uint(v), nil
	case uint64:
		return uint(v), nil
	default:
		return 0, errors.Errorf("cannot convert primary key to uint: unsupported type %T", value)
	}
}

func (t *Table) Where(query any, args ...any) *Table {
	t.db = t.db.Where(query, args...)
	return t
}

//	func (t *Table) Set(column string, value any) *Table {
//		tx := t.db.Set(column, value)
//		return &Table{db: tx}
//	}
func (t *Table) Offset(i int) *Table {
	tx := t.db.Offset(i)
	return &Table{db: tx}
}

func (t *Table) Order(query any) *Table {
	tx := t.db.Order(query)
	return &Table{db: tx}
}

func (t *Table) Limit(size int) *Table {
	tx := t.db.Limit(size)
	return &Table{db: tx}
}

func (t *Table) Find(dest any, conds ...any) error {
	tx := t.db.Find(dest, conds...)
	return tx.Error

}

func (t *Table) First(dest any, conds ...any) error {
	tx := t.db.First(dest, conds...)
	return tx.Error
}

func (t *Table) Count(i *int64) error {
	tx := t.db.Count(i)
	return tx.Error
}

func (t *Table) Updates(values any) error {
	tx := t.db.Updates(values)
	return tx.Error
}

func (t *Table) UpdateColumn(column string, value any) error {
	tx := t.db.UpdateColumn(column, value)
	return tx.Error
}

type DB struct {
	db *gorm.DB
}

func (d *DB) Transaction(fc func(tx *DB) error) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		return fc(&DB{db: tx})
	})
}

func (d *DB) Migrator() gorm.Migrator {
	return d.db.Migrator()
}

func (d *DB) Table(name string) *Table {
	tx := d.db.Table(name)
	return &Table{db: tx}
}

type noConfigDBError struct {
}

func (e *noConfigDBError) Error() string {
	return "no config db"
}

var NoConfigDBError = &noConfigDBError{}

type IConfig interface {
	Connection() (*DB, error)
}
type Config struct {
	Type string
}

const ConfigKey = "web.db"

func CreateDB(c config.IConfig) (*DB, error) {
	var config2 Config
	err := c.UnmarshalKey(ConfigKey, &config2)
	if err != nil {
		return nil, err
	}
	if util.IsNotBlank(config2.Type) {
		if util.EqualsAnyIgnoreCase(config2.Type, MYSQL) {
			var mysqlConfig MysqlConfig
			err := c.UnmarshalKey(ConfigKey, &mysqlConfig)
			if err != nil {
				return nil, err
			}
			return mysqlConfig.Connection()
		}
		if util.EqualsAnyIgnoreCase(config2.Type, SQLITE) {
			var sqliteConfig SQLiteConfig
			err := c.UnmarshalKey(ConfigKey, &sqliteConfig)
			if err != nil {
				return nil, err
			}
			return sqliteConfig.Connection()
		}
	}
	return nil, errors.WithStackIf(NoConfigDBError)
}

// 建议：不要直接使用本包中的 db.DB / db.Table 操作数据库，推荐使用 model.EntryModel 或 model.Model，
// 它们提供了类型安全、更高层级的 CRUD 封装，代码更简洁且不易出错。
//
// Recommendation: Prefer model.EntryModel or model.Model over raw db.DB / db.Table.
// They provide type-safe, higher-level CRUD operations — less boilerplate and fewer mistakes.
//
// 推奨: 本パッケージの生の db.DB / db.Table ではなく model.EntryModel または model.Model を使用してください。
// 型安全で抽象度の高い CRUD 操作が提供され、コードが簡潔になりミスも減ります。
//
// 권장: 이 패키지의 raw db.DB / db.Table 대신 model.EntryModel 또는 model.Model 을 사용하세요.
// 타입 안전하고 더 높은 수준의 CRUD 연산을 제공하므로 코드가 간결해지고 실수가 줄어듭니다.
//
// Рекомендация: используйте model.EntryModel или model.Model вместо прямого db.DB / db.Table из этого пакета.
// Они предоставляют типобезопасные CRUD-операции высокого уровня — меньше шаблонного кода и ошибок.
package db

import (
	"context"
	"reflect"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/util"
	"gorm.io/gorm"
)

// ConnectionPoolConfig defines the interface for database connection pool configuration
type ConnectionPoolConfig interface {
	GetMaxOpenConns() int
	GetMaxIdleConns() int
	GetConnMaxLifetime() int
}

// ApplyConnectionPool applies connection pool settings to the underlying sql.DB
func ApplyConnectionPool(db *gorm.DB, cfg ConnectionPoolConfig) error {
	sqlDB, err := db.DB()
	if err != nil {
		return errors.WithStackIf(err)
	}
	if cfg.GetMaxOpenConns() > 0 {
		sqlDB.SetMaxOpenConns(cfg.GetMaxOpenConns())
	}
	if cfg.GetMaxIdleConns() > 0 {
		sqlDB.SetMaxIdleConns(cfg.GetMaxIdleConns())
	}
	if cfg.GetConnMaxLifetime() > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(cfg.GetConnMaxLifetime()) * time.Second)
	}
	return nil
}

const (
	MYSQL      = "mysql"
	POSTGRES   = "postgres"
	POSTGRESQL = "postgresql"
	SQLITE     = "sqlite"
)

// Source defines the interface for database connection sources.
type Source interface {
	Connection(cfg config.IConfig) (db *gorm.DB, err error)
}

// Session wraps a GORM database session for direct operations.
type Session struct {
	db *gorm.DB
}

// Delete deletes records matching the given conditions.
func (s *Session) Delete(value any, conds ...any) error {
	tx := s.db.Delete(value, conds...)
	return tx.Error
}

// 建议：不要直接使用 db.Table，推荐使用 model.EntryModel 或 model.Model，
// 它们提供了类型安全、更高层级的 CRUD 封装，代码更简洁且不易出错。
//
// Recommendation: Prefer model.EntryModel or model.Model over raw db.Table.
// They provide type-safe, higher-level CRUD operations — less boilerplate and fewer mistakes.
//
// 推奨: 生の db.Table ではなく model.EntryModel または model.Model を使用してください。
// 型安全で抽象度の高い CRUD 操作が提供され、コードが簡潔になりミスも減ります。
//
// 권장: raw db.Table 대신 model.EntryModel 또는 model.Model 을 사용하세요.
// 타입 안전하고 더 높은 수준의 CRUD 연산을 제공하므로 코드가 간결해지고 실수가 줄어듭니다.
//
// Рекомендация: используйте model.EntryModel или model.Model вместо прямого db.Table.
// Они предоставляют типобезопасные CRUD-операции высокого уровня — меньше шаблонного кода и ошибок.
type Table struct {
	db        *gorm.DB
	tableName string
	raw       *gorm.DB
}

// NewTable creates a fresh Table instance with the original GORM connection.
func (t *Table) NewTable() *Table {
	return &Table{db: t.raw.Table(t.tableName), tableName: t.tableName}
}

// WithContext returns a new Table whose underlying *gorm.DB carries the given context.
// The original Table is unchanged; safe for concurrent use.
func (t *Table) WithContext(ctx context.Context) *Table {
	return &Table{
		db:        t.db.WithContext(ctx),
		tableName: t.tableName,
		raw:       t.raw.WithContext(ctx),
	}
}

// Session creates a GORM session with the given options.
func (t *Table) Session(g *gorm.Session) *Session {
	return &Session{db: t.db.Session(g)}
}

// AutoMigrate runs auto migration for the given models.
func (t *Table) AutoMigrate(v ...any) error {
	return t.db.AutoMigrate(v...)
}

// Delete deletes records matching the conditions from the table.
func (t *Table) Delete(value any, conds ...any) error {
	tx := t.db.Delete(value, conds...)
	return tx.Error
}

// Save creates or updates a record in the table.
func (t *Table) Save(entry any) error {
	tx := t.db.Save(entry)
	return tx.Error
}

// Create inserts a new record into the table.
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

	if util.IsNotBlank(keyName) && !strings.HasPrefix(keyName, "@") {
		keyName = "@" + keyName
	}

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

// Where adds a WHERE condition to the query chain.
func (t *Table) Where(query any, args ...any) *Table {
	t.db = t.db.Where(query, args...)
	return t
}

//	func (t *Table) Set(column string, value any) *Table {
//		tx := t.db.Set(column, value)
//		return &Table{db: tx}
//	}
// Offset sets the number of records to skip.
func (t *Table) Offset(i int) *Table {
	tx := t.db.Offset(i)
	return &Table{db: tx, tableName: t.tableName, raw: t.raw}
}

// Order adds an ORDER BY clause.
func (t *Table) Order(query any) *Table {
	tx := t.db.Order(query)
	return &Table{db: tx, tableName: t.tableName, raw: t.raw}
}

// Preload preloads associations for the query (GORM foreign key support)
// Usage: db.Table("users").Preload("Profile").Preload("Role").Find(&users)
func (t *Table) Preload(query string, args ...any) *Table {
	tx := t.db.Preload(query, args...)
	return &Table{db: tx, tableName: t.tableName, raw: t.raw}
}

// Joins performs a join operation (GORM join support)
// Usage: db.Table("users").Joins("Profile").Find(&users)
func (t *Table) Joins(query string, args ...any) *Table {
	tx := t.db.Joins(query, args...)
	return &Table{db: tx, tableName: t.tableName, raw: t.raw}
}

// Limit sets the maximum number of records to return.
func (t *Table) Limit(size int) *Table {
	tx := t.db.Limit(size)
	return &Table{db: tx, tableName: t.tableName, raw: t.raw}
}

// Find retrieves all matching records into dest.
func (t *Table) Find(dest any, conds ...any) error {
	tx := t.db.Find(dest, conds...)
	return tx.Error

}

// First retrieves the first matching record into dest.
func (t *Table) First(dest any, conds ...any) error {
	tx := t.db.First(dest, conds...)
	return tx.Error
}

// Count returns the number of matching records.
func (t *Table) Count(i *int64) error {
	tx := t.db.Count(i)
	return tx.Error
}

// Updates applies the given values to matching records.
func (t *Table) Updates(values any) error {
	tx := t.db.Updates(values)
	return tx.Error
}

// UpdateColumn updates a single column for matching records.
func (t *Table) UpdateColumn(column string, value any) error {
	tx := t.db.UpdateColumn(column, value)
	return tx.Error
}

// Raw executes a raw SQL query and returns a Table for further chaining (e.g., with Scan)
func (t *Table) Raw(sql string, args ...any) *Table {
	tx := t.db.Raw(sql, args...)
	return &Table{db: tx, tableName: t.tableName, raw: t.raw}
}

// Scan scans query results into dest
func (t *Table) Scan(dest any) error {
	return t.db.Scan(dest).Error
}

// 建议：不要直接使用 db.DB / db.Table 操作数据库，推荐使用 model.EntryModel 或 model.Model，
// 它们提供了类型安全、更高层级的 CRUD 封装，代码更简洁且不易出错。
//
// Recommendation: Prefer model.EntryModel or model.Model over raw db.DB / db.Table.
// They provide type-safe, higher-level CRUD operations — less boilerplate and fewer mistakes.
//
// 推奨: 生の db.DB / db.Table ではなく model.EntryModel または model.Model を使用してください。
// 型安全で抽象度の高い CRUD 操作が提供され、コードが簡潔になりミスも減ります。
//
// 권장: raw db.DB / db.Table 대신 model.EntryModel 또는 model.Model 을 사용하세요.
// 타입 안전하고 더 높은 수준의 CRUD 연산을 제공하므로 코드가 간결해지고 실수가 줄어듭니다.
//
// Рекомендация: используйте model.EntryModel или model.Model вместо прямого db.DB / db.Table.
// Они предоставляют типобезопасные CRUD-операции высокого уровня — меньше шаблонного кода и ошибок.
type DB struct {
	db *gorm.DB
}

// New creates a fresh DB instance with the original GORM connection.
func (d *DB) New() *DB {
	return &DB{db: d.db}
}

// WithContext returns a new DB whose underlying *gorm.DB carries the given context.
// The original DB instance is unchanged; safe for concurrent use across requests.
func (d *DB) WithContext(ctx context.Context) *DB {
	return &DB{db: d.db.WithContext(ctx)}
}

// Transaction executes fc within a database transaction.
// The transaction is committed if fc returns nil, rolled back otherwise.
func (d *DB) Transaction(fc func(tx *DB) error) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		return fc(&DB{db: tx})
	})
}

// Migrator returns the GORM migrator for schema management.
func (d *DB) Migrator() gorm.Migrator {
	return d.db.Migrator()
}

// Table returns a Table instance for the given table name.
func (d *DB) Table(name string) *Table {
	tx := d.db.Table(name)
	return &Table{db: tx, tableName: name, raw: d.db}
}

type noConfigDBError struct {
}

func (e *noConfigDBError) Error() string {
	return "no config db"
}

// NoConfigDBError is returned when no database configuration is found.
var NoConfigDBError = &noConfigDBError{}

// IConfig defines the interface for database configuration sources.
type IConfig interface {
	Connection() (*DB, error)
}
// Config holds the basic database configuration.
type Config struct {
	Type string
}

const ConfigKey = "web.db"

// CreateDB creates a database connection based on the configuration.
// It supports MySQL, PostgreSQL, and SQLite database types.
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
		if util.EqualsAnyIgnoreCase(config2.Type, POSTGRES) || util.EqualsAnyIgnoreCase(config2.Type, POSTGRESQL) {
			var pgConfig PostgresConfig
			err := c.UnmarshalKey(ConfigKey, &pgConfig)
			if err != nil {
				return nil, err
			}
			return pgConfig.Connection()
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
